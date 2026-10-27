package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Timer — таймер цикла событий. Точность — как у цикла: не раньше срока,
// но может опоздать, пока занят UI-поток.
type Timer struct {
	id, src uint32
	fn      func()
	once    bool
}

// After вызывает fn один раз примерно через ms миллисекунд; ms ≤ 0 — при первой возможности.
func After(ms int, fn func()) *Timer { return newTimer(ms, fn, true) }

// Every вызывает fn каждые ms миллисекунд, пока не вызван Stop.
func Every(ms int, fn func()) *Timer { return newTimer(ms, fn, false) }

func newTimer(ms int, fn func(), once bool) *Timer {
	core.CheckThread()
	t := &Timer{fn: fn, once: once}
	t.id = core.Register((*timerEv)(t))
	t.src = drv.TimerAdd(ms, core.UD(t.id, core.KTimer))
	return t
}

// Stop останавливает таймер. Можно вызывать изнутри fn; повторный Stop
// и Stop после срабатывания After ничего не делают.
func (t *Timer) Stop() {
	core.CheckThread()
	if t.id == 0 {
		return
	}
	drv.SourceRemove(t.src) // GLib разрешает удалять источник и изнутри его колбэка
	core.Release(t.id)
	t.id = 0
}

type timerEv Timer

func (t *timerEv) Event(core.Kind, uintptr) uintptr {
	if t.fn != nil {
		t.fn()
	}
	if t.id == 0 {
		return 0 // остановлен изнутри fn
	}
	if t.once {
		core.Release(t.id)
		t.id = 0
		return 0 // источник удалит сам GLib
	}
	return 1
}
