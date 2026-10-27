package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Entry — однострочное поле ввода.
type Entry struct {
	widget
	onChange, onEnter func()
}

// NewEntry создаёт пустое поле ввода.
func NewEntry() *Entry {
	checkCreate()
	e := &Entry{}
	e.id = core.Register((*entryEv)(e))
	e.h = drv.EntryNew(e.id)
	return e
}

// SetText задаёт содержимое; OnChange при этом не вызывается (ADR-8).
func (e *Entry) SetText(s string) {
	if !e.dead() {
		e.quiet++
		drv.EntrySetText(e.h, s)
		e.quiet--
	}
}

// Text — содержимое поля.
func (e *Entry) Text() string {
	if e.dead() {
		return ""
	}
	return drv.EntryText(e.h)
}

// SetPlaceholder — серая подсказка в пустом поле.
func (e *Entry) SetPlaceholder(s string) {
	if !e.dead() {
		drv.EntrySetPlaceholder(e.h, s)
	}
}

// SetPassword скрывает ввод.
func (e *Entry) SetPassword(on bool) {
	if !e.dead() {
		drv.EntrySetPassword(e.h, on)
	}
}

// SetReadOnly: менять нельзя, но можно выделить и скопировать.
func (e *Entry) SetReadOnly(on bool) {
	if !e.dead() {
		drv.EntrySetReadOnly(e.h, on)
	}
}

// SetWidthChars — естественная ширина в символах.
func (e *Entry) SetWidthChars(n int) {
	if !e.dead() {
		drv.EntrySetWidthChars(e.h, n)
	}
}

// SetColor — цвет текста в поле; NoColor — цвет темы.
func (e *Entry) SetColor(c Color) {
	if !e.dead() {
		drv.SetForeground(e.h, uint32(c))
	}
}

// SetBackground — фон поля; NoColor — фон темы.
func (e *Entry) SetBackground(c Color) {
	if !e.dead() {
		drv.SetBackground(e.h, uint32(c), false)
	}
}

// OnChange: текст изменён пользователем (ввод, вставка, удаление).
func (e *Entry) OnChange(fn func()) { e.onChange = fn }

// OnEnter: нажат Enter. Если обработчик задан, кнопка окна по умолчанию не срабатывает;
// если не задан — срабатывает.
func (e *Entry) OnEnter(fn func()) {
	e.onEnter = fn
	if !e.dead() {
		drv.EntrySetActivatesDefault(e.h, fn == nil)
	}
}

type entryEv Entry

func (e *entryEv) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvChange:
		if e.quiet == 0 && e.onChange != nil {
			e.onChange()
		}
	case core.EvEnter:
		if e.quiet == 0 && e.onEnter != nil {
			e.onEnter()
		}
	case core.EvDestroy:
		e.destroyed()
	}
	return 0
}
