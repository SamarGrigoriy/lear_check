//go:build gui

package nui

import (
	"slices"
	"strings"
	"testing"
	"time"

	"nui/internal/drv"
)

// LoadJSON: свой формат и формат API моделей; ошибки не трогают ленту; JSON() — туда и
// обратно; Stream попадает в историю.
func TestChatJSON(t *testing.T) {
	ui(func() {
		c := NewChat()
		w := NewWindow("чат из JSON", c)
		w.SetSize(400, 300)
		w.Show()

		own := `[
			{"who": "Вы", "text": "Привет! Как сказать \"кот\"?"},
			{"who": "ИИ", "text": "Cat 🙂\nЕщё строка", "лишнее": 1},
			{"who": "", "text": "служебное"},
			{"who": "Вы", "text": ""}
		]`
		if err := c.LoadJSON([]byte(own)); err != nil {
			t.Fatalf("свой формат: %v", err)
		}
		want := []ChatMessage{{"Вы", `Привет! Как сказать "кот"?`}, {"ИИ", "Cat 🙂\nЕщё строка"}, {"", "служебное"}}
		if got := c.Messages(); !slices.Equal(got, want) {
			t.Errorf("Messages = %q, ожидалось %q", got, want)
		}
		if lent := "Вы:\nПривет! Как сказать \"кот\"?\n\nИИ:\nCat 🙂\nЕщё строка\n\nслужебное\n\n"; c.Text() != lent {
			t.Errorf("лента %q, ожидалось %q", c.Text(), lent)
		}
		if c.LastMessage() != "служебное" {
			t.Errorf("LastMessage = %q", c.LastMessage())
		}

		c.SetLabels("Я", "", "")
		api := `{"model": "x", "messages": [
			{"role": "system", "content": "Ты помощник."},
			{"role": "user", "content": [{"type": "text", "text": "Что "}, {"type": "image_url", "image_url": {"url": "data:..."}}, {"type": "text", "text": "на фото?"}]},
			{"role": "assistant", "content": null, "tool_calls": []},
			{"role": "tool", "content": "{\"ok\": true}"},
			{"role": "assistant", "content": "Кот."},
			{"role": "user", "name": "Гриша", "content": "Спасибо"}
		]}`
		if err := c.LoadJSON([]byte(api)); err != nil {
			t.Fatalf("формат API: %v", err)
		}
		want = []ChatMessage{{"Я", "Что на фото?"}, {"ИИ", "Кот."}, {"Гриша", "Спасибо"}}
		if got := c.Messages(); !slices.Equal(got, want) {
			t.Errorf("Messages = %q, ожидалось %q", got, want)
		}

		before := c.Text()
		for _, bad := range []string{`[{"who": "Вы", "text": "обрыв`, `{"who": "Вы"}`, `[1, 2]`, `"строка"`} {
			if err := c.LoadJSON([]byte(bad)); err == nil {
				t.Errorf("%s: ожидалась ошибка", bad)
			} else if !strings.HasPrefix(err.Error(), "nui: JSON чата: ") {
				t.Errorf("текст ошибки: %v", err)
			}
		}
		if c.Text() != before {
			t.Error("ошибка LoadJSON изменила ленту")
		}

		saved := c.JSON() // туда и обратно
		c.Clear()
		if err := c.LoadJSON(saved); err != nil || !slices.Equal(c.Messages(), want) {
			t.Errorf("LoadJSON(JSON()) = %q, %v; JSON:\n%s", c.Messages(), err, saved)
		}

		st := c.Stream("ИИ")
		st.WriteString("часть 1, ")
		st.WriteString("часть 2")
		st.Close()
		c.flush()
		if m := c.Messages(); m[len(m)-1] != (ChatMessage{"ИИ", "часть 1, часть 2"}) {
			t.Errorf("после Stream: %q", m[len(m)-1])
		}
		w.Close()
		if c.Messages() != nil || c.LoadJSON([]byte("[]")) != nil {
			t.Error("уничтоженный чат: Messages не nil или LoadJSON с ошибкой")
		}
	})
}

// OnReply: ответ из горутины — один вызов с полным текстом, история уже записана;
// «Стоп» — OnReply раньше OnStop; новое сообщение поверх — OnReply до него; Add без
// Stream и уничтоженный чат — без вызова.
func TestChatOnReply(t *testing.T) {
	var c *Chat
	var w *Window
	var events []string
	ui(func() {
		c = NewChat()
		c.OnReply(func(m ChatMessage) {
			saved := string(c.JSON())
			events = append(events, "reply:"+m.Who+":"+m.Text)
			if !strings.Contains(saved, m.Text) || c.Busy() {
				t.Errorf("к OnReply история без ответа или чат занят: %s", saved)
			}
		})
		c.OnStop(func() { events = append(events, "stop") })
		w = NewWindow("OnReply", c)
		w.SetSize(400, 300)
		w.Show()
		st := c.Stream("ИИ")
		go func() {
			st.WriteString("Ответ ")
			st.WriteString("целиком.")
			st.Close()
		}()
	})
	waitUntil(t, 2000, func() bool { return len(events) == 1 })
	ui(func() {
		if events[0] != "reply:ИИ:Ответ целиком." {
			t.Errorf("OnReply: %q", events)
		}
		c.Add("Вы", "без Stream") // обычное сообщение — без OnReply
		st := c.Stream("ИИ")
		st.WriteString("на полуслове")
		c.flush()
		drv.SimClick(c.stop.h) // «Стоп»
		st = c.Stream("ИИ")
		st.WriteString("прервано")
		c.flush()
		c.Add("Вы", "новое") // заканчивает ответ — OnReply до нового сообщения
	})
	ui(func() {
		want := []string{"reply:ИИ:Ответ целиком.", "reply:ИИ:на полуслове", "stop", "reply:ИИ:прервано"}
		if !slices.Equal(events, want) {
			t.Errorf("события %q, ожидалось %q", events, want)
		}
		if m := c.Messages(); m[len(m)-1] != (ChatMessage{"Вы", "новое"}) || m[len(m)-2].Text != "прервано" {
			t.Errorf("история: %q", m)
		}
		c.Stream("ИИ").WriteString("не дойдёт")
		w.Close() // чат уничтожен: таймер молча останавливается, OnReply нет
	})
	time.Sleep(200 * time.Millisecond) // таймер чата — каждые 40 мс
	ui(func() {
		if len(events) != 4 {
			t.Errorf("после уничтожения чата: %q", events)
		}
	})
}
