// Package drv — драйвер ОС. Один пакет, реализации в файлах с суффиксом ОС (ADR-5);
// набор функций одинаков для всех ОС — это контракт драйвера (docs/driver-gtk3.md, §3).
// drv не знает о типах nui: только хендлы H, числа и строки.
package drv

// Флаги размера C-части (-ffunction-sections, --gc-sections …) в #cgo писать нельзя:
// gc их не пропускает. Они приходят из Makefile через CGO_CFLAGS/CGO_LDFLAGS
// (docs/driver-gtk3.md, §9). Здесь — только библиотеки, по soname: -dev не нужен.

/*
#cgo LDFLAGS: -l:libgtk-3.so.0 -l:libgobject-2.0.so.0 -l:libglib-2.0.so.0 -l:libpango-1.0.so.0 -l:libgio-2.0.so.0
#cgo LDFLAGS: -l:libgdk-3.so.0 -l:libgdk_pixbuf-2.0.so.0 -l:libcairo.so.2 -l:libpangocairo-1.0.so.0
#include "gtk3_linux.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// H — хендл ОС: на Linux GtkWidget*.
type H = unsafe.Pointer

// Nil — нулевой хендл. nui обнуляет и сравнивает хендлы только через него:
// на Windows H — целое, и nil там не годится.
var Nil H

// cstr копирует строку с завершающим нулём в память Go: указатель годен только
// на время вызова C, GTK копирует всё, что хранит (ADR-13). Нулевой байт внутри
// обрежет строку; битые последовательности UTF-8 заменяются на U+FFFD.
func cstr(s string) *C.gchar { return cpath(core.FixUTF8(s)) }

// cpath — без проверки UTF-8: для путей файловой системы, они не обязаны быть UTF-8.
func cpath(s string) *C.gchar {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return (*C.gchar)(unsafe.Pointer(&b[0]))
}

// gostr копирует строку C в Go. Без C.GoString — одинаково в gc и TinyGo.
func gostr(p *C.gchar) string {
	if p == nil {
		return ""
	}
	n := int(C.strlen((*C.char)(unsafe.Pointer(p))))
	return string(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
}

// gostrFree — для строк, которые принадлежат вызывающему: копирует и освобождает.
func gostrFree(p *C.gchar) string {
	s := gostr(p)
	C.g_free(C.gpointer(p))
	return s
}

func b2i(b bool) C.gboolean {
	if b {
		return 1
	}
	return 0
}

func ud(id uint32, k core.Kind) C.uintptr_t { return C.uintptr_t(core.UD(id, k)) }

// connect подписывает сигнал объекта на трамплин nuiSignal с user_data = UD(id, k).
func connect(h H, sig string, id uint32, k core.Kind) {
	C.nui_connect(C.gpointer(h), cstr(sig), ud(id, k))
}

// connectRet — то же для сигналов с ответом (nuiSignalRet).
func connectRet(h H, sig string, id uint32, k core.Kind) {
	C.nui_connect_ret(C.gpointer(h), cstr(sig), ud(id, k))
}

func widget(h H) *C.GtkWidget { return (*C.GtkWidget)(h) }
