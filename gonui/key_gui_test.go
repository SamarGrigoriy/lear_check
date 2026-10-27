//go:build gui

package nui

import (
	"strings"
	"testing"

	"nui/internal/drv"
)

// OnKey: окно видит клавишу раньше поля с фокусом; обработанная (true) до поля не
// доходит, остальные доходят; nil снимает обработчик.
func TestOnKey(t *testing.T) {
	type press struct {
		k Key
		m Mod
	}
	var w *Window
	var e *Entry
	var got []press
	enters := 0
	ui(func() {
		e = NewEntry()
		e.OnEnter(func() { enters++ })
		w = NewWindow("клавиши", VBox(e, NewLabel("Ctrl+S, F5")))
		w.OnKey(func(k Key, m Mod) bool {
			got = append(got, press{k, m})
			return k == 'S' && m == ModCtrl || k == KeyF1+4 || k == KeyEnter && m == ModCtrl
		})
		w.Show()
		e.Focus()
	})
	ui(func() {
		drv.SimKey(w.h, 'S', drv.ModCtrl) // окно забирает
		drv.SimKey(w.h, drv.KeyF1+4, 0)   // окно забирает
		drv.SimKey(w.h, 'X', 0)           // окно отдаёт — буква в поле
		drv.SimKey(w.h, drv.KeyEnter, drv.ModCtrl)
		drv.SimKey(w.h, '1', drv.ModShift|drv.ModAlt)
	})
	waitUntil(t, 2000, func() bool { return e.Text() == "x" })
	ui(func() {
		want := []press{{'S', ModCtrl}, {KeyF1 + 4, 0}, {'X', 0}, {KeyEnter, ModCtrl}, {'1', ModShift | ModAlt}}
		if len(got) != len(want) {
			t.Fatalf("OnKey получил %v, ожидалось %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("нажатие %d: %v, ожидалось %v", i, got[i], want[i])
			}
		}
		if enters != 0 {
			t.Errorf("Ctrl+Enter обработан окном, а OnEnter поля вызван %d раз", enters)
		}
		drv.SimKeyPosted(w.h, drv.KeyF1+4) // через цикл сообщений (на Windows — WM_KEYDOWN)
	})
	waitUntil(t, 2000, func() bool { return len(got) == 6 })
	ui(func() {
		if got[5] != (press{KeyF1 + 4, 0}) {
			t.Errorf("клавиша через цикл: %v", got[5])
		}
		w.OnKey(nil)
		drv.SimKey(w.h, 'Y', 0)
	})
	waitUntil(t, 2000, func() bool { return e.Text() == "xy" }) // без обработчика всё — полю
	ui(func() {
		if len(got) != 6 {
			t.Errorf("после OnKey(nil) обработчик вызван ещё %d раз", len(got)-6)
		}
		w.Close()
	})
}

// TextView.OnEnter: Enter без Shift — обработчик, текст не меняется; Shift+Enter — перевод
// строки; nil — Enter снова переводит строку. Подсказка в пустом поле не мешает тексту.
func TestTextViewEnter(t *testing.T) {
	var w *Window
	var tv *TextView
	enters := 0
	ui(func() {
		tv = NewTextView()
		tv.SetWrap(true)
		tv.SetPlaceholder("Сообщение… (Shift+Enter — новая строка)")
		tv.OnEnter(func() { enters++ })
		w = NewWindow("Enter", Expand(tv))
		w.SetSize(320, 120)
		w.Show()
		tv.Focus()
	})
	shot(t, "textview-placeholder")
	ui(func() {
		drv.SimType(tv.view, "привет")
		drv.SimEnter(tv.view)
		if enters != 1 || tv.Text() != "привет" {
			t.Errorf("Enter: OnEnter %d раз, текст %q", enters, tv.Text())
		}
		drv.SimKey(w.h, drv.KeyEnter, drv.ModShift)
	})
	waitUntil(t, 2000, func() bool { return strings.Count(tv.Text(), "\n") == 1 })
	ui(func() {
		if enters != 1 {
			t.Errorf("Shift+Enter вызвал OnEnter")
		}
		tv.SetWrap(false) // на Windows поле пересоздаётся: OnEnter должен остаться
		drv.SimEnter(tv.view)
		if enters != 2 {
			t.Errorf("после SetWrap Enter: OnEnter %d раз, ожидалось 2", enters)
		}
		tv.OnEnter(nil)
		tv.SetPlaceholder("")
		w.Close()
	})
}
