package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"nui/internal/core"
	"nui/internal/layout"
)

// Radio (driver-win32.md, §5): узел-ряд, как Box; варианты — узлы kRadioItem с BUTTON
// BS_RADIOBUTTON. Выбор ведёт nui (n.last), а не Windows: у BS_AUTORADIOBUTTON группа —
// соседи по Z-порядку до следующего WS_GROUP, и в неё попали бы чужие контролы. Tab
// попадает на выбранный вариант (WS_TABSTOP только у него), стрелки выбирают соседний
// вариант и у края группы останавливаются (nav_key), как в GTK.

// Промежутки между вариантами — как в драйвере GTK.
const (
	radioGapV = 6
	radioGapH = 12
)

func RadioNew(id uint32) H {
	radioHooks.clicked, radioHooks.key = radioClicked, radioKey
	h, n := newNode(kRadio, id)
	n.lay.Kind, n.spacing, n.last = layout.VBox, radioGapV, -1
	return h
}

func RadioAdd(h H, id uint32, text string) {
	c, cn := newNode(kRadioItem, 0)
	cn.text = osText(text)
	control(c, "BUTTON", amp(text), C.NUI_BS_RADIOBUTTON, 0)
	attach(h, c)
	if n := nodes[h]; n.last < 0 {
		radioCheck(h, 0)
	}
}

func RadioSelected(h H) int       { return nodes[h].last }
func RadioSetSelected(h H, i int) { radioCheck(h, i) }

func RadioSetHorizontal(h H, on bool) {
	n := nodes[h]
	n.lay.Kind, n.spacing = layout.VBox, radioGapV
	if on {
		n.lay.Kind, n.spacing = layout.HBox, radioGapH
	}
	invalidate(h)
}

func RadioFocus(h H) {
	if n := nodes[h]; n.last >= 0 {
		C.SetFocus(nodes[n.kids[n.last]].hwnd)
	}
}

// radioCheck отмечает вариант i, снимает остальные; WS_TABSTOP — только у отмеченного.
// Фокус, если он на варианте, переходит на отмеченный: снятый вариант, получив фокус снова
// (окно стало активным), прислал бы BN_CLICKED и сам себя выбрал.
func radioCheck(h H, i int) {
	n := nodes[h]
	n.last = i
	for j, k := range n.kids {
		hw := nodes[k].hwnd
		send(hw, C.NUI_BM_SETCHECK, uintptr(b2i(j == i)), 0)
		style := uintptr(C.GetWindowLongPtrW(hw, C.NUI_GWL_STYLE)) &^ C.NUI_WS_TABSTOP
		if j == i {
			style |= C.NUI_WS_TABSTOP
		}
		C.SetWindowLongPtrW(hw, C.NUI_GWL_STYLE, C.LONG_PTR(style))
	}
	if _, f := nodeOf(C.GetFocus()); f != nil && f.parent == h && f != nodes[n.kids[i]] {
		C.SetFocus(nodes[n.kids[i]].hwnd)
	}
}

// radioClicked — BN_CLICKED варианта: щелчок мышью или фокус на невыбранном варианте
// (так кнопка-переключатель сообщает о выборе с клавиатуры).
func radioClicked(item H) {
	p := nodes[item].parent
	n := get(p)
	if n == nil {
		return
	}
	for i, k := range n.kids {
		if k == item && i != n.last {
			radioCheck(p, i)
			emit(n, core.EvChange)
		}
	}
}

// radioKey — стрелка на варианте focus: соседний вариант (back — назад) выбран и в фокусе;
// у края группы — ничего, как в GTK. false — focus не вариант nui.
func radioKey(focus C.HWND, back bool) bool {
	item, c := nodeOf(focus)
	if c == nil || c.kind != kRadioItem {
		return false
	}
	p := c.parent
	n := nodes[p]
	cur := 0
	for i, k := range n.kids {
		if k == item {
			cur = i
		}
	}
	next := cur + 1
	if back {
		next = cur - 1
	}
	if next < 0 || next >= len(n.kids) {
		return true
	}
	if next != n.last {
		radioCheck(p, next)
		emit(n, core.EvChange)
	}
	C.SetFocus(nodes[n.kids[next]].hwnd)
	return true
}
