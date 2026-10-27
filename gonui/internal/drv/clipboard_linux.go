package drv

/*
#include "gtk3_linux.h"
*/
import "C"

// ClipboardSet кладёт текст в буфер обмена. set_can_store: при выходе из программы GTK
// передаст текст менеджеру буфера обмена, если он есть, — иначе текст пропадёт вместе
// с программой (так устроен буфер обмена X11 и Wayland).
func ClipboardSet(s string) {
	cb := C.nui_clipboard()
	C.gtk_clipboard_set_text(cb, cstr(s), -1)
	C.gtk_clipboard_set_can_store(cb, nil, 0)
}

// ClipboardText — текст из буфера обмена; "" — пусто или там не текст. Ждёт ответа
// владельца буфера во вложенном цикле событий (gtk_clipboard_wait_for_text).
func ClipboardText() string {
	p := C.gtk_clipboard_wait_for_text(C.nui_clipboard())
	if p == nil {
		return ""
	}
	return gostrFree(p)
}
