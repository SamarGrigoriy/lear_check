//go:build tinygo && windows

package main

// TinyGo не подхватывает .syso (spike/s4/RESULT.md): тот же объект с манифестом уходит
// линковщику отсюда. rsrc.o делает make win из app.rc.

// #cgo LDFLAGS: -L. -l:rsrc.o
import "C"
