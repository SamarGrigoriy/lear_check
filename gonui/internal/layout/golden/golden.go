// Package golden — общий код golden-тестов раскладки (layout.md, §8): язык случаев,
// формат golden-файлов и сверка движка с ними. Им пользуются два теста:
//   - TestLayoutGTK (пакет nui, тег gui) строит случай из настоящих виджетов nui на GTK,
//     снимает размеры и прямоугольники, сверяет с ними движок и по NUI_GOLDEN=write
//     пишет golden-файлы;
//   - TestGolden (internal/layout) сверяет движок с golden-файлами без GTK, на любой ОС.
//
// Это код тестов: библиотека и приложения его не импортируют (tools/depcheck.sh).
//
// Язык случаев (testdata/cases/*.txt). Первая строка — окно, дальше — дерево, вложенность
// задаётся отступом пробелами. Комментарий — от # до конца строки.
//
//	window 400x300 padding=12      # размер окна и Window.SetPadding (по умолчанию 12)
//	vbox spacing=6                 # корень — один узел
//	  fixed 100x20                 # лист ровно 100×20 (Separator + SetMinSize)
//	  label expand ellipsis "Длинная подпись"
//	  grid colgap=6 rowgap=6 expandcol=1
//	    fixed @0,0 60x20
//	    fixed @0,1 span=2x1 300x40
//
// Контейнеры: vbox, hbox, grid, frame "заголовок", scroll. Листья: fixed WxH,
// label "текст", button "текст", check "текст", entry, combo "a|b", slider, progress,
// textview, separator, vseparator, stretch. Атрибуты: expand (ребёнок Box),
// @кол,стр и span=КxС (ребёнок Grid), align=гор,верт (fill|start|center|end),
// min=WxH (SetMinSize, -1 — не задавать), hidden, ellipsis (label), spacing=N,
// padding=N (Box, Grid), colgap=N, rowgap=N, expandcol=1,2, expandrow=0 (Grid).
// У окна ещё tolerance=1: допуск 1 px, причину пишут в комментарии.
//
// Ребёнок frame и scroll должен иметь выравнивание fill: по его месту golden-генератор
// измеряет оформление контейнера (Inset).
package golden

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"nui/internal/layout"
)

// Item — узел случая.
type Item struct {
	Path string // номер в родителе через точку: "0", "0.2", "0.2.1"
	Type string
	Text string
	Line int // строка в файле случая

	Size                     layout.Size // fixed WxH
	MinW, MinH               int         // min=WxH; -1 — не задано
	Expand, Hidden, Ellipsis bool
	HAlign, VAlign           uint8
	Col, Row, Cols, Rows     int
	Spacing, Padding         int
	ColGap, RowGap           int
	ExpandCols, ExpandRows   []int

	Children []*Item
}

// Case — случай из testdata/cases.
type Case struct {
	Name          string
	W, H, Padding int
	Tolerance     int
	Root          *Item
}

// Container — может ли узел этого типа иметь детей.
func Container(typ string) bool {
	switch typ {
	case "vbox", "hbox", "grid", "frame", "scroll":
		return true
	}
	return false
}

var leaves = map[string]bool{
	"fixed": true, "label": true, "button": true, "check": true, "entry": true, "combo": true,
	"slider": true, "progress": true, "textview": true, "separator": true, "vseparator": true,
	"stretch": true,
}

// Cases читает все случаи из каталога dir.
func Cases(dir string) ([]*Case, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.txt"))
	if err != nil {
		return nil, err
	}
	var cs []*Case
	for _, f := range files {
		c, err := ParseCaseFile(f)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}

// ParseCaseFile читает случай; имя случая — имя файла без .txt.
func ParseCaseFile(path string) (*Case, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c, err := ParseCase(string(data))
	if err != nil {
		return nil, fmt.Errorf("%s:%v", path, err)
	}
	c.Name = strings.TrimSuffix(filepath.Base(path), ".txt")
	return c, nil
}

type frame struct {
	indent int
	it     *Item
}

// ParseCase разбирает текст случая.
func ParseCase(src string) (*Case, error) {
	c := &Case{Padding: 12}
	var stack []frame
	haveWindow := false
	for i, raw := range strings.Split(src, "\n") {
		ln := i + 1
		toks, err := tokens(raw)
		if err != nil {
			return nil, fmt.Errorf("%d: %v", ln, err)
		}
		if len(toks) == 0 {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		if strings.HasPrefix(raw[indent:], "\t") {
			return nil, fmt.Errorf("%d: отступ — только пробелами", ln)
		}
		if !haveWindow {
			if toks[0] != "window" {
				return nil, fmt.Errorf("%d: первая строка — window WxH", ln)
			}
			if err := parseWindow(c, toks[1:]); err != nil {
				return nil, fmt.Errorf("%d: %v", ln, err)
			}
			haveWindow = true
			continue
		}
		it, err := parseItem(toks)
		if err != nil {
			return nil, fmt.Errorf("%d: %v", ln, err)
		}
		it.Line = ln
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			if c.Root != nil {
				return nil, fmt.Errorf("%d: корень у окна один", ln)
			}
			it.Path = "0"
			c.Root = it
		} else {
			p := stack[len(stack)-1].it
			if !Container(p.Type) {
				return nil, fmt.Errorf("%d: у %s не бывает детей", ln, p.Type)
			}
			if (p.Type == "frame" || p.Type == "scroll") && len(p.Children) > 0 {
				return nil, fmt.Errorf("%d: у %s один ребёнок", ln, p.Type)
			}
			it.Path = p.Path + "." + strconv.Itoa(len(p.Children))
			p.Children = append(p.Children, it)
			if err := checkPlacement(p, it); err != nil {
				return nil, fmt.Errorf("%d: %v", ln, err)
			}
		}
		stack = append(stack, frame{indent, it})
	}
	if c.Root == nil {
		return nil, fmt.Errorf("пустой случай")
	}
	return c, nil
}

// tokens делит строку на слова; "строка в кавычках" — одно слово (с кавычками).
func tokens(s string) ([]string, error) {
	var out []string
	for i := 0; i < len(s); {
		switch {
		case s[i] == ' ' || s[i] == '\t' || s[i] == '\r':
			i++
		case s[i] == '#':
			return out, nil
		case s[i] == '"':
			j := strings.IndexByte(s[i+1:], '"')
			if j < 0 {
				return nil, fmt.Errorf("нет закрывающей кавычки")
			}
			out = append(out, s[i:i+j+2])
			i += j + 2
		default:
			j := i
			for j < len(s) && s[j] != ' ' && s[j] != '\t' && s[j] != '\r' && s[j] != '#' {
				j++
			}
			out = append(out, s[i:j])
			i = j
		}
	}
	return out, nil
}

func parseWindow(c *Case, toks []string) error {
	sizeSet := false
	for _, t := range toks {
		k, v, kv := strings.Cut(t, "=")
		switch {
		case !kv && !sizeSet:
			w, h, err := parseWxH(t)
			if err != nil || w <= 0 || h <= 0 {
				return fmt.Errorf("размер окна %q", t)
			}
			c.W, c.H, sizeSet = w, h, true
		case k == "padding":
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return fmt.Errorf("padding=%q", v)
			}
			c.Padding = n
		case k == "tolerance":
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 || n > 1 {
				return fmt.Errorf("tolerance — 0 или 1")
			}
			c.Tolerance = n
		default:
			return fmt.Errorf("непонятно у window: %q", t)
		}
	}
	if !sizeSet {
		return fmt.Errorf("нет размера окна")
	}
	return nil
}

func parseItem(toks []string) (*Item, error) {
	it := &Item{Type: toks[0], MinW: -1, MinH: -1, Spacing: 6, ColGap: 6, RowGap: 6}
	if !Container(it.Type) && !leaves[it.Type] {
		return nil, fmt.Errorf("неизвестный узел %q", it.Type)
	}
	sizeSet := false
	for _, t := range toks[1:] {
		k, v, kv := strings.Cut(t, "=")
		var err error
		switch {
		case strings.HasPrefix(t, `"`):
			it.Text = strings.Trim(t, `"`)
		case strings.HasPrefix(t, "@"):
			it.Col, it.Row, err = parsePair(t[1:], ",")
		case t == "expand":
			it.Expand = true
		case t == "hidden":
			it.Hidden = true
		case t == "ellipsis" && it.Type == "label":
			it.Ellipsis = true
		case !kv && it.Type == "fixed" && !sizeSet:
			it.Size.W, it.Size.H, err = parseWxH(t)
			sizeSet = true
		case k == "span":
			it.Cols, it.Rows, err = parseWxH(v)
		case k == "min":
			it.MinW, it.MinH, err = parseWxH(v)
		case k == "align":
			it.HAlign, it.VAlign, err = parseAlign(v)
		case k == "spacing" && (it.Type == "vbox" || it.Type == "hbox"):
			it.Spacing, err = strconv.Atoi(v)
		case k == "spacing" && it.Type == "grid":
			it.ColGap, err = strconv.Atoi(v)
			it.RowGap = it.ColGap
		case k == "padding" && (it.Type == "vbox" || it.Type == "hbox" || it.Type == "grid"):
			it.Padding, err = strconv.Atoi(v)
		case k == "colgap" && it.Type == "grid":
			it.ColGap, err = strconv.Atoi(v)
		case k == "rowgap" && it.Type == "grid":
			it.RowGap, err = strconv.Atoi(v)
		case k == "expandcol" && it.Type == "grid":
			it.ExpandCols, err = parseList(v)
		case k == "expandrow" && it.Type == "grid":
			it.ExpandRows, err = parseList(v)
		default:
			return nil, fmt.Errorf("непонятно у %s: %q", it.Type, t)
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %q: %v", it.Type, t, err)
		}
	}
	if it.Type == "fixed" && !sizeSet {
		return nil, fmt.Errorf("fixed: нет размера WxH")
	}
	if it.Type == "fixed" && (it.MinW >= 0 || it.MinH >= 0) {
		return nil, fmt.Errorf("fixed: размер уже задан, min= не нужен")
	}
	return it, nil
}

// checkPlacement: @ и span — только у детей Grid, expand — только у детей Box.
func checkPlacement(p, it *Item) error {
	inGrid := p.Type == "grid"
	if !inGrid && (it.Col != 0 || it.Row != 0 || it.Cols != 0 || it.Rows != 0) {
		return fmt.Errorf("@ и span — только в grid")
	}
	if it.Expand && p.Type != "vbox" && p.Type != "hbox" && it.Type != "stretch" {
		return fmt.Errorf("expand — только у детей vbox и hbox")
	}
	return nil
}

func parseWxH(s string) (int, int, error) { return parsePair(s, "x") }

func parsePair(s, sep string) (int, int, error) {
	a, b, ok := strings.Cut(s, sep)
	if !ok {
		return 0, 0, fmt.Errorf("ожидалось A%sB", sep)
	}
	x, err1 := strconv.Atoi(a)
	y, err2 := strconv.Atoi(b)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("не числа")
	}
	return x, y, nil
}

func parseList(s string) ([]int, error) {
	var out []int
	for _, f := range strings.Split(s, ",") {
		n, err := strconv.Atoi(f)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("список номеров")
		}
		out = append(out, n)
	}
	return out, nil
}

var alignNames = map[string]uint8{
	"fill": layout.AlignFill, "start": layout.AlignStart, "center": layout.AlignCenter, "end": layout.AlignEnd,
}

func parseAlign(s string) (uint8, uint8, error) {
	a, b, ok := strings.Cut(s, ",")
	h, ok1 := alignNames[a]
	v, ok2 := alignNames[b]
	if !ok || !ok1 || !ok2 {
		return 0, 0, fmt.Errorf("align=гор,верт из fill|start|center|end")
	}
	return h, v, nil
}

// Walk обходит дерево случая в прямом порядке.
func (c *Case) Walk(fn func(*Item)) { walk(c.Root, fn) }

func walk(it *Item, fn func(*Item)) {
	fn(it)
	for _, ch := range it.Children {
		walk(ch, fn)
	}
}

// --- golden-файлы ---

// Rec — то, что GTK показал для одного узла.
type Rec struct {
	Path, Type string
	Hidden     bool          // узел или его предок скрыт: не сравнивается
	Min, Nat   layout.Size   // у листьев — вход движка, у контейнеров — ожидаемый результат
	Inset      layout.Insets // frame, scroll: вход движка
	Floor      layout.Size   // frame, scroll: вход движка
	Out        layout.Rect
}

// Golden — снимок одного случая с GTK.
type Golden struct {
	Comment  string      // откуда снято: версия GTK, тема
	W, H     int         // окно по факту
	Min, Nat layout.Size // предпочтительный размер окна
	Recs     []*Rec      // в порядке обхода
}

// Find ищет запись по пути.
func (g *Golden) Find(path string) *Rec {
	for _, r := range g.Recs {
		if r.Path == path {
			return r
		}
	}
	return nil
}

func size(s layout.Size) string { return fmt.Sprintf("%dx%d", s.W, s.H) }

// Write сохраняет снимок в файл.
func (g *Golden) Write(path string) error {
	var b strings.Builder
	if g.Comment != "" {
		fmt.Fprintf(&b, "# %s\n", g.Comment)
	}
	fmt.Fprintf(&b, "window %dx%d min=%s nat=%s\n", g.W, g.H, size(g.Min), size(g.Nat))
	for _, r := range g.Recs {
		if r.Hidden {
			fmt.Fprintf(&b, "%s %s hidden\n", r.Path, r.Type)
			continue
		}
		fmt.Fprintf(&b, "%s %s min=%s nat=%s", r.Path, r.Type, size(r.Min), size(r.Nat))
		if r.Type == "frame" || r.Type == "scroll" {
			in := r.Inset
			fmt.Fprintf(&b, " inset=%d,%d,%d,%d floor=%s", in.L, in.T, in.R, in.B, size(r.Floor))
		}
		o := r.Out
		fmt.Fprintf(&b, " out=%d,%d,%d,%d\n", o.X, o.Y, o.W, o.H)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// ReadGolden читает снимок из файла.
func ReadGolden(path string) (*Golden, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	g := &Golden{}
	sc := bufio.NewScanner(f)
	for ln := 1; sc.Scan(); ln++ {
		s := strings.TrimSpace(sc.Text())
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "#") {
			if g.Comment == "" {
				g.Comment = strings.TrimSpace(s[1:])
			}
			continue
		}
		if err := g.parseLine(strings.Fields(s)); err != nil {
			return nil, fmt.Errorf("%s:%d: %v", path, ln, err)
		}
	}
	return g, sc.Err()
}

func (g *Golden) parseLine(f []string) error {
	if f[0] == "window" {
		if len(f) != 4 {
			return fmt.Errorf("window WxH min= nat=")
		}
		var err error
		if g.W, g.H, err = parseWxH(f[1]); err != nil {
			return err
		}
		return kvs(f[2:], map[string]any{"min": &g.Min, "nat": &g.Nat})
	}
	if len(f) < 3 {
		return fmt.Errorf("короткая строка")
	}
	r := &Rec{Path: f[0], Type: f[1]}
	g.Recs = append(g.Recs, r)
	if f[2] == "hidden" {
		r.Hidden = true
		return nil
	}
	return kvs(f[2:], map[string]any{"min": &r.Min, "nat": &r.Nat, "inset": &r.Inset, "floor": &r.Floor, "out": &r.Out})
}

func kvs(f []string, dst map[string]any) error {
	for _, t := range f {
		k, v, ok := strings.Cut(t, "=")
		p, known := dst[k]
		if !ok || !known {
			return fmt.Errorf("непонятно: %q", t)
		}
		var n []int
		for _, s := range strings.FieldsFunc(v, func(r rune) bool { return r == 'x' || r == ',' }) {
			x, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("%s: не число", k)
			}
			n = append(n, x)
		}
		switch p := p.(type) {
		case *layout.Size:
			if len(n) != 2 {
				return fmt.Errorf("%s=WxH", k)
			}
			*p = layout.Size{W: n[0], H: n[1]}
		case *layout.Insets:
			if len(n) != 4 {
				return fmt.Errorf("%s=L,T,R,B", k)
			}
			*p = layout.Insets{L: n[0], T: n[1], R: n[2], B: n[3]}
		case *layout.Rect:
			if len(n) != 4 {
				return fmt.Errorf("%s=X,Y,W,H", k)
			}
			*p = layout.Rect{X: n[0], Y: n[1], W: n[2], H: n[3]}
		}
	}
	return nil
}

// --- сверка движка ---

// Engine строит дерево layout.Node по случаю и снимку, измеряет и раскладывает его
// в окне снимка. Возвращает корень и соответствие путь → узел.
func Engine(c *Case, g *Golden) (*layout.Node, map[string]*layout.Node, error) {
	nodes := map[string]*layout.Node{}
	var build func(it *Item) (*layout.Node, error)
	build = func(it *Item) (*layout.Node, error) {
		r := g.Find(it.Path)
		if r == nil || r.Type != it.Type {
			return nil, fmt.Errorf("%s %s: в golden-файле нет такого узла — перегенерируйте (make golden)", it.Path, it.Type)
		}
		n := &layout.Node{
			Hidden: it.Hidden, Expand: it.Expand || it.Type == "stretch",
			HAlign: it.HAlign, VAlign: it.VAlign,
			Col: it.Col, Row: it.Row, Cols: it.Cols, Rows: it.Rows,
			Floor: layout.Size{W: maxInt(it.MinW, 0), H: maxInt(it.MinH, 0)},
		}
		switch it.Type {
		case "vbox", "hbox":
			n.Kind = layout.HBox
			if it.Type == "vbox" {
				n.Kind = layout.VBox
			}
			n.Spacing, n.Padding = it.Spacing, it.Padding
		case "grid":
			n.Kind = layout.Grid
			n.ColGap, n.RowGap, n.Padding = it.ColGap, it.RowGap, it.Padding
			n.ExpandCols, n.ExpandRows = flags(it.ExpandCols), flags(it.ExpandRows)
		case "frame", "scroll":
			n.Kind = layout.Frame
			if it.Type == "scroll" {
				n.Kind = layout.Scroll
			}
			n.Inset, n.Floor = r.Inset, r.Floor
		default:
			n.Min, n.Nat = r.Min, r.Nat
		}
		nodes[it.Path] = n
		for _, ch := range it.Children {
			cn, err := build(ch)
			if err != nil {
				return nil, err
			}
			n.Children = append(n.Children, cn)
		}
		return n, nil
	}
	root, err := build(c.Root)
	if err != nil {
		return nil, nil, err
	}
	layout.Measure(root)
	p := c.Padding
	layout.Arrange(root, layout.Rect{X: p, Y: p, W: maxInt(1, g.W-2*p), H: maxInt(1, g.H-2*p)})
	return root, nodes, nil
}

func flags(idx []int) []bool {
	var f []bool
	for _, i := range idx {
		for len(f) <= i {
			f = append(f, false)
		}
		f[i] = true
	}
	return f
}

// Check сверяет движок со снимком GTK; возвращает расхождения.
func Check(c *Case, g *Golden) []string {
	root, nodes, err := Engine(c, g)
	if err != nil {
		return []string{err.Error()}
	}
	var errs []string
	bad := func(format string, a ...any) { errs = append(errs, fmt.Sprintf(format, a...)) }
	tol := c.Tolerance
	p := c.Padding
	if wm := (layout.Size{W: root.Min.W + 2*p, H: root.Min.H + 2*p}); !nearSize(wm, g.Min, tol) {
		bad("окно: минимум движка %s, GTK %s", size(wm), size(g.Min))
	}
	c.Walk(func(it *Item) {
		r, n := g.Find(it.Path), nodes[it.Path]
		if r.Hidden {
			return
		}
		if Container(it.Type) && (!nearSize(n.Min, r.Min, tol) || !nearSize(n.Nat, r.Nat, tol)) {
			bad("%s %s (строка %d): min/nat движка %s/%s, GTK %s/%s",
				it.Path, it.Type, it.Line, size(n.Min), size(n.Nat), size(r.Min), size(r.Nat))
		}
		if !nearRect(n.Out, r.Out, tol) {
			bad("%s %s (строка %d): движок %v, GTK %v", it.Path, it.Type, it.Line, n.Out, r.Out)
		}
	})
	return errs
}

func near(a, b, tol int) bool { return a-b <= tol && b-a <= tol }

func nearSize(a, b layout.Size, tol int) bool { return near(a.W, b.W, tol) && near(a.H, b.H, tol) }

func nearRect(a, b layout.Rect, tol int) bool {
	return near(a.X, b.X, tol) && near(a.Y, b.Y, tol) && near(a.W, b.W, tol) && near(a.H, b.H, tol)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
