package main

// В файле с //export преамбула может содержать только объявления.

/*
#include "s1.h"
*/
import "C"

import "unsafe"

// Четыре трамплина — по одному на сигнатуру колбэка (ADR-6).
// ud приходит числом, а не указателем: в Go не появляется «поддельных» указателей.

//export nuiSignal
func nuiSignal(obj unsafe.Pointer, ud C.uintptr_t) { dispatch(uintptr(ud), 0) }

//export nuiSignalRet
func nuiSignalRet(obj, arg unsafe.Pointer, ud C.uintptr_t) C.int {
	return C.int(dispatch(uintptr(ud), uintptr(arg)))
}

//export nuiSource
func nuiSource(ud C.uintptr_t) C.int { return C.int(dispatch(uintptr(ud), 0)) }

//export nuiWatch
func nuiWatch(fd C.int, cond C.uint, ud C.uintptr_t) C.int {
	return C.int(dispatch(uintptr(ud), uintptr(cond)))
}
