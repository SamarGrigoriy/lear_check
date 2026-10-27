package drv

/*
#include "win32_windows.h"
*/
import "C"

import "unsafe"

// OnDropFiles (driver-win32.md, §3): DragAcceptFiles у окна, WM_DROPFILES → EvDrop с HDROP,
// пути — DragQueryFileW. Бросить файлы можно и на контролы: оболочка ищет окно с
// WS_EX_ACCEPTFILES вверх по родителям.

func WindowSetDrop(h H, id uint32, on bool) {
	hwnd := nodes[h].hwnd
	C.DragAcceptFiles(hwnd, b2i(on))
	if on {
		// Программа, запущенная от администратора, иначе не получит файлы из Проводника
		// (UIPI): WM_DROPFILES и служебные сообщения перетаскивания пропускаются.
		for _, m := range []C.UINT{C.NUI_WM_DROPFILES, C.NUI_WM_COPYDATA, 0x0049 /* WM_COPYGLOBALDATA */} {
			C.ChangeWindowMessageFilterEx(hwnd, m, C.NUI_MSGFLT_ALLOW, nil)
		}
	}
}

// DropPaths — пути из HDROP (arg события EvDrop).
func DropPaths(arg uintptr) []string {
	d := hdropOf(arg)
	n := int(C.DragQueryFileW(d, 0xFFFFFFFF, nil, 0))
	paths := make([]string, 0, n)
	for i := 0; i < n; i++ {
		l := int(C.DragQueryFileW(d, C.UINT(i), nil, 0))
		buf := make([]uint16, l+1)
		C.DragQueryFileW(d, C.UINT(i), wptr(buf), C.UINT(len(buf)))
		paths = append(paths, fromUTF16(buf[:l]))
	}
	return paths
}

func hdropOf(v uintptr) C.HDROP { return *(*C.HDROP)(unsafe.Pointer(&v)) }
