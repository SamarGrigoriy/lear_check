// Спайк S1: gc + cgo без системных заголовков + GTK3 (docs/roadmap.md).
// Одноразовый код: в библиотеку не переносится, переносятся только выводы (RESULT.md).
package main

/*
#cgo LDFLAGS: -l:libgtk-3.so.0 -l:libgobject-2.0.so.0 -l:libglib-2.0.so.0
#include "s1.h"
*/
import "C"

import "unsafe"

type H = unsafe.Pointer

// cstr копирует строку с завершающим нулём в память Go. Указатель годен
// только на время вызова C: GTK копирует всё, что сохраняет (ADR-13).
func cstr(s string) *C.gchar {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return (*C.gchar)(unsafe.Pointer(&b[0]))
}

// gostr копирует строку C в Go. Без C.GoString — так же будет и в TinyGo.
func gostr(p *C.gchar) string {
	if p == nil {
		return ""
	}
	n := int(C.strlen((*C.char)(unsafe.Pointer(p))))
	return string(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
}

func isMainThread() bool { return C.gettid() == C.getpid() }
