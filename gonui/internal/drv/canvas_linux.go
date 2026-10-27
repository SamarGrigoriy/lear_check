package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// Canvas (api.md, §21): GtkDrawingArea; draw → EvDraw с cairo_t* в arg, нажатия,
// отпускания, движение и колесо → EvMouse с GdkEvent* (ответ 1 — событие обработано).
// Рисование — nui_paint_* в wrap_linux.c, в логических пикселях.

func CanvasNew(id uint32) H {
	w := C.gtk_drawing_area_new()
	C.gtk_widget_add_events(w, C.NUI_GDK_POINTER_MOTION_MASK|C.NUI_GDK_BUTTON_PRESS_MASK|
		C.NUI_GDK_BUTTON_RELEASE_MASK|C.NUI_GDK_SCROLL_MASK)
	h := adopt(w, id, true)
	connectRet(h, "draw", id, core.EvDraw)
	for _, sig := range []string{"button-press-event", "button-release-event", "motion-notify-event", "scroll-event"} {
		connectRet(h, sig, id, core.EvMouse)
	}
	return h
}

// CanvasRedraw — нарисовать заново (при следующей перерисовке окна).
func CanvasRedraw(h H) { C.gtk_widget_queue_draw(widget(h)) }

func cairo(ctx uintptr) *C.cairo_t { return *(**C.cairo_t)(unsafe.Pointer(&ctx)) }

// PaintBegin: цвет — текста темы, линия — 1.
func PaintBegin(ctx uintptr, h H)           { C.nui_paint_begin(cairo(ctx), widget(h)) }
func PaintThemeColor(ctx uintptr, h H)      { C.nui_paint_theme_color(cairo(ctx), widget(h)) }
func PaintColor(ctx uintptr, rgb uint32)    { C.nui_paint_color(cairo(ctx), C.uint(rgb)) }
func PaintLineWidth(ctx uintptr, w float64) { C.cairo_set_line_width(cairo(ctx), C.double(w)) }
func PaintLine(ctx uintptr, x1, y1, x2, y2 float64) {
	C.nui_paint_line(cairo(ctx), C.double(x1), C.double(y1), C.double(x2), C.double(y2))
}
func PaintRect(ctx uintptr, x, y, w, h float64, fill bool) {
	C.nui_paint_rect(cairo(ctx), C.double(x), C.double(y), C.double(w), C.double(h), C.int(b2i(fill)))
}
func PaintEllipse(ctx uintptr, x, y, w, h float64, fill bool) {
	C.nui_paint_ellipse(cairo(ctx), C.double(x), C.double(y), C.double(w), C.double(h), C.int(b2i(fill)))
}

// PaintText — текст шрифтом интерфейса, (x, y) — левый верхний угол.
func PaintText(ctx uintptr, h H, x, y float64, s string) {
	C.nui_paint_text(cairo(ctx), widget(h), C.double(x), C.double(y), cstr(s))
}

// TextSize — размер текста шрифтом интерфейса, логические пиксели.
func TextSize(ctx uintptr, h H, s string) (w, ht float64) {
	var out [2]C.int
	C.nui_paint_text_size(widget(h), cstr(s), &out[0])
	return float64(out[0]), float64(out[1])
}

// MouseDecode — событие мыши из arg EvMouse: вид (Mouse*; 0 — пропустить), координаты,
// кнопка, модификаторы, колесо (x, y: −1, 0, +1).
func MouseDecode(arg uintptr) (kind uint8, x, y float64, button int, mods uint8, wheelX, wheelY float64) {
	var xy [2]C.double
	var info [3]C.int
	k := C.nui_mouse_decode(*(**C.GdkEvent)(unsafe.Pointer(&arg)), &xy[0], &info[0])
	switch w := int(info[2]); w {
	case -1, 1:
		wheelY = float64(w)
	case -2, 2:
		wheelX = float64(w / 2)
	}
	return uint8(k), float64(xy[0]), float64(xy[1]), int(info[0]), uint8(info[1]), wheelX, wheelY
}
