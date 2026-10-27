//go:build gui

package nui

import (
	"testing"

	"nui/internal/drv"
)

// Radio: выбран первый; SetSelected без OnChange, вне диапазона — ничего; щелчок по
// варианту — OnChange, по выбранному — нет; стрелки — соседний вариант, у края — ничего; Add не
// меняет выбора; в ряд — ниже, чем в столбик; пустой переключатель.
func TestRadio(t *testing.T) {
	var w *Window
	var r, empty *Radio
	changes := 0
	ui(func() {
		r = NewRadio("Быстро", "Качественно", "Как в прошлый раз")
		r.OnChange(func() { changes++ })
		empty = NewRadio()
		w = NewWindow("переключатель", VBox(r, NewEntry(), empty))
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, rh := r.Size(); return rh > 1 })
	shot(t, "radio")
	ui(func() {
		if r.Selected() != 0 || r.Len() != 3 || r.Item(1) != "Качественно" || r.Item(5) != "" {
			t.Errorf("начало: Selected %d, Len %d", r.Selected(), r.Len())
		}
		r.SetSelected(2)
		r.SetSelected(9)
		r.SetSelected(-1)
		if r.Selected() != 2 || changes != 0 {
			t.Errorf("SetSelected(2): Selected %d, OnChange ×%d", r.Selected(), changes)
		}
		drv.SimRadioSelect(r.h, 1)
		drv.SimRadioSelect(r.h, 1) // уже выбран — не изменение
		if r.Selected() != 1 || changes != 1 {
			t.Errorf("щелчок: Selected %d, OnChange ×%d", r.Selected(), changes)
		}
		r.Focus()
	})
	for _, c := range []struct {
		key  uint32
		want int
	}{{drv.KeyDown, 2}, {drv.KeyDown, 2}, {drv.KeyUp, 1}, {drv.KeyUp, 0}, {drv.KeyUp, 0}, {drv.KeyDown, 1}} {
		ui(func() {
			drv.SimKey(w.h, c.key, 0)
			if r.Selected() != c.want {
				t.Errorf("стрелка %#x: Selected %d, ожидалось %d", c.key, r.Selected(), c.want)
			}
		})
	}
	ui(func() {
		if changes != 5 {
			t.Errorf("стрелки: OnChange ×%d, ожидалось 5", changes)
		}
		r.SetSelected(2) // фокус на варианте 1: переходит на выбранный, OnChange нет
		r.SetSelected(1)
		r.Add("Свой")
		if r.Len() != 4 || r.Selected() != 1 || changes != 5 {
			t.Errorf("Add: Len %d, Selected %d, OnChange ×%d", r.Len(), r.Selected(), changes)
		}
		if empty.Selected() != -1 {
			t.Errorf("пустой: Selected %d", empty.Selected())
		}
		empty.Add("первый")
		if empty.Selected() != 0 {
			t.Errorf("первый вариант не выбран: %d", empty.Selected())
		}
	})
	var colW, colH int
	waitUntil(t, 2000, func() bool { colW, colH = r.Size(); return colH > 1 })
	ui(func() {
		r.SetHorizontal(true)
		r.SetAlign(AlignStart, AlignFill)
	})
	waitUntil(t, 2000, func() bool { rw, rh := r.Size(); return rh < colH && rw > colW })
	ui(func() { // окно выросло под ряд — на Windows тоже, как в GTK
		rw, _ := r.Size()
		if ww, _ := w.Size(); ww < rw+24 {
			t.Errorf("окно шириной %d уже ряда вариантов (%d) с отступами", ww, rw)
		}
	})
	shot(t, "radio-row")
	ui(func() { w.Close() })
}
