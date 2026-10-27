//go:build gui

package drv

// Помощники интеграционных тестов: имитируют действия пользователя — вызывают GTK
// напрямую, мимо счётчика quiet, поэтому обработчики On* срабатывают.
// #cgo CFLAGS действует на все C-файлы пакета: тестовые C-помощники в wrap_linux.c
// попадают только в сборку с тегом gui.

/*
#cgo CFLAGS: -DNUI_GUI_TEST
#include "gtk3_linux.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// SimClick — как нажатие кнопки мышью.
func SimClick(h H) { C.gtk_button_clicked((*C.GtkButton)(h)) }

// SimClose — как крестик: delete-event приходит не сразу, а через idle-источник.
func SimClose(h H) { C.gtk_window_close((*C.GtkWindow)(h)) }

// Allocation — прямоугольник, который GTK выделил виджету при раскладке.
func Allocation(h H) (x, y, w, ht int) {
	var a [4]C.int
	C.nui_allocation(widget(h), &a[0])
	return int(a[0]), int(a[1]), int(a[2]), int(a[3])
}

// Preferred — минимальный и естественный размер виджета, как его видит GTK.
func Preferred(h H) (minW, natW, minH, natH int) {
	var a [4]C.int
	C.nui_preferred(widget(h), &a[0])
	return int(a[0]), int(a[1]), int(a[2]), int(a[3])
}

// WindowPos — положение виджета в окне (сквозь прокрутку) и начало окна на экране.
func WindowPos(h H) (x, y, screenX, screenY int) {
	var a [4]C.int
	C.nui_window_pos(widget(h), &a[0])
	return int(a[0]), int(a[1]), int(a[2]), int(a[3])
}

// SimType — как ввод текста в поле: changed приходит один раз.
func SimType(h H, s string) {
	if isTextView(h) {
		TextSetText(h, s)
		return
	}
	C.gtk_entry_set_text((*C.GtkEntry)(h), cstr(s))
}

func isTextView(h H) bool {
	return C.g_type_check_instance_is_a((*C.GTypeInstance)(h), C.gtk_text_view_get_type()) != 0
}

// SimComboType — ввод в поле ComboEntry.
func SimComboType(h H, s string) { SimType(comboEntry(h), s) }

// SimEnter — как Enter: activate; у TextView — нажатие Enter (key-press-event поля).
func SimEnter(h H) {
	if isTextView(h) {
		C.nui_sim_key(widget(h), 0xff0d, 0xff0d, 0)
		return
	}
	C.gtk_widget_activate(widget(h))
}

// SimToggle — как щелчок по флажку.
func SimToggle(h H) {
	t := (*C.GtkToggleButton)(h)
	C.gtk_toggle_button_set_active(t, b2i(C.gtk_toggle_button_get_active(t) == 0))
}

// SimSelect — как выбор пункта в списке.
func SimSelect(h H, i int) { C.gtk_combo_box_set_active((*C.GtkComboBox)(h), C.gint(i)) }

// SimSpinStep — ↑ (up) или ↓ в числовом поле: шаг, как от стрелок.
func SimSpinStep(h H, up bool) {
	dir := C.GtkSpinType(C.NUI_GTK_SPIN_STEP_BACKWARD)
	if up {
		dir = C.NUI_GTK_SPIN_STEP_FORWARD
	}
	C.gtk_spin_button_spin((*C.GtkSpinButton)(h), dir, 1)
}

// SimRadioSelect — щелчок по варианту i переключателя.
func SimRadioSelect(h H, i int) {
	C.gtk_button_clicked((*C.GtkButton)(C.nui_radio_nth(radioBox(h), C.int(i))))
}

// SimSlide — как перетаскивание ползунка.
func SimSlide(h H, v float64) { C.gtk_range_set_value((*C.GtkRange)(h), C.gdouble(v)) }

// SimEdit — как правка многострочного текста: на непустом буфере changed приходит
// дважды (удаление и вставка) — так ведёт себя GTK.
func SimEdit(view H, s string) { C.gtk_text_buffer_set_text(textBuffer(view), cstr(s), -1) }

// ComboEntryForTest — поле ввода ComboEntry (чтобы имитировать Enter в нём).
func ComboEntryForTest(h H) H { return comboEntry(h) }

// SimAnswer — ответ в открытом сообщении: OK/«Да» (true) или Esc/«Нет» (false). false — диалога нет.
func SimAnswer(yes bool) bool {
	r := C.int(C.NUI_GTK_RESPONSE_NO)
	if yes {
		// у Message/Alert кнопка OK, у Ask — «Да»: на «Да» GTK отвечает YES, на OK — OK.
		r = C.NUI_GTK_RESPONSE_YES
	}
	return C.nui_sim_dialog_response(0, r) != 0
}

// SimOK — нажать OK в открытом сообщении.
func SimOK() bool { return C.nui_sim_dialog_response(0, C.NUI_GTK_RESPONSE_OK) != 0 }

// SimEsc — закрыть открытое сообщение, как Esc.
func SimEsc() bool { return C.nui_sim_dialog_response(0, C.NUI_GTK_RESPONSE_DELETE_EVENT) != 0 }

// SimChooserSet выбирает путь в открытом файловом диалоге; false — диалога нет.
func SimChooserSet(path string) bool { return C.nui_sim_file_chooser_set(cpath(path)) != 0 }

// SimChooserAdd добавляет файл к выбранным (OpenFiles), как Ctrl+щелчок. Файл — из того
// же каталога, что выбран SimChooserSet, и каталог уже загружен (дайте ему время).
func SimChooserAdd(path string) bool { return C.nui_sim_file_chooser_add(cpath(path)) != 0 }

// SimChooserDone — «Открыть/Сохранить» (accept) или «Отмена» в открытом файловом диалоге.
func SimChooserDone(accept bool) bool {
	r := C.int(C.NUI_GTK_RESPONSE_CANCEL)
	if accept {
		r = C.NUI_GTK_RESPONSE_ACCEPT
	}
	return C.nui_sim_dialog_response(1, r) != 0
}

// DialogOpen — открыт ли сейчас диалог: файловый (chooser) или сообщение.
func DialogOpen(chooser bool) bool { return C.nui_dialog_open(C.int(b2i(chooser))) != 0 }

// ListCellText — текст ячейки так, как его получает GtkTreeView (через модель).
func ListCellText(view H, row, col int) string {
	return gostrFree((*C.gchar)(C.nui_list_cell_text((*C.GtkTreeView)(view), C.int(row), C.int(col))))
}

// SimListSelect — как щелчок по строке.
func SimListSelect(view H, row int) { C.nui_list_select((*C.GtkTreeView)(view), C.int(row)) }

// SimListActivate — как двойной щелчок по строке (или Enter на ней).
func SimListActivate(view H, row int) {
	C.nui_list_select((*C.GtkTreeView)(view), C.int(row))
	C.nui_sim_list_activate((*C.GtkTreeView)(view), C.int(row))
}

// SimKey — нажатие клавиши в окне win, как с клавиатуры: сначала OnKey окна, потом виджет
// с фокусом. key — код nui ('A'…'Z', '0'…'9', ' ', Key*), mods — Mod*.
func SimKey(win H, key uint32, mods uint8) {
	kv := keyKeyval(key)
	if key >= 'A' && key <= 'Z' && mods&ModShift != 0 {
		kv = key // с Shift keyval буквы — заглавная
	}
	C.nui_sim_key(widget(win), C.guint(kv), C.guint(kv), C.guint(gdkMods(mods)))
}

// SimKeyval — нажатие с keyval GDK как есть, на клавише, где в раскладке по умолчанию
// стоит codeOf: например, «ы» (0x6d9) на клавише 's' — как в русской раскладке.
func SimKeyval(win H, keyval, codeOf uint32, mods uint8) {
	var st C.guint
	if mods&ModCtrl != 0 {
		st |= C.NUI_GDK_CONTROL_MASK
	}
	C.nui_sim_key(widget(win), C.guint(keyval), C.guint(codeOf), st)
}

// SimKeyPosted — на GTK то же, что SimKey без модификаторов (на Windows — через цикл).
func SimKeyPosted(win H, key uint32) { SimKey(win, key, 0) }

// SimDropFiles — на окно бросили файлы: то же, что делает drag-data-received с URI
// (сам протокол перетаскивания X11/Wayland в тестах не участвует).
func SimDropFiles(win H, id uint32, paths []string) {
	uris := make([]*C.gchar, 0, len(paths)+1)
	for _, p := range paths {
		uris = append(uris, C.g_filename_to_uri(cpath(p), nil, nil))
	}
	uris = append(uris, nil)
	carr := (**C.gchar)(C.g_malloc(C.gsize(len(uris)) * C.gsize(unsafe.Sizeof(uintptr(0)))))
	copy(unsafe.Slice(carr, len(uris)), uris)
	C.nui_drop_uris(widget(win), carr, ud(id, core.EvDrop))
	C.g_strfreev(carr)
}

// DefaultButton — кнопка по умолчанию окна; Nil — нет.
func DefaultButton(win H) H { return H(C.gtk_window_get_default_widget((*C.GtkWindow)(win))) }

// SimMenuActivate — пункт меню выбран мышью (у флажка состояние переключается).
func SimMenuActivate(item H) { C.gtk_menu_item_activate((*C.GtkMenuItem)(item)) }

// MenuItemText — текст пункта; сочетание (если есть) — после табуляции, как в меню Windows.
func MenuItemText(item H) string {
	s := gostr(C.gtk_menu_item_get_label((*C.GtkMenuItem)(item)))
	var key C.guint
	var mods C.GdkModifierType
	label := C.gtk_bin_get_child((*C.GtkBin)(item))
	C.gtk_accel_label_get_accel((*C.GtkAccelLabel)(H(label)), &key, &mods)
	if key != 0 {
		s += "\t" + core.Itoa(int(keyvalKey(uint32(key)))) + "/" + core.Itoa(int(mods))
	}
	return s
}

// SimMouse — событие мыши над виджетом, как от GDK: kind — Mouse*; у MouseWheel button —
// направление (−1 вверх, +1 вниз). x, y — логические пиксели.
func SimMouse(h H, kind uint8, x, y float64, button int, mods uint8) {
	t := map[uint8]C.int{MouseDown: C.NUI_GDK_BUTTON_PRESS, MouseUp: C.NUI_GDK_BUTTON_RELEASE,
		MouseMove: C.NUI_GDK_MOTION_NOTIFY, MouseWheel: C.NUI_GDK_SCROLL}[kind]
	b := C.uint(button)
	if kind == MouseWheel {
		b = C.NUI_GDK_SCROLL_DOWN
		if button < 0 {
			b = C.NUI_GDK_SCROLL_UP
		}
	}
	C.nui_sim_mouse(widget(h), t, C.double(x), C.double(y), b, C.uint(gdkMods(mods)))
}

// PixelAt — цвет точки (x, y) виджета на экране, 0xRRGGBB (логические пиксели = точки
// при масштабе 1).
func PixelAt(h H, x, y int) uint32 { return uint32(C.nui_pixel_at(widget(h), C.int(x), C.int(y))) }

// SimSplitDrag — перетаскивание границы на delta: то, что делает GtkPaned, пока её тянут.
func SimSplitDrag(h H, delta int) { SplitSetPosition(h, SplitPosition(h)+delta) }

// SimTabsSelect — пользователь выбрал ярлык i.
func SimTabsSelect(h H, i int) { TabsSetCurrent(h, i) }

// ContextMenuOwner — чьё контекстное меню откроется правым щелчком в точке (x, y) окна
// win (логические пиксели от левого верхнего угла содержимого окна); Nil — ничьё.
func ContextMenuOwner(win H, x, y int) H {
	return H(C.nui_ctx_owner_at(widget(win), C.int(x), C.int(y)))
}
