package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// Spin (api.md, §11): GtkSpinButton с шагом 1, только цифры и знак (numeric). Значение
// принимается стрелками, кнопками, Enter и уходом фокуса — тогда value-changed → EvChange.
// Enter, как в Entry, нажимает кнопку окна по умолчанию (значение уже принято).

func SpinNew(id uint32, min, max int) H {
	w := C.gtk_spin_button_new_with_range(C.gdouble(min), C.gdouble(max), 1)
	C.gtk_spin_button_set_numeric((*C.GtkSpinButton)(w), 1)
	C.gtk_entry_set_activates_default((*C.GtkEntry)(w), 1)
	h := adopt(w, id, true)
	connect(h, "value-changed", id, core.EvChange)
	return h
}

func SpinSetValue(h H, v int) { C.gtk_spin_button_set_value((*C.GtkSpinButton)(h), C.gdouble(v)) }
func SpinValue(h H) int       { return int(C.gtk_spin_button_get_value_as_int((*C.GtkSpinButton)(h))) }
