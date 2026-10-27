//go:build tinygo

package main

// TinyGo не подхватывает .syso: объект с манифестом (тот же app.rc, собранный windres
// в rsrc.o) уходит линковщику через #cgo LDFLAGS. ${SRCDIR} TinyGo не раскрывает, но
// относительный -L сам отсчитывает от каталога пакета; -l:файл lld подключает целиком.
// gc этот файл не видит — ему хватает .syso.

// #cgo LDFLAGS: -L. -l:rsrc.o
import "C"
