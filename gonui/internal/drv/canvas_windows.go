package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// Canvas (driver-win32.md, §5): дочернее окно класса nuiWindow. WM_PAINT — двойная
// буферизация и GDI+ (nui_canvas_*), EvDraw с NuiPaint* в arg; мышь — EvMouse, событие
// лежит в lastMouse (MouseDecode его отдаёт).

func CanvasNew(id uint32) H {
	h, _ := newNode(kCanvas, id)
	control(h, "nuiWindow", "", 0, 0)
	return h
}

func CanvasRedraw(h H) { C.InvalidateRect(nodes[h].hwnd, nil, 0) }

func paint(ctx uintptr) *C.NuiPaint { return *(**C.NuiPaint)(unsafe.Pointer(&ctx)) }

// themeText — цвет текста темы Windows.
func themeText() C.COLORREF { return C.COLORREF(C.GetSysColor(C.NUI_COLOR_BTNTEXT)) }

func PaintBegin(ctx uintptr, h H) {
	C.nui_paint_color(paint(ctx), themeText())
	C.nui_paint_width(paint(ctx), 1)
}
func PaintThemeColor(ctx uintptr, h H)      { C.nui_paint_color(paint(ctx), themeText()) }
func PaintColor(ctx uintptr, rgb uint32)    { C.nui_paint_color(paint(ctx), bgr(rgb)) }
func PaintLineWidth(ctx uintptr, w float64) { C.nui_paint_width(paint(ctx), C.float(w)) }
func PaintLine(ctx uintptr, x1, y1, x2, y2 float64) {
	C.nui_paint_line(paint(ctx), C.float(x1), C.float(y1), C.float(x2), C.float(y2))
}
func PaintRect(ctx uintptr, x, y, w, h float64, fill bool) {
	C.nui_paint_rect(paint(ctx), C.float(x), C.float(y), C.float(w), C.float(h), C.int(b2i(fill)))
}
func PaintEllipse(ctx uintptr, x, y, w, h float64, fill bool) {
	C.nui_paint_ellipse(paint(ctx), C.float(x), C.float(y), C.float(w), C.float(h), C.int(b2i(fill)))
}

func PaintText(ctx uintptr, h H, x, y float64, s string) {
	b := utf16z(crlf(s))
	C.nui_paint_text(paint(ctx), C.float(x), C.float(y), wptr(b), C.int(len(b)-1))
}

func TextSize(ctx uintptr, h H, s string) (w, ht float64) {
	b := utf16z(crlf(s))
	var out [2]C.float
	C.nui_paint_text_size(paint(ctx), wptr(b), C.int(len(b)-1), &out[0])
	return float64(out[0]), float64(out[1])
}

type mouseEvent struct {
	kind           uint8
	x, y           float64
	button         int
	mods           uint8
	wheelX, wheelY float64
}

var lastMouse mouseEvent

func MouseDecode(uintptr) (kind uint8, x, y float64, button int, mods uint8, wheelX, wheelY float64) {
	m := lastMouse
	return m.kind, m.x, m.y, m.button, m.mods, m.wheelX, m.wheelY
}

// canvasProc — сообщения холста: рисование и мышь.
func canvasProc(n *node, m uint32, wp, lp uintptr) (uintptr, bool) {
	switch m {
	case C.NUI_WM_PAINT:
		w := dpiOf(n)
		bg := C.GetSysColorBrush(C.NUI_COLOR_BTNFACE)
		if c := inheritedBg(n); c != noColor {
			bg = brushFor(c)
		}
		p := C.nui_canvas_begin(n.hwnd, bg, fontsFor(n)[fontNormal], C.UINT(w.dpi), themeText())
		core.Dispatch(core.UD(n.id, core.EvDraw), uintptr(unsafe.Pointer(p)))
		C.nui_canvas_end(n.hwnd)
		return 0, true
	}
	ev := mouseEvent{}
	switch m {
	case C.NUI_WM_LBUTTONDOWN, C.NUI_WM_MBUTTONDOWN, C.NUI_WM_RBUTTONDOWN:
		ev.kind = MouseDown
		C.SetCapture(n.hwnd) // движение и отпускание — и за пределами холста, как в GTK
	case C.NUI_WM_LBUTTONUP, C.NUI_WM_MBUTTONUP, C.NUI_WM_RBUTTONUP:
		ev.kind = MouseUp
		if wp&(0x0001|0x0002|0x0010) == 0 { // MK_LBUTTON, MK_RBUTTON, MK_MBUTTON — больше не нажаты
			C.ReleaseCapture()
		}
	case C.NUI_WM_MOUSEMOVE:
		ev.kind = MouseMove
	case C.NUI_WM_MOUSEWHEEL, C.NUI_WM_MOUSEHWHEEL:
		ev.kind = MouseWheel
		steps := -float64(int16(wp>>16)) / 120 // от себя — вверх (−)
		if m == C.NUI_WM_MOUSEHWHEEL {
			ev.wheelX = -steps
		} else {
			ev.wheelY = steps
		}
		pt := C.POINT{C.LONG(int16(lp)), C.LONG(int16(lp >> 16))} // у колеса — экранные координаты
		C.ScreenToClient(n.hwnd, &pt)
		lp = uintptr(uint16(pt.x)) | uintptr(uint16(pt.y))<<16
	default:
		return 0, false
	}
	switch m {
	case C.NUI_WM_LBUTTONDOWN, C.NUI_WM_LBUTTONUP:
		ev.button = 1
	case C.NUI_WM_MBUTTONDOWN, C.NUI_WM_MBUTTONUP:
		ev.button = 2
	case C.NUI_WM_RBUTTONDOWN, C.NUI_WM_RBUTTONUP:
		ev.button = 3
	}
	if wp&C.NUI_MK_SHIFT != 0 {
		ev.mods |= ModShift
	}
	if wp&C.NUI_MK_CONTROL != 0 {
		ev.mods |= ModCtrl
	}
	if C.GetKeyState(C.NUI_VK_MENU) < 0 {
		ev.mods |= ModAlt
	}
	w := dpiOf(n)
	s := float64(w.dpi) / 96
	ev.x, ev.y = float64(int16(lp))/s, float64(int16(lp>>16))/s
	lastMouse = ev
	core.Dispatch(core.UD(n.id, core.EvMouse), 0)
	return 0, true
}
