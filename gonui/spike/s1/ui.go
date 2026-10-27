package main

/*
#include "s1.h"
*/
import "C"

// Минимальные Do, таймеры, Watch и виджеты — ровно столько, сколько нужно сценарию S1.

func Do(fn func()) {
	if post(fn) {
		C.nui_idle(C.uintptr_t(UD(0, KDo)))
	}
}

type Timer struct {
	id, src uint32
	fn      func()
	once    bool
}

func After(ms int, fn func()) *Timer { return addTimer(ms, fn, true) }
func Every(ms int, fn func()) *Timer { return addTimer(ms, fn, false) }

func addTimer(ms int, fn func(), once bool) *Timer {
	if ms < 0 {
		ms = 0
	}
	t := &Timer{fn: fn, once: once}
	t.id = register(t)
	t.src = uint32(C.nui_timeout(C.guint(ms), C.uintptr_t(UD(t.id, KTimer))))
	return t
}

func (t *Timer) Event(Kind, uintptr) uintptr {
	t.fn()
	if t.id == 0 {
		return 0 // остановлен изнутри fn
	}
	if t.once {
		release(t.id)
		t.id = 0
		return 0 // источник удалит сам GLib
	}
	return 1
}

func (t *Timer) Stop() {
	if t.id == 0 {
		return
	}
	C.g_source_remove(C.guint(t.src))
	release(t.id)
	t.id = 0
}

const ioRead = 1 // G_IO_IN

type Watcher struct {
	id, src uint32
	fn      func(got uint)
}

func Watch(fd int, cond uint, fn func(got uint)) *Watcher {
	w := &Watcher{fn: fn}
	w.id = register(w)
	w.src = uint32(C.nui_watch(C.int(fd), C.uint(cond), C.uintptr_t(UD(w.id, KWatch))))
	return w
}

func (w *Watcher) Event(_ Kind, arg uintptr) uintptr {
	w.fn(uint(arg))
	if w.id == 0 {
		return 0
	}
	return 1
}

func (w *Watcher) Stop() {
	if w.id == 0 {
		return
	}
	C.g_source_remove(C.guint(w.src))
	release(w.id)
	w.id = 0
}

// Виджеты.

type Widget interface{ handle() H }

type widget struct {
	id uint32 // 0 — уничтожен
	h  H
}

func (w *widget) handle() H { return w.h }

// adopt регистрирует виджет и подписывает destroy. Окна не показываются при создании (ADR-14).
func (w *widget) adopt(r Receiver, h *C.GtkWidget, show bool) {
	w.h = H(h)
	w.id = register(r)
	C.nui_connect(C.gpointer(w.h), cstr("destroy"), C.uintptr_t(UD(w.id, EvDestroy)))
	if show {
		C.gtk_widget_show(h)
	}
}

func (w *widget) destroyed() {
	release(w.id)
	w.id = 0
	w.h = nil
}

type Label struct{ widget }

func NewLabel(text string) *Label {
	l := &Label{}
	l.adopt(l, C.gtk_label_new(cstr(text)), true)
	return l
}

func (l *Label) SetText(s string) {
	if l.id != 0 {
		C.gtk_label_set_text((*C.GtkLabel)(l.h), cstr(s))
	}
}

func (l *Label) Text() string {
	if l.id == 0 {
		return ""
	}
	return gostr(C.gtk_label_get_text((*C.GtkLabel)(l.h)))
}

func (l *Label) Event(k Kind, _ uintptr) uintptr {
	if k == EvDestroy {
		l.destroyed()
	}
	return 0
}

type Button struct {
	widget
	onClick func()
}

func NewButton(text string, onClick func()) *Button {
	b := &Button{onClick: onClick}
	b.adopt(b, C.gtk_button_new_with_label(cstr(text)), true)
	C.nui_connect(C.gpointer(b.h), cstr("clicked"), C.uintptr_t(UD(b.id, EvClick)))
	return b
}

func (b *Button) Event(k Kind, _ uintptr) uintptr {
	switch k {
	case EvClick:
		if b.onClick != nil {
			b.onClick()
		}
	case EvDestroy:
		b.destroyed()
	}
	return 0
}

type Box struct{ widget }

func VBox(items ...Widget) *Box {
	b := &Box{}
	b.adopt(b, C.gtk_box_new(C.NUI_GTK_ORIENTATION_VERTICAL, 6), true)
	for _, it := range items {
		C.gtk_box_pack_start((*C.GtkBox)(b.h), (*C.GtkWidget)(it.handle()), 0, 1, 0)
	}
	return b
}

func (b *Box) Event(k Kind, _ uintptr) uintptr {
	if k == EvDestroy {
		b.destroyed()
	}
	return 0
}

var windows int // живые окна

type Window struct {
	widget
	onClose func() bool
}

func NewWindow(title string, child Widget) *Window {
	w := &Window{}
	w.adopt(w, C.gtk_window_new(C.NUI_GTK_WINDOW_TOPLEVEL), false)
	win := (*C.GtkWindow)(w.h)
	C.gtk_window_set_title(win, cstr(title))
	C.gtk_window_set_position(win, C.NUI_GTK_WIN_POS_CENTER)
	C.gtk_window_set_default_size(win, 420, -1)
	C.gtk_container_set_border_width((*C.GtkContainer)(w.h), 12)
	C.nui_connect_ret(C.gpointer(w.h), cstr("delete-event"), C.uintptr_t(UD(w.id, EvClose)))
	C.gtk_container_add((*C.GtkContainer)(w.h), (*C.GtkWidget)(child.handle()))
	windows++
	return w
}

func (w *Window) Show() {
	if w.id != 0 {
		C.gtk_widget_show((*C.GtkWidget)(w.h))
		C.gtk_window_present((*C.GtkWindow)(w.h))
	}
}

func (w *Window) SetTitle(s string) {
	if w.id != 0 {
		C.gtk_window_set_title((*C.GtkWindow)(w.h), cstr(s))
	}
}

func (w *Window) Title() string {
	if w.id == 0 {
		return ""
	}
	return gostr(C.gtk_window_get_title((*C.GtkWindow)(w.h)))
}

func (w *Window) Event(k Kind, _ uintptr) uintptr {
	switch k {
	case EvClose:
		if w.onClose != nil && !w.onClose() {
			return 1 // окно остаётся
		}
	case EvDestroy:
		w.destroyed()
		if windows--; windows == 0 {
			C.gtk_main_quit()
		}
	}
	return 0
}

func Ask(parent *Window, text string) bool {
	var p *C.GtkWindow
	if parent != nil && parent.id != 0 {
		p = (*C.GtkWindow)(parent.h)
	}
	return C.nui_ask(p, cstr(text)) != 0
}
