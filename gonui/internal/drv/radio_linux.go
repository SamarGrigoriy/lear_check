package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// Radio (api.md, §9): GtkBox с GtkRadioButton одной группы (nui_radio_add). EvChange —
// только от кнопки, которая включилась.

// Промежутки между вариантами: в столбик — как у VBox, в ряд — шире, чтобы варианты не
// сливались.
const (
	radioGapV = 6
	radioGapH = 12
)

func RadioNew(id uint32) H {
	return adopt(C.gtk_box_new(C.NUI_GTK_ORIENTATION_VERTICAL, radioGapV), id, true)
}

func radioBox(h H) *C.GtkBox { return (*C.GtkBox)(h) }

func RadioAdd(h H, id uint32, text string) {
	C.nui_radio_add(radioBox(h), cstr(text), ud(id, core.EvChange))
}
func RadioSelected(h H) int { return int(C.nui_radio_selected(radioBox(h))) }

func RadioSetSelected(h H, i int) {
	C.gtk_toggle_button_set_active((*C.GtkToggleButton)(C.nui_radio_nth(radioBox(h), C.int(i))), 1)
}

func RadioSetHorizontal(h H, on bool) {
	o, gap := C.GtkOrientation(C.NUI_GTK_ORIENTATION_VERTICAL), radioGapV
	if on {
		o, gap = C.NUI_GTK_ORIENTATION_HORIZONTAL, radioGapH
	}
	C.gtk_orientable_set_orientation((*C.GtkOrientable)(h), o)
	C.gtk_box_set_spacing(radioBox(h), C.gint(gap))
}

// RadioFocus — фокус на выбранный вариант.
func RadioFocus(h H) {
	if b := C.nui_radio_nth(radioBox(h), C.nui_radio_selected(radioBox(h))); b != nil {
		C.gtk_widget_grab_focus(b)
	}
}
