package drv

/*
#include "win32_windows.h"
*/
import "C"

import "nui/internal/layout"

// Контейнеры (driver-win32.md, §2, §5): Box, Grid, Stretch — виртуальные узлы без HWND;
// Frame — рамка BS_GROUPBOX, её дети — соседи по окну внутри её прямоугольника;
// Scroll — своё дочернее окно с полосами прокрутки, дети — его дочерние окна.

// BoxNew — промежуток 6 (layout.md, §2).
func BoxNew(id uint32, vertical bool) H {
	h, n := newNode(kBox, id)
	n.spacing = 6
	n.lay.Kind = layout.HBox
	if vertical {
		n.lay.Kind = layout.VBox
	}
	return h
}

func BoxSetSpacing(h H, px int) {
	nodes[h].spacing = max(px, 0)
	invalidate(h)
}

// BoxPack: ребёнок забирает свободное место, только если помечен Expand (layout.md, §3).
// Parent — контейнер, в который вставлен виджет; Nil — никуда.
func Parent(h H) H { return nodes[h].parent }

// ContainerClear уничтожает всех детей Box или Grid.
func ContainerClear(h H) {
	for _, k := range append([]H(nil), nodes[h].kids...) {
		Destroy(k)
	}
}

func BoxPack(h, child H, expand bool) {
	nodes[child].lay.Expand = expand
	attach(h, child)
}

// SpacerNew — пустая вставка для Stretch: растягивается всегда.
func SpacerNew(id uint32) H {
	h, n := newNode(kSpacer, id)
	n.lay.Expand = true
	return h
}

// GridNew — промежутки 6 по колонкам и строкам (layout.md, §2).
func GridNew(id uint32) H {
	h, n := newNode(kGrid, id)
	n.colGap, n.rowGap = 6, 6
	n.lay.Kind = layout.Grid
	return h
}

func GridAttach(h, child H, col, row, cols, rows int) {
	l := &nodes[child].lay
	l.Col, l.Row, l.Cols, l.Rows = col, row, cols, rows
	attach(h, child)
}

func GridSetSpacing(h H, colGap, rowGap int) {
	n := nodes[h]
	n.colGap, n.rowGap = max(colGap, 0), max(rowGap, 0)
	invalidate(h)
}

// SeparatorNew: горизонтальная линия для VBox, вертикальная — для HBox (SS_ETCHED*).
func SeparatorNew(id uint32, vertical bool) H {
	h, n := newNode(kSeparator, id)
	n.vertical = vertical
	style := uint32(C.NUI_SS_ETCHEDHORZ)
	if vertical {
		style = C.NUI_SS_ETCHEDVERT
	}
	control(h, "STATIC", "", style, 0)
	return h
}

// FrameNew — рамка с заголовком; "" — без заголовка.
func FrameNew(id uint32, title string) H {
	h, n := newNode(kFrame, id)
	n.text = osText(title)
	n.lay.Kind = layout.Frame
	control(h, "BUTTON", amp(title), C.NUI_BS_GROUPBOX, 0)
	return h
}

func FrameSetTitle(h H, s string) {
	n := nodes[h]
	n.text = osText(s)
	C.SetWindowTextW(n.hwnd, wstr(amp(s)))
	invalidate(h)
}

// ScrollNew — полосы прокрутки появляются, только если ребёнок не помещается;
// естественный размер — как у ребёнка (layout.md, §5).
func ScrollNew(id uint32) H {
	h, n := newNode(kScroll, id)
	n.lay.Kind = layout.Scroll
	n.hwnd = C.nui_scroll_new(hostOf(h), C.LONG_PTR(h))
	return h
}

// ContainerSetChild — Frame, Scroll: старый ребёнок уничтожается, новый вставляется.
func ContainerSetChild(h, child H) { replaceChild(h, child) }
