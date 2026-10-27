package nui

import (
	"errors"
	"sync"
	"unicode/utf8"

	"nui/internal/core"
	"nui/internal/drv"
)

// Chat — готовый чат (api.md, §18): лента сообщений, поле ввода, кнопки «Отправить»
// и «Стоп». Собран из TextView и Button, поэтому выглядит родным на каждой ОС. Поле ввода
// многострочное: Enter отправляет, Shift+Enter переводит строку.
// Ответ, который приходит частями (например, от ИИ), пишется через Stream — из любой
// горутины.
//
//	chat := nui.NewChat()
//	chat.OnSend(func(q string) {
//		st := chat.Stream("ИИ")
//		go func() { defer st.Close(); ask(q, st) }() // ask пишет ответ в io.Writer
//	})
type Chat struct {
	*widget // корень — VBox: общие методы виджета (SetVisible, SetMinSize, Alive …)

	log         *TextView
	input       *TextView
	send, stop  *Button
	you         string
	onSend      func(text string)
	onStop      func()
	onReply     func(m ChatMessage)
	busy        bool
	cur         *ChatStream   // сообщение, которое сейчас дописывается
	tick        *Timer        // переносит текст cur в ленту
	lastNewline bool          // лента кончается переводом строки
	msgs        []ChatMessage // история: Messages, JSON, LastMessage
}

// ChatMessage — сообщение чата: подпись и текст.
type ChatMessage struct {
	Who  string // подпись: «Вы», «ИИ»…; "" — без подписи
	Text string
}

// NewChat создаёт чат. Лента растягивается: задайте окну размер (SetSize) или вставьте
// чат через Expand.
func NewChat() *Chat {
	c := &Chat{you: "Вы", lastNewline: true}
	c.log = NewTextView()
	c.log.SetReadOnly(true)
	c.log.SetWrap(true)
	c.log.SetAutoScroll(true)
	c.input = NewTextView()
	c.input.SetWrap(true)
	c.input.OnEnter(c.submit)
	c.send = NewButton("Отправить", c.submit)
	c.stop = NewButton("Стоп", c.pressStop)
	c.stop.SetVisible(false)
	c.send.SetAlign(AlignFill, AlignEnd) // кнопки — внизу строки ввода, как в мессенджерах
	c.stop.SetAlign(AlignFill, AlignEnd)
	root := VBox(Expand(c.log), HBox(Expand(c.input), c.stop, c.send))
	c.widget = &root.widget
	return c
}

// OnSend: пользователь отправил непустой текст (Enter или «Отправить»; Shift+Enter —
// новая строка в сообщении). К вызову fn
// поле уже очищено, а сообщение пользователя уже в ленте. Пока чат занят (SetBusy,
// идёт Stream), отправить нельзя.
func (c *Chat) OnSend(fn func(text string)) { c.onSend = fn }

// OnReply: сообщение, начатое Stream, закончено — Close, «Стоп» или новое сообщение
// (Add, Stream, Clear, LoadJSON) поверх, — и весь его текст уже в ленте и в истории
// (Messages, JSON). m — это сообщение целиком. Вызывается в UI-потоке, один раз на
// сообщение; если сообщение закончило новое, — до нового. Здесь удобно сохранять историю:
//
//	chat.OnReply(func(nui.ChatMessage) { os.WriteFile("история.json", chat.JSON(), 0o644) })
func (c *Chat) OnReply(fn func(m ChatMessage)) { c.onReply = fn }

// OnStop: пользователь нажал «Стоп». Кнопка видна, только пока чат занят и fn задан.
// Текущее сообщение (Stream) к вызову fn уже закончено: дальнейшая запись в него вернёт
// ErrChatClosed — так io.Copy сам остановится.
func (c *Chat) OnStop(fn func()) {
	c.onStop = fn
	if !c.dead() {
		c.stop.SetVisible(c.busy && fn != nil)
	}
}

// Add добавляет готовое сообщение: «who:», текст и пустую строку. who == "" — без
// подписи. Сообщение, которое дописывалось (Stream), сначала заканчивается.
func (c *Chat) Add(who, text string) {
	if c.dead() {
		return
	}
	c.finish()
	c.appendLog(header(who) + text)
	c.endMessage()
	c.msgs = append(c.msgs, ChatMessage{who, text})
}

// Stream начинает сообщение, которое приходит частями, и делает чат занятым. Писать
// в него можно из любой горутины; когда ответ закончен — Close. Предыдущее
// незаконченное сообщение заканчивается.
func (c *Chat) Stream(who string) *ChatStream {
	s := &ChatStream{}
	if c.dead() {
		s.closed = true
		return s
	}
	c.finish()
	c.appendLog(header(who))
	c.msgs = append(c.msgs, ChatMessage{Who: who})
	c.cur = s
	c.setBusy(true)
	c.tick = Every(40, c.flush)
	return s
}

// SetBusy — чат ждёт ответа: «Отправить» недоступна, «Стоп» видна (если задан OnStop).
// Stream включает это сам и выключает, когда сообщение закончено.
func (c *Chat) SetBusy(on bool) {
	if !c.dead() {
		c.setBusy(on)
	}
}

// Busy — чат ждёт ответа.
func (c *Chat) Busy() bool { return c.busy && !c.dead() }

// Clear очищает ленту; незаконченное сообщение заканчивается.
func (c *Chat) Clear() {
	if c.dead() {
		return
	}
	c.finish()
	c.log.Clear()
	c.lastNewline = true
	c.msgs = nil
}

// LastMessage — текст последнего сообщения без подписи (у незаконченного — то, что уже
// в ленте). Например, для кнопки «Скопировать ответ»: nui.SetClipboard(chat.LastMessage()).
func (c *Chat) LastMessage() string {
	if c.dead() || len(c.msgs) == 0 {
		return ""
	}
	return c.msgs[len(c.msgs)-1].Text
}

// Messages — история: все сообщения ленты по порядку (у незаконченного Stream — то, что
// уже пришло). Её же удобно отправлять модели как контекст разговора.
func (c *Chat) Messages() []ChatMessage {
	if c.dead() {
		return nil
	}
	return append([]ChatMessage(nil), c.msgs...)
}

// JSON — история в формате LoadJSON: [{"who": "Вы", "text": "…"}, …], по сообщению в
// строке. Сохраните её в файл (os.WriteFile), а при запуске загрузите LoadJSON.
func (c *Chat) JSON() []byte {
	b := []byte("[")
	for i, m := range c.Messages() {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, "\n  {\"who\": "...)
		b = append(b, core.QuoteJSON(m.Who)...)
		b = append(b, ", \"text\": "...)
		b = append(b, core.QuoteJSON(m.Text)...)
		b = append(b, '}')
	}
	return append(b, "\n]\n"...)
}

// LoadJSON показывает историю из JSON: лента очищается, затем все сообщения — по порядку,
// как Add. Ошибка — лента не меняется. Понимает два вида (api.md, §18):
//
//	[{"who": "Вы", "text": "Привет"}, {"who": "ИИ", "text": "Здравствуйте!"}]
//	{"messages": [{"role": "user", "content": "Привет"}, {"role": "assistant", "content": "…"}]}
//
// Второй — формат API моделей. Подпись — поле "who" (или "name"); без него — по "role":
// user — подпись пользователя (SetLabels, по умолчанию «Вы»), assistant — «ИИ»; сообщения
// system, developer и tool не показываются. Текст — "text" или "content": строка или
// массив частей [{"type": "text", "text": "…"}] (картинки пропускаются). Сообщения без
// текста пропускаются, неизвестные поля — тоже.
func (c *Chat) LoadJSON(data []byte) error {
	if c.dead() {
		return nil
	}
	v, err := core.ParseJSON(data)
	if err != nil {
		return errors.New("nui: JSON чата: " + err.Error())
	}
	if m, ok := v.Get("messages"); ok && v.Kind == 'o' {
		v = m
	}
	if v.Kind != 'a' {
		return errors.New("nui: JSON чата: нужен массив сообщений или объект с полем \"messages\"")
	}
	var msgs []ChatMessage
	for i, m := range v.Arr {
		if m.Kind != 'o' {
			return errors.New("nui: JSON чата: сообщение " + core.Itoa(i+1) + " — не объект")
		}
		who, show := c.jsonWho(m)
		text := jsonText(m)
		if show && text != "" {
			msgs = append(msgs, ChatMessage{who, text})
		}
	}
	c.Clear()
	for _, m := range msgs {
		c.Add(m.Who, m.Text)
	}
	return nil
}

// jsonWho — подпись сообщения; show = false — сообщение не для ленты (system, tool…).
func (c *Chat) jsonWho(m core.JSON) (who string, show bool) {
	for _, k := range []string{"who", "name"} {
		if v, ok := m.Get(k); ok && v.Kind == 's' {
			return v.Str, true
		}
	}
	role, ok := m.Get("role")
	if !ok {
		return "", true
	}
	switch role.Str {
	case "user":
		return c.you, true
	case "assistant":
		return "ИИ", true
	case "system", "developer", "tool", "function":
		return "", false
	}
	return role.Str, true
}

// jsonText — текст сообщения: "text" или "content", строка или массив частей с "text".
func jsonText(m core.JSON) string {
	for _, k := range []string{"text", "content"} {
		v, ok := m.Get(k)
		switch {
		case !ok:
			continue
		case v.Kind == 's':
			return v.Str
		case v.Kind == 'a':
			s := ""
			for _, part := range v.Arr {
				if t, ok := part.Get("text"); ok && t.Kind == 's' {
					s += t.Str
				}
			}
			return s
		}
	}
	return ""
}

// Text — вся лента.
func (c *Chat) Text() string {
	if c.dead() {
		return ""
	}
	return c.log.Text()
}

// SetPlaceholder — подсказка в пустом поле ввода.
func (c *Chat) SetPlaceholder(s string) {
	if !c.dead() {
		c.input.SetPlaceholder(s)
	}
}

// SetLabels — подпись сообщений пользователя и надписи кнопок; "" — оставить как было.
// По умолчанию «Вы», «Отправить», «Стоп».
func (c *Chat) SetLabels(you, send, stop string) {
	if c.dead() {
		return
	}
	if you != "" {
		c.you = you
	}
	if send != "" {
		c.send.SetText(send)
	}
	if stop != "" {
		c.stop.SetText(stop)
	}
}

// SetColor — цвет текста ленты; NoColor — цвет темы.
func (c *Chat) SetColor(col Color) {
	if !c.dead() {
		c.log.SetColor(col)
	}
}

// SetBackground — фон чата: лента и полоса с полем ввода; NoColor — фон темы.
func (c *Chat) SetBackground(col Color) {
	if c.dead() {
		return
	}
	c.log.SetBackground(col)
	drv.SetBackground(c.h, uint32(col), true)
}

// SetContextMenu — меню по правому щелчку по ленте сообщений (Widget.SetContextMenu).
// Например, «Скопировать ответ»: nui.Item("Скопировать ответ", func() { nui.SetClipboard(chat.LastMessage()) }).
func (c *Chat) SetContextMenu(items ...*MenuItem) {
	if !c.dead() {
		c.log.SetContextMenu(items...)
	}
}

// Focus передаёт фокус полю ввода.
func (c *Chat) Focus() {
	if !c.dead() {
		c.input.Focus()
	}
}

func (c *Chat) submit() {
	if c.dead() || c.busy {
		return
	}
	text := c.input.Text()
	if blank(text) {
		return
	}
	c.input.SetText("")
	c.Add(c.you, text)
	if c.onSend != nil {
		c.onSend(text)
	}
}

func (c *Chat) pressStop() {
	if c.cur != nil {
		c.cur.Close()
		c.flush()
	}
	if c.onStop != nil {
		c.onStop()
	}
}

func (c *Chat) setBusy(on bool) {
	c.busy = on
	c.send.SetEnabled(!on)
	c.stop.SetVisible(on && c.onStop != nil)
}

// flush — таймер UI-потока: накопленное в cur — в ленту; закрытый cur — закончить.
func (c *Chat) flush() {
	s := c.cur
	if s == nil || !c.Alive() { // чат уничтожен вместе с окном: молча остановиться
		c.stopTick()
		c.cur = nil
		return
	}
	text, done := s.take(false)
	c.appendLog(text)
	c.streamed(text)
	if done {
		c.cur = nil
		c.stopTick()
		c.endMessage()
		c.setBusy(false)
		c.replied()
	}
}

// finish заканчивает незаконченное сообщение сразу: всё накопленное — в ленту.
func (c *Chat) finish() {
	s := c.cur
	if s == nil {
		return
	}
	s.Close()
	text, _ := s.take(true)
	c.appendLog(text)
	c.streamed(text)
	c.cur = nil
	c.stopTick()
	c.endMessage()
	c.setBusy(false)
	c.replied()
}

// replied — OnReply с последним сообщением истории (его только что закончил Stream).
func (c *Chat) replied() {
	if c.onReply != nil && len(c.msgs) > 0 {
		c.onReply(c.msgs[len(c.msgs)-1])
	}
}

// streamed дописывает пришедший текст Stream к последнему сообщению истории.
func (c *Chat) streamed(text string) {
	if n := len(c.msgs); n > 0 {
		c.msgs[n-1].Text += text
	}
}

func (c *Chat) stopTick() {
	if c.tick != nil {
		c.tick.Stop()
		c.tick = nil
	}
}

func (c *Chat) appendLog(s string) {
	if s == "" {
		return
	}
	c.log.Append(s)
	c.lastNewline = s[len(s)-1] == '\n'
}

// endMessage — пустая строка после сообщения.
func (c *Chat) endMessage() {
	if c.lastNewline {
		c.appendLog("\n")
	} else {
		c.appendLog("\n\n")
	}
}

func header(who string) string {
	if who == "" {
		return ""
	}
	return who + ":\n"
}

func blank(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\n', '\r':
		default:
			return false
		}
	}
	return true
}

// ErrChatClosed — запись в законченное сообщение чата: Close, «Стоп» или новое сообщение.
var ErrChatClosed = errors.New("nui: сообщение чата уже закончено")

// ChatStream — сообщение чата, которое приходит частями (Chat.Stream). Write, WriteString
// и Close можно вызывать из любой горутины — это исключение из правила api.md, §0:
// текст копится в буфере, а в ленту его переносит таймер UI-потока (примерно 25 раз
// в секунду). Символ UTF-8, разрезанный между двумя Write, не ломается.
type ChatStream struct {
	mu     sync.Mutex
	buf    []byte
	closed bool
}

// Write дописывает p к сообщению (io.Writer). После Close — ErrChatClosed.
func (s *ChatStream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, ErrChatClosed
	}
	s.buf = append(s.buf, p...)
	return len(p), nil
}

// WriteString — то же для строки (io.StringWriter).
func (s *ChatStream) WriteString(str string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, ErrChatClosed
	}
	s.buf = append(s.buf, str...)
	return len(str), nil
}

// Close заканчивает сообщение (io.Closer): остаток уйдёт в ленту, чат перестанет быть
// занятым. Повторный Close ничего не делает.
func (s *ChatStream) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	return nil
}

// take забирает накопленное. Пока сообщение не закончено (и не all), недописанный
// последний символ UTF-8 остаётся в буфере. done — закончено и всё забрано.
func (s *ChatStream) take(all bool) (text string, done bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.buf)
	if !s.closed && !all {
		n = fullRunes(s.buf)
	}
	text = string(s.buf[:n])
	s.buf = append(s.buf[:0], s.buf[n:]...)
	return text, s.closed && len(s.buf) == 0
}

// fullRunes — длина начала b, которое кончается целым символом UTF-8.
func fullRunes(b []byte) int {
	i := len(b) - 1
	for i >= 0 && len(b)-i < utf8.UTFMax && !utf8.RuneStart(b[i]) {
		i--
	}
	if i < 0 || utf8.FullRune(b[i:]) {
		return len(b)
	}
	return i
}
