package drv

// В файле с //export преамбула может содержать только объявления.

/*
#include "gtk3_linux.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// Четыре трамплина — по одному на сигнатуру колбэка (ADR-6). Все ведут в core.Dispatch.
// ud приходит числом, а не указателем: в Go не появляется «поддельных» указателей.

//export nuiSignal
func nuiSignal(obj unsafe.Pointer, ud C.uintptr_t) { core.Dispatch(uintptr(ud), 0) }

//export nuiSignalRet
func nuiSignalRet(obj, arg unsafe.Pointer, ud C.uintptr_t) C.int {
	return C.int(core.Dispatch(uintptr(ud), uintptr(arg))) // arg — указатель C (GdkEvent* и т. п.)
}

//export nuiSource
func nuiSource(ud C.uintptr_t) C.int { return C.int(core.Dispatch(uintptr(ud), 0)) }

//export nuiWatch
func nuiWatch(fd C.int, cond C.uint, ud C.uintptr_t) C.int {
	return C.int(core.Dispatch(uintptr(ud), uintptr(cond)))
}

// Модель List (list_linux.go): текст ячейки и уничтожение модели. Это не события
// виджета, а вопросы модели к данным — поэтому мимо core.Dispatch.

//export nuiListCell
func nuiListCell(key C.uintptr_t, row, col C.int) *C.gchar {
	s := ""
	if cell := listCells[uintptr(key)]; cell != nil {
		s = cell(int(row), int(col))
	}
	return C.g_strndup(cstr(s), C.gsize(len(s)))
}

//export nuiListDrop
func nuiListDrop(key C.uintptr_t) { delete(listCells, uintptr(key)) }
