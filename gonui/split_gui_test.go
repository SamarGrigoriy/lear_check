//go:build gui

package nui

import (
	"runtime"
	"testing"

	"nui/internal/drv"
)

// Split: начальное положение — естественный размер первой панели; SetPosition; упор в
// минимум; первая панель держит размер при растягивании окна; перетаскивание; скрытая
// панель; VSplit.
func TestSplit(t *testing.T) {
	gap := 1 // граница: в GTK линия в 1 точку, на Windows полоса 5 точек
	if runtime.GOOS == "windows" {
		gap = 5
	}
	var w *Window
	var a *Label
	var b *TextView
	var s *Split
	ui(func() {
		a = NewLabel("левая панель")
		a.SetMinSize(80, -1)
		b = NewTextView()
		s = HSplit(a, b)
		w = NewWindow("разделитель", s)
		w.SetSize(400, 200)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { aw, _ := a.Size(); return aw > 0 })
	widths := func() (int, int) { aw, _ := a.Size(); bw, _ := b.Size(); return aw, bw }
	var nat int
	ui(func() {
		aw, bw := widths()
		nat = aw
		if s.Position() != aw || aw+gap+bw != 400-24 {
			t.Errorf("начало: первая %d, вторая %d, Position %d", aw, bw, s.Position())
		}
		s.SetPosition(150)
	})
	check := func(name string, want int) {
		t.Helper()
		waitUntil(t, 2000, func() bool { aw, _ := widths(); return aw == want })
		ui(func() {
			aw, bw := widths()
			ww, _ := w.Size()
			if s.Position() != want || aw+gap+bw != ww-24 {
				t.Errorf("%s: первая %d, вторая %d, Position %d (ожидалось %d)", name, aw, bw, s.Position(), want)
			}
		})
	}
	check("SetPosition(150)", 150)
	ui(func() { s.SetPosition(10) }) // меньше минимума первой панели
	check("упор в минимум", nat)
	ui(func() { s.SetPosition(150); w.SetSize(600, 200) })
	check("окно шире", 150) // первая держит размер, растёт вторая
	ui(func() { drv.SimSplitDrag(s.h, 30) })
	check("перетаскивание", 180)
	ui(func() { a.SetVisible(false) })
	waitUntil(t, 2000, func() bool { bw, _ := b.Size(); return bw == 600-24 })
	ui(func() { w.Close() })

	ui(func() { // VSplit: то же по высоте
		top, bottom := NewTextView(), NewTextView()
		v := VSplit(top, bottom).SetPosition(60)
		w = NewWindow("по высоте", v)
		w.SetSize(300, 300)
		w.Show()
		s = v
		b = bottom
		_ = top
	})
	waitUntil(t, 2000, func() bool { return s.Position() == 60 })
	ui(func() {
		_, bh := b.Size()
		if bh != 300-24-60-gap {
			t.Errorf("VSplit: нижняя высотой %d, ожидалось %d", bh, 300-24-60-gap)
		}
		w.Close()
	})
}
