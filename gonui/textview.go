package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// TextView — многострочный текст, уже с полосами прокрутки.
type TextView struct {
	widget           // h — прокрутка: раскладка и общие методы
	view       drv.H // сам текст
	autoscroll bool
	maxLines   int
	onChange   func()
	onEnter    func()
}

// NewTextView создаёт пустое поле многострочного текста.
func NewTextView() *TextView {
	checkCreate()
	t := &TextView{}
	t.id = core.Register((*textViewEv)(t))
	t.h, t.view = drv.TextNew(t.id)
	return t
}

// SetText заменяет всё содержимое; OnChange не вызывается (ADR-8).
func (t *TextView) SetText(s string) {
	if !t.dead() {
		t.quiet++
		drv.TextSetText(t.view, s)
		t.quiet--
	}
}

// Text — всё содержимое.
func (t *TextView) Text() string {
	if t.dead() {
		return ""
	}
	return drv.TextText(t.view)
}

// Append дописывает в конец — быстрее, чем SetText(Text()+s). Учитывает
// SetMaxLines и SetAutoScroll. OnChange не вызывается.
func (t *TextView) Append(s string) {
	if !t.dead() {
		t.quiet++
		drv.TextAppend(t.view, s, t.autoscroll, t.maxLines)
		t.quiet--
	}
}

// Clear очищает содержимое.
func (t *TextView) Clear() { t.SetText("") }

// SetReadOnly — только просмотр.
func (t *TextView) SetReadOnly(on bool) {
	if !t.dead() {
		drv.TextSetReadOnly(t.view, on)
	}
}

// SetMonospace — моноширинный шрифт (логи, код).
func (t *TextView) SetMonospace(on bool) {
	if !t.dead() {
		drv.TextSetMonospace(t.view, on)
	}
}

// SetWrap — перенос по словам; без него появляется горизонтальная прокрутка.
func (t *TextView) SetWrap(on bool) {
	if !t.dead() {
		drv.TextSetWrap(t.view, on)
	}
}

// SetAutoScroll: после Append прокручивать в конец (по умолчанию выключено).
// Курсор и выделение пользователя при этом не сбрасываются.
func (t *TextView) SetAutoScroll(on bool) { t.autoscroll = on }

// SetMaxLines: при Append удалять старые строки сверх n; 0 — без ограничения.
// Строка — текст до \n включительно: завершающий \n не порождает лишней строки.
func (t *TextView) SetMaxLines(n int) {
	if n < 0 {
		n = 0
	}
	t.maxLines = n
}

// Focus передаёт фокус ввода самому тексту.
func (t *TextView) Focus() {
	if !t.dead() {
		drv.Focus(t.view)
	}
}

// SetColor — цвет текста; NoColor — цвет темы.
func (t *TextView) SetColor(c Color) {
	if !t.dead() {
		drv.SetForeground(t.view, uint32(c))
	}
}

// SetBackground — фон поля текста; NoColor — фон темы.
func (t *TextView) SetBackground(c Color) {
	if !t.dead() {
		drv.SetBackground(t.view, uint32(c), false)
	}
}

// OnEnter: Enter без Shift вызывает fn вместо перевода строки, Shift+Enter переводит
// строку — как поле ввода в чатах. Без обработчика (nil) Enter переводит строку.
// Enter, которым заканчивается набор в методе ввода (китайский, японский…), остаётся ему.
func (t *TextView) OnEnter(fn func()) {
	if !t.dead() {
		t.onEnter = fn
		drv.TextSetEnter(t.view, t.id, fn != nil)
	}
}

// SetPlaceholder — серая подсказка в пустом поле; "" убирает её.
func (t *TextView) SetPlaceholder(s string) {
	if !t.dead() {
		drv.TextSetPlaceholder(t.view, s)
	}
}

// OnChange: текст изменён пользователем. Вставка поверх выделения — это удаление
// и вставка, поэтому обработчик может сработать дважды.
func (t *TextView) OnChange(fn func()) { t.onChange = fn }

type textViewEv TextView

func (t *textViewEv) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvChange:
		if t.quiet == 0 && t.onChange != nil {
			t.onChange()
		}
	case core.EvEnter:
		if t.onEnter != nil {
			t.onEnter()
		}
	case core.EvDestroy:
		t.destroyed()
		t.view = drv.Nil
	}
	return 0
}
