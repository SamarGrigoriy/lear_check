package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// OnDropFiles (api.md, §5): окно — приёмник text/uri-list (gtk_drag_dest_set),
// drag-data-received переводит URI в пути (nui_drop_uris) и шлёт EvDrop.

// WindowSetDrop делает окно приёмником файлов (один раз; без обработчика nui молчит).
func WindowSetDrop(h H, id uint32, on bool) {
	if on {
		C.nui_drop_enable(widget(h), ud(id, core.EvDrop))
	}
}

// DropPaths — пути из arg события EvDrop: строка «путь\0путь\0\0».
func DropPaths(arg uintptr) []string {
	p := *(**C.char)(unsafe.Pointer(&arg))
	var paths []string
	for *p != 0 {
		s := gostr((*C.gchar)(p))
		paths = append(paths, s)
		p = (*C.char)(unsafe.Add(unsafe.Pointer(p), len(s)+1))
	}
	return paths
}
