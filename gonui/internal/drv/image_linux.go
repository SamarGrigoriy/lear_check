package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "unsafe"

// ImageNew — GtkImage; картинку декодирует GdkPixbuf (nui_image_* в wrap_linux.c).
func ImageNew(id uint32) H { return adopt(C.nui_image_new(), id, true) }

// ImageLoad декодирует картинку и запоминает её байты (из них она перерисовывается при
// смене масштаба экрана). Размер — после поворота по EXIF. data пустой — убрать картинку.
// false — не картинка, прежняя остаётся.
func ImageLoad(h H, data []byte) (w, ht int, ok bool) {
	var wh [2]C.int
	var p *C.guchar
	if len(data) > 0 {
		p = (*C.guchar)(unsafe.Pointer(&data[0]))
	}
	if C.nui_image_load(widget(h), p, C.gsize(len(data)), &wh[0]) == 0 {
		return 0, 0, false
	}
	return int(wh[0]), int(wh[1]), true
}

// ImageShow показывает загруженную картинку размером w×h логических пикселей.
func ImageShow(h H, w, ht int) { C.nui_image_show(widget(h), C.int(w), C.int(ht)) }

// ReadFile читает файл целиком (g_file_get_contents): у библиотеки нет пакета os (ADR-7).
func ReadFile(path string) ([]byte, bool) {
	var p *C.gchar
	var n C.gsize
	if C.g_file_get_contents(cpath(path), &p, &n, nil) == 0 {
		return nil, false
	}
	b := make([]byte, int(n))
	copy(b, unsafe.Slice((*byte)(unsafe.Pointer(p)), int(n)))
	C.g_free(C.gpointer(p))
	return b, true
}
