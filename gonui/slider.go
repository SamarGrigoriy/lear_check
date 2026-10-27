package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Slider — горизонтальный ползунок с целыми значениями.
type Slider struct {
	widget
	min, max int
	onChange func()
}

// NewSlider создаёт ползунок с диапазоном [min, max]; min должен быть меньше max.
func NewSlider(min, max int) *Slider {
	checkCreate()
	if min >= max {
		panic("nui: NewSlider: min должен быть меньше max")
	}
	s := &Slider{min: min, max: max}
	s.id = core.Register((*sliderEv)(s))
	s.h = drv.SliderNew(s.id, min, max)
	return s
}

// SetValue задаёт значение, ограничивая его диапазоном; OnChange не вызывается (ADR-8).
func (s *Slider) SetValue(v int) {
	if s.dead() {
		return
	}
	if v < s.min {
		v = s.min
	}
	if v > s.max {
		v = s.max
	}
	s.quiet++
	drv.SliderSetValue(s.h, v)
	s.quiet--
}

// Value — текущее значение.
func (s *Slider) Value() int {
	if s.dead() {
		return 0
	}
	return drv.SliderValue(s.h)
}

// OnChange: пользователь сдвинул ползунок.
func (s *Slider) OnChange(fn func()) { s.onChange = fn }

type sliderEv Slider

func (s *sliderEv) Event(k core.Kind, _ uintptr) uintptr {
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
