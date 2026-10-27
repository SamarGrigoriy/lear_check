package drv

/*
#include "win32_windows.h"
*/
import "C"

// Show, Hide — окна верхнего уровня (у остальных виджетов — SetVisible).
func Show(h H) { WindowShow(h) }
func Hide(h H) { C.ShowWindow(nodes[h].hwnd, C.NUI_SW_HIDE) }

// Destroy уничтожает виджет и его поддерево (EvDestroy каждому узлу); окно — DestroyWindow,
// и тогда EvDestroy приходят из WM_DESTROY.
// Size — размер узла после последней раскладки, логические пиксели; не показан на
// экране (скрыт сам, скрыт родитель, окно не показано) — 0, 0.
func Size(h H) (w, ht int) {
	n := nodes[h]
	if effectiveHidden(h) { // скрыт сам, предок или страница вкладок не текущая
		return 0, 0
	}
	win := get(n.win)
	if win == nil || C.IsWindowVisible(win.hwnd) == 0 {
		return 0, 0
	}
	return win.unpx(n.lay.Out.W), win.unpx(n.lay.Out.H)
}

func Destroy(h H) {
	n := get(h)
	if n == nil {
		return
	}
	if n.kind == kWindow {
		C.DestroyWindow(n.hwnd)
		return
	}
	detach(h)
	destroyTree(h, true)
}

// SetEnabled: как в GTK, выключенный контейнер выключает всё поддерево, а свой флаг
// каждого виджета сохраняется.
func SetEnabled(h H, on bool) {
	n := nodes[h]
	n.disabled = !on
	applyEnabled(h, effectiveDisabled(h))
}

func applyEnabled(h H, disabled bool) {
	n := nodes[h]
	disabled = disabled || n.disabled
	if n.hwnd != nil {
		C.EnableWindow(n.hwnd, b2i(!disabled))
	}
	for _, k := range n.kids {
		applyEnabled(k, disabled)
	}
}

func Enabled(h H) bool { return !nodes[h].disabled }

func SetVisible(h H, on bool) {
	n := nodes[h]
	if n.kind == kWindow {
		if on {
			WindowShow(h)
		} else {
			Hide(h)
		}
		return
	}
	n.hidden = !on
	if p := get(n.parent); p != nil && p.kind == kTabs {
		tabsHooks.visibility(n.parent, h)
	}
	invalidate(n.parent)
	if n.parent == Nil && n.hwnd != nil && !on {
		C.ShowWindow(n.hwnd, C.NUI_SW_HIDE)
	}
}

func Visible(h H) bool {
	n := nodes[h]
	if n.kind == kWindow {
		return C.IsWindowVisible(n.hwnd) != 0
	}
	return !n.hidden
}

// SetTooltip — всплывающая подсказка (tooltips_class32); "" убирает её. У Box и Grid
// подсказка видна над детьми без своей.
func SetTooltip(h H, s string) {
	nodes[h].tooltip = osText(s)
	applyTooltipTree(h)
}

func SetAlign(h H, hz, v uint8) {
	n := nodes[h]
	n.lay.HAlign, n.lay.VAlign = hz, v
	invalidate(n.parent)
}

func SetMinSize(h H, w, ht int) {
	n := nodes[h]
	n.minW, n.minH = w, ht
	invalidate(h)
}

func SetExpand(h H, hz, v bool) {
	n := nodes[h]
	n.expandH, n.expandV = hz, v
	invalidate(n.parent)
}

func SetPadding(h H, px int) {
	n := nodes[h]
	n.padding = max(px, 0)
	invalidate(h)
}

func Focus(h H) {
	if n := get(h); n != nil && n.hwnd != nil {
		C.SetFocus(n.hwnd)
	}
}

// SetForeground — цвет текста (Button, Entry, TextView); noColor — цвет темы (ADR-19).
func SetForeground(h H, color uint32) {
	n := nodes[h]
	n.fg = themeColor(color)
	restyleColors(n)
}

// SetBackground — фон. Box, Grid (panel) и окно красит WM_ERASEBKGND вместе с отступом;
// подписи и флажки внутри берут этот фон (WM_CTLCOLORSTATIC).
func SetBackground(h H, color uint32, panel bool) {
	n := nodes[h]
	n.bg = themeColor(color)
	restyleColors(n)
}

func themeColor(c uint32) uint32 {
	if c > noColor {
		return noColor
	}
	return c
}

// restyleColors: цветная кнопка рисуется драйвером (BS_OWNERDRAW) — визуальный стиль
// Windows её не перекрашивает; без цветов — снова обычная. Затем — перерисовка окна.
func restyleColors(n *node) {
	if n.kind == kButton {
		colored := n.fg != noColor || n.bg != noColor
		if colored != n.ownerDraw {
			n.ownerDraw = colored
			style := uint32(C.NUI_BS_PUSHBUTTON)
			switch {
			case colored:
				style = C.NUI_BS_OWNERDRAW
			case n.isDefault:
				style = C.NUI_BS_DEFPUSHBUTTON
			}
			restyle(n, style)
			n.measured = false
		}
	}
	w := n
	if n.kind != kWindow {
		w = get(n.win)
	}
	if w != nil {
		invalidate(n.parent)
		C.RedrawWindow(w.hwnd, nil, nil, C.NUI_RDW_INVALIDATE|C.NUI_RDW_ERASE|C.NUI_RDW_ALLCHILDREN)
	} else if n.hwnd != nil {
		C.InvalidateRect(n.hwnd, nil, 1)
	}
}
