package drv

/*
#include "gtk3_linux.h"
*/
import "C"

// BoxNew — GtkBox со стандартным промежутком 6 (layout.md, §2).
func BoxNew(id uint32, vertical bool) H {
	o := C.GtkOrientation(C.NUI_GTK_ORIENTATION_HORIZONTAL)
	if vertical {
		o = C.NUI_GTK_ORIENTATION_VERTICAL
	}
	return adopt(C.gtk_box_new(o, 6), id, true)
}

func BoxSetSpacing(h H, px int) {
	if px < 0 {
		px = 0
	}
	C.gtk_box_set_spacing((*C.GtkBox)(h), C.gint(px))
}

// BoxPack: ребёнок занимает свою ячейку целиком (fill = TRUE) и забирает свободное
// место, только если помечен Expand (layout.md, §6).
// Parent — контейнер, в который вставлен виджет; Nil — никуда.
func Parent(h H) H { return H(C.gtk_widget_get_parent(widget(h))) }

// ContainerClear уничтожает всех детей Box или Grid.
func ContainerClear(h H) { C.nui_container_clear((*C.GtkContainer)(h)) }

func BoxPack(h, child H, expand bool) {
	C.gtk_box_pack_start((*C.GtkBox)(h), widget(child), b2i(expand), 1, 0)
}

// SpacerNew — пустая вставка для Stretch: упаковывается с expand = TRUE.
func SpacerNew(id uint32) H {
	return adopt(C.gtk_box_new(C.NUI_GTK_ORIENTATION_HORIZONTAL, 0), id, true)
}

// GridNew — GtkGrid с промежутками 6 по колонкам и строкам (layout.md, §2).
func GridNew(id uint32) H {
	w := C.gtk_grid_new()
	C.gtk_grid_set_column_spacing((*C.GtkGrid)(w), 6)
	C.gtk_grid_set_row_spacing((*C.GtkGrid)(w), 6)
	return adopt(w, id, true)
}

func GridAttach(h, child H, col, row, cols, rows int) {
	C.gtk_grid_attach((*C.GtkGrid)(h), widget(child), C.gint(col), C.gint(row), C.gint(cols), C.gint(rows))
}

func GridSetSpacing(h H, colGap, rowGap int) {
	if colGap < 0 {
		colGap = 0
	}
	if rowGap < 0 {
		rowGap = 0
	}
	C.gtk_grid_set_column_spacing((*C.GtkGrid)(h), C.guint(colGap))
	C.gtk_grid_set_row_spacing((*C.GtkGrid)(h), C.guint(rowGap))
}

// SeparatorNew: горизонтальная линия для VBox, вертикальная — для HBox.
func SeparatorNew(id uint32, vertical bool) H {
	o := C.GtkOrientation(C.NUI_GTK_ORIENTATION_HORIZONTAL)
	if vertical {
		o = C.NUI_GTK_ORIENTATION_VERTICAL
	}
	return adopt(C.gtk_separator_new(o), id, true)
}

// FrameNew — рамка с заголовком; "" — без заголовка.
func FrameNew(id uint32, title string) H {
	h := adopt(C.gtk_frame_new(nil), id, true)
	FrameSetTitle(h, title)
	return h
}

func FrameSetTitle(h H, s string) {
	if s == "" {
		C.gtk_frame_set_label((*C.GtkFrame)(h), nil)
		return
	}
	C.gtk_frame_set_label((*C.GtkFrame)(h), cstr(s))
}

// ScrollNew — полосы прокрутки появляются, только если ребёнок не помещается.
// Естественный размер — как у ребёнка (layout.md, §5): без SetSize окно покажет его
// целиком, если хватит экрана; минимальный — маленький, по полосам прокрутки.
func ScrollNew(id uint32) H {
	w := C.gtk_scrolled_window_new(nil, nil)
	sw := (*C.GtkScrolledWindow)(w)
	C.gtk_scrolled_window_set_policy(sw, C.NUI_GTK_POLICY_AUTOMATIC, C.NUI_GTK_POLICY_AUTOMATIC)
	C.gtk_scrolled_window_set_propagate_natural_width(sw, 1)
	C.gtk_scrolled_window_set_propagate_natural_height(sw, 1)
	return adopt(w, id, true)
}
