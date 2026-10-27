package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// WindowNew — окно верхнего уровня: скрытое (ADR-14), отступ 12 (layout.md, §2), шрифты
// под его dpi. По центру экрана и по размеру содержимого — при первом показе.
func WindowNew(id uint32, title string) H {
	h, n := newNode(kWindow, id)
	n.padding = 12
	n.hwnd = C.nui_window_new(wstr(title), C.NUI_WS_OVERLAPPEDWINDOW, C.LONG_PTR(h))
	n.dpi = uint32(C.GetDpiForWindow(n.hwnd))
	C.nui_fonts(C.UINT(n.dpi), &n.fonts[0])
	return h
}

func WindowSetTitle(h H, s string) { C.SetWindowTextW(nodes[h].hwnd, wstr(s)) }
func WindowTitle(h H) string       { return windowText(nodes[h].hwnd) }

// WindowSetChild заменяет содержимое; старый ребёнок уничтожается (EvDestroy).
func WindowSetChild(h, child H) { replaceChild(h, child) }

// WindowSetSize (api.md, §5): растягиваемое — начальный размер (после показа — сразу
// новый), нерастягиваемое — фиксированный, но не меньше естественного размера содержимого.
// WindowSize — клиентская область окна, логические пиксели.
func WindowSize(h H) (w, ht int) {
	n := nodes[h]
	var cs [2]C.int
	C.nui_client_size(n.hwnd, &cs[0])
	return n.unpx(int(cs[0])), n.unpx(int(cs[1]))
}

func WindowSetSize(h H, w, ht int) {
	n := nodes[h]
	n.reqW, n.reqH = max(w, 0), max(ht, 0)
	if n.sized {
		resize(h, false)
	}
}

// WindowSetResizable меняет рамку и заново применяет последний SetSize: его смысл от неё зависит.
func WindowSetResizable(h H, on bool) {
	n := nodes[h]
	n.fixed = !on
	style := uintptr(C.GetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE))
	const resizing = C.NUI_WS_THICKFRAME | C.NUI_WS_MAXIMIZEBOX
	if on {
		style |= resizing
	} else {
		style &^= resizing
	}
	C.SetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE, C.LONG_PTR(style))
	C.SetWindowPos(n.hwnd, nil, 0, 0, 0, 0,
		C.NUI_SWP_NOMOVE|C.NUI_SWP_NOSIZE|C.NUI_SWP_NOZORDER|C.NUI_SWP_NOACTIVATE|C.NUI_SWP_FRAMECHANGED)
	if n.sized {
		resize(h, false)
	}
}

// WindowSetDefault — кнопка по умолчанию: BS_DEFPUSHBUTTON и id IDOK, его IsDialogMessage
// узнаёт через DM_GETDEFID и нажимает по Enter (driver-win32.md, §5). btn == Nil — убрать.
func WindowSetDefault(h, btn H) {
	n := nodes[h]
	if old := get(n.defBtn); old != nil && old.hwnd != nil {
		old.isDefault = false
		if !old.ownerDraw { // у цветной кнопки стиль BS_OWNERDRAW: «по умолчанию» рисует драйвер
			send(old.hwnd, C.NUI_BM_SETSTYLE, C.NUI_BS_PUSHBUTTON, 1)
		}
		C.InvalidateRect(old.hwnd, nil, 1)
		C.SetWindowLongPtrW(old.hwnd, C.NUI_GWLP_ID, 0)
	}
	n.defBtn = btn
	if b := get(btn); b != nil && b.hwnd != nil {
		b.isDefault = true
		if !b.ownerDraw {
			send(b.hwnd, C.NUI_BM_SETSTYLE, C.NUI_BS_DEFPUSHBUTTON, 1)
		}
		C.InvalidateRect(b.hwnd, nil, 1)
		C.SetWindowLongPtrW(b.hwnd, C.NUI_GWLP_ID, C.NUI_IDOK)
	}
}

// WindowSetIconName — иконка из ресурсов exe по имени; нет такой — остаётся иконка
// приложения (ресурс 1, driver-win32.md, §10).
func WindowSetIconName(h H, name string) {
	icon := C.LoadIconW(C.GetModuleHandleW(nil), wstr(name))
	if icon == nil {
		return
	}
	n := nodes[h]
	send(n.hwnd, C.NUI_WM_SETICON, 0, uintptr(unsafe.Pointer(icon)))
	send(n.hwnd, C.NUI_WM_SETICON, 1, uintptr(unsafe.Pointer(icon)))
}

// WindowShow показывает окно, а если оно уже показано — поднимает наверх. Первый показ:
// размер по SetSize или по содержимому, по центру экрана.
func WindowShow(h H) {
	n := nodes[h]
	show := C.int(C.NUI_SW_SHOW)
	if !n.sized {
		n.sized = true
		resize(h, true) // размер «восстановленного» окна — и для развёрнутого
		if n.fullscreen {
			enterFullscreen(n)
		} else if n.maximized {
			show = C.NUI_SW_SHOWMAXIMIZED
		}
	}
	C.ShowWindow(n.hwnd, show)
	C.SetForegroundWindow(n.hwnd)
}

// WindowSetMaximized — развернуть (SW_MAXIMIZE) или восстановить; до показа — окно
// откроется развёрнутым.
func WindowSetMaximized(h H, on bool) {
	n := nodes[h]
	n.maximized = on
	if !n.sized || n.fullscreen {
		return
	}
	cmd := C.int(C.NUI_SW_RESTORE)
	if on {
		cmd = C.NUI_SW_MAXIMIZE
	}
	C.ShowWindow(n.hwnd, cmd)
}

// WindowSetFullscreen — без рамки и заголовка, на весь монитор окна; false — прежние
// рамка и место. До показа — окно откроется полноэкранным.
func WindowSetFullscreen(h H, on bool) {
	n := nodes[h]
	if n.fullscreen == on {
		return
	}
	n.fullscreen = on
	if !n.sized {
		return
	}
	if on {
		enterFullscreen(n)
		return
	}
	C.SetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE, C.LONG_PTR(n.savedStyle))
	r := n.savedRect
	C.SetWindowPos(n.hwnd, nil, C.int(r.left), C.int(r.top), C.int(r.right-r.left), C.int(r.bottom-r.top),
		C.NUI_SWP_NOZORDER|C.NUI_SWP_NOACTIVATE|C.NUI_SWP_FRAMECHANGED)
	if n.maximized {
		C.ShowWindow(n.hwnd, C.NUI_SW_MAXIMIZE)
	}
}

// wsPopup — WS_POPUP: в перечислении C 0x80000000 отрицательное, uintptr из него не получить.
const wsPopup uintptr = 0x80000000

func enterFullscreen(n *node) {
	if C.IsWindowVisible(n.hwnd) != 0 {
		C.ShowWindow(n.hwnd, C.NUI_SW_RESTORE) // из развёрнутого — к обычному: запомнить его место
	}
	n.savedStyle = uintptr(C.GetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE))
	C.GetWindowRect(n.hwnd, &n.savedRect)
	style := n.savedStyle&^uintptr(C.NUI_WS_OVERLAPPEDWINDOW) | wsPopup
	C.SetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE, C.LONG_PTR(style))
	var m [4]C.int
	C.nui_monitor_rect(n.hwnd, &m[0])
	C.SetWindowPos(n.hwnd, nil, m[0], m[1], m[2], m[3], C.NUI_SWP_NOACTIVATE|C.NUI_SWP_FRAMECHANGED)
}

// resize ставит окну размер по правилам SetSize; center — по центру рабочей области.
func resize(h H, center bool) {
	n := nodes[h]
	relayout(h) // свежие минимальный и естественный размеры
	natW, natH := n.minClient[0], n.minClient[1]
	if root := get(firstKid(n)); root != nil {
		p := n.px(n.padding)
		natW, natH = root.lay.Nat.W+2*p, root.lay.Nat.H+2*p
	}
	cw, ch := natW, natH
	if n.reqW > 0 {
		cw = n.px(n.reqW)
	}
	if n.reqH > 0 {
		ch = n.px(n.reqH)
	}
	if n.fixed {
		cw, ch = max(cw, natW), max(ch, natH)
	} else {
		cw, ch = max(cw, n.minClient[0]), max(ch, n.minClient[1])
	}
	var r [4]C.int
	C.nui_window_frame(n.hwnd, C.int(cw), C.int(ch), b2i(center), &r[0])
	C.SetWindowPos(n.hwnd, nil, r[0], r[1], r[2], r[3], C.NUI_SWP_NOZORDER|C.NUI_SWP_NOACTIVATE)
}

// windowProc — сообщения окна верхнего уровня.
func windowProc(h H, n *node, m uint32, wp, lp uintptr) (uintptr, bool) {
	switch m {
	case C.NUI_WM_CLOSE:
		if emit(n, core.EvClose) == 0 { // 1 — обработчик OnClose сказал «не закрывать»
			C.DestroyWindow(n.hwnd)
		}
		return 0, true
	case C.NUI_WM_DESTROY:
		destroyTree(h, false) // HWND детей Windows уничтожит сама
		return 0, true
	case C.NUI_WM_SIZE, C.NUI_WM_RELAYOUT:
		relayout(h)
		return 0, true
	case C.NUI_WM_GETMINMAXINFO:
		var r [4]C.int
		C.nui_window_frame(n.hwnd, C.int(n.minClient[0]), C.int(n.minClient[1]), 0, &r[0])
		mmi := minmaxOf(lp)
		mmi.ptMinTrackSize.x, mmi.ptMinTrackSize.y = C.LONG(r[2]), C.LONG(r[3])
		if n.fixed && n.sized && !n.fullscreen && !n.growing { // нерастягиваемое — но не меньше содержимого
			var wr C.RECT
			C.GetWindowRect(n.hwnd, &wr)
			mmi.ptMaxTrackSize.x, mmi.ptMaxTrackSize.y = wr.right-wr.left, wr.bottom-wr.top
		}
		return 0, true
	case C.NUI_WM_DROPFILES: // OnDropFiles: arg события — HDROP, пути читает DropPaths
		core.Dispatch(core.UD(n.id, core.EvDrop), wp)
		C.DragFinish(hdropOf(wp))
		return 0, true
	case C.NUI_WM_DPICHANGED:
		// Окно переехало на монитор с другим масштабом (driver-win32.md, §7): шрифты под
		// новый dpi, всем — WM_SETFONT, листья перемерить, окно — в предложенный прямоугольник.
		n.dpi = uint32(wp & 0xFFFF)
		old := n.fonts
		C.nui_fonts(C.UINT(n.dpi), &n.fonts[0])
		refont(firstKid(n))
		for _, f := range old {
			C.DeleteObject(C.HGDIOBJ(f))
		}
		r := rectOf(lp)
		C.SetWindowPos(n.hwnd, nil, C.int(r.left), C.int(r.top), C.int(r.right-r.left), C.int(r.bottom-r.top),
			C.NUI_SWP_NOZORDER|C.NUI_SWP_NOACTIVATE)
		relayout(h)
		return 0, true
	case C.NUI_DM_GETDEFID:
		if b := get(n.defBtn); b != nil && b.hwnd != nil {
			return C.NUI_IDOK | C.NUI_DC_HASDEFID<<16, true
		}
		return 0, true
	}
	return 0, false
}

func refont(h H) {
	n := get(h)
	if n == nil {
		return
	}
	n.measured = false
	setFont(n)
	for _, k := range n.kids {
		refont(k)
	}
}
