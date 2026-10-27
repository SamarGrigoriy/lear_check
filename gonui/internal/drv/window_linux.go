package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// WindowNew создаёт окно: скрытое (ADR-14), по центру экрана, с отступом 12 (layout.md, §2).
// delete-event → EvClose с ответом, destroy → EvDestroy.
func WindowNew(id uint32, title string) H {
	w := C.gtk_window_new(C.NUI_GTK_WINDOW_TOPLEVEL)
	win := (*C.GtkWindow)(w)
	C.gtk_window_set_title(win, cstr(title))
	C.gtk_window_set_position(win, C.NUI_GTK_WIN_POS_CENTER)
	C.gtk_container_set_border_width((*C.GtkContainer)(w), 12)
	h := adopt(w, id, false)
	connectRet(h, "delete-event", id, core.EvClose)
	connectRet(h, "key-press-event", id, core.EvKey) // OnKey (key_linux.go)
	return h
}

func WindowSetTitle(h H, s string) { C.gtk_window_set_title((*C.GtkWindow)(h), cstr(s)) }
func WindowTitle(h H) string       { return gostr(C.gtk_window_get_title((*C.GtkWindow)(h))) }

// WindowSetChild заменяет содержимое окна. Старый ребёнок удаляется из контейнера,
// ссылок на него не остаётся, и GTK его уничтожает (destroy → EvDestroy).
// child == Nil — только удалить.
// У окна с меню содержимое лежит под строкой меню, в своём GtkBin (nui_window_content).
func WindowSetChild(h, child H) { containerSetChild(H(C.nui_window_content((*C.GtkWindow)(h))), child) }

func containerSetChild(h, child H) {
	if old := C.gtk_bin_get_child((*C.GtkBin)(h)); old != nil {
		C.gtk_container_remove((*C.GtkContainer)(h), old)
	}
	if child != Nil {
		C.gtk_container_add((*C.GtkContainer)(h), widget(child))
	}
}

// WindowSetPadding — отступ окна: border-width окна, а у окна с меню — ряда под меню.
func WindowSetPadding(h H, px int) {
	C.gtk_container_set_border_width(C.nui_window_content((*C.GtkWindow)(h)), C.guint(max(px, 0)))
}

// WindowShow показывает окно, а если оно уже показано — поднимает наверх.
func WindowShow(h H) {
	C.gtk_widget_show(widget(h))
	C.gtk_window_present((*C.GtkWindow)(h))
}

// WindowSetSize (api.md, §5): растягиваемое окно — начальный размер (до показа —
// default size, после — resize), пользователь может его менять; нерастягиваемое —
// фиксированный размер через size request: окно = max(запрос, размер содержимого).
// Последний SetSize хранится в default size самого окна — у nui.Window нет лишних полей.
// WindowSize — размер окна без рамки, заголовка и строки меню, логические пиксели: та же
// область, что клиентская область на Windows.
func WindowSize(h H) (w, ht int) {
	var cw, ch C.gint
	C.gtk_window_get_size((*C.GtkWindow)(h), &cw, &ch)
	return int(cw), int(ch) - menubarHeight(h)
}

// menubarHeight — высота строки меню окна (её запоминает nui_window_menubar_resize); 0 — нет меню.
func menubarHeight(h H) int {
	return int(uintptr(C.g_object_get_data((*C.GObject)(h), cstr("nui-menubar-h"))))
}

func WindowSetSize(h H, w, ht int) {
	win := (*C.GtkWindow)(h)
	if ht > 0 {
		ht += menubarHeight(h) // размер — под строкой меню
	}
	C.gtk_window_set_default_size(win, C.gint(w), C.gint(ht))
	if C.gtk_window_get_resizable(win) == 0 {
		C.gtk_widget_set_size_request(widget(h), C.gint(w), C.gint(ht))
		return
	}
	C.gtk_widget_set_size_request(widget(h), -1, -1)
	if C.gtk_widget_get_visible(widget(h)) != 0 {
		C.gtk_window_resize(win, C.gint(w), C.gint(ht))
	}
}

// WindowSetResizable меняет растягиваемость и заново применяет последний SetSize:
// его смысл от неё зависит (начальный размер или фиксированный).
func WindowSetResizable(h H, on bool) {
	win := (*C.GtkWindow)(h)
	C.gtk_window_set_resizable(win, b2i(on))
	var w, ht C.gint
	C.gtk_window_get_default_size(win, &w, &ht)
	if w > 0 || ht > 0 {
		WindowSetSize(h, int(w), int(ht)-menubarHeight(h))
	}
}

// WindowSetDefault делает кнопку кнопкой по умолчанию: её нажимает Enter в поле ввода
// с activates_default. btn == Nil — убрать.
func WindowSetDefault(h, btn H) {
	if btn != Nil {
		C.gtk_widget_set_can_default(widget(btn), 1)
	}
	C.gtk_window_set_default((*C.GtkWindow)(h), widget(btn))
}

func WindowSetIconName(h H, name string) { C.gtk_window_set_icon_name((*C.GtkWindow)(h), cstr(name)) }

// WindowSetMaximized: просьбу GTK запоминает и до показа — окно откроется развёрнутым.
// Разворачивает окно оконный менеджер (композитор): без него просьба не действует.
func WindowSetMaximized(h H, on bool) {
	if on {
		C.gtk_window_maximize((*C.GtkWindow)(h))
	} else {
		C.gtk_window_unmaximize((*C.GtkWindow)(h))
	}
}

// WindowSetFullscreen — полноэкранный режим: без рамки, поверх панелей; тоже можно до показа.
func WindowSetFullscreen(h H, on bool) {
	if on {
		C.gtk_window_fullscreen((*C.GtkWindow)(h))
	} else {
		C.gtk_window_unfullscreen((*C.GtkWindow)(h))
	}
}

// ContainerSetChild — Frame, Scroll: удалить старого ребёнка (он уничтожится), добавить нового.
// Прокрутка сама оборачивает ребёнка без собственной прокрутки в GtkViewport.
func ContainerSetChild(h, child H) { containerSetChild(h, child) }
