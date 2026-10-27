package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Separator — линия-разделитель.
type Separator struct{ widget }

// NewSeparator — горизонтальная линия (для VBox).
func NewSeparator() *Separator { return newSeparator(false) }

// NewVSeparator — вертикальная линия (для HBox).
func NewVSeparator() *Separator { return newSeparator(true) }

func newSeparator(vertical bool) *Separator {
	checkCreate()
	s := &Separator{}
	s.id = core.Register((*separatorEv)(s))
	s.h = drv.SeparatorNew(s.id, vertical)
	return s
}

type separatorEv Separator

func (s *separatorEv) Event(k core.Kind, _ uintptr) uintptr {
	if k == core.EvDestroy {
		s.destroyed()
	}
	return 0
}

// Frame — рамка с заголовком вокруг одного виджета.
type Frame struct{ widget }

// NewFrame создаёт рамку; title может быть "", child — nil.
func NewFrame(title string, child Widget) *Frame {
	checkCreate()
	f := &Frame{}
	f.id = core.Register((*frameEv)(f))
	f.h = drv.FrameNew(f.id, title)
	if child != nil {
		f.SetChild(child)
	}
	return f
}

// SetTitle задаёт заголовок рамки; "" убирает его.
func (f *Frame) SetTitle(s string) {
	if !f.dead() {
		drv.FrameSetTitle(f.h, s)
	}
}

// SetChild заменяет содержимое рамки; старый ребёнок уничтожается. nil — только убрать.
func (f *Frame) SetChild(c Widget) { setChild(&f.widget, c) }

type frameEv Frame

func (f *frameEv) Event(k core.Kind, _ uintptr) uintptr {
	if k == core.EvDestroy {
		f.destroyed()
	}
	return 0
}

// Scroll — полосы прокрутки вокруг виджета, если он не помещается.
type Scroll struct{ widget }

// NewScroll создаёт прокрутку вокруг child (может быть nil).
func NewScroll(child Widget) *Scroll {
	checkCreate()
	s := &Scroll{}
	s.id = core.Register((*scrollEv)(s))
	s.h = drv.ScrollNew(s.id)
	if child != nil {
		setChild(&s.widget, child)
	}
	return s
}

type scrollEv Scroll

func (s *scrollEv) Event(k core.Kind, _ uintptr) uintptr {
	if k == core.EvDestroy {
		s.destroyed()
	}
	return 0
}

// setChild — общее для Frame и Scroll: заменить единственного ребёнка.
func setChild(parent *widget, c Widget) {
	if parent.dead() {
		return
	}
	ch := drv.Nil
	if c != nil {
		b := c.base()
		if b.dead() {
			return
		}
		b.place()
		ch = b.h
	}
	drv.ContainerSetChild(parent.h, ch)
}
