package drv

// Трамплин окон (driver-win32.md, §3): указатель на nuiWndProc ставит nui_init в класс
// nuiWindow; уведомления стандартных контролов приходят родителю — тоже сюда.

/*
#include "win32_windows.h"
*/
import "C"

//export nuiWndProc
func nuiWndProc(hwnd C.HWND, msg C.UINT, wp C.WPARAM, lp C.LPARAM) C.LRESULT {
	if r, ok := wndProc(hwnd, uint32(msg), uintptr(wp), uintptr(lp)); ok {
		return C.LRESULT(r)
	}
	return C.DefWindowProcW(hwnd, msg, wp, lp)
}
