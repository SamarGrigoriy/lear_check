package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Check — флажок с подписью.
type Check struct {
	widget
	onChange func()
}

// NewCheck создаёт снятый флажок.
func NewCheck(text string) *Check {
	checkCreate()
	c := &Check{}
	c.id = core.Register((*checkEv)(c))
	c.h = drv.CheckNew(c.id, text)
	return c
}

// SetChecked задаёт состояние; OnChange при этом не вызывается (ADR-8).
func (c *Check) SetChecked(on bool) {
	if !c.dead() {
		c.quiet++
		drv.CheckSetChecked(c.h, on)
		c.quiet--
	}
}

// Checked — состояние флажка.
func (c *Check) Checked() bool { return !c.dead() && drv.CheckChecked(c.h) }

// SetText задаёт подпись.
func (c *Check) SetText(s string) {
	if !c.dead() {
		drv.ButtonSetText(c.h, s)
	}
}

// Text — подпись.
func (c *Check) Text() string {
	if c.dead() {
		return ""
	}
	return drv.ButtonText(c.h)
}

// OnChange: пользователь переключил флажок.
func (c *Check) OnChange(fn func()) { c.onChange = fn }

type checkEv Check

func (c *checkEv) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvChange:
		if c.quiet == 0 && c.onChange != nil {
			c.onChange()
		}
	case core.EvDestroy:
		c.destroyed()
	}
	return 0
}
