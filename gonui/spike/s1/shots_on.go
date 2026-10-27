//go:build shots

package main

/*
int system(const char *command);
*/
import "C"

import "unsafe"

var shotN int

// shot снимает весь экран (под Xvfb это только наши окна) в shotDir/NN-name.png.
// Выполняется синхронно в UI-потоке: окно уже нарисовано, X-сервер хранит его содержимое.
// Снимок — через rootshot (tools/rootshot.c), а не xwd -root: xwd теряет окна
// с не-умолчательным визуалом, и модальный диалог выходил чёрным.
// Команда запускается через system() из libc, а не os/exec: так снимки работают и в TinyGo.
// Под тегом shots, чтобы этот код не попадал в бинарник, по которому меряют размер.
func shot(name string) {
	if shotDir == "" {
		return
	}
	shotN++
	n := itoa(shotN)
	if shotN < 10 {
		n = "0" + n
	}
	path := shotDir + "/" + n + "-" + name + ".png"
	if C.system((*C.char)(unsafe.Pointer(cstr("rootshot | pnmtopng > "+path)))) != 0 {
		println("s1: снимок не удался:", path)
	}
}
