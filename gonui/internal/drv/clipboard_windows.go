package drv

/*
#include "win32_windows.h"
*/
import "C"

// ClipboardSet — текст в буфер обмена (CF_UNICODETEXT); переводы строк — \r\n, как ждут
// программы Windows.
func ClipboardSet(s string) {
	b := utf16z(crlf(s))
	C.nui_clip_set(wptr(b), C.int(len(b)-1))
}

// ClipboardText — текст из буфера обмена с переводами строк \n; "" — пусто или не текст.
func ClipboardText() string {
	n := int(C.nui_clip_get(nil, 0))
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	n = int(C.nui_clip_get(wptr(buf), C.int(len(buf))))
	return lf(fromUTF16(buf[:min(n, len(buf))]))
}
