// Package drv — драйвер ОС. Один пакет, реализации в файлах с суффиксом ОС (ADR-5);
// набор функций одинаков для всех ОС — это контракт драйвера (docs/driver-gtk3.md, §3).
// drv не знает о типах nui: только хендлы H, числа и строки.
package drv

// Win32 (docs/driver-win32.md). comctl32 не импортируется: её драйвер грузит сам (ADR-16).
// Флаги размера C-части — через CGO_CFLAGS из Makefile, как на Linux.

/*
#cgo LDFLAGS: -luser32 -lgdi32 -lshell32 -lole32 -lcomdlg32
#include "win32_windows.h"
*/
import "C"

import (
	"runtime"
	"unicode/utf8"
	"unsafe"

	"nui/internal/core"
)

// H — хендл ОС: на Windows индекс узла в таблице драйвера, а не HWND (driver-win32.md, §2).
type H = uintptr

// Nil — нулевой хендл.
var Nil H

// utf16z — строка в UTF-16 с нулём на конце, в памяти Go: годна на время вызова C (ADR-13).
// Битый UTF-8 становится U+FFFD (так работает range по строке), нулевой символ обрезает.
func utf16z(s string) []uint16 {
	buf := make([]uint16, 0, len(s)+1)
	for _, r := range s {
		if r == 0 {
			break
		}
		if r >= 0x10000 {
			r -= 0x10000
			buf = append(buf, uint16(0xD800+(r>>10)), uint16(0xDC00+(r&0x3FF)))
			continue
		}
		buf = append(buf, uint16(r))
	}
	return append(buf, 0)
}

func wptr(b []uint16) *C.WCHAR { return (*C.WCHAR)(unsafe.Pointer(&b[0])) }

// wstr — указатель на UTF-16 строки s для вызова C.
func wstr(s string) *C.WCHAR { return wptr(utf16z(s)) }

// fromUTF16 — UTF-16 (без нуля) в строку Go; одиночные суррогаты становятся U+FFFD.
func fromUTF16(w []uint16) string {
	b := make([]byte, 0, len(w))
	for i := 0; i < len(w); i++ {
		r := rune(w[i])
		if r >= 0xD800 && r < 0xDC00 && i+1 < len(w) && w[i+1] >= 0xDC00 && w[i+1] < 0xE000 {
			r = 0x10000 + (r-0xD800)<<10 + rune(w[i+1]-0xDC00)
			i++
		} else if r >= 0xD800 && r < 0xE000 {
			r = utf8.RuneError
		}
		b = utf8.AppendRune(b, r)
	}
	return string(b)
}

// osText — строка такой, какой её увидит и вернёт ОС: до первого нулевого символа,
// битый UTF-8 — U+FFFD (как cstr у GTK). Её хранят узлы, которые отдают свой текст.
func osText(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			s = s[:i]
			break
		}
	}
	return core.FixUTF8(s)
}

// crlf: "\n" → "\r\n" для многострочного EDIT; уже готовые "\r\n" не удваиваются.
func crlf(s string) string {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' && (i == 0 || s[i-1] != '\r') {
			n++
		}
	}
	if n == 0 {
		return s
	}
	b := make([]byte, 0, len(s)+n)
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' && (i == 0 || s[i-1] != '\r') {
			b = append(b, '\r')
		}
		b = append(b, s[i])
	}
	return string(b)
}

// lf: "\r\n" → "\n" — текст из EDIT в виде, как на Linux.
func lf(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
			continue
		}
		b = append(b, s[i])
	}
	return string(b)
}

// amp удваивает &: Win32 считает его префиксом мнемоники, а на GTK текст виден как есть.
func amp(s string) string {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '&' {
			n++
		}
	}
	if n == 0 {
		return s
	}
	b := make([]byte, 0, len(s)+n)
	for i := 0; i < len(s); i++ {
		if s[i] == '&' {
			b = append(b, '&')
		}
		b = append(b, s[i])
	}
	return string(b)
}

// windowText — текст окна или контрола (GetWindowTextW).
func windowText(hwnd C.HWND) string {
	n := int(C.GetWindowTextLengthW(hwnd))
	if n <= 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	n = int(C.GetWindowTextW(hwnd, wptr(buf), C.int(len(buf))))
	return fromUTF16(buf[:n])
}

func b2i(b bool) C.BOOL {
	if b {
		return 1
	}
	return 0
}

func send(hwnd C.HWND, msg C.UINT, wp, lp uintptr) uintptr {
	return uintptr(C.SendMessageW(hwnd, msg, C.WPARAM(wp), C.LPARAM(lp)))
}

// sendStr — сообщение со строкой в lParam. Указатель, превращённый в uintptr, не держит
// буфер живым, поэтому KeepAlive — до конца синхронного SendMessageW.
func sendStr(hwnd C.HWND, msg C.UINT, wp uintptr, s string) uintptr {
	b := utf16z(s)
	r := send(hwnd, msg, wp, uintptr(unsafe.Pointer(&b[0])))
	runtime.KeepAlive(b)
	return r
}

func emit(n *node, k core.Kind) uintptr { return core.Dispatch(core.UD(n.id, k), 0) }

// Целые из wParam/lParam — хендлы и указатели Windows, а не память Go. Они
// переинтерпретируются через адрес переменной, а не преобразуются uintptr → unsafe.Pointer
// (go vet: possible misuse of unsafe.Pointer).
func hwndOf(v uintptr) C.HWND                { return *(*C.HWND)(unsafe.Pointer(&v)) }
func hdcOf(v uintptr) C.HDC                  { return *(*C.HDC)(unsafe.Pointer(&v)) }
func handleOf(v uintptr) C.HANDLE            { return *(*C.HANDLE)(unsafe.Pointer(&v)) }
func minmaxOf(v uintptr) *C.MINMAXINFO       { return *(**C.MINMAXINFO)(unsafe.Pointer(&v)) }
func rectOf(v uintptr) *C.RECT               { return *(**C.RECT)(unsafe.Pointer(&v)) }
func drawItemOf(v uintptr) *C.DRAWITEMSTRUCT { return *(**C.DRAWITEMSTRUCT)(unsafe.Pointer(&v)) }
func nmhdrOf(v uintptr) *C.NMHDR             { return *(**C.NMHDR)(unsafe.Pointer(&v)) }
