//go:build tinygo

package drv

// TinyGo несёт библиотеки импорта только для kernel32, advapi32 и UCRT; user32, gdi32 и
// прочие берутся из mingw-w64 (spike/s4/RESULT.md). gc их находит сам: это путь его gcc.

// #cgo LDFLAGS: -L/usr/x86_64-w64-mingw32/lib
import "C"
