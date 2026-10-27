package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Combo — выпадающий список: только выбор (NewCombo) или выбор плюс свободный ввод
// (NewComboEntry). Изначально ничего не выбрано: Selected() == -1.
type Combo struct {
	widget
	editable bool
	items    []string // копия на стороне Go: Len и Item не вызывают C
	onChange func()
}

// NewCombo — только выбор из списка.
func NewCombo(items ...string) *Combo { return newCombo(false, items) }

// NewComboEntry — список и свободный ввод. Enter в поле нажимает кнопку окна по умолчанию.
func NewComboEntry(items ...string) *Combo { return newCombo(true, items) }

func newCombo(editable bool, items []string) *Combo {
	checkCreate()
	c := &Combo{editable: editable}
	c.id = core.Register((*comboEv)(c))
	c.h = drv.ComboNew(c.id, editable)
	for _, s := range items {
		c.Add(s)
	}
	return c
}

// Add добавляет элемент в конец.
func (c *Combo) Add(text string) {
	if !c.dead() {
		c.items = append(c.items, text)
		c.quiet++
		drv.ComboAdd(c.h, text)
		c.quiet--
	}
}

// Clear удаляет все элементы; выбор сбрасывается в -1.
func (c *Combo) Clear() {
	if !c.dead() {
		c.items = nil
		c.quiet++
		drv.ComboClear(c.h)
		c.quiet--
	}
}

// Len — число элементов.
func (c *Combo) Len() int { return len(c.items) }

// Item — текст элемента i; вне диапазона — "".
func (c *Combo) Item(i int) string {
	if i < 0 || i >= len(c.items) {
		return ""
	}
	return c.items[i]
}

// Selected — индекс выбранного элемента; -1 — ничего.
func (c *Combo) Selected() int {
	if c.dead() {
		return -1
	}
	return drv.ComboSelected(c.h)
}

// SetSelected выбирает элемент; -1 или индекс вне диапазона — ничего не выбрано.
func (c *Combo) SetSelected(i int) {
	if c.dead() {
		return
	}
	if i < 0 || i >= len(c.items) {
		i = -1
	}
	c.quiet++
	drv.ComboSetSelected(c.h, i)
	c.quiet--
}

// Text: для NewCombo — текст выбранного элемента или ""; для NewComboEntry — текст поля.
func (c *Combo) Text() string {
	if c.dead() {
		return ""
	}
	if c.editable {
		return drv.ComboText(c.h)
	}
	return c.Item(drv.ComboSelected(c.h))
}

// SetText: для NewComboEntry — текст поля; для NewCombo — выбрать элемент, равный s,
// или -1, если такого нет.
func (c *Combo) SetText(s string) {
	if c.dead() {
		return
	}
	if c.editable {
		c.quiet++
		drv.ComboSetText(c.h, s)
		c.quiet--
		return
	}
	idx := -1
	for i, it := range c.items {
		if it == s {
			idx = i
			break
		}
	}
	c.SetSelected(idx)
}

// OnChange: пользователь выбрал элемент или (у ComboEntry) изменил текст.
func (c *Combo) OnChange(fn func()) { c.onChange = fn }

type comboEv Combo

func (c *comboEv) Event(k core.Kind, _ uintptr) uintptr {
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
