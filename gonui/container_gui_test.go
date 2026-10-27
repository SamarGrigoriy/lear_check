//go:build gui

package nui

import (
	"testing"
	"time"

	"nui/internal/drv"
)

// Box.Remove и Clear: удалённые уничтожены, соседи занимают место, чужой виджет — паника.
func TestBoxRemoveClear(t *testing.T) {
	var w *Window
	var box *Box
	var a, b, c *Label
	var chat *Chat
	ui(func() {
		a, b, c = NewLabel("первая"), NewLabel("вторая"), NewLabel("третья")
		chat = NewChat()
		box = VBox(a, b, c, Expand(chat))
		w = NewWindow("удаление", box)
		w.SetSize(300, 300)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, _, ww, _ := alloc(c); return ww > 1 })
	var by int
	ui(func() {
		_, by, _, _ = alloc(b)
		box.Remove(b, chat)
		if b.Alive() || chat.Alive() || !a.Alive() || !c.Alive() {
			t.Errorf("после Remove: a %v, b %v, c %v, чат %v", a.Alive(), b.Alive(), c.Alive(), chat.Alive())
		}
		box.Remove(b) // уже удалён — ничего
		expectPanic(t, "nui: Remove: виджет не в этом контейнере", func() { box.Remove(NewLabel("чужая")) })
	})
	waitUntil(t, 2000, func() bool { _, y, _, _ := alloc(c); return y == by }) // третья — на месте второй
	ui(func() {
		box.Clear()
		if a.Alive() || c.Alive() {
			t.Error("после Clear дети живы")
		}
		box.Add(NewLabel("новая"))
	})
	ui(func() { w.Close() })
}

// Удалённая кнопка по умолчанию больше не нажимается Enter — и новая кнопка не становится
// кнопкой по умолчанию вместо неё (на Windows хендлы узлов переиспользуются).
func TestRemoveDefaultButton(t *testing.T) {
	var w *Window
	var box *Box
	var e *Entry
	clicks := 0
	ui(func() {
		e = NewEntry()
		old := NewButton("OK", func() { t.Error("нажата удалённая кнопка") })
		box = VBox(e, old)
		w = NewWindow("кнопка по умолчанию", box)
		w.SetDefault(old)
		w.Show()
		box.Remove(old)
		if d := drv.DefaultButton(w.h); d != drv.Nil {
			t.Errorf("окно всё ещё ссылается на удалённую кнопку по умолчанию (%v)", d)
		}
		box.Add(NewButton("Новая", func() { clicks++ }))
		e.Focus()
		drv.SimEnter(e.h)
	})
	time.Sleep(300 * time.Millisecond) // Enter на Windows идёт через очередь
	ui(func() {
		if clicks != 0 {
			t.Errorf("Enter нажал новую кнопку %d раз: она не кнопка по умолчанию", clicks)
		}
		w.Close()
	})
}

// Grid.Remove освобождает ячейку (в неё можно положить другой виджет), Clear — все.
func TestGridRemoveClear(t *testing.T) {
	ui(func() {
		l1, l2 := NewLabel("1"), NewLabel("2")
		g := NewGrid().Add(0, 0, l1).Add(1, 0, l2)
		w := NewWindow("сетка", g)
		w.Show()
		g.Remove(l1)
		if l1.Alive() || !l2.Alive() {
			t.Errorf("после Remove: l1 %v, l2 %v", l1.Alive(), l2.Alive())
		}
		g.Add(0, 0, NewLabel("вместо 1")) // ячейка свободна: в отладочной сборке без паники о пересечении
		expectPanic(t, "nui: Remove: виджет не в этом контейнере", func() { g.Remove(NewLabel("чужая")) })
		g.Clear()
		if l2.Alive() {
			t.Error("после Clear l2 жива")
		}
		g.Add(0, 0, NewLabel("заново"))
		w.Close()
	})
}
