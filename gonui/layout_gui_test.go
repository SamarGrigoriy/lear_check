//go:build gui && linux

package nui

// Golden-тесты раскладки, сторона GTK (layout.md, §8). Каждый случай из
// internal/layout/testdata/cases строится из настоящих виджетов nui, GTK его раскладывает,
// тест снимает размеры и прямоугольники и тут же сверяет с ними движок internal/layout —
// на шрифтах и теме этой машины. NUI_GOLDEN=write ещё и записывает снимки в testdata/golden
// (make golden — в контейнере, чтобы снимки не зависели от машины). С NUI_SHOTS делается
// снимок окна, поверх которого красным нарисованы прямоугольники движка (фотоотчёт).

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"nui/internal/drv"
	"nui/internal/layout"
	"nui/internal/layout/golden"
)

const (
	casesDir  = "internal/layout/testdata/cases"
	goldenDir = "internal/layout/testdata/golden"
)

func TestLayoutGTK(t *testing.T) {
	cases, err := golden.Cases(casesDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("нет случаев в " + casesDir)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) { layoutCase(t, c) })
	}
}

func layoutCase(t *testing.T, c *golden.Case) {
	ws := map[string]Widget{}
	var win *Window
	g := &golden.Golden{Comment: "снято с GTK тестом TestLayoutGTK (make golden); руками не править"}
	ui(func() {
		root := buildCase(c.Root, ws)
		win = NewWindow("раскладка: "+c.Name, root)
		win.SetPadding(c.Padding)
		win.SetSize(c.W, c.H)
		// Минимум Frame без ребёнка (ширина заголовка) — до показа окна, чтобы не мелькало.
		frameFloors(c, ws, g)
		win.Show()
	})
	defer ui(func() { win.Close() })
	waitLayout(t, win, ws["0"], c)
	ui(func() { recordCase(c, ws, win, g) })
	if os.Getenv("NUI_GOLDEN") == "write" {
		if err := g.Write(filepath.Join(goldenDir, c.Name+".txt")); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range golden.Check(c, g) {
		t.Error(e)
	}
	if dir := os.Getenv("NUI_SHOTS"); dir != "" {
		overlayShot(t, c, g, win, filepath.Join(dir, "layout-"+c.Name+".png"))
	}
}

// buildCase строит виджеты nui по случаю; ws — виджет по пути узла.
func buildCase(it *golden.Item, ws map[string]Widget) Widget {
	var w Widget
	child := func() Widget {
		if len(it.Children) == 0 {
			return nil
		}
		return buildCase(it.Children[0], ws)
	}
	switch it.Type {
	case "vbox", "hbox":
		var kids []Widget
		for _, ch := range it.Children {
			k := buildCase(ch, ws)
			if ch.Expand {
				k = Expand(k)
			}
			kids = append(kids, k)
		}
		box := HBox
		if it.Type == "vbox" {
			box = VBox
		}
		w = box(kids...).Spacing(it.Spacing).Padding(it.Padding)
	case "grid":
		g := NewGrid().Spacing(it.ColGap, it.RowGap).Padding(it.Padding)
		for _, i := range it.ExpandCols {
			g.ExpandCol(i)
		}
		for _, i := range it.ExpandRows {
			g.ExpandRow(i)
		}
		for _, ch := range it.Children {
			g.AddSpan(ch.Col, ch.Row, ch.Cols, ch.Rows, buildCase(ch, ws))
		}
		w = g
	case "frame":
		w = NewFrame(it.Text, child())
	case "scroll":
		w = NewScroll(child())
	case "fixed", "separator":
		w = NewSeparator()
	case "vseparator":
		w = NewVSeparator()
	case "label":
		l := NewLabel(it.Text)
		if it.Ellipsis {
			l.SetEllipsis(EllipsisEnd)
		}
		w = l
	case "button":
		w = NewButton(it.Text, nil)
	case "check":
		w = NewCheck(it.Text)
	case "entry":
		w = NewEntry()
	case "combo":
		w = NewCombo(strings.Split(it.Text, "|")...)
	case "slider":
		w = NewSlider(0, 100)
	case "progress":
		w = NewProgress()
	case "textview":
		w = NewTextView()
	case "stretch":
		w = Stretch()
	default:
		panic("buildCase: " + it.Type)
	}
	b := w.base()
	if it.Type == "fixed" {
		b.SetMinSize(it.Size.W, it.Size.H)
	}
	if it.MinW >= 0 || it.MinH >= 0 {
		b.SetMinSize(it.MinW, it.MinH)
	}
	if it.HAlign != 0 || it.VAlign != 0 {
		b.SetAlign(Align(it.HAlign), Align(it.VAlign))
	}
	if it.Hidden {
		b.SetVisible(false)
	}
	ws[it.Path] = w
	return w
}

// frameFloors — минимум каждой Frame без ребёнка: в нём ширина заголовка (вход движка).
func frameFloors(c *golden.Case, ws map[string]Widget, g *golden.Golden) {
	floors := map[string]layout.Size{}
	c.Walk(func(it *golden.Item) {
		if it.Type != "frame" {
			return
		}
		var kid *widget
		if len(it.Children) > 0 && !it.Children[0].Hidden {
			kid = ws[it.Children[0].Path].base()
			kid.SetVisible(false)
		}
		mw, _, mh, _ := drv.Preferred(ws[it.Path].base().h)
		floors[it.Path] = layout.Size{W: mw, H: mh}
		if kid != nil {
			kid.SetVisible(true)
		}
	})
	g.Recs = nil
	c.Walk(func(it *golden.Item) {
		g.Recs = append(g.Recs, &golden.Rec{Path: it.Path, Type: it.Type, Floor: floors[it.Path]})
	})
}

// waitLayout ждёт, пока окно получит размер случая и раскладка перестанет меняться.
func waitLayout(t *testing.T, win *Window, root Widget, c *golden.Case) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	var last [4]int
	same := 0
	for {
		var cur [4]int
		ui(func() {
			_, _, cur[0], cur[1] = drv.Allocation(win.h)
			_, _, cur[2], cur[3] = drv.Allocation(root.base().h)
		})
		if cur[0] == c.W && cur[1] == c.H && cur == last {
			if same++; same >= 3 {
				return
			}
		} else {
			same = 0
		}
		last = cur
		if time.Now().After(deadline) {
			t.Fatalf("окно %dx%d вместо %dx%d: размер меньше минимума или WM вмешался", cur[0], cur[1], c.W, c.H)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// recordCase снимает с GTK размеры и прямоугольники всех узлов.
func recordCase(c *golden.Case, ws map[string]Widget, win *Window, g *golden.Golden) {
	_, _, g.W, g.H = drv.Allocation(win.h)
	mw, nw, mh, nh := drv.Preferred(win.h)
	g.Min, g.Nat = layout.Size{W: mw, H: mh}, layout.Size{W: nw, H: nh}
	hidden := map[string]bool{}
	c.Walk(func(it *golden.Item) {
		r := g.Find(it.Path)
		if parent := it.Path[:max(0, strings.LastIndexByte(it.Path, '.'))]; it.Hidden || hidden[parent] {
			r.Hidden, hidden[it.Path] = true, true
			return
		}
		h := ws[it.Path].base().h
		mw, nw, mh, nh := drv.Preferred(h)
		r.Min, r.Nat = layout.Size{W: mw, H: mh}, layout.Size{W: nw, H: nh}
		x, y, _, _ := drv.WindowPos(h)
		_, _, w, ht := drv.Allocation(h)
		r.Out = layout.Rect{X: x, Y: y, W: w, H: ht}
	})
	// Оформление Frame и Scroll — по месту ребёнка (у него выравнивание fill) за вычетом
	// его Padding: allocation в GTK, как и Out движка, отступ не включает. Ребёнок прокрутки
	// может быть больше видимой области, поэтому справа и снизу — как слева и сверху.
	// Минимум Scroll — его собственный (полосы прокрутки): вход движка, как и у листьев.
	c.Walk(func(it *golden.Item) {
		r := g.Find(it.Path)
		if r.Hidden || (it.Type != "frame" && it.Type != "scroll") {
			return
		}
		if it.Type == "scroll" {
			r.Floor = r.Min
		}
		if len(it.Children) == 0 || it.Children[0].Hidden {
			return
		}
		o, p := g.Find(it.Children[0].Path).Out, it.Children[0].Padding
		o = layout.Rect{X: o.X - p, Y: o.Y - p, W: o.W + 2*p, H: o.H + 2*p}
		in := layout.Insets{L: o.X - r.Out.X, T: o.Y - r.Out.Y}
		in.R, in.B = in.L, in.T
		if it.Type == "frame" {
			in.R = r.Out.X + r.Out.W - o.X - o.W
			in.B = r.Out.Y + r.Out.H - o.Y - o.H
		}
		r.Inset = in
	})
}

// overlayShot — снимок клиентской области окна, поверх него красные рамки листьев
// и синие — контейнеров, как их разложил движок. Совпали с краями виджетов — движок
// разложил так же, как GTK.
func overlayShot(t *testing.T, c *golden.Case, g *golden.Golden, win *Window, path string) {
	_, nodes, err := golden.Engine(c, g)
	if err != nil {
		t.Error(err)
		return
	}
	var sx, sy int
	ui(func() { _, _, sx, sy = drv.WindowPos(win.h) })
	time.Sleep(300 * time.Millisecond) // UI-поток дорисовывает окно
	out, err := exec.Command("rootshot").Output()
	if err != nil {
		t.Logf("снимок: %v", err)
		return
	}
	screen, err := readPPM(out)
	if err != nil {
		t.Logf("снимок: %v", err)
		return
	}
	img := image.NewRGBA(image.Rect(0, 0, g.W, g.H))
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			img.Set(x, y, screen.At(sx+x, sy+y))
		}
	}
	red, blue := color.RGBA{230, 0, 0, 255}, color.RGBA{0, 90, 230, 255}
	c.Walk(func(it *golden.Item) {
		if r := g.Find(it.Path); r.Hidden || golden.Container(it.Type) {
			return
		}
		outline(img, nodes[it.Path].Out, red)
	})
	c.Walk(func(it *golden.Item) {
		if r := g.Find(it.Path); !r.Hidden && (it.Type == "frame" || it.Type == "scroll") {
			outline(img, nodes[it.Path].Out, blue)
		}
	})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err == nil {
		err = os.WriteFile(path, buf.Bytes(), 0o644)
	}
	if err != nil {
		t.Logf("снимок: %v", err)
	}
}

func outline(img *image.RGBA, r layout.Rect, col color.RGBA) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	for x := r.X; x < r.X+r.W; x++ {
		img.SetRGBA(x, r.Y, col)
		img.SetRGBA(x, r.Y+r.H-1, col)
	}
	for y := r.Y; y < r.Y+r.H; y++ {
		img.SetRGBA(r.X, y, col)
		img.SetRGBA(r.X+r.W-1, y, col)
	}
}

// readPPM читает P6 с maxval 255 — то, что пишет tools/rootshot.
func readPPM(data []byte) (*image.RGBA, error) {
	rd := bufio.NewReader(bytes.NewReader(data))
	var hdr [4]int
	for i := 0; i < 4; {
		var word []byte
		for {
			b, err := rd.ReadByte()
			if err != nil {
				return nil, err
			}
			if b == '#' {
				rd.ReadString('\n')
				continue
			}
			if b == ' ' || b == '\n' || b == '\r' || b == '\t' {
				if len(word) > 0 {
					break
				}
				continue
			}
			word = append(word, b)
		}
		if i == 0 {
			if string(word) != "P6" {
				return nil, os.ErrInvalid
			}
		} else {
			n, err := strconv.Atoi(string(word))
			if err != nil {
				return nil, err
			}
			hdr[i] = n
		}
		i++
	}
	w, h := hdr[1], hdr[2]
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	px := make([]byte, 3*w*h)
	if _, err := io.ReadFull(rd, px); err != nil {
		return nil, err
	}
	for i := 0; i < w*h; i++ {
		img.Pix[4*i], img.Pix[4*i+1], img.Pix[4*i+2], img.Pix[4*i+3] = px[3*i], px[3*i+1], px[3*i+2], 255
	}
	return img, nil
}
