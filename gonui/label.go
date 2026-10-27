package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Label — метка: текст как есть, без разметки, выровненный по началу.
type Label struct{ widget }

// Ellipsis — как обрезать текст, который не помещается.
type Ellipsis uint8

const (
	EllipsisNone   Ellipsis = iota // не обрезать: метка не даёт окну стать уже текста (по умолчанию)
	EllipsisStart                  // «…конец»
	EllipsisMiddle                 // «нача…нец» — для путей к файлам
	EllipsisEnd                    // «начало…»
)

// NewLabel создаёт метку с текстом.
func NewLabel(text string) *Label {
	checkCreate()
	l := &Label{}
	l.id = core.Register((*labelEv)(l))
	l.h = drv.LabelNew(l.id, text)
	return l
}

// SetText задаёт текст метки.
func (l *Label) SetText(s string) {
	if !l.dead() {
		drv.LabelSetText(l.h, s)
	}
}

// Text — текст метки.
func (l *Label) Text() string {
	if l.dead() {
		return ""
	}
	return drv.LabelText(l.h)
}

// SetEllipsis: если текст не помещается, он обрезается многоточием.
func (l *Label) SetEllipsis(e Ellipsis) {
	if !l.dead() {
		drv.LabelSetEllipsis(l.h, uint8(e))
	}
}

// SetWrap включает перенос по словам.
func (l *Label) SetWrap(on bool) {
	if !l.dead() {
		drv.LabelSetWrap(l.h, on)
	}
}

// SetSelectable: текст можно выделить и скопировать.
func (l *Label) SetSelectable(on bool) {
	if !l.dead() {
		drv.LabelSetSelectable(l.h, on)
	}
}

// SetTextAlign — выравнивание текста внутри метки: Start (по умолчанию), Center или End;
// Fill считается Start.
func (l *Label) SetTextAlign(a Align) {
	if !l.dead() {
		drv.LabelSetTextAlign(l.h, uint8(a))
	}
}

// SetBold — жирное начертание. Вместе с SetColor — единственные средства оформления в nui.
func (l *Label) SetBold(on bool) {
	if !l.dead() {
		drv.LabelSetBold(l.h, on)
	}
}

// SetColor — цвет текста; NoColor возвращает цвет темы.
func (l *Label) SetColor(c Color) {
	if !l.dead() {
		drv.LabelSetColor(l.h, uint32(c))
	}
}

// SetBackground — фон подписи; NoColor — фон контейнера или окна (по умолчанию).
func (l *Label) SetBackground(c Color) {
	if !l.dead() {
		drv.SetBackground(l.h, uint32(c), false)
	}
}

// SetFontScale — размер шрифта относительно обычного: 1.5 — заголовок, 0.85 — мелкий
// текст. Жирность и цвет сохраняются.
func (l *Label) SetFontScale(f float64) {
	if !l.dead() {
		drv.LabelSetFontScale(l.h, int(f*100+0.5))
	}
}

type labelEv Label

func (l *labelEv) Event(k core.Kind, _ uintptr) uintptr {
	if k == core.EvDestroy {
		l.destroyed()
	}
	return 0
}
