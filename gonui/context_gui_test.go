//go:build gui

package nui

import (
	"slices"
	"testing"

	"nui/internal/drv"
)

// Контекстное меню: чьё меню у точки (виджет, контейнер выше, окно), выбор пунктов,
// снятие меню, пункты уничтожаются вместе с окном.
func TestContextMenu(t *testing.T) {
	var w *Window
	var inBox, plain *Label
	var box *Box
	var tv *TextView
	var got []string
	say := func(s string) func() { return func() { got = append(got, s) } }
	boxIt, tvIt, winIt := Item("Ряд", say("ряд")), Item("Поле", say("поле")), Item("Окно", say("окно"))
	ui(func() {
		inBox, plain = NewLabel("подпись в ряду с меню"), NewLabel("подпись без меню")
		box = VBox(inBox)
		box.SetContextMenu(boxIt, Sep(), SubMenu("Ещё", Item("Вложенный", say("вложенный"))))
		tv = NewTextView()
		tv.SetContextMenu(tvIt)
		w = NewWindow("контекстное меню", VBox(box, Expand(tv), plain))
		w.SetContextMenu(winIt)
		w.SetSize(360, 300)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, _, pw, _ := alloc(plain); return pw > 1 })
	center := func(x Widget) (int, int) { ax, ay, aw, ah := alloc(x); return ax + aw/2, ay + ah/2 }
	ui(func() {
		for _, c := range []struct {
			name string
			at   Widget
			want drv.H
		}{{"подпись в ряду", inBox, box.h}, {"поле", tv, tv.h}, {"подпись без меню", plain, w.h}} {
			x, y := center(c.at)
			if got := drv.ContextMenuOwner(w.h, x, y); got != c.want {
				t.Errorf("%s (%d, %d): меню у %v, ожидалось у %v", c.name, x, y, got, c.want)
			}
		}
		drv.SimMenuActivate(boxIt.h)
		drv.SimMenuActivate(tvIt.h)
		drv.SimMenuActivate(winIt.h)
		if !slices.Equal(got, []string{"ряд", "поле", "окно"}) {
			t.Errorf("выбор пунктов: %v", got)
		}
		box.SetContextMenu() // снять: дальше — меню окна
		if boxIt.h != drv.Nil {
			t.Error("пункты снятого меню не уничтожены")
		}
		x, y := center(inBox)
		if got := drv.ContextMenuOwner(w.h, x, y); got != w.h {
			t.Errorf("после снятия меню ряда: меню у %v, ожидалось у окна", got)
		}
		expectPanic(t, "nui: пункт меню уже в меню", func() { plain.SetContextMenu(tvIt) })
		w.Close()
		if tvIt.h != drv.Nil || winIt.h != drv.Nil {
			t.Error("пункты контекстных меню не уничтожены вместе с окном")
		}
	})
}
