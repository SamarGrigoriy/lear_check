package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Spin — числовое поле (api.md, §11): целое число в диапазоне, со стрелками «больше» и
// «меньше». Значение принимается стрелками (и клавишами ↑ ↓), а набранное — по Enter или
// когда фокус уходит из поля: тогда оно ограничивается диапазоном, пустое поле — 0 (тоже
// в диапазоне). Буквы в поле не набираются.
//
//	port := nui.NewSpin(1, 65535)
//	port.SetValue(8080)
type Spin struct {
	widget
	min, max int
	onChange func()
}

// NewSpin — поле с диапазоном [min, max] и значением min; min должен быть меньше max.
func NewSpin(min, max int) *Spin {
	checkCreate()
	if min >= max {
		panic("nui: NewSpin: min должен быть меньше max")
	}
	s := &Spin{min: min, max: max}
	s.id = core.Register((*spinEv)(s))
	s.h = drv.SpinNew(s.id, min, max)
	return s
}

// SetValue задаёт значение, ограничивая его диапазоном; OnChange не вызывается (ADR-8).
func (s *Spin) SetValue(v int) {
	if s.dead() {
		return
	}
	s.quiet++
	drv.SpinSetValue(s.h, max(s.min, min(v, s.max)))
	s.quiet--
}

// Value — принятое значение: набранное, но ещё не принятое (нет Enter, фокус в поле) в
// него не входит.
func (s *Spin) Value() int {
	if s.dead() {
		return 0
	}
	return drv.SpinValue(s.h)
}

// OnChange: пользователь изменил значение — стрелками или набрал и подтвердил.
func (s *Spin) OnChange(fn func()) { s.onChange = fn }

type spinEv Spin

func (s *spinEv) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvChange:
		if s.quiet == 0 && s.onChange != nil {
			s.onChange()
		}
	case core.EvDestroy:
		s.destroyed()
	}
	return 0
}
