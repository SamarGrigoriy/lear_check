package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
	"nui/internal/layout"
)

// Spin (driver-win32.md, §5): узел kSpin — поле EDIT, ребёнок kSpinArrows — стрелки
// msctls_updown32 (UDS_SETBUDDYINT: шаг читает и пишет текст поля; UDS_ARROWKEYS: ↑ и ↓
// в поле). Стрелки справа ставит placeSpin. Значение (n.last) принимается, как в
// GtkSpinButton: стрелками (WM_VSCROLL от стрелок), по Enter (метка nuiEnter = 3: цикл
// сообщает и пускает Enter дальше, к кнопке по умолчанию) и при уходе фокуса (EN_KILLFOCUS).
// Принятое ограничивается диапазоном; пустое поле — 0, как в GTK; не число (буквы можно
// набрать, только если в диапазоне есть минус: иначе поле ES_NUMBER) — откат.

func SpinNew(id uint32, min, max int) H {
	spinHooks.measure, spinHooks.place, spinHooks.commit = measureSpin, placeSpin, spinCommit
	h, n := newNode(kSpin, id)
	n.spinMin, n.spinMax, n.last = min, max, min
	style := uint32(C.NUI_ES_AUTOHSCROLL | C.NUI_WS_TABSTOP)
	if min >= 0 {
		style |= C.NUI_ES_NUMBER // только цифры: минуса в диапазоне нет
	}
	control(h, "EDIT", core.Itoa(min), style, C.NUI_WS_EX_CLIENTEDGE)
	C.SetPropW(n.hwnd, wstr("nuiEnter"), handleOf(3))
	a, an := newNode(kSpinArrows, 0)
	control(a, "msctls_updown32", "", C.NUI_UDS_SETBUDDYINT|C.NUI_UDS_ARROWKEYS|C.NUI_UDS_NOTHOUSANDS, 0)
	send(an.hwnd, C.NUI_UDM_SETBUDDY, uintptr(unsafe.Pointer(n.hwnd)), 0)
	send(an.hwnd, C.NUI_UDM_SETRANGE32, uintptr(min), uintptr(max))
	send(an.hwnd, C.NUI_UDM_SETPOS32, 0, uintptr(min))
	attach(h, a)
	return h
}

func SpinSetValue(h H, v int) {
	n := nodes[h]
	n.last = v
	C.SetWindowTextW(n.hwnd, wstr(core.Itoa(v)))
	send(nodes[n.kids[0]].hwnd, C.NUI_UDM_SETPOS32, 0, uintptr(v))
}

func SpinValue(h H) int { return nodes[h].last }

// spinCommit принимает текст поля: число — в диапазон, пустое — 0 (в диапазон), не
// число — прежнее значение; изменилось — EvChange.
func spinCommit(h H) {
	n := nodes[h]
	text := windowText(n.hwnd)
	v, ok := core.Atoi(text)
	switch {
	case !ok && blank(text):
		v = 0
	case !ok:
		v = n.last
	}
	v = clamp(v, n.spinMin, n.spinMax)
	if s := core.Itoa(v); text != s {
		C.SetWindowTextW(n.hwnd, wstr(s))
	}
	send(nodes[n.kids[0]].hwnd, C.NUI_UDM_SETPOS32, 0, uintptr(v))
	if v != n.last {
		n.last = v
		emit(n, core.EvChange)
	}
}

// blank — в строке только пробелы.
func blank(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return false
		}
	}
	return true
}

// measureSpin — поле под самое длинное число диапазона и стрелки.
func measureSpin(w *node, n *node) {
	_, lh := charMetrics(n)
	tw, _ := textSize(n, core.Itoa(n.spinMin), 0)
	tw2, _ := textSize(n, core.Itoa(n.spinMax), 0)
	width := max(tw, tw2) + w.px(12) + spinArrowsWidth(w)
	n.lay.Min = layout.Size{W: width, H: lh + w.px(8)}
	n.lay.Nat = n.lay.Min
}

func spinArrowsWidth(w *node) int {
	return int(C.GetSystemMetricsForDpi(C.NUI_SM_CXVSCROLL, C.UINT(w.dpi)))
}

// placeSpin ставит стрелки к правому краю места узла; ответ — ширина поля.
func placeSpin(w *node, n *node) int {
	o := n.lay.Out
	aw := min(spinArrowsWidth(w), max(o.W, 0))
	nodes[n.kids[0]].lay.Out = layout.Rect{X: o.X + o.W - aw, Y: o.Y, W: aw, H: o.H}
	return o.W - aw
}
