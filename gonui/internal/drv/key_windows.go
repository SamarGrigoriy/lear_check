package drv

/*
#include "win32_windows.h"
*/
import "C"

// OnKey (driver-win32.md, §3): цикл сообщений до IsDialogMessage отдаёт WM_KEYDOWN и
// WM_SYSKEYDOWN окна с меткой "nuiKeys" служебному окну (NUI_WM_KEY), то — в EvKey.
// arg события: виртуальная клавиша | модификаторы << 16.

// WindowSetKeys включает перехват клавиш окна.
func WindowSetKeys(h H, on bool) {
	hwnd := nodes[h].hwnd
	if on {
		C.SetPropW(hwnd, wstr("nuiKeys"), handleOf(1))
	} else {
		C.RemovePropW(hwnd, wstr("nuiKeys"))
	}
}

// KeyDecode — клавиша nui из виртуальной клавиши Windows. Коды букв VK_A…VK_Z — по
// положению клавиши в раскладке, как ждёт OnKey: в русской раскладке Ctrl+S — тоже 'S'.
func KeyDecode(arg uintptr) (key uint32, mods uint8) {
	vk, mods := uint32(arg&0xFFFF), uint8(arg>>16)
	switch {
	case vk >= 'A' && vk <= 'Z', vk >= '0' && vk <= '9', vk == ' ':
		return vk, mods
	case vk >= 0x60 && vk <= 0x69: // VK_NUMPAD0…9
		return '0' + vk - 0x60, mods
	case vk >= 0x70 && vk <= 0x7B: // VK_F1…F12
		return KeyF1 + vk - 0x70, mods
	}
	switch vk {
	case 0x0D:
		return KeyEnter, mods
	case 0x1B:
		return KeyEscape, mods
	case 0x09:
		return KeyTab, mods
	case 0x08:
		return KeyBackspace, mods
	case 0x2E:
		return KeyDelete, mods
	case 0x2D:
		return KeyInsert, mods
	case 0x24:
		return KeyHome, mods
	case 0x23:
		return KeyEnd, mods
	case 0x21:
		return KeyPageUp, mods
	case 0x22:
		return KeyPageDown, mods
	case 0x25:
		return KeyLeft, mods
	case 0x27:
		return KeyRight, mods
	case 0x26:
		return KeyUp, mods
	case 0x28:
		return KeyDown, mods
	}
	return 0, mods
}
