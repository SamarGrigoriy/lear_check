package drv

import "nui/internal/core"

// KeyLabel — подпись сочетания клавиш в меню Windows: «Ctrl+Shift+O», «F5», «Del».
// Названия — английские, как в меню самой Windows на любом языке.
func KeyLabel(key uint32, mods uint8) string {
	s := ""
	if mods&ModCtrl != 0 {
		s += "Ctrl+"
	}
	if mods&ModShift != 0 {
		s += "Shift+"
	}
	if mods&ModAlt != 0 {
		s += "Alt+"
	}
	switch {
	case key >= KeyF1 && key < KeyF1+12:
		return s + "F" + core.Itoa(int(key-KeyF1+1))
	case key >= KeyEnter && key <= KeyDown:
		return s + [...]string{"Enter", "Esc", "Tab", "Backspace", "Del", "Ins", "Home", "End", "PgUp", "PgDn",
			"Left", "Right", "Up", "Down"}[key-KeyEnter]
	case key == ' ':
		return s + "Space"
	}
	return s + string(rune(key))
}
