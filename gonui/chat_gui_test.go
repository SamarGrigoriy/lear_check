//go:build gui

package nui

import (
	"io"
	"strings"
	"testing"

	"nui/internal/drv"
)

// Отправка: Enter в поле — сообщение пользователя в ленте, поле пустое, OnSend получил
// текст; пустой ввод и ввод, пока чат занят, не отправляются.
func TestChatSend(t *testing.T) {
	var c *Chat
	var w *Window
	var got []string
	ui(func() {
		c = NewChat()
		c.OnSend(func(s string) { got = append(got, s) })
		w = NewWindow("чат", c)
		w.SetSize(400, 300)
		w.Show()
		drv.SimType(c.input.view, "  ")
		drv.SimEnter(c.input.view)
		drv.SimType(c.input.view, "Привет 🙂")
		drv.SimEnter(c.input.view)
		if len(got) != 1 || got[0] != "Привет 🙂" || c.input.Text() != "" {
			t.Errorf("отправка: OnSend %q, поле %q", got, c.input.Text())
		}
		if want := "Вы:\nПривет 🙂\n\n"; c.Text() != want {
			t.Errorf("лента %q, ожидалось %q", c.Text(), want)
		}
		c.SetBusy(true)
		drv.SimType(c.input.view, "ещё")
		drv.SimEnter(c.input.view)
		if len(got) != 1 || c.send.Enabled() {
			t.Errorf("занятый чат: OnSend %d раз, «Отправить» доступна: %v", len(got), c.send.Enabled())
		}
		c.SetBusy(false)
		drv.SimClick(c.send.h) // кнопка отправляет то же, что Enter
		if len(got) != 2 || got[1] != "ещё" {
			t.Errorf("кнопка: OnSend %q", got)
		}
		c.SetLabels("Я", "Спросить", "")
		c.Add("", "служебное")
		if !strings.HasSuffix(c.Text(), "Вы:\nещё\n\nслужебное\n\n") || c.send.Text() != "Спросить" {
			t.Errorf("Add без подписи: %q, кнопка %q", c.Text(), c.send.Text())
		}
	})
	shot(t, "chat-send")
	ui(func() { w.Close() })
}

// Ответ частями из горутины: io.Copy в поток, символ UTF-8 разрезан между Write;
// после Close чат свободен, запись в закрытый поток — ErrChatClosed.
func TestChatStream(t *testing.T) {
	var c *Chat
	var w *Window
	var st *ChatStream
	ui(func() {
		c = NewChat()
		w = NewWindow("чат", c)
		w.SetSize(400, 300)
		w.Show()
		st = c.Stream("ИИ")
		if !c.Busy() || c.send.Enabled() {
			t.Error("Stream должен делать чат занятым")
		}
	})
	done := make(chan struct{})
	go func() { // не UI-поток: так пишет ответ сетевой клиент
		defer close(done)
		b := []byte("Здравствуйте! ")
		st.Write(b[:1]) // половина «З»
		st.Write(b[1:])
		io.Copy(st, strings.NewReader("Чем помочь?"))
		st.Close()
	}()
	<-done
	waitUntil(t, 2000, func() bool { return !c.Busy() })
	ui(func() {
		if want := "ИИ:\nЗдравствуйте! Чем помочь?\n\n"; c.Text() != want {
			t.Errorf("лента %q, ожидалось %q", c.Text(), want)
		}
		if _, err := st.WriteString("поздно"); err != ErrChatClosed {
			t.Errorf("запись после Close: %v", err)
		}
		if !c.send.Enabled() {
			t.Error("после Close «Отправить» должна стать доступной")
		}
	})
	ui(func() { w.Close() })
}

// «Стоп»: видна, пока чат занят и задан OnStop; нажатие заканчивает сообщение.
// Новое сообщение посреди потока сначала заканчивает поток. Clear очищает ленту.
func TestChatStopAndInterleave(t *testing.T) {
	var c *Chat
	var w *Window
	stopped := 0
	ui(func() {
		c = NewChat()
		w = NewWindow("чат", c)
		w.SetSize(400, 300)
		w.Show()
		st := c.Stream("ИИ")
		if c.stop.Visible() {
			t.Error("«Стоп» без OnStop видна")
		}
		c.OnStop(func() { stopped++ })
		if !c.stop.Visible() {
			t.Error("«Стоп» не видна, пока чат занят")
		}
		st.WriteString("начало ответа")
		drv.SimClick(c.stop.h)
		if stopped != 1 || c.Busy() || c.stop.Visible() {
			t.Errorf("Стоп: OnStop %d, занят %v, кнопка видна %v", stopped, c.Busy(), c.stop.Visible())
		}
		if _, err := st.WriteString("хвост"); err != ErrChatClosed {
			t.Errorf("запись после «Стоп»: %v", err)
		}
		if want := "ИИ:\nначало ответа\n\n"; c.Text() != want {
			t.Errorf("после «Стоп» лента %q, ожидалось %q", c.Text(), want)
		}

		st = c.Stream("ИИ")
		st.WriteString("первый")
		c.Add("Вы", "перебиваю")
		if want := "ИИ:\nначало ответа\n\nИИ:\nпервый\n\nВы:\nперебиваю\n\n"; c.Text() != want {
			t.Errorf("Add посреди потока: %q", c.Text())
		}
		c.Clear()
		if c.Text() != "" || c.Busy() {
			t.Errorf("Clear: лента %q, занят %v", c.Text(), c.Busy())
		}
	})
	ui(func() { w.Close() })
}

// Уничтоженный чат: методы ничего не делают, поток не пишет в ленту и не падает.
func TestChatDestroyed(t *testing.T) {
	var c *Chat
	var st *ChatStream
	ui(func() {
		c = NewChat()
		w := NewWindow("чат", c)
		st = c.Stream("ИИ")
		w.Close()
	})
	st.WriteString("в пустоту") // горутина ещё пишет, а окна уже нет
	st.Close()
	waitUntil(t, 1000, func() bool { return !c.Alive() })
	ui(func() {
		c.Add("Вы", "x")
		c.Clear()
		c.SetLabels("a", "b", "c")
		c.Focus()
		if c.Text() != "" || c.Busy() {
			t.Error("уничтоженный чат должен молчать")
		}
		c.Stream("ИИ").WriteString("ничего")
	})
}

// Многострочное сообщение: Shift+Enter переводит строку, Enter отправляет всё сообщение.
func TestChatMultiline(t *testing.T) {
	var c *Chat
	var w *Window
	var got []string
	ui(func() {
		c = NewChat()
		c.OnSend(func(s string) { got = append(got, s) })
		w = NewWindow("чат", c)
		w.SetSize(400, 300)
		w.Show()
		c.Focus()
		drv.SimType(c.input.view, "первая строка")
		drv.SimKey(w.h, drv.KeyEnter, drv.ModShift) // перевод строки, не отправка
	})
	waitUntil(t, 2000, func() bool { return strings.Contains(c.input.Text(), "\n") })
	ui(func() {
		if len(got) != 0 {
			t.Fatalf("Shift+Enter отправил сообщение: %q", got)
		}
		drv.SimType(c.input.view, "первая строка\nвторая строка")
		drv.SimEnter(c.input.view)
		if len(got) != 1 || got[0] != "первая строка\nвторая строка" {
			t.Errorf("OnSend: %q", got)
		}
		if want := "Вы:\nпервая строка\nвторая строка\n\n"; c.Text() != want {
			t.Errorf("лента %q, ожидалось %q", c.Text(), want)
		}
		w.Close()
	})
}
