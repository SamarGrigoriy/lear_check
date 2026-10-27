package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Progress — индикатор выполнения.
type Progress struct{ widget }

// NewProgress создаёт пустой индикатор.
func NewProgress() *Progress {
	checkCreate()
	p := &Progress{}
	p.id = core.Register((*progressEv)(p))
	p.h = drv.ProgressNew(p.id)
	return p
}

// SetValue — доля от 0 до 1; значения вне диапазона обрезаются. Выключает режим Pulse.
func (p *Progress) SetValue(f float64) {
	if p.dead() {
		return
	}
	if !(f >= 0) { // и NaN
		f = 0
	}
	if f > 1 {
		f = 1
	}
	drv.ProgressSetValue(p.h, f)
}

// Value — текущая доля.
func (p *Progress) Value() float64 {
	if p.dead() {
		return 0
	}
	return drv.ProgressValue(p.h)
}

// Pulse — режим «неизвестно, сколько осталось»: каждый вызов сдвигает бегунок.
func (p *Progress) Pulse() {
	if !p.dead() {
		drv.ProgressPulse(p.h)
	}
}

// SetText — надпись на индикаторе; "" убирает её. На Windows не отображается.
func (p *Progress) SetText(s string) {
	if !p.dead() {
		drv.ProgressSetText(p.h, s)
	}
}

type progressEv Progress

func (p *progressEv) Event(k core.Kind, _ uintptr) uintptr {
	if k == core.EvDestroy {
		p.destroyed()
	}
	return 0
}
