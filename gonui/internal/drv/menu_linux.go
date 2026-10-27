package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// Меню окна (api.md, §5): GtkMenuBar над содержимым (nui_window_menubar), меню —
// GtkMenu под пунктом строки меню, пункты — GtkMenuItem, GtkCheckMenuItem и
// разделители. Сочетания клавиш только показываются (gtk_accel_label_set_accel):
// срабатывают они в nui — тем же разбором клавиш, что OnKey.

// MenuBar — новая пустая строка меню окна; прежняя уничтожается вместе с пунктами.
func MenuBar(win H) H { return H(C.nui_window_menubar((*C.GtkWindow)(win))) }

// MenuBarRemove убирает строку меню (пункты уничтожаются).
func MenuBarRemove(win H) {
	if bar := C.g_object_get_data((*C.GObject)(win), cstr("nui-menubar")); bar != nil {
		C.gtk_widget_destroy((*C.GtkWidget)(bar))
		C.g_object_set_data_full((*C.GObject)(win), cstr("nui-menubar"), nil, nil)
		C.nui_window_menubar_resize((*C.GtkWindow)(win))
	}
}

// MenuSub — пункт с подменю в parent (строке меню или меню): item — сам пункт (id 0 —
// без событий), menu — подменю, куда добавляются пункты.
func MenuSub(parent H, id uint32, title string) (item, menu H) {
	it := C.gtk_menu_item_new_with_label(cstr(title))
	m := C.gtk_menu_new()
	C.gtk_menu_item_set_submenu((*C.GtkMenuItem)(H(it)), m)
	C.gtk_menu_shell_append((*C.GtkMenuShell)(parent), it)
	if id != 0 {
		connect(H(it), "destroy", id, core.EvDestroy)
	}
	return H(it), H(m)
}

// MenuItemNew — пункт меню (check — с флажком): activate → EvClick (у флажка состояние
// к этому моменту уже переключено), destroy → EvDestroy. key != 0 — сочетание в подписи.
func MenuItemNew(menu H, id uint32, check bool, text string, key uint32, mods uint8) H {
	var it *C.GtkWidget
	if check {
		it = C.gtk_check_menu_item_new_with_label(cstr(text))
	} else {
		it = C.gtk_menu_item_new_with_label(cstr(text))
	}
	if key != 0 {
		label := C.gtk_bin_get_child((*C.GtkBin)(H(it))) // GtkAccelLabel
		C.gtk_accel_label_set_accel((*C.GtkAccelLabel)(H(label)), C.guint(keyKeyval(key)),
			C.GdkModifierType(gdkMods(mods)))
	}
	C.gtk_menu_shell_append((*C.GtkMenuShell)(menu), it)
	connect(H(it), "activate", id, core.EvClick)
	connect(H(it), "destroy", id, core.EvDestroy)
	return H(it)
}

// MenuSepNew — разделитель.
func MenuSepNew(menu H) {
	C.gtk_menu_shell_append((*C.GtkMenuShell)(menu), C.gtk_separator_menu_item_new())
}

// MenuDone — меню построено: показать; окно — на высоту строки меню больше, чтобы
// область под меню осталась прежней.
func MenuDone(win H) { C.nui_window_menubar_resize((*C.GtkWindow)(win)) }

func MenuItemSetEnabled(h H, on bool) { C.gtk_widget_set_sensitive(widget(h), b2i(on)) }

// MenuItemSetChecked — флажок пункта; GTK при этом шлёт activate (EvClick) — nui гасит его quiet.
func MenuItemSetChecked(h H, on bool) {
	C.gtk_check_menu_item_set_active((*C.GtkCheckMenuItem)(h), b2i(on))
}

func MenuItemChecked(h H) bool {
	return C.gtk_check_menu_item_get_active((*C.GtkCheckMenuItem)(h)) != 0
}

func MenuItemSetText(h H, text string, key uint32, mods uint8) {
	C.gtk_menu_item_set_label((*C.GtkMenuItem)(h), cstr(text))
}

// ContextMenu — новое пустое контекстное меню виджета (прежнее уничтожается); пункты —
// MenuItemNew и др., затем ContextMenuDone.
func ContextMenu(owner H) H {
	m := C.gtk_menu_new()
	C.nui_ctx_set(widget(owner), m)
	return H(m)
}

func ContextMenuDone(owner H, menu H) { C.gtk_widget_show_all(widget(menu)) }

func ContextMenuRemove(owner H) { C.nui_ctx_set(widget(owner), nil) }
