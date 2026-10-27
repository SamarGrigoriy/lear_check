//go:build gui

package nui

import (
	"testing"

	"nui/internal/drv"
)

// Spin: значение min; SetValue в диапазон без OnChange; стрелки — шаг, у края — ничего;
// набранное принимается по Enter, до того Value прежнее; пустое — 0, то есть min; вне
// диапазона — край; отрицательный диапазон; паника при min >= max.
func TestSpin(t *testing.T) {
	var w *Window
	var s, neg *Spin
	changes := 0
	ui(func() {
		s = NewSpin(1, 100)
		s.OnChange(func() { changes++ })
		neg = NewSpin(-50, 50)
		w = NewWindow("числовое поле", VBox(s, neg))
		w.Show()
	})
	waitUntil(t, 2000, func() bool { sw, _ := s.Size(); return sw > 1 })
	ui(func() {
		text := func() string { return drv.EntryText(s.h) }
		check := func(what string, value, n int, txt string) {
			t.Helper()
			if s.Value() != value || changes != n || text() != txt {
				t.Errorf("%s: Value %d, текст %q, OnChange ×%d; ожидалось %d, %q, ×%d",
					what, s.Value(), text(), changes, value, txt, n)
			}
		}
		check("начало", 1, 0, "1")
		s.SetValue(50)
		check("SetValue(50)", 50, 0, "50")
		s.SetValue(500)
		check("SetValue(500)", 100, 0, "100")
		s.SetValue(-5)
		check("SetValue(-5)", 1, 0, "1")
		drv.SimSpinStep(s.h, false) // у нижнего края
		check("↓ у края", 1, 0, "1")
		drv.SimSpinStep(s.h, true)
		check("↑", 2, 1, "2")
		drv.SimSpinStep(s.h, false)
		check("↓", 1, 2, "1")
		drv.SimType(s.h, "42")
		check("набрано, не принято", 1, 2, "42")
		drv.SimEnter(s.h)
		check("Enter", 42, 3, "42")
		drv.SimType(s.h, "")
		drv.SimEnter(s.h)
		check("пустое", 1, 4, "1") // 0 — меньше min
		drv.SimType(s.h, "1000")
		drv.SimEnter(s.h)
		check("больше max", 100, 5, "100")
		if neg.Value() != -50 {
			t.Errorf("отрицательный диапазон: Value %d, ожидалось -50", neg.Value())
		}
		drv.SimType(neg.h, "-7")
		drv.SimEnter(neg.h)
		if neg.Value() != -7 {
			t.Errorf("набрано -7: Value %d", neg.Value())
		}
		expectPanic(t, "nui: NewSpin: min должен быть меньше max", func() { NewSpin(5, 5) })
		w.Close()
	})
}
