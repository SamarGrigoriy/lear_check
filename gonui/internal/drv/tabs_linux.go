package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// Tabs (api.md, §15): GtkNotebook. Ярлык — GtkLabel; спрятанная страница прячет и свой
// ярлык. EvChange — только переключение пользователем (nui_tabs_new).

func TabsNew(id uint32) H { return adopt(C.nui_tabs_new(ud(id, core.EvChange)), id, true) }

func notebook(h H) *C.GtkNotebook { return (*C.GtkNotebook)(h) }

func TabsAdd(h, page H, title string) {
	l := C.gtk_label_new(cstr(title))
	C.gtk_widget_show(l)
	C.gtk_notebook_append_page(notebook(h), widget(page), l)
}

func TabsSetCurrent(h H, i int) { C.gtk_notebook_set_current_page(notebook(h), C.gint(i)) }
func TabsCurrent(h H) int       { return int(C.gtk_notebook_get_current_page(notebook(h))) }

func TabsSetTitle(h, page H, title string) {
	C.gtk_notebook_set_tab_label_text(notebook(h), widget(page), cstr(title))
}
