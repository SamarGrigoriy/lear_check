// Спайк S4 (driver-win32.md, §11): hello на Win32 под gc и TinyGo. Проверяем cgo без
// заголовков, WndProc-трамплин через //export, манифест из .syso, подсистему GUI, размер.
package main

/*
#cgo LDFLAGS: -L/usr/x86_64-w64-mingw32/lib -luser32 -lgdi32
#include "win32.h"
*/
import "C"

import "unsafe"

var (
	win, label, button, edit C.HWND
	clicks                   int
	selftest                 bool
)

func main() {
	selftest = C.s4_selftest() != 0
	actctx := "не нужен"
	if C.s4_has_manifest() == 0 {
		actctx = "ошибка " + itoa(int(C.s4_actctx()))
		if actctx == "ошибка 0" {
			actctx = "включён"
		}
	}
	win = C.s4_window(wstr("nui S4 · Win32"), 420, 220)
	if win == nil {
		panic("s4: окно не создано")
	}
	manifest := "нет"
	if C.s4_has_manifest() != 0 {
		manifest = "есть"
	}
	label = child("STATIC", "Привет, Windows! Манифест в exe: "+manifest, C.NUI_SS_NOPREFIX, 0, 16, 16, 370, 20)
	edit = child("EDIT", "", C.NUI_WS_TABSTOP|C.NUI_ES_AUTOHSCROLL, C.NUI_WS_EX_CLIENTEDGE, 16, 48, 370, 24)
	// Подсказка в пустом поле — только у comctl32 v6, то есть если манифест сработал.
	cue := C.SendMessageW(edit, C.NUI_EM_SETCUEBANNER, 1, C.LPARAM(uintptr(unsafe.Pointer(wstr("Подсказка: comctl32 v6 включён")))))
	if cue == 0 {
		C.SetWindowTextW(edit, wstr("comctl32 v6 не включён (нет подсказки)"))
	}
	button = child("BUTTON", "Нажми меня", C.NUI_WS_TABSTOP|C.NUI_BS_DEFPUSHBUTTON, 0, 16, 88, 140, 30)
	child("STATIC", "Манифест в exe: "+manifest+"; контекст shell32: "+actctx, C.NUI_SS_NOPREFIX, 0, 16, 132, 370, 20)
	C.ShowWindow(win, C.NUI_SW_SHOWNORMAL)
	if selftest {
		C.SetTimer(win, 1, 500, nil)  // нажать кнопку
		C.SetTimer(win, 2, 4000, nil) // закрыть окно
	}
	if code := C.s4_loop(win); code != 0 {
		panic("s4: цикл завершился с ошибкой")
	}
}

func child(class, text string, style, exstyle C.DWORD, x, y, w, h C.int) C.HWND {
	return C.s4_child(win, wstr(class), wstr(text), style, exstyle, x, y, w, h)
}

//export nuiWndProc
func nuiWndProc(hwnd C.HWND, msg C.UINT, wp C.WPARAM, lp C.LPARAM) C.LRESULT {
	switch msg {
	case C.NUI_WM_COMMAND:
		if C.HWND(unsafe.Pointer(uintptr(lp))) == button && wp>>16 == C.NUI_BN_CLICKED {
			clicks++
			C.SetWindowTextW(label, wstr("Нажато раз: "+itoa(clicks)+" — WndProc на Go работает 🙂"))
		}
		return 0
	case C.NUI_WM_TIMER:
		C.KillTimer(hwnd, C.UINT_PTR(wp))
		if wp == 1 {
			C.SendMessageW(button, C.NUI_BM_CLICK, 0, 0)
		} else {
			C.DestroyWindow(hwnd)
		}
		return 0
	case C.NUI_WM_DESTROY:
		if hwnd == win {
			C.PostQuitMessage(0)
		}
		return 0
	}
	return C.DefWindowProcW(hwnd, msg, wp, lp)
}

// wstr — строка в UTF-16 с нулём на конце (без unicode/utf16: правила кода nui).
func wstr(s string) *C.WCHAR {
	buf := make([]uint16, 0, len(s)+1)
	for _, r := range s {
		if r >= 0x10000 {
			r -= 0x10000
			buf = append(buf, uint16(0xD800+(r>>10)), uint16(0xDC00+(r&0x3FF)))
			continue
		}
		buf = append(buf, uint16(r))
	}
	buf = append(buf, 0)
	return (*C.WCHAR)(unsafe.Pointer(&buf[0]))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for ; n > 0; n /= 10 {
		i--
		b[i] = byte('0' + n%10)
	}
	return string(b[i:])
}
