package drv

/*
#include "win32_windows.h"
*/
import "C"

import "unsafe"

// Args — аргументы командной строки из GetCommandLineW: в UTF-16, поэтому без потерь
// при любой кодовой странице (TinyGo берёт os.Args из __getmainargs, в кодировке ANSI).
func Args() []string {
	var n C.int
	argv := C.CommandLineToArgvW(C.GetCommandLineW(), &n)
	if argv == nil {
		return nil
	}
	defer C.LocalFree(C.HANDLE(unsafe.Pointer(argv)))
	args := make([]string, 0, int(n))
	for _, p := range unsafe.Slice(argv, int(n)) {
		l := 0
		for *(*uint16)(unsafe.Add(unsafe.Pointer(p), 2*l)) != 0 {
			l++
		}
		args = append(args, fromUTF16(unsafe.Slice((*uint16)(unsafe.Pointer(p)), l)))
	}
	return args
}
