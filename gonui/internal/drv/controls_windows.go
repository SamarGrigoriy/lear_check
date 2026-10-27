package drv

/*
#include "win32_windows.h"
*/
import "C"

// Контролы (driver-win32.md, §5). Текст, который участвует в измерении, узел хранит сам:
// Text() возвращает исходную строку, а не то, что видит Win32 (например, с удвоенным &).

// --- Label ---

func LabelNew(id uint32, text string) H {
	h, n := newNode(kLabel, id)
	n.text = osText(text)
	control(h, "STATIC", text, labelStyle(n), 0)
	return h
}

// labelStyle: перенос — SS_LEFT/CENTER/RIGHT (они переносят слова); без переноса слева —
// SS_LEFTNOWORDWRAP, по центру и справа — те же SS_CENTER/RIGHT: уже текста метка без
// многоточия не бывает (её минимум — ширина текста), так что переносить нечего.
// Многоточие: End → SS_ENDELLIPSIS, Middle и Start → SS_PATHELLIPSIS (§10).
// Однострочный текст — по центру ячейки по вертикали (SS_CENTERIMAGE), как у GtkLabel.
func labelStyle(n *node) uint32 {
	s := uint32(C.NUI_SS_NOPREFIX)
	if !n.wrap && !hasNewline(n.text) {
		s |= C.NUI_SS_CENTERIMAGE
	}
	switch n.textAlign {
	case 2: // Center
		s |= C.NUI_SS_CENTER
	case 3: // End
		s |= C.NUI_SS_RIGHT
	default:
		if !n.wrap {
			s |= C.NUI_SS_LEFTNOWORDWRAP
		}
	}
	switch n.ellipsis {
	case 3:
		s |= C.NUI_SS_ENDELLIPSIS
	case 1, 2:
		s |= C.NUI_SS_PATHELLIPSIS
	}
	return s
}

// labelEditStyle — выделяемая подпись: read-only EDIT без рамки, с тем же выравниванием;
// многострочная — при переносе или переводах строк.
func labelEditStyle(n *node) uint32 {
	s := uint32(C.NUI_ES_READONLY | C.NUI_ES_AUTOHSCROLL | C.NUI_WS_TABSTOP)
	if n.wrap || hasNewline(n.text) {
		s |= C.NUI_ES_MULTILINE
	}
	switch n.textAlign {
	case 2:
		s |= C.NUI_ES_CENTER
	case 3:
		s |= C.NUI_ES_RIGHT
	}
	return s
}

// labelRestyle применяет стиль подписи: STATIC меняется на лету, EDIT — пересоздаётся.
func labelRestyle(h H) {
	n := nodes[h]
	if n.selectable {
		recreate(h, "EDIT", crlf(n.text), labelEditStyle(n), 0)
		return
	}
	restyle(n, labelStyle(n))
}

func restyle(n *node, style uint32) {
	old := uintptr(C.GetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE))
	keep := old & 0xFFFF0000 // WS_*: видимость, WS_CHILD, WS_TABSTOP…
	C.SetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE, C.LONG_PTR(keep|uintptr(style)&0xFFFF))
	C.InvalidateRect(n.hwnd, nil, 1)
}

func LabelSetText(h H, s string) {
	n := nodes[h]
	multi := hasNewline(n.text)
	n.text = osText(s)
	if n.selectable && multi != hasNewline(n.text) {
		labelRestyle(h) // EDIT стал или перестал быть многострочным
	} else if !n.selectable {
		restyle(n, labelStyle(n))
	}
	if n.selectable {
		C.SetWindowTextW(n.hwnd, wstr(crlf(n.text)))
	} else {
		C.SetWindowTextW(n.hwnd, wstr(n.text))
	}
	remeasure(h)
}

func hasNewline(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return true
		}
	}
	return false
}

func LabelText(h H) string { return nodes[h].text }

func LabelSetEllipsis(h H, e uint8) {
	nodes[h].ellipsis = e
	labelRestyle(h)
	remeasure(h)
}

func LabelSetWrap(h H, on bool) {
	nodes[h].wrap = on
	labelRestyle(h)
	remeasure(h)
}

func LabelSetTextAlign(h H, a uint8) {
	nodes[h].textAlign = a
	labelRestyle(h)
}

// LabelSetFontScale — свой шрифт подписи: шрифт окна в pct процентов (setFont).
func LabelSetFontScale(h H, pct int) {
	n := nodes[h]
	if pct < 1 {
		pct = 100
	}
	n.fontPct = pct
	setFont(n)
	remeasure(h)
}

// LabelSetSelectable: текст можно выделить и скопировать — read-only EDIT без рамки
// (driver-win32.md, §5). Многоточие EDIT не рисует.
func LabelSetSelectable(h H, on bool) {
	n := nodes[h]
	if n.selectable == on {
		return
	}
	n.selectable = on
	if on {
		recreate(h, "EDIT", crlf(n.text), labelEditStyle(n), 0)
	} else {
		recreate(h, "STATIC", n.text, labelStyle(n), 0)
	}
}

func LabelSetBold(h H, on bool) {
	n := nodes[h]
	n.bold = on
	setFont(n)
	remeasure(h)
}

func LabelSetColor(h H, color uint32) {
	n := nodes[h]
	n.color = color
	C.InvalidateRect(n.hwnd, nil, 1)
}

// --- Button, Check ---

func ButtonNew(id uint32, text string) H {
	h, n := newNode(kButton, id)
	n.text = osText(text)
	control(h, "BUTTON", amp(text), C.NUI_BS_PUSHBUTTON|C.NUI_WS_TABSTOP, 0)
	return h
}

func CheckNew(id uint32, text string) H {
	h, n := newNode(kCheck, id)
	n.text = osText(text)
	control(h, "BUTTON", amp(text), C.NUI_BS_AUTOCHECKBOX|C.NUI_WS_TABSTOP, 0)
	return h
}

func ButtonSetText(h H, s string) {
	n := nodes[h]
	n.text = osText(s)
	C.SetWindowTextW(n.hwnd, wstr(amp(s)))
	remeasure(h)
}

func ButtonText(h H) string { return nodes[h].text }

func CheckSetChecked(h H, on bool) { send(nodes[h].hwnd, C.NUI_BM_SETCHECK, uintptr(b2i(on)), 0) }
func CheckChecked(h H) bool        { return send(nodes[h].hwnd, C.NUI_BM_GETCHECK, 0, 0) == 1 }

// --- Entry ---

func EntryNew(id uint32) H {
	h, _ := newNode(kEntry, id)
	control(h, "EDIT", "", C.NUI_ES_AUTOHSCROLL|C.NUI_WS_TABSTOP, C.NUI_WS_EX_CLIENTEDGE)
	return h
}

func EntrySetText(h H, s string) { C.SetWindowTextW(nodes[h].hwnd, wstr(s)) }
func EntryText(h H) string       { return windowText(nodes[h].hwnd) }

// EntrySetPlaceholder — EM_SETCUEBANNER (только comctl32 v6, то есть с манифестом).
func EntrySetPlaceholder(h H, s string) { sendStr(nodes[h].hwnd, C.NUI_EM_SETCUEBANNER, 1, s) }

func EntrySetPassword(h H, on bool) {
	ch := uintptr(0)
	if on {
		ch = 0x25CF // «●», как у полей паролей Windows
	}
	n := nodes[h]
	send(n.hwnd, C.NUI_EM_SETPASSWORDCHAR, ch, 0)
	C.InvalidateRect(n.hwnd, nil, 1)
}

func EntrySetReadOnly(h H, on bool) { send(nodes[h].hwnd, C.NUI_EM_SETREADONLY, uintptr(b2i(on)), 0) }

func EntrySetWidthChars(h H, n int) {
	nodes[h].widthChars = n
	remeasure(h)
}

// EntrySetActivatesDefault: false — Enter в поле даёт EvEnter (OnEnter), кнопка по
// умолчанию не нажимается. Метку "nuiEnter" проверяет цикл до IsDialogMessage.
func EntrySetActivatesDefault(h H, on bool) {
	hwnd := nodes[h].hwnd
	if on {
		C.RemovePropW(hwnd, wstr("nuiEnter"))
	} else {
		C.SetPropW(hwnd, wstr("nuiEnter"), handleOf(1))
	}
}

// --- Combo ---

func ComboNew(id uint32, editable bool) H {
	h, n := newNode(kCombo, id)
	n.editable = editable
	style := uint32(C.NUI_CBS_DROPDOWNLIST)
	if editable {
		style = C.NUI_CBS_DROPDOWN | C.NUI_CBS_AUTOHSCROLL
	}
	control(h, "COMBOBOX", "", style|C.NUI_WS_VSCROLL|C.NUI_WS_TABSTOP, 0)
	return h
}

func ComboAdd(h H, s string) {
	n := nodes[h]
	n.items = append(n.items, s)
	sendStr(n.hwnd, C.NUI_CB_ADDSTRING, 0, s)
	remeasure(h)
}

func ComboClear(h H) {
	n := nodes[h]
	n.items = nil
	send(n.hwnd, C.NUI_CB_RESETCONTENT, 0, 0)
	remeasure(h)
}

func ComboSetSelected(h H, i int) { send(nodes[h].hwnd, C.NUI_CB_SETCURSEL, uintptr(i), 0) }

// ComboSelected: у ComboEntry, как в GTK, ввод своего текста снимает выбор — Win32 же
// помнит последний выбранный пункт, поэтому выбор засчитывается, только пока текст совпадает.
func ComboSelected(h H) int {
	n := nodes[h]
	r := int(int32(send(n.hwnd, C.NUI_CB_GETCURSEL, 0, 0)))
	if r < 0 || r >= len(n.items) || n.editable && windowText(n.hwnd) != n.items[r] {
		return -1
	}
	return r
}

func ComboSetText(h H, s string) { C.SetWindowTextW(nodes[h].hwnd, wstr(s)) }
func ComboText(h H) string       { return windowText(nodes[h].hwnd) }

// --- Slider, Progress ---

func SliderNew(id uint32, min, max int) H {
	h, n := newNode(kSlider, id)
	control(h, "msctls_trackbar32", "", C.NUI_TBS_NOTICKS|C.NUI_WS_TABSTOP, 0)
	send(n.hwnd, C.NUI_TBM_SETRANGEMIN, 0, uintptr(min))
	send(n.hwnd, C.NUI_TBM_SETRANGEMAX, 1, uintptr(max))
	n.last = min
	send(n.hwnd, C.NUI_TBM_SETPOS, 1, uintptr(min))
	return h
}

func SliderSetValue(h H, v int) {
	n := nodes[h]
	n.last = v
	send(n.hwnd, C.NUI_TBM_SETPOS, 1, uintptr(v))
}

func SliderValue(h H) int { return int(int32(send(nodes[h].hwnd, C.NUI_TBM_GETPOS, 0, 0))) }

func ProgressNew(id uint32) H {
	h, n := newNode(kProgress, id)
	control(h, "msctls_progress32", "", 0, 0)
	send(n.hwnd, C.NUI_PBM_SETRANGE32, 0, 1000)
	return h
}

func ProgressSetValue(h H, f float64) {
	n := nodes[h]
	if n.marquee {
		n.marquee = false
		send(n.hwnd, C.NUI_PBM_SETMARQUEE, 0, 0)
		restyle(n, uint32(C.GetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE))&0xFFFF&^C.NUI_PBS_MARQUEE)
	}
	n.value = f
	send(n.hwnd, C.NUI_PBM_SETPOS, uintptr(int(f*1000+0.5)), 0)
}

func ProgressValue(h H) float64 { return nodes[h].value }

// ProgressPulse — стиль PBS_MARQUEE: бегунок Windows двигает сам; следующий SetValue
// выключает режим.
func ProgressPulse(h H) {
	n := nodes[h]
	if !n.marquee {
		n.marquee = true
		restyle(n, uint32(C.GetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE))&0xFFFF|C.NUI_PBS_MARQUEE)
		send(n.hwnd, C.NUI_PBM_SETMARQUEE, 1, 30)
	}
}

// ProgressSetText: у прогресс-бара Win32 нет надписи (api.md, §12; driver-win32.md, §10).
func ProgressSetText(h H, s string) {}

// --- TextView ---

// TextNew: на Windows outer и view — один и тот же узел (EDIT с полосами прокрутки).
func TextNew(id uint32) (outer, view H) {
	h, _ := newNode(kText, id)
	textControl(h, "")
	return h, h
}

// textControl создаёт многострочный EDIT: без переноса — с горизонтальной прокруткой.
// Стиль переноса на лету не меняется — SetWrap пересоздаёт контрол.
func textControl(h H, text string) {
	n := nodes[h]
	control(h, "EDIT", crlf(text), textStyle(n), C.NUI_WS_EX_CLIENTEDGE)
	send(n.hwnd, C.NUI_EM_SETLIMITTEXT, 0x7FFFFFFE, 0)
}

func textStyle(n *node) uint32 {
	style := uint32(C.NUI_ES_MULTILINE | C.NUI_ES_AUTOVSCROLL | C.NUI_ES_WANTRETURN | C.NUI_ES_NOHIDESEL |
		C.NUI_WS_VSCROLL | C.NUI_WS_TABSTOP)
	if !n.wrapText {
		style |= C.NUI_WS_HSCROLL | C.NUI_ES_AUTOHSCROLL
	}
	return style
}

func TextSetText(view H, s string) { C.SetWindowTextW(nodes[view].hwnd, wstr(crlf(s))) }
func TextText(view H) string       { return lf(windowText(nodes[view].hwnd)) }

func TextAppend(view H, s string, autoscroll bool, maxLines int) {
	C.nui_text_append(nodes[view].hwnd, wstr(crlf(s)), b2i(autoscroll), C.int(maxLines))
}

func TextSetReadOnly(view H, on bool) {
	send(nodes[view].hwnd, C.NUI_EM_SETREADONLY, uintptr(b2i(on)), 0)
}

func TextSetMonospace(view H, on bool) {
	n := nodes[view]
	n.mono = on
	setFont(n)
}

func TextSetWrap(view H, on bool) {
	n := nodes[view]
	if n.wrapText == on {
		return
	}
	n.wrapText = on
	text := windowText(n.hwnd)
	readOnly := uintptr(C.GetWindowLongPtrW(n.hwnd, C.NUI_GWL_STYLE))&C.NUI_ES_READONLY != 0
	recreate(view, "EDIT", text, textStyle(n), C.NUI_WS_EX_CLIENTEDGE)
	send(n.hwnd, C.NUI_EM_SETLIMITTEXT, 0x7FFFFFFE, 0)
	TextSetReadOnly(view, readOnly)
	TextSetEnter(view, n.id, n.enterText) // новый контрол — метка, подсказка и меню заново
	TextSetPlaceholder(view, n.hint)
	ctxForward(n)
}

// TextSetEnter: on — Enter без Shift даёт EvEnter вместо перевода строки (метку
// nuiEnter = 2 проверяет цикл до IsDialogMessage); Shift+Enter — перевод строки.
func TextSetEnter(view H, id uint32, on bool) {
	n := nodes[view]
	n.enterText = on
	if on {
		C.SetPropW(n.hwnd, wstr("nuiEnter"), handleOf(2))
	} else {
		C.RemovePropW(n.hwnd, wstr("nuiEnter"))
	}
}

// TextSetPlaceholder — подсказка в пустом поле (nui_edit_hint: у многострочного EDIT
// нет EM_SETCUEBANNER); "" — без подсказки.
func TextSetPlaceholder(view H, s string) {
	n := nodes[view]
	n.hint = s
	C.nui_edit_hint(n.hwnd, wstr(crlf(osText(s))))
}
