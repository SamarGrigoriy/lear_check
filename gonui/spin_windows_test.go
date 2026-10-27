//go:build gui

package nui

import (
	"testing"

	"nui/internal/drv"
)

// Win32: набранное в Spin принимается, когда фокус уходит из поля (EN_KILLFOCUS) — так
// щелчок по кнопке «OK» читает уже новое значение. В GTK то же делает focus-out, но под
// Xvfb без оконного менеджера окно не получает фокуса, и проверить это нельзя.
func TestSpinFocusOutWindows(t *testing.T) {
	var w *Window
	var s *Spin
	var e *Entry
	changes := 0
	ui(func() {
		s = NewSpin(0, 10)
		s.OnChange(func() { changes++ })
		e = NewEntry()
		w = NewWindow("уход фокуса", VBox(s, e))
		w.Show()
	})
	waitUntil(t, 2000, func() bool { sw, _ := s.Size(); return sw > 1 })
	ui(func() {
		s.Focus()
		drv.SimType(s.h, "7")
		e.Focus()
		if s.Value() != 7 || changes != 1 {
			t.Errorf("фокус ушёл: Value %d, OnChange ×%d; ожидалось 7, ×1", s.Value(), changes)
		}
		w.Close()
	})
}
