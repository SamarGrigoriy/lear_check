package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
)

// List (driver-win32.md, §5): ListView с LVS_OWNERDATA — строки не хранятся, текст
// ячейки ListView спрашивает уведомлением LVN_GETDISPINFOW, когда строка видна.

func ListNew(id uint32, columns []string) (outer, view H) {
	h, n := newNode(kList, id)
	if len(columns) == 0 {
		columns = []string{""}
	}
	n.columns = columns
	style := uint32(C.NUI_LVS_REPORT | C.NUI_LVS_SINGLESEL | C.NUI_LVS_SHOWSELALWAYS | C.NUI_LVS_OWNERDATA |
		C.NUI_WS_TABSTOP)
	if !listHeaders(n) {
		style |= C.NUI_LVS_NOCOLUMNHEADER
	}
	hwnd := control(h, "SysListView32", "", style, C.NUI_WS_EX_CLIENTEDGE)
	ex := uintptr(C.NUI_LVS_EX_FULLROWSELECT | C.NUI_LVS_EX_LABELTIP | C.NUI_LVS_EX_DOUBLEBUFFER)
	send(hwnd, C.NUI_LVM_SETEXTENDEDLISTVIEWSTYLE, ex, ex)
	for i, t := range columns {
		C.nui_list_column(hwnd, C.int(i), wstr(t))
	}
	listAutosize(n)
	C.SetPropW(hwnd, wstr("nuiEnter"), handleOf(1)) // Enter — строке (OnActivate), а не кнопке по умолчанию
	return h, h
}

// listHeaders — заголовки видны, если хоть один не пуст (как на GTK).
func listHeaders(n *node) bool {
	for _, t := range n.columns {
		if t != "" {
			return true
		}
	}
	return false
}

// ListSetRows — новые данные: выделение снимается, прокрутка — в начало.
func ListSetRows(view H, rows int, cell func(row, col int) string) {
	n := nodes[view]
	n.cell, n.rows = cell, rows
	C.nui_list_select(n.hwnd, -1)
	send(n.hwnd, C.NUI_LVM_SETITEMCOUNT, uintptr(rows), 0)
	if rows > 0 {
		send(n.hwnd, C.NUI_LVM_ENSUREVISIBLE, 0, 0)
	}
	listAutosize(n)
	listFillLast(n)
	C.InvalidateRect(n.hwnd, nil, 1)
}

// ListUpdate — текст строк поменялся: перерисовать, ListView заново спросит видимые.
func ListUpdate(view H) { C.InvalidateRect(nodes[view].hwnd, nil, 1) }

func ListSelected(view H) int {
	return int(int32(send(nodes[view].hwnd, C.NUI_LVM_GETNEXTITEM, ^uintptr(0), C.NUI_LVNI_SELECTED)))
}

func ListSelect(view H, row int) { C.nui_list_select(nodes[view].hwnd, C.int(row)) }

// listAutosize — ширина колонок по заголовку и первым 100 строкам (все строки не
// измерить), не больше 400 логических пикселей. Как на GTK.
func listAutosize(n *node) {
	if n.hwnd == nil {
		return
	}
	d := dpiOf(n)
	for c, title := range n.columns {
		best := d.px(40)
		if title != "" {
			best = max(best, strWidth(n, title)+d.px(24))
		}
		for r := 0; r < n.rows && r < 100 && n.cell != nil; r++ {
			best = max(best, strWidth(n, n.cell(r, c))+d.px(16))
		}
		send(n.hwnd, C.NUI_LVM_SETCOLUMNWIDTH, uintptr(c), uintptr(min(best, d.px(400))))
	}
	n.colWidths = n.colWidths[:0]
	for c := range n.columns {
		n.colWidths = append(n.colWidths, int(send(n.hwnd, C.NUI_LVM_GETCOLUMNWIDTH, uintptr(c), 0)))
	}
}

// listFillLast — последняя колонка забирает свободную ширину (как expand у GTK), но не
// становится уже своей ширины по тексту.
func listFillLast(n *node) {
	last := len(n.colWidths) - 1
	if n.hwnd == nil || last < 0 {
		return
	}
	var cs [2]C.int
	C.nui_client_size(n.hwnd, &cs[0])
	rest := int(cs[0])
	for _, w := range n.colWidths[:last] {
		rest -= w
	}
	send(n.hwnd, C.NUI_LVM_SETCOLUMNWIDTH, uintptr(last), uintptr(max(n.colWidths[last], rest)))
}

func strWidth(n *node, s string) int {
	return int(send(n.hwnd, C.NUI_LVM_GETSTRINGWIDTHW, 0, uintptr(unsafe.Pointer(wstr(s)))))
}

// listNotify — WM_NOTIFY от ListView: текст ячейки, смена выделения, активация строки.
func listNotify(n *node, lp uintptr) uintptr {
	switch int32(nmhdrOf(lp).code) {
	case C.NUI_LVN_GETDISPINFOW:
		it := &(*(**C.NMLVDISPINFOW)(unsafe.Pointer(&lp))).item
		if it.mask&C.NUI_LVIF_TEXT != 0 && it.cchTextMax > 0 && it.pszText != nil {
			s := ""
			if row := int(it.iItem); n.cell != nil && row >= 0 && row < n.rows {
				s = n.cell(row, int(it.iSubItem))
			}
			copyText(unsafe.Slice((*uint16)(unsafe.Pointer(it.pszText)), int(it.cchTextMax)), s)
		}
	case C.NUI_LVN_ITEMCHANGED:
		nm := *(**C.NMLISTVIEW)(unsafe.Pointer(&lp))
		if nm.uChanged&C.NUI_LVIF_STATE != 0 && (nm.uNewState^nm.uOldState)&C.NUI_LVIS_SELECTED != 0 {
			if nm.uNewState&C.NUI_LVIS_SELECTED != 0 {
				emit(n, core.EvChange)
			} else {
				// Снятие приходит раньше нового выбора: сообщить потом, когда станет ясно,
				// выбрано ли что-то (повтор nui отбрасывает сам).
				C.PostMessageW(msg(), C.NUI_WM_LISTSEL, 0, C.LPARAM(uintptr(unsafe.Pointer(n.hwnd))))
			}
		}
	case C.NUI_LVN_ITEMACTIVATE: // двойной щелчок; Enter ловит цикл (метка nuiEnter)
		emit(n, core.EvEnter)
	}
	return 0
}

// copyText — s в буфер UTF-16 с завершающим нулём; не влезло — обрезается, не разрезая
// суррогатную пару.
func copyText(buf []uint16, s string) {
	b := utf16z(s)
	n := min(len(b)-1, len(buf)-1)
	if n > 0 && b[n-1] >= 0xD800 && b[n-1] < 0xDC00 {
		n--
	}
	copy(buf, b[:n])
	buf[n] = 0
}
