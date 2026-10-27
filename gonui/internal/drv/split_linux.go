package drv

/*
#include "gtk3_linux.h"
*/
import "C"

// Split (api.md, §15): GtkPaned. Первая панель — pack1(resize=FALSE, shrink=FALSE): держит
// свой размер, когда окно растягивается; вторая — pack2(resize=TRUE, shrink=FALSE):
// забирает остальное. Сжать панель меньше её минимума нельзя.

func SplitNew(id uint32, vertical bool) H {
	o := C.GtkOrientation(C.NUI_GTK_ORIENTATION_HORIZONTAL)
	if vertical {
		o = C.NUI_GTK_ORIENTATION_VERTICAL
	}
	return adopt(C.gtk_paned_new(o), id, true)
}

// SplitSet кладёт панели.
func SplitSet(h, first, second H) {
	C.gtk_paned_pack1((*C.GtkPaned)(h), widget(first), 0, 0)
	C.gtk_paned_pack2((*C.GtkPaned)(h), widget(second), 1, 0)
}

// SplitSetPosition — размер первой панели, логические пиксели.
func SplitSetPosition(h H, px int) { C.gtk_paned_set_position((*C.GtkPaned)(h), C.gint(px)) }

func SplitPosition(h H) int { return int(C.gtk_paned_get_position((*C.GtkPaned)(h))) }
