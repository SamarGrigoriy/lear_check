package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// LabelNew: у GtkLabel по умолчанию xalign = 0.5; nui выравнивает текст по началу.
func LabelNew(id uint32, text string) H {
	w := C.gtk_label_new(cstr(text))
	C.gtk_label_set_xalign((*C.GtkLabel)(w), 0)
	return adopt(w, id, true)
}

func LabelSetText(h H, s string) { C.gtk_label_set_text((*C.GtkLabel)(h), cstr(s)) }
func LabelText(h H) string       { return gostr(C.gtk_label_get_text((*C.GtkLabel)(h))) }

// ButtonNew: clicked → EvClick. Подчёркивания в тексте не становятся мнемониками.
func ButtonNew(id uint32, text string) H {
	h := adopt(C.gtk_button_new_with_label(cstr(text)), id, true)
	connect(h, "clicked", id, core.EvClick)
	return h
}

func ButtonSetText(h H, s string) { C.gtk_button_set_label((*C.GtkButton)(h), cstr(s)) }
func ButtonText(h H) string       { return gostr(C.gtk_button_get_label((*C.GtkButton)(h))) }

// Значения nui.Ellipsis (None 0, Start 1, Middle 2, End 3) совпадают с PangoEllipsizeMode.
func LabelSetEllipsis(h H, e uint8) {
	C.gtk_label_set_ellipsize((*C.GtkLabel)(h), C.PangoEllipsizeMode(e))
}

func LabelSetWrap(h H, on bool)       { C.gtk_label_set_line_wrap((*C.GtkLabel)(h), b2i(on)) }
func LabelSetSelectable(h H, on bool) { C.gtk_label_set_selectable((*C.GtkLabel)(h), b2i(on)) }

// LabelSetTextAlign: xalign двигает блок текста целиком, justify выравнивает строки
// внутри многострочного текста. a — nui.Align; Fill считается Start.
func LabelSetTextAlign(h H, a uint8) {
	l := (*C.GtkLabel)(h)
	switch a {
	case 2: // Center
		C.gtk_label_set_xalign(l, 0.5)
		C.gtk_label_set_justify(l, C.NUI_GTK_JUSTIFY_CENTER)
	case 3: // End
		C.gtk_label_set_xalign(l, 1)
		C.gtk_label_set_justify(l, C.NUI_GTK_JUSTIFY_RIGHT)
	default:
		C.gtk_label_set_xalign(l, 0)
		C.gtk_label_set_justify(l, C.NUI_GTK_JUSTIFY_LEFT)
	}
}

// LabelSetBold — жирное начертание. Цвет при этом сохраняется (он хранится в самой метке).
func LabelSetBold(h H, on bool) { C.nui_label_style((*C.GtkLabel)(h), 1, C.int(b2i(on)), 0, 0, 0) }

// LabelSetFontScale — размер шрифта относительно обычного, в процентах (Pango scale).
func LabelSetFontScale(h H, pct int) {
	if pct < 1 {
		pct = 100
	}
	C.nui_label_style((*C.GtkLabel)(h), 0, 0, 0, 0, C.int(pct))
}

// LabelSetColor — цвет 0xRRGGBB; 1<<24 (nui.NoColor) — цвет темы. Жирность сохраняется.
func LabelSetColor(h H, color uint32) {
	if color > 1<<24 {
		color = 1 << 24
	}
	C.nui_label_style((*C.GtkLabel)(h), 0, 0, 1, C.uint(color), 0)
}

// CheckNew: toggled → EvChange. Текст — через ButtonSetText/ButtonText.
func CheckNew(id uint32, text string) H {
	h := adopt(C.gtk_check_button_new_with_label(cstr(text)), id, true)
	connect(h, "toggled", id, core.EvChange)
	return h
}

func CheckSetChecked(h H, on bool) { C.gtk_toggle_button_set_active((*C.GtkToggleButton)(h), b2i(on)) }
func CheckChecked(h H) bool        { return C.gtk_toggle_button_get_active((*C.GtkToggleButton)(h)) != 0 }

// EntryNew: changed → EvChange, activate (Enter) → EvEnter. Enter по умолчанию нажимает
// кнопку окна по умолчанию; nui выключает это, когда задан OnEnter.
func EntryNew(id uint32) H {
	w := C.gtk_entry_new()
	C.gtk_entry_set_activates_default((*C.GtkEntry)(w), 1)
	h := adopt(w, id, true)
	connect(h, "changed", id, core.EvChange)
	connect(h, "activate", id, core.EvEnter)
	return h
}

func EntrySetText(h H, s string) { C.gtk_entry_set_text((*C.GtkEntry)(h), cstr(s)) }
func EntryText(h H) string       { return gostr(C.gtk_entry_get_text((*C.GtkEntry)(h))) }

func EntrySetPlaceholder(h H, s string) {
	C.gtk_entry_set_placeholder_text((*C.GtkEntry)(h), cstr(s))
}

func EntrySetPassword(h H, on bool) { C.gtk_entry_set_visibility((*C.GtkEntry)(h), b2i(!on)) }
func EntrySetReadOnly(h H, on bool) { C.gtk_editable_set_editable((*C.GtkEditable)(h), b2i(!on)) }
func EntrySetWidthChars(h H, n int) { C.gtk_entry_set_width_chars((*C.GtkEntry)(h), C.gint(n)) }

func EntrySetActivatesDefault(h H, on bool) {
	C.gtk_entry_set_activates_default((*C.GtkEntry)(h), b2i(on))
}

// ComboNew: changed → EvChange. У ComboEntry текст поля меняется вместе с выбором, и его
// changed уже приходит в changed комбобокса. Enter в поле нажимает кнопку по умолчанию.
func ComboNew(id uint32, editable bool) H {
	var w *C.GtkWidget
	if editable {
		w = C.gtk_combo_box_text_new_with_entry()
		C.gtk_entry_set_activates_default((*C.GtkEntry)(comboEntry(H(w))), 1)
	} else {
		w = C.gtk_combo_box_text_new()
	}
	h := adopt(w, id, true)
	connect(h, "changed", id, core.EvChange)
	return h
}

// comboEntry — поле ввода ComboEntry (ребёнок GtkBin).
func comboEntry(h H) H { return H(C.gtk_bin_get_child((*C.GtkBin)(h))) }

func ComboAdd(h H, s string) { C.gtk_combo_box_text_append_text((*C.GtkComboBoxText)(h), cstr(s)) }
func ComboClear(h H)         { C.gtk_combo_box_text_remove_all((*C.GtkComboBoxText)(h)) }

func ComboSetSelected(h H, i int) { C.gtk_combo_box_set_active((*C.GtkComboBox)(h), C.gint(i)) }
func ComboSelected(h H) int       { return int(C.gtk_combo_box_get_active((*C.GtkComboBox)(h))) }

// ComboSetText и ComboText — только для ComboEntry: текст поля.
func ComboSetText(h H, s string) { EntrySetText(comboEntry(h), s) }
func ComboText(h H) string       { return EntryText(comboEntry(h)) }

// SliderNew: горизонтальная шкала с шагом 1 и без дробной части; value-changed → EvChange.
func SliderNew(id uint32, min, max int) H {
	w := C.gtk_scale_new_with_range(C.NUI_GTK_ORIENTATION_HORIZONTAL, C.gdouble(min), C.gdouble(max), 1)
	C.gtk_scale_set_digits((*C.GtkScale)(w), 0)
	h := adopt(w, id, true)
	connect(h, "value-changed", id, core.EvChange)
	return h
}

func SliderSetValue(h H, v int) { C.gtk_range_set_value((*C.GtkRange)(h), C.gdouble(v)) }

// SliderValue округляет до целого на стороне Go.
func SliderValue(h H) int {
	f := float64(C.gtk_range_get_value((*C.GtkRange)(h)))
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

func ProgressNew(id uint32) H { return adopt(C.gtk_progress_bar_new(), id, true) }

func ProgressSetValue(h H, f float64) {
	C.gtk_progress_bar_set_fraction((*C.GtkProgressBar)(h), C.gdouble(f))
}

func ProgressValue(h H) float64 {
	return float64(C.gtk_progress_bar_get_fraction((*C.GtkProgressBar)(h)))
}

func ProgressPulse(h H) { C.gtk_progress_bar_pulse((*C.GtkProgressBar)(h)) }

// ProgressSetText: "" — убрать надпись.
func ProgressSetText(h H, s string) {
	p := (*C.GtkProgressBar)(h)
	if s == "" {
		C.gtk_progress_bar_set_show_text(p, 0)
		return
	}
	C.gtk_progress_bar_set_text(p, cstr(s))
	C.gtk_progress_bar_set_show_text(p, 1)
}

// TextNew: outer — прокрутка (для раскладки и общих методов), view — сам текст.
// destroy подписан только у outer; changed у буфера → EvChange с тем же id.
func TextNew(id uint32) (outer, view H) {
	sw := C.gtk_scrolled_window_new(nil, nil)
	C.gtk_scrolled_window_set_policy((*C.GtkScrolledWindow)(sw), C.NUI_GTK_POLICY_AUTOMATIC, C.NUI_GTK_POLICY_AUTOMATIC)
	v := C.gtk_text_view_new()
	// Поля внутри текста: без них строки прижаты к рамке (так же на Windows — EM_SETMARGINS).
	tv := (*C.GtkTextView)(v)
	C.gtk_text_view_set_left_margin(tv, 6)
	C.gtk_text_view_set_right_margin(tv, 6)
	C.gtk_text_view_set_top_margin(tv, 4)
	C.gtk_text_view_set_bottom_margin(tv, 4)
	C.gtk_container_add((*C.GtkContainer)(sw), v)
	C.gtk_widget_show(v)
	outer = adopt(sw, id, true)
	connect(H(C.gtk_text_view_get_buffer((*C.GtkTextView)(v))), "changed", id, core.EvChange)
	return outer, H(v)
}

// TextSetEnter: on — Enter без Shift даёт EvEnter вместо перевода строки (TextView.OnEnter).
func TextSetEnter(view H, id uint32, on bool) {
	C.nui_text_enter((*C.GtkTextView)(view), C.int(b2i(on)), ud(id, core.EvEnter))
}

// TextSetPlaceholder — подсказка в пустом поле; "" — без подсказки.
func TextSetPlaceholder(view H, s string) { C.nui_text_placeholder((*C.GtkTextView)(view), cstr(s)) }

func textBuffer(view H) *C.GtkTextBuffer { return C.gtk_text_view_get_buffer((*C.GtkTextView)(view)) }

func TextSetText(view H, s string) { C.gtk_text_buffer_set_text(textBuffer(view), cstr(s), -1) }
func TextText(view H) string       { return gostrFree((*C.gchar)(C.nui_text_get(textBuffer(view)))) }

func TextAppend(view H, s string, autoscroll bool, maxLines int) {
	C.nui_text_append((*C.GtkTextView)(view), cstr(s), -1, C.int(b2i(autoscroll)), C.int(maxLines))
}

func TextSetReadOnly(view H, on bool) {
	v := (*C.GtkTextView)(view)
	C.gtk_text_view_set_editable(v, b2i(!on))
	C.gtk_text_view_set_cursor_visible(v, b2i(!on))
}

func TextSetMonospace(view H, on bool) {
	C.gtk_text_view_set_monospace((*C.GtkTextView)(view), b2i(on))
}

func TextSetWrap(view H, on bool) {
	m := C.GtkWrapMode(C.NUI_GTK_WRAP_NONE)
	if on {
		m = C.NUI_GTK_WRAP_WORD_CHAR
	}
	C.gtk_text_view_set_wrap_mode((*C.GtkTextView)(view), m)
}
