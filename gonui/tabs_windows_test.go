//go:build gui

package nui

import (
	"testing"

	"nui/internal/drv"
)

// Win32: область страниц вырезана из SysTabControl32 (он лежит над страницами) —
// щелчок по полю на странице достаётся полю, щелчок по ярлыку — вкладкам.
func TestTabsHitWindows(t *testing.T) {
	var w *Window
	var e *Entry
	var tabs *Tabs
	ui(func() {
		e = NewEntry()
		tabs = NewTabs().Add("Поле", VBox(e)).Add("Другая", NewLabel("x"))
		w = NewWindow("щелчок сквозь вкладки", tabs)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { ew, _ := e.Size(); return ew > 0 })
	ui(func() {
		ex, ey, ew, eh := alloc(e)
		if got := drv.HwndAt(w.h, ex+ew/2, ey+eh/2); got != e.h {
			t.Errorf("щелчок по полю на странице достался %v, ожидалось поле %v", got, e.h)
		}
		tx, ty, _, _ := alloc(tabs)
		if got := drv.HwndAt(w.h, tx+8, ty+8); got != tabs.h {
			t.Errorf("щелчок по ярлыку достался %v, ожидались вкладки %v", got, tabs.h)
		}
		w.Close()
	})
}
