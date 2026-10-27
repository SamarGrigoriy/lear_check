package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"nui/internal/layout"
)

// Split (driver-win32.md, §5): у Windows нет готового разделителя. Узел — ряд из двух
// панелей с промежутком-полосой (движок меряет его как HBox или VBox), HWND узла — эта
// полоса: курсор «↔», за неё тянут. Панели расставляет placeSplit по правилу GtkPaned:
// первая — SetPosition или свой естественный размер, не меньше своего минимума; вторая —
// остальное, тоже не меньше минимума.

// splitGap — ширина полосы, логические пиксели: как у разделителей Проводника (в GTK —
// линия в 1 точку с невидимой полосой захвата поверх панелей).
const splitGap = 5

func SplitNew(id uint32, vertical bool) H {
	h, n := newNode(kSplit, id)
	n.vertical, n.splitPos = vertical, -1
	n.lay.Kind = layout.HBox
	if vertical {
		n.lay.Kind = layout.VBox
	}
	control(h, "nuiWindow", "", 0, 0)
	return h
}

func SplitSet(h, first, second H) {
	nodes[first].lay.Expand, nodes[second].lay.Expand = false, true
	attach(h, first)
	attach(h, second)
}

func SplitSetPosition(h H, px int) {
	nodes[h].splitPos = px
	invalidate(h)
}

// SplitPosition — размер первой панели после раскладки; до неё — заданный (или 0).
func SplitPosition(h H) int {
	n := nodes[h]
	if w := get(n.win); w != nil && n.splitCur > 0 {
		return w.unpx(n.splitCur)
	}
	return max(n.splitPos, 0)
}

// main и cross — размеры вдоль и поперёк деления.
func (n *node) splitMain(s layout.Size) int {
	if n.vertical {
		return s.H
	}
	return s.W
}

// placeSplit расставляет панели (заново раскладывая их поддеревья), полосу и HWND детей.
func placeSplit(w *node, h H, ox, oy int, hide bool) {
	n := nodes[h]
	o := n.lay.Out
	a, b := nodes[n.kids[0]], nodes[n.kids[1]]
	gap := w.px(splitGap)
	total := n.splitMain(layout.Size{W: o.W, H: o.H})
	pos := n.splitMain(a.lay.Nat)
	if n.splitPos >= 0 {
		pos = w.px(n.splitPos)
	}
	switch { // одна панель скрыта — другая на всё место, полосы нет
	case a.hidden:
		pos = -gap
	case b.hidden:
		pos, gap = total, 0
	default:
		pos = max(min(pos, total-gap-n.splitMain(b.lay.Min)), n.splitMain(a.lay.Min))
	}
	n.splitCur = max(pos, 0)
	r1, r2, bar := o, o, o
	if n.vertical {
		r1.H, r2.Y, r2.H = max(pos, 0), o.Y+pos+gap, max(o.H-pos-gap, 0)
		bar.Y, bar.H = o.Y+pos, gap
	} else {
		r1.W, r2.X, r2.W = max(pos, 0), o.X+pos+gap, max(o.W-pos-gap, 0)
		bar.X, bar.W = o.X+pos, gap
	}
	layout.Arrange(&a.lay, r1)
	layout.Arrange(&b.lay, r2)
	barHidden := hide || a.hidden || b.hidden
	C.SetWindowPos(n.hwnd, nil, C.int(bar.X-ox), C.int(bar.Y-oy), C.int(max(bar.W, 0)), C.int(max(bar.H, 0)),
		C.NUI_SWP_NOZORDER|C.NUI_SWP_NOACTIVATE)
	show := C.int(C.NUI_SW_SHOWNA)
	if barHidden {
		show = C.NUI_SW_HIDE
	}
	C.ShowWindow(n.hwnd, show)
	for _, k := range n.kids {
		place(w, k, ox, oy, hide)
	}
}

// splitProc — полоса разделителя: курсор, перетаскивание, фон.
func splitProc(h H, n *node, m uint32, wp, lp uintptr) (uintptr, bool) {
	switch m {
	case C.NUI_WM_SETCURSOR:
		id := 32644 // IDC_SIZEWE
		if n.vertical {
			id = 32645 // IDC_SIZENS
		}
		C.SetCursor(C.nui_cursor(C.int(id)))
		return 1, true
	case C.NUI_WM_LBUTTONDOWN:
		var pt C.POINT
		C.GetCursorPos(&pt)
		n.dragStart = [2]int{n.splitCur, n.splitMainPoint(pt)}
		C.SetCapture(n.hwnd)
		return 0, true
	case C.NUI_WM_MOUSEMOVE:
		if C.GetCapture() != n.hwnd {
			return 0, true
		}
		var pt C.POINT
		C.GetCursorPos(&pt)
		if w := get(n.win); w != nil {
			n.splitPos = w.unpx(max(n.dragStart[0]+n.splitMainPoint(pt)-n.dragStart[1], 0))
			relayout(n.win) // сразу, а не через очередь: граница идёт за мышью
		}
		return 0, true
	case C.NUI_WM_LBUTTONUP:
		C.ReleaseCapture()
		return 0, true
	case C.NUI_WM_ERASEBKGND:
		var r C.RECT
		C.GetClientRect(n.hwnd, &r)
		bg := C.GetSysColorBrush(C.NUI_COLOR_BTNFACE)
		if c := inheritedBg(n); c != noColor {
			bg = brushFor(c)
		}
		C.FillRect(hdcOf(wp), &r, bg)
		return 1, true
	}
	return 0, false
}

func (n *node) splitMainPoint(pt C.POINT) int {
	if n.vertical {
		return int(pt.y)
	}
	return int(pt.x)
}
