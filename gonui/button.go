package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Button — кнопка с надписью.
type Button struct {
	widget
	onClick func()
}

// NewButton создаёт кнопку; onClick может быть nil.
func NewButton(text string, onClick func()) *Button {
	checkCreate()
	b := &Button{onClick: onClick}
	b.id = core.Register((*buttonEv)(b))
	b.h = drv.ButtonNew(b.id, text)
	return b
}

// SetText задаёт надпись.
func (b *Button) SetText(s string) {
	if !b.dead() {
		b.quiet++
		drv.ButtonSetText(b.h, s)
		b.quiet--
	}
}

// Text — надпись.
func (b *Button) Text() string {
	if b.dead() {
		return ""
	}
	return drv.ButtonText(b.h)
}

// SetColor — цвет текста кнопки; NoColor — цвет темы.
func (b *Button) SetColor(c Color) {
	if !b.dead() {
		drv.SetForeground(b.h, uint32(c))
	}
}

// SetBackground — цвет кнопки; NoColor — вид темы. На Windows цветная кнопка рисуется
// драйвером (плоская, без визуального стиля системы).
func (b *Button) SetBackground(c Color) {
	if !b.dead() {
		drv.SetBackground(b.h, uint32(c), false)
	}
}

// OnClick: нажатие мышью, пробелом или Enter (если кнопка по умолчанию).
// Повторный вызов заменяет обработчик, nil снимает его (ADR-8).
func (b *Button) OnClick(fn func()) { b.onClick = fn }

type buttonEv Button

func (b *buttonEv) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvClick:
		if b.quiet == 0 && b.onClick != nil {
			b.onClick()
		}
	case core.EvDestroy:
		b.destroyed()
	}
	return 0
}
