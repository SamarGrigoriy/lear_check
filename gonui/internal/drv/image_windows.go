package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// Картинка (driver-win32.md, §5): STATIC с SS_OWNERDRAW, рисует WM_DRAWITEM окна.
// Декодирует GDI+ (nui_image_* в wrap_windows.c); узел хранит закодированные байты и
// при смене dpi окна рисует картинку заново — чёткой при любом масштабе.

func ImageNew(id uint32) H {
	h, _ := newNode(kImage, id)
	control(h, "STATIC", "", C.NUI_SS_OWNERDRAW, 0)
	return h
}

// ImageLoad декодирует картинку и запоминает её байты. Размер — после поворота по EXIF.
// data пустой — убрать картинку. false — не картинка (или нет GDI+), прежняя остаётся.
func ImageLoad(h H, data []byte) (w, ht int, ok bool) {
	n := nodes[h]
	if len(data) == 0 {
		freeImage(n)
		n.imgW, n.imgH = 0, 0
		C.InvalidateRect(n.hwnd, nil, 1)
		remeasure(h)
		return 0, 0, true
	}
	var wh [2]C.int
	pic := C.nui_image_decode(bytePtr(data), C.int(len(data)), C.int(core.Orientation(data)), &wh[0])
	if pic == nil {
		return 0, 0, false
	}
	freeImage(n)
	n.pic = pic
	n.img = append([]byte(nil), data...)
	return int(wh[0]), int(wh[1]), true
}

// ImageShow показывает загруженную картинку размером w×h логических пикселей.
func ImageShow(h H, w, ht int) {
	n := nodes[h]
	n.imgW, n.imgH = w, ht
	renderImage(n)
	C.InvalidateRect(n.hwnd, nil, 1)
	remeasure(h)
}

// renderImage — картинка под dpi окна: из только что декодированной или заново из байтов.
func renderImage(n *node) {
	if n.img == nil || n.imgW <= 0 {
		return
	}
	pic := n.pic
	n.pic = nil
	if pic == nil {
		var wh [2]C.int
		if pic = C.nui_image_decode(bytePtr(n.img), C.int(len(n.img)), C.int(core.Orientation(n.img)), &wh[0]); pic == nil {
			return
		}
	}
	d := dpiOf(n)
	w, ht := d.px(n.imgW), d.px(n.imgH)
	bmp := C.nui_image_render(pic, C.int(w), C.int(ht))
	C.nui_image_free(pic)
	if n.bmp != nil {
		C.DeleteObject(C.HGDIOBJ(n.bmp))
	}
	n.bmp, n.bmpW, n.bmpH, n.bmpDpi = bmp, w, ht, d.dpi
	if bmp == nil {
		n.bmpW, n.bmpH = 0, 0
	}
	n.measured = false
}

// freeImage освобождает всё, что держит картинка (и при уничтожении узла).
func freeImage(n *node) {
	if n.pic != nil {
		C.nui_image_free(n.pic)
		n.pic = nil
	}
	if n.bmp != nil {
		C.DeleteObject(C.HGDIOBJ(n.bmp))
		n.bmp = nil
	}
	n.img, n.bmpW, n.bmpH = nil, 0, 0
}

// drawImage — WM_DRAWITEM картинки: фон ближайшей цветной панели или окна, сверху картинка.
func drawImage(n *node, d *C.DRAWITEMSTRUCT) {
	bg := C.GetSysColorBrush(C.NUI_COLOR_BTNFACE)
	if c := inheritedBg(n); c != noColor {
		bg = brushFor(c)
	}
	C.nui_draw_image(d, n.bmp, C.int(n.bmpW), C.int(n.bmpH), bg)
}

// ReadFile читает файл целиком: у библиотеки нет пакета os (ADR-7). Файлы больше 4 ГБ —
// не картинки, они не читаются.
func ReadFile(path string) ([]byte, bool) {
	f := C.CreateFileW(wstr(path), genericRead, C.NUI_FILE_SHARE_READ, nil, C.NUI_OPEN_EXISTING,
		C.NUI_FILE_ATTRIBUTE_NORMAL, nil)
	if uintptr(unsafe.Pointer(f)) == ^uintptr(0) { // INVALID_HANDLE_VALUE
		return nil, false
	}
	defer C.CloseHandle(f)
	var hi C.DWORD
	n := C.GetFileSize(f, &hi)
	if hi != 0 || n == 0xFFFFFFFF {
		return nil, false
	}
	b := make([]byte, int(n))
	if n == 0 {
		return b, true
	}
	var got C.DWORD
	if C.ReadFile(f, C.LPVOID(unsafe.Pointer(&b[0])), n, &got, nil) == 0 || got != n {
		return nil, false
	}
	return b, true
}

// GENERIC_READ (сверяется как NUI_GENERIC_READ): в C это enum, то есть int, и TinyGo
// видит его отрицательным.
const genericRead C.DWORD = 0x80000000

func bytePtr(b []byte) *C.BYTE { return (*C.BYTE)(unsafe.Pointer(&b[0])) }
