package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Radio — переключатель (api.md, §9): несколько вариантов, выбран ровно один. Варианты —
// в столбик (или в ряд, SetHorizontal); выбран первый.
//
//	quality := nui.NewRadio("Быстро", "Качественно", "Как в прошлый раз")
//	quality.OnChange(func() { mode = quality.Selected() })
type Radio struct {
	widget
	items    []string // копия на стороне Go: Len и Item не вызывают C
	onChange func()
}

// NewRadio — переключатель с вариантами items; выбран первый.
func NewRadio(items ...string) *Radio {
	checkCreate()
	r := &Radio{}
	r.id = core.Register((*radioEv)(r))
	r.h = drv.RadioNew(r.id)
	for _, s := range items {
		r.Add(s)
	}
	return r
}

// Add добавляет вариант в конец; первый добавленный становится выбранным.
func (r *Radio) Add(text string) {
	if !r.dead() {
		r.items = append(r.items, text)
		r.quiet++
		drv.RadioAdd(r.h, r.id, text)
		r.quiet--
	}
}

// Len — число вариантов.
func (r *Radio) Len() int { return len(r.items) }

// Item — текст варианта i; вне диапазона — "".
func (r *Radio) Item(i int) string {
	if i < 0 || i >= len(r.items) {
		return ""
	}
	return r.items[i]
}

// Selected — номер выбранного варианта; -1 — вариантов нет.
func (r *Radio) Selected() int {
	if r.dead() {
		return -1
	}
	return drv.RadioSelected(r.h)
}

// SetSelected выбирает вариант i; OnChange не вызывается (ADR-8). Вне диапазона — ничего:
// выбран всегда один из вариантов.
func (r *Radio) SetSelected(i int) {
	if r.dead() || i < 0 || i >= len(r.items) {
		return
	}
	r.quiet++
	drv.RadioSetSelected(r.h, i)
	r.quiet--
}

// SetHorizontal — варианты в ряд (true) или в столбик (false, по умолчанию).
func (r *Radio) SetHorizontal(on bool) {
	if !r.dead() {
		drv.RadioSetHorizontal(r.h, on)
	}
}

// Focus передаёт фокус выбранному варианту.
func (r *Radio) Focus() {
	if !r.dead() {
		drv.RadioFocus(r.h)
	}
}

// OnChange: пользователь выбрал другой вариант — щелчком или стрелками (они выбирают
// соседний вариант и у края группы останавливаются).
func (r *Radio) OnChange(fn func()) { r.onChange = fn }

type radioEv Radio

func (r *radioEv) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvChange:
		if r.quiet == 0 && r.onChange != nil {
			r.onChange()
		}
	case core.EvDestroy:
		r.destroyed()
	}
	return 0
}
