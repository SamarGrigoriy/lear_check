//go:build gui

package nui

import (
	"runtime"
	"slices"
	"testing"

	"nui/internal/drv"
)

// Меню окна: выбор пунктов, сочетания раньше OnKey, выключенный пункт, флажок, подменю,
// подпись сочетания, размер окна под меню, замена и удаление меню, повторная вставка.
func TestMenu(t *testing.T) {
	var w *Window
	var e *Entry
	opened, saved, keys := 0, 0, 0
	var panel []bool
	openIt := Item("Открыть…", func() { opened++ }).Shortcut('O', ModCtrl)
	saveIt := Item("Сохранить", func() { saved++ }).Shortcut('S', ModCtrl)
	recent := Item("отчёт.txt", func() { opened += 10 })
	sub := SubMenu("Недавние", recent)
	panelIt := CheckItem("Боковая панель", func(on bool) { panel = append(panel, on) })
	ui(func() {
		saveIt.SetEnabled(false) // до SetMenu — тоже применяется
		panelIt.SetChecked(true)
		e = NewEntry()
		w = NewWindow("меню", VBox(e))
		w.SetSize(400, 250)
		w.OnKey(func(Key, Mod) bool { keys++; return false })
		w.SetMenu(
			NewMenu("Файл", openIt, saveIt, sub, Sep(), Item("Выход", func() {})),
			NewMenu("Вид", panelIt),
		)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { ew, _ := e.Size(); return ew > 0 })
	shot(t, "menu")
	ui(func() {
		if ww, wh := w.Size(); ww != 400 || wh != 250 {
			t.Errorf("окно с меню %d×%d, ожидалось 400×250 (область под меню)", ww, wh)
		}
		want := "Открыть…\tCtrl+O"
		if runtime.GOOS != "windows" {
			want = "Открыть…\t79/4" // GTK: 'O' и GDK_CONTROL_MASK в GtkAccelLabel
		}
		if got := drv.MenuItemText(openIt.h); got != want {
			t.Errorf("подпись пункта %q, ожидалась %q", got, want)
		}
		if !drv.MenuItemChecked(panelIt.h) {
			t.Error("SetChecked до SetMenu не применился")
		}
		drv.SimMenuActivate(openIt.h)
		drv.SimKey(w.h, 'O', drv.ModCtrl) // сочетание меню — OnKey его не видит
		drv.SimKey(w.h, 'S', drv.ModCtrl) // пункт выключен — сочетание уходит в OnKey
		drv.SimMenuActivate(saveIt.h)     // выключенный не выбирается
		drv.SimMenuActivate(recent.h)
		drv.SimMenuActivate(panelIt.h)
	})
	ui(func() {
		if opened != 12 || saved != 0 || keys != 1 {
			t.Errorf("открыть %d (ожидалось 12), сохранить %d, OnKey %d (ожидалось 1)", opened, saved, keys)
		}
		if !slices.Equal(panel, []bool{false}) || panelIt.Checked() {
			t.Errorf("флажок: вызовы %v, Checked %v", panel, panelIt.Checked())
		}
		panelIt.SetChecked(true) // программно — без вызова
		openIt.SetText("Открыть файл…")
		sub.SetEnabled(false)
		saveIt.SetEnabled(true)
		drv.SimKey(w.h, 'S', drv.ModCtrl)
		if !drv.MenuItemChecked(panelIt.h) || len(panel) != 1 || saved != 1 {
			t.Errorf("после SetChecked и SetEnabled: флажок %v, вызовы %v, сохранить %d", drv.MenuItemChecked(panelIt.h), panel, saved)
		}
		if got := drv.MenuItemText(openIt.h); got[:len("Открыть файл…")] != "Открыть файл…" {
			t.Errorf("SetText: %q", got)
		}
		expectPanic(t, "nui: пункт меню уже в меню", func() { w.SetMenu(NewMenu("снова", openIt)) })
		helpIt := Item("О программе", func() {})
		w.SetMenu(NewMenu("Справка", helpIt)) // замена: прежние пункты уничтожены
		if openIt.h != drv.Nil || panelIt.h != drv.Nil {
			t.Error("пункты прежнего меню не уничтожены")
		}
		drv.SimKey(w.h, 'O', drv.ModCtrl) // сочетания прежнего меню больше нет
		if opened != 12 || keys != 2 {
			t.Errorf("после замены меню: открыть %d, OnKey %d", opened, keys)
		}
		w.SetMenu() // убрать меню
		if helpIt.h != drv.Nil {
			t.Error("SetMenu() не уничтожил пункты")
		}
	})
	waitUntil(t, 2000, func() bool { ww, wh := w.Size(); return ww == 400 && wh == 250 })
	ui(func() {
		e.SetText("ещё жив")
		if e.Text() != "ещё жив" {
			t.Error("содержимое окна пострадало от SetMenu")
		}
		w.Close()
	})
}
