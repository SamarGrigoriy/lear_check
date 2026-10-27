//go:build unix

package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// IOCond — состояние файлового дескриптора.
type IOCond uint8

const (
	IORead  IOCond = 1 << iota // есть что читать
	IOWrite                    // можно писать
	IOHup                      // другая сторона закрыта / устройство пропало
	IOErr                      // ошибка
)

// Watcher — наблюдатель за файловым дескриптором.
type Watcher struct {
	id, src uint32
	fn      func(got IOCond)
}

// Watch вызывает fn в UI-потоке, когда fd готов; got — что именно произошло.
// IOHup и IOErr приходят всегда, даже если их нет в cond (как в poll).
// Watch не закрывает fd и не читает из него. Перед close(fd) обязательно вызовите Stop:
// после IOHup/IOErr наблюдатель продолжает срабатывать. Открывайте fd с O_NONBLOCK,
// чтобы read в обработчике не повесил интерфейс. Только unix.
func Watch(fd int, cond IOCond, fn func(got IOCond)) *Watcher {
	core.CheckThread()
	w := &Watcher{fn: fn}
	w.id = core.Register((*watcherEv)(w))
	w.src = drv.WatchAdd(fd, uint8(cond), core.UD(w.id, core.KWatch))
	return w
}

// Stop прекращает наблюдение. Можно вызывать изнутри fn; повторный Stop ничего не делает.
func (w *Watcher) Stop() {
	core.CheckThread()
	if w.id == 0 {
		return
	}
	drv.SourceRemove(w.src)
	core.Release(w.id)
	w.id = 0
}

type watcherEv Watcher

func (w *watcherEv) Event(_ core.Kind, arg uintptr) uintptr {
	if w.fn != nil {
		w.fn(IOCond(drv.IOCond(arg)))
	}
	if w.id == 0 {
		return 0 // остановлен изнутри fn
	}
	return 1
}
