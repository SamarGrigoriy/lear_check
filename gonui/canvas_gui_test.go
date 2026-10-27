//go:build gui

package nui

import (
	"testing"
	"time"

	"nui/internal/drv"
)

// Canvas: рисунок на экране (цвета точек), размер, TextSize, Redraw, мышь, Painter вне draw.
func TestCanvas(t *testing.T) {
	var w *Window
	var c *Canvas
	var kept *Painter
	draws, n := 0, 0
	var size [2]float64
	var text [2]float64
	var events []MouseEvent
	ui(func() {
		c = NewCanvas(func(p *Painter) {
			draws++
			size[0], size[1] = p.Size()
			p.SetColor(RGB(255, 255, 255))
			p.FillRect(0, 0, size[0], size[1])
			p.SetColor(RGB(255, 0, 0))
			p.FillRect(10, 10, 60, 40)
			p.SetColor(RGB(40, 80, 200))
			p.FillEllipse(110, 20, 60, 60)
			p.SetColor(NoColor)
			p.SetLineWidth(2)
			p.Line(10, 100, 190, 100)
			p.Rect(5.5, 5.5, 189, 109)
			p.Ellipse(110, 20, 60, 60)
			p.Text(12, 60, "Привет")
			text[0], text[1] = p.TextSize("Привет")
			kept = p
		})
		c.SetMinSize(200, 120)
		c.SetAlign(AlignStart, AlignStart)
		c.OnMouse(func(e MouseEvent) { events = append(events, e) })
		w = NewWindow("холст", VBox(c))
		w.SetSize(260, 180)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { return draws > 0 })
	time.Sleep(200 * time.Millisecond) // рисунок доходит до экрана
	shot(t, "canvas")
	ui(func() {
		if size != [2]float64{200, 120} {
			t.Errorf("Painter.Size = %v, ожидалось 200×120", size)
		}
		if text[0] < 20 || text[1] < 8 {
			t.Errorf("TextSize = %v", text)
		}
		for _, px := range []struct {
			x, y int
			want uint32
		}{{40, 15, 0xFF0000}, {140, 50, 0x2850C8}, {90, 40, 0xFFFFFF}} {
			if got := drv.PixelAt(c.h, px.x, px.y); got != px.want {
				t.Errorf("точка (%d, %d) = %06X, ожидалось %06X", px.x, px.y, got, px.want)
			}
		}
		kept.FillRect(0, 0, 10, 10) // Painter вне draw — ничего
		w.Show()                    // наверх: окно, закрытое другими (Wine без композитора), не перерисовывается
		n = draws                   // здесь, в UI-потоке: draws меняет draw
		c.Redraw()
		c.Redraw() // несколько подряд — одна перерисовка
	})
	waitUntil(t, 2000, func() bool { return draws > n })
	ui(func() {
		drv.SimMouse(c.h, drv.MouseDown, 30, 40, 1, drv.ModShift) // Win32 — в целых точках
		drv.SimMouse(c.h, drv.MouseMove, 50, 60, 0, 0)
		drv.SimMouse(c.h, drv.MouseUp, 50, 60, 1, 0)
		drv.SimMouse(c.h, drv.MouseWheel, 10, 10, 1, 0)
		drv.SimMouse(c.h, drv.MouseWheel, 10, 10, -1, drv.ModCtrl)
	})
	ours := func() []MouseEvent { // до первого нажатия — движение настоящего указателя (Wine)
		for i, e := range events {
			if e.Kind == MouseDown {
				return events[i:]
			}
		}
		return nil
	}
	waitUntil(t, 2000, func() bool { return len(ours()) >= 5 })
	ui(func() {
		events := ours()
		want := []MouseEvent{
			{Kind: MouseDown, X: 30, Y: 40, Button: 1, Mod: ModShift},
			{Kind: MouseMove, X: 50, Y: 60},
			{Kind: MouseUp, X: 50, Y: 60, Button: 1},
			{Kind: MouseWheel, X: 10, Y: 10, WheelY: 1},
			{Kind: MouseWheel, X: 10, Y: 10, WheelY: -1, Mod: ModCtrl},
		}
		if len(events) != len(want) {
			t.Fatalf("события %+v", events)
		}
		for i := range want {
			if events[i] != want[i] {
				t.Errorf("событие %d: %+v, ожидалось %+v", i, events[i], want[i])
			}
		}
		w.Close()
	})
}
