package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// Init задаёт имя программы (WM_CLASS, app_id на Wayland) и инициализирует GTK.
// false — нет дисплея или GTK не инициализировался.
func Init(appID string) bool {
	if appID != "" {
		C.g_set_prgname(cstr(appID))
	}
	return C.gtk_init_check(nil, nil) != 0
}

func Main() { C.gtk_main() }
func Quit() { C.gtk_main_quit() }

// Wake планирует выполнение очереди Do в UI-потоке. Потокобезопасна.
func Wake() { C.nui_idle(ud(0, core.KDo)) }

// IsMainThread — вызов из главного потока процесса.
func IsMainThread() bool { return C.gettid() == C.getpid() }

// TimerAdd — источник GLib, который вызывает Dispatch(ud) каждые ms миллисекунд.
func TimerAdd(ms int, u uintptr) uint32 {
	if ms < 0 {
		ms = 0
	}
	return uint32(C.nui_timeout(C.guint(ms), C.uintptr_t(u)))
}

// SourceRemove удаляет источник таймера или Watch. GLib разрешает это и изнутри колбэка.
func SourceRemove(src uint32) { C.g_source_remove(C.guint(src)) }

// Биты nui.IOCond: IORead 1, IOWrite 2, IOHup 4, IOErr 8 (драйвер знает значения констант nui).
const (
	ioRead  = 1
	ioWrite = 2
	ioHup   = 4
	ioErr   = 8
)

// WatchAdd — источник GLib на файловый дескриптор; cond — биты nui.IOCond.
func WatchAdd(fd int, cond uint8, u uintptr) uint32 {
	var g C.uint
	if cond&ioRead != 0 {
		g |= C.NUI_G_IO_IN | C.NUI_G_IO_PRI
	}
	if cond&ioWrite != 0 {
		g |= C.NUI_G_IO_OUT
	}
	return uint32(C.nui_watch(C.int(fd), g, C.uintptr_t(u)))
}

// IOCond переводит GIOCondition в биты nui.IOCond.
func IOCond(gcond uintptr) uint8 {
	var c uint8
	if gcond&(C.NUI_G_IO_IN|C.NUI_G_IO_PRI) != 0 {
		c |= ioRead
	}
	if gcond&C.NUI_G_IO_OUT != 0 {
		c |= ioWrite
	}
	if gcond&C.NUI_G_IO_HUP != 0 {
		c |= ioHup
	}
	if gcond&(C.NUI_G_IO_ERR|C.NUI_G_IO_NVAL) != 0 {
		c |= ioErr
	}
	return c
}
