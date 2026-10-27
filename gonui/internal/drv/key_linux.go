package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "unsafe"

// OnKey (api.md, §5): key-press-event окна — его обработчики идут раньше, чем окно
// передаст клавишу виджету с фокусом; ответ 1 — клавиша обработана.

// WindowSetKeys: на GTK окно подписано на клавиши с создания (WindowNew), без
// обработчика nui отвечает 0. Функция — для Windows, где перехват клавиш включается.
func WindowSetKeys(h H, on bool) {}

// KeyDecode — клавиша и модификаторы из GdkEventKey* (arg события EvKey). 0 — клавиша,
// о которой nui не сообщает.
func KeyDecode(arg uintptr) (key uint32, mods uint8) {
	e := *(**C.GdkEventKey)(unsafe.Pointer(&arg))
	if key = keyvalKey(uint32(e.keyval)); key == 0 {
		key = keyvalKey(uint32(C.nui_key_by_code(e)))
	}
	st := uint32(e.state)
	if st&C.NUI_GDK_SHIFT_MASK != 0 {
		mods |= ModShift
	}
	if st&C.NUI_GDK_CONTROL_MASK != 0 {
		mods |= ModCtrl
	}
	if st&C.NUI_GDK_MOD1_MASK != 0 {
		mods |= ModAlt
	}
	return key, mods
}

// keyvalKey — keyval GDK (gdkkeysyms.h) в код клавиши nui.
func keyvalKey(kv uint32) uint32 {
	switch {
	case kv >= 'a' && kv <= 'z':
		return kv - 'a' + 'A'
	case kv >= 'A' && kv <= 'Z', kv >= '0' && kv <= '9', kv == ' ':
		return kv
	case kv >= 0xffb0 && kv <= 0xffb9: // KP_0…KP_9
		return '0' + kv - 0xffb0
	case kv >= 0xffbe && kv <= 0xffc9: // F1…F12
		return KeyF1 + kv - 0xffbe
	}
	switch kv {
	case 0xff0d, 0xff8d, 0xfe34: // Return, KP_Enter, ISO_Enter
		return KeyEnter
	case 0xff1b:
		return KeyEscape
	case 0xff09, 0xfe20: // Tab, ISO_Left_Tab (Shift+Tab)
		return KeyTab
	case 0xff08:
		return KeyBackspace
	case 0xffff, 0xff9f:
		return KeyDelete
	case 0xff63, 0xff9e:
		return KeyInsert
	case 0xff50, 0xff95:
		return KeyHome
	case 0xff57, 0xff9c:
		return KeyEnd
	case 0xff55, 0xff9a:
		return KeyPageUp
	case 0xff56, 0xff9b:
		return KeyPageDown
	case 0xff51, 0xff96:
		return KeyLeft
	case 0xff53, 0xff98:
		return KeyRight
	case 0xff52, 0xff97:
		return KeyUp
	case 0xff54, 0xff99:
		return KeyDown
	case 0xff80: // KP_Space
		return ' '
	}
	return 0
}

// keyKeyval — keyval GDK для клавиши nui (обратное keyvalKey): буквы — строчные, как в
// сочетаниях GTK.
func keyKeyval(key uint32) uint32 {
	switch {
	case key >= 'A' && key <= 'Z':
		return key - 'A' + 'a'
	case key >= KeyF1 && key < KeyF1+12:
		return 0xffbe + key - KeyF1
	case key >= KeyEnter && key <= KeyDown:
		return [...]uint32{0xff0d, 0xff1b, 0xff09, 0xff08, 0xffff, 0xff63, 0xff50, 0xff57, 0xff55, 0xff56,
			0xff51, 0xff53, 0xff52, 0xff54}[key-KeyEnter]
	}
	return key // цифры, пробел
}

// gdkMods — модификаторы nui в маску GDK.
func gdkMods(mods uint8) uint32 {
	var st uint32
	if mods&ModShift != 0 {
		st |= C.NUI_GDK_SHIFT_MASK
	}
	if mods&ModCtrl != 0 {
		st |= C.NUI_GDK_CONTROL_MASK
	}
	if mods&ModAlt != 0 {
		st |= C.NUI_GDK_MOD1_MASK
	}
	return st
}
