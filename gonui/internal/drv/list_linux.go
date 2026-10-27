package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// List (driver-gtk3.md, §7): GtkTreeView в прокрутке со своей моделью — строки не
// хранятся, текст ячейки модель спрашивает у Go (nuiListCell) при отрисовке.

// listCells — функции ячеек живых моделей. Ключ — номер модели: каждая SetRows создаёт
// новую, старую GTK отпускает, и nuiListDrop удаляет её функцию.
var (
	listCells   = map[uintptr]func(row, col int) string{}
	listNextKey uintptr
)

// ListNew — список с колонками columns (заголовки; все пустые — без заголовков).
// outer — прокрутка (в раскладку), view — GtkTreeView.
func ListNew(id uint32, columns []string) (outer, view H) {
	var v *C.GtkWidget
	sw := C.nui_list_new(&v, C.uintptr_t(core.UD(id, core.EvEnter)))
	if len(columns) == 0 {
		columns = []string{""}
	}
	for i, t := range columns {
		C.nui_list_column((*C.GtkTreeView)(H(v)), cstr(t), C.int(i))
	}
	outer = adopt(sw, id, true)
	connect(H(C.gtk_tree_view_get_selection((*C.GtkTreeView)(H(v)))), "changed", id, core.EvChange)
	return outer, H(v)
}

// ListSetRows — новые данные: n строк, текст ячейки — cell. Выделение снимается,
// прокрутка — в начало.
func ListSetRows(view H, n int, cell func(row, col int) string) {
	listNextKey++
	listCells[listNextKey] = cell
	C.nui_list_set_rows(treeView(view), C.int(n), C.uintptr_t(listNextKey))
}

// ListUpdate — текст строк поменялся (их число то же): перерисовать.
func ListUpdate(view H) { C.gtk_widget_queue_draw(widget(view)) }

// ListSelected — выбранная строка; -1 — нет.
func ListSelected(view H) int { return int(C.nui_list_selected(treeView(view))) }

// ListSelect выбирает строку и прокручивает к ней; -1 — снять выделение.
func ListSelect(view H, row int) { C.nui_list_select(treeView(view), C.int(row)) }

func treeView(h H) *C.GtkTreeView { return (*C.GtkTreeView)(h) }
