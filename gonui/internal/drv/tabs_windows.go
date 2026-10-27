package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"nui/internal/core"
	"nui/internal/layout"
)

// Tabs (driver-win32.md, §5): SysTabControl32 — ярлыки и рамка. Как у Frame, HWND лежит
// в том же хосте над своими страницами, но область страниц из него вырезана (SetWindowRgn):
// щелчки по ней достаются контролам страниц, а её фон рисует хост — цветом области страниц
// у темы (tabsPane). Страницы — дети узла: все раскладываются в одно место (layout.Stack),
// видна текущая, HWND остальных скрыты. Ярлыки — у видимых страниц, как в GTK: спрятанная
// страница прячет и свой ярлык.

// tcnSelChange — TCN_SELCHANGE (TCN_FIRST - 1 = 0U - 551U), сверяется в win32_windows.h.
const tcnSelChange = 0xFFFFFDD9

func TabsNew(id uint32) H {
	t := &tabsHooks
	t.measure, t.hole, t.pane = measureTabs, tabsHole, tabsPane
	t.detach, t.sync, t.visibility = tabsDetach, tabsSync, tabsVisibility
	t.clicked, t.key = tabsClicked, tabsKey
	h, n := newNode(kTabs, id)
	n.lay.Kind = layout.Stack
	control(h, "SysTabControl32", "", C.NUI_WS_TABSTOP, 0)
	return h
}

func TabsAdd(h, page H, title string) {
	n := nodes[h]
	nodes[page].tabTitle = title
	attach(h, page)
	if n.tabCur == Nil && !nodes[page].hidden {
		n.tabCur = page
	}
	tabsSync(h)
}

func TabsSetCurrent(h H, i int) {
	n := nodes[h]
	if k := n.kids[i]; !nodes[k].hidden {
		tabsSelect(h, k, false)
	}
}

func TabsCurrent(h H) int {
	n := nodes[h]
	for i, k := range n.kids {
		if k == n.tabCur {
			return i
		}
	}
	return -1
}

func TabsSetTitle(h, page H, title string) {
	nodes[page].tabTitle = title
	tabsSync(h)
}

// tabsSync — ярлыки заново по видимым страницам, выбран ярлык текущей.
func tabsSync(h H) {
	n := nodes[h]
	send(n.hwnd, C.NUI_TCM_DELETEALLITEMS, 0, 0)
	n.tabShown = n.tabShown[:0]
	sel := -1
	for _, k := range n.kids {
		if nodes[k].hidden {
			continue
		}
		if k == n.tabCur {
			sel = len(n.tabShown)
		}
		C.nui_tabs_insert(n.hwnd, C.int(len(n.tabShown)), wstr(amp(nodes[k].tabTitle)))
		n.tabShown = append(n.tabShown, k)
	}
	send(n.hwnd, C.NUI_TCM_SETCURSEL, uintptr(sel), 0)
	invalidate(h)
}

// tabsAway — какая страница станет текущей вместо page, которую убирают или прячут:
// следующая видимая, иначе предыдущая (как GtkNotebook).
func tabsAway(n *node, page H) H {
	i := 0
	for i < len(n.kids) && n.kids[i] != page {
		i++
	}
	for j := i + 1; j < len(n.kids); j++ {
		if !nodes[n.kids[j]].hidden {
			return n.kids[j]
		}
	}
	for j := i - 1; j >= 0; j-- {
		if !nodes[n.kids[j]].hidden {
			return n.kids[j]
		}
	}
	return Nil
}

// tabsDetach — страницу убирают (detach, до удаления из детей).
func tabsDetach(h, page H) {
	if n := nodes[h]; n.tabCur == page {
		n.tabCur = tabsAway(n, page)
	}
}

// tabsVisibility — страницу спрятали или показали.
func tabsVisibility(h, page H) {
	n := nodes[h]
	switch {
	case nodes[page].hidden && n.tabCur == page:
		n.tabCur = tabsAway(n, page)
	case !nodes[page].hidden && n.tabCur == Nil:
		n.tabCur = page
	}
	tabsSync(h)
}

// tabsSelect делает page текущей; user — выбор пользователя (EvChange). Фокус со
// скрываемой страницы переходит на ярлыки, иначе он остался бы на невидимом контроле.
func tabsSelect(h, page H, user bool) {
	n := nodes[h]
	old := n.tabCur
	if old == page {
		return
	}
	n.tabCur = page
	for i, k := range n.tabShown {
		if k == page {
			send(n.hwnd, C.NUI_TCM_SETCURSEL, uintptr(i), 0)
		}
	}
	if _, f := nodeOf(C.GetFocus()); f != nil && old != Nil && within(f, old) {
		C.SetFocus(n.hwnd)
	}
	if get(n.win) != nil {
		relayout(n.win) // сразу: страница сменилась на глазах у пользователя
	} else {
		invalidate(h)
	}
	if user {
		emit(n, core.EvChange)
	}
}

// within — узел f лежит в поддереве h.
func within(f *node, h H) bool {
	for p := f; p != nil; p = get(p.parent) {
		if p == nodes[h] {
			return true
		}
	}
	return false
}

// tabsClicked — TCN_SELCHANGE: пользователь выбрал ярлык мышью или стрелками.
func tabsClicked(h H) {
	n := nodes[h]
	i := int(int32(send(n.hwnd, C.NUI_TCM_GETCURSEL, 0, 0)))
	if i >= 0 && i < len(n.tabShown) {
		tabsSelect(h, n.tabShown[i], true)
	}
}

// tabsKey — Ctrl+Tab, Ctrl+PgDn (back — с Shift, Ctrl+PgUp) у контрола focus: соседняя
// видимая страница ближайших вкладок выше (по кругу). false — фокус не во вкладках.
func tabsKey(focus C.HWND, back bool) bool {
	h := Nil
	for hw := focus; hw != nil && h == Nil; hw = C.GetParent(hw) {
		h, _ = nodeOf(hw)
	}
	for ; h != Nil && nodes[h].kind != kTabs; h = nodes[h].parent {
	}
	if h == Nil {
		return false
	}
	n := nodes[h]
	if len(n.tabShown) == 0 {
		return true
	}
	cur := 0
	for i, k := range n.tabShown {
		if k == n.tabCur {
			cur = i
		}
	}
	step := 1
	if back {
		step = len(n.tabShown) - 1
	}
	tabsSelect(h, n.tabShown[(cur+step)%len(n.tabShown)], true)
	return true
}

// offPage — h — страница вкладок, но не текущая (её HWND скрыты).
func offPage(h H) bool {
	p := get(nodes[h].parent)
	return p != nil && p.kind == kTabs && p.tabCur != h
}

// measureTabs — оформление: ярлыки сверху и рамка (TCM_ADJUSTRECT). Ширина — не меньше
// самого широкого ярлыка со стрелками прокрутки ярлыков, как у GtkNotebook: остальные
// ярлыки прокручиваются.
func measureTabs(w *node, n *node) {
	const big = 1000
	var in [4]C.int
	C.nui_tabs_inner(n.hwnd, big, big, &in[0])
	l, t := int(in[0]), int(in[1])
	n.lay.Inset = layout.Insets{L: l, T: t, R: big - l - int(in[2]), B: big - t - int(in[3])}
	widest := 0
	for i := range n.tabShown {
		widest = max(widest, int(C.nui_tabs_item_width(n.hwnd, C.int(i))))
	}
	if widest > 0 {
		if len(n.tabShown) > 1 {
			widest += 2 * int(C.GetSystemMetricsForDpi(C.NUI_SM_CXHSCROLL, C.UINT(w.dpi)))
		}
		n.lay.Floor.W = max(n.lay.Floor.W, widest+n.lay.Inset.L+n.lay.Inset.R)
	}
}

// tabsHole вырезает из окна контрола область страниц (после SetWindowPos).
func tabsHole(n *node) {
	o, in := n.lay.Out, n.lay.Inset
	hole := [6]int{o.W, o.H, in.L, in.T, o.W - in.L - in.R, o.H - in.T - in.B}
	if hole == n.tabHole {
		return
	}
	n.tabHole = hole
	r := [4]C.int{C.int(hole[2]), C.int(hole[3]), C.int(hole[4]), C.int(hole[5])}
	C.nui_tabs_hole(n.hwnd, C.int(o.W), C.int(o.H), &r[0])
}

var (
	paneColor uint32
	paneKnown bool
)

// tabsPane — цвет области страниц у темы, 0xRRGGBB: фон для неё и для подписей на ней.
func tabsPane() uint32 {
	if !paneKnown {
		paneColor, paneKnown = uint32(bgr(uint32(C.nui_tabs_pane()))), true
	}
	return paneColor
}
