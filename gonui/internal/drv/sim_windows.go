//go:build gui

package drv

// Помощники интеграционных тестов на Windows (тот же набор, что sim_linux.go): имитируют
// действия пользователя — шлют контролам и их родителям те же сообщения, что Windows
// при вводе мышью и клавиатурой, мимо счётчика quiet, поэтому обработчики On* срабатывают.

/*
#cgo CFLAGS: -DNUI_GUI_TEST
#include "win32_windows.h"
*/
import "C"

import "unsafe"

// notifyParent — уведомление WM_COMMAND родителю контрола, как от пользователя.
func notifyParent(n *node, code uintptr) {
	send(C.GetParent(n.hwnd), C.NUI_WM_COMMAND, code<<16, uintptrOf(n.hwnd))
}

func uintptrOf(h C.HWND) uintptr { return uintptr(unsafe.Pointer(h)) }

// SimClick — как нажатие кнопки мышью: BN_CLICKED родителю.
func SimClick(h H) { notifyParent(nodes[h], C.NUI_BN_CLICKED) }

// SimClose — как крестик: WM_CLOSE приходит не сразу, а через очередь.
func SimClose(h H) { C.PostMessageW(nodes[h].hwnd, C.NUI_WM_CLOSE, 0, 0) }

// Allocation — место узла после раскладки: у окна — клиентская область, у виджета — Out
// движка, в координатах окна (как allocation в GTK).
func Allocation(h H) (x, y, w, ht int) {
	n := nodes[h]
	if n.kind == kWindow {
		var cs [2]C.int
		C.nui_client_size(n.hwnd, &cs[0])
		return 0, 0, int(cs[0]), int(cs[1])
	}
	o := n.lay.Out
	return o.X, o.Y, o.W, o.H
}

// SimType — как ввод текста в поле: EN_CHANGE приходит один раз.
func SimType(h H, s string) { C.SetWindowTextW(nodes[h].hwnd, wstr(s)) }

// SimComboType — ввод в поле ComboEntry: текст в дочерний EDIT комбобокса и, как при
// вводе с клавиатуры, CBN_EDITCHANGE родителю (на WM_SETTEXT комбобокс его не шлёт).
func SimComboType(h H, s string) {
	n := nodes[h]
	C.SetWindowTextW(C.GetWindow(n.hwnd, C.NUI_GW_CHILD), wstr(s))
	notifyParent(n, C.NUI_CBN_EDITCHANGE)
}

// SimEnter — как Enter: в поле с OnEnter — EvEnter сразу; иначе нажатие клавиши идёт
// через настоящий цикл (IsDialogMessage → кнопка по умолчанию), то есть позже.
func SimEnter(h H) {
	n := nodes[h]
	target := n.hwnd
	if n.kind == kCombo {
		target = C.GetWindow(n.hwnd, C.NUI_GW_CHILD)
	}
	if C.GetPropW(target, wstr("nuiEnter")) != nil {
		send(msg(), C.NUI_WM_ENTER, 0, uintptrOf(target))
		return
	}
	C.PostMessageW(target, C.NUI_WM_KEYDOWN, C.NUI_VK_RETURN, 0)
}

// SimToggle — как щелчок по флажку: состояние меняется, BN_CLICKED родителю.
func SimToggle(h H) {
	n := nodes[h]
	send(n.hwnd, C.NUI_BM_SETCHECK, uintptr(b2i(!CheckChecked(h))), 0)
	notifyParent(n, C.NUI_BN_CLICKED)
}

// SimSelect — как выбор пункта в списке: CBN_SELCHANGE родителю.
func SimSelect(h H, i int) {
	n := nodes[h]
	send(n.hwnd, C.NUI_CB_SETCURSEL, uintptr(i), 0)
	notifyParent(n, C.NUI_CBN_SELCHANGE)
}

// SimSlide — как перетаскивание ползунка: WM_HSCROLL родителю.
func SimSlide(h H, v int) {
	n := nodes[h]
	send(n.hwnd, C.NUI_TBM_SETPOS, 1, uintptr(v))
	send(C.GetParent(n.hwnd), C.NUI_WM_HSCROLL, 4 /* TB_THUMBPOSITION */, uintptrOf(n.hwnd))
}

// SimEdit — как правка многострочного текста: новый текст и EN_CHANGE родителю. На
// WM_SETTEXT многострочный EDIT EN_CHANGE не шлёт; пользовательская правка даёт одно
// уведомление и на непустом буфере (на GTK — два: удаление и вставка).
func SimEdit(view H, s string) {
	n := nodes[view]
	C.SetWindowTextW(n.hwnd, wstr(crlf(s)))
	notifyParent(n, C.NUI_EN_CHANGE)
}

// ComboEntryForTest — поле ComboEntry для имитации Enter в нём: на Windows SimEnter
// сам находит поле по узлу комбобокса.
func ComboEntryForTest(h H) H { return h }

// dialogButton — нажать кнопку id в открытом сообщении: TDM_CLICK_BUTTON у TaskDialog,
// WM_COMMAND у MessageBoxW.
func dialogButton(id uintptr) bool {
	d := C.nui_find_dialog()
	if d == nil || C.nui_dialog_kind() > 2 {
		return false
	}
	m := C.UINT(C.NUI_WM_COMMAND)
	if C.nui_task_dialog() != 0 {
		m = C.NUI_TDM_CLICK_BUTTON
	}
	C.PostMessageW(d, m, C.WPARAM(id), 0)
	return true
}

// SimAnswer — ответ в открытом сообщении: OK/«Да» (true) или «Нет» (false). false — диалога нет.
func SimAnswer(yes bool) bool {
	ask := C.nui_dialog_kind() == 2
	switch {
	case yes && ask:
		return dialogButton(C.NUI_IDYES)
	case ask:
		return dialogButton(C.NUI_IDNO)
	}
	return dialogButton(C.NUI_IDOK)
}

// SimOK — нажать OK в открытом сообщении.
func SimOK() bool { return dialogButton(C.NUI_IDOK) }

// SimEsc — закрыть открытое сообщение, как Esc (TDF_ALLOW_DIALOG_CANCELLATION).
func SimEsc() bool { return dialogButton(C.NUI_IDCANCEL) }

// SimChooserSet выбирает путь в открытом файловом диалоге или диалоге папки.
func SimChooserSet(path string) bool {
	d := C.nui_find_dialog()
	switch {
	case d == nil:
		return false
	case C.nui_dialog_kind() == 4: // SHBrowseForFolderW: BFFM_SETSELECTIONW
		sendStr(d, 0x400+103, 1, path)
		return true
	case C.nui_dialog_kind() == 3:
		e := fileNameBox(d)
		if e == nil {
			return false
		}
		C.SetWindowTextW(e, wstr(path))
		return true
	}
	return false
}

// SimChooserAdd добавляет файл к выбранным (OpenFiles) — того же каталога, что выбран
// SimChooserSet. Несколько файлов диалог принимает только из текущего каталога в виде
// "имя1" "имя2", поэтому в первый раз — переход в каталог (его путь в поле имени и OK:
// диалог остаётся открытым), а полный путь первого файла становится именем в кавычках.
func SimChooserAdd(path string) bool {
	d := C.nui_find_dialog()
	if d == nil || C.nui_dialog_kind() != 3 {
		return false
	}
	e := fileNameBox(d)
	if e == nil {
		return false
	}
	text := windowText(e)
	if text == "" {
		return false
	}
	if text[0] != '"' {
		dir, name := splitPath(text)
		C.SetWindowTextW(e, wstr(dir))
		C.SendMessageW(d, C.NUI_WM_COMMAND, C.NUI_IDOK, 0)
		text = `"` + name + `"`
	}
	_, name := splitPath(path)
	C.SetWindowTextW(e, wstr(text+` "`+name+`"`))
	return true
}

// fileNameBox — поле имени файла: cmb13 (диалог в стиле проводника) или edt1.
func fileNameBox(d C.HWND) C.HWND {
	if e := C.GetDlgItem(d, 0x47C); e != nil {
		return e
	}
	return C.GetDlgItem(d, 0x480)
}

func splitPath(p string) (dir, name string) {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '\\' || p[i] == '/' {
			return p[:i], p[i+1:]
		}
	}
	return "", p
}

// SimChooserDone — «Открыть/Сохранить/OK» (accept) или «Отмена» в файловом диалоге.
func SimChooserDone(accept bool) bool {
	d := C.nui_find_dialog()
	if d == nil || C.nui_dialog_kind() < 3 {
		return false
	}
	id := C.WPARAM(C.NUI_IDCANCEL)
	if accept {
		id = C.NUI_IDOK
	}
	C.PostMessageW(d, C.NUI_WM_COMMAND, id, 0)
	return true
}

// DialogOpen — открыт ли сейчас диалог: файловый (chooser) или сообщение.
func DialogOpen(chooser bool) bool {
	k := C.nui_dialog_kind()
	return C.nui_find_dialog() != nil && (k >= 3) == chooser
}

// ListCellText — текст ячейки так, как его получает ListView (LVN_GETDISPINFOW).
func ListCellText(view H, row, col int) string {
	buf := make([]uint16, 512)
	n := int(C.nui_list_item_text(nodes[view].hwnd, C.int(row), C.int(col), wptr(buf), C.int(len(buf))))
	return fromUTF16(buf[:n])
}

// SimListSelect — как щелчок по строке: уведомления LVN_ITEMCHANGED идут родителю.
func SimListSelect(view H, row int) { C.nui_list_select(nodes[view].hwnd, C.int(row)) }

// SimListActivate — строка выбрана и нажат Enter (метка nuiEnter, как в цикле).
func SimListActivate(view H, row int) {
	C.nui_list_select(nodes[view].hwnd, C.int(row))
	SimEnter(view)
}

// SimKey — нажатие клавиши в окне win: тот же путь, что у цикла (NUI_WM_KEY, затем
// NUI_WM_NAVKEY), только модификаторы — из mods, а не из GetKeyState. Не забрало окно — символ (WM_CHAR)
// получает виджет с фокусом.
func SimKey(win H, key uint32, mods uint8) {
	n := nodes[win]
	vk := key
	switch {
	case key >= KeyF1 && key < KeyF1+12:
		vk = 0x70 + key - KeyF1
	case key >= KeyEnter:
		vk = [...]uint32{0x0D, 0x1B, 0x09, 0x08, 0x2E, 0x2D, 0x24, 0x23, 0x21, 0x22, 0x25, 0x27, 0x26, 0x28}[key-KeyEnter]
	}
	if C.GetPropW(n.hwnd, wstr("nuiKeys")) != nil &&
		send(msg(), C.NUI_WM_KEY, uintptr(vk)|uintptr(mods)<<16, uintptrOf(n.hwnd)) != 0 {
		return
	}
	// nav_key: Ctrl+Tab, Ctrl+PgUp/PgDn — вкладки; стрелки на варианте переключателя.
	nav, how := false, uintptr(0)
	switch {
	case mods&ModCtrl != 0 && (key == KeyTab || key == KeyPageUp || key == KeyPageDown):
		nav, how = true, uintptr(b2i(key == KeyPageUp || key == KeyTab && mods&ModShift != 0))
	case mods&ModCtrl == 0 && key >= KeyLeft && key <= KeyDown:
		if _, f := nodeOf(C.GetFocus()); f != nil && f.kind == kRadioItem {
			nav, how = true, 2|uintptr(b2i(key == KeyLeft || key == KeyUp))
		}
	}
	if nav && send(msg(), C.NUI_WM_NAVKEY, how, uintptrOf(C.GetFocus())) != 0 {
		return
	}
	ch := key
	switch {
	case key >= 'A' && key <= 'Z' && mods&ModShift == 0:
		ch = key - 'A' + 'a'
	case key == KeyEnter && mods == ModShift: // Shift+Enter: перевод строки в многострочном поле
		ch = '\r'
	}
	if ch < 0x100 && mods&(ModCtrl|ModAlt) == 0 {
		C.PostMessageW(C.GetFocus(), C.NUI_WM_CHAR, C.WPARAM(ch), 0)
	}
}

// SimKeyPosted — клавиша без модификаторов через настоящий цикл: WM_KEYDOWN в очередь
// виджету с фокусом, перехват OnKey — в nui_loop (window_key).
func SimKeyPosted(win H, key uint32) {
	vk := key
	if key == KeyF1+4 {
		vk = 0x74
	}
	C.PostMessageW(C.GetFocus(), C.NUI_WM_KEYDOWN, C.WPARAM(vk), 0)
}

// SimDropFiles — на окно бросили файлы: HDROP, как у оболочки, и WM_DROPFILES в очередь.
func SimDropFiles(win H, id uint32, paths []string) {
	var b []uint16
	for _, p := range paths {
		b = append(b, utf16z(p)...)
	}
	b = append(b, 0)
	C.nui_sim_drop(nodes[win].hwnd, wptr(b), C.int(len(b)))
}

// DefaultButton — кнопка по умолчанию окна; Nil — нет.
func DefaultButton(win H) H { return nodes[win].defBtn }

// SimMenuActivate — пункт меню выбран мышью: строки меню — WM_COMMAND окну, контекстного —
// команда, как её возвращает TrackPopupMenu.
func SimMenuActivate(item H) {
	n := nodes[item]
	if n.ctx {
		menuCommand(n.cmd)
		return
	}
	send(nodes[n.win].hwnd, C.NUI_WM_COMMAND, uintptr(n.cmd), 0)
}

// MenuItemText — текст пункта, как его показывает Windows (с сочетанием после табуляции).
func MenuItemText(item H) string {
	n := nodes[item]
	ref, by := menuRef(n)
	buf := make([]uint16, 256)
	l := int(C.GetMenuStringW(n.hmenu, ref, wptr(buf), C.int(len(buf)), by))
	return fromUTF16(buf[:l])
}

// SimMouse — событие мыши над холстом: настоящее сообщение окну холста (у колеса —
// экранные координаты, как у Windows). x, y — логические пиксели; у MouseWheel button —
// направление (−1 вверх, +1 вниз).
func SimMouse(h H, kind uint8, x, y float64, button int, mods uint8) {
	n := nodes[h]
	s := float64(dpiOf(n).dpi) / 96
	px, py := int(x*s+0.5), int(y*s+0.5)
	var wp uintptr
	if mods&ModShift != 0 {
		wp |= C.NUI_MK_SHIFT
	}
	if mods&ModCtrl != 0 {
		wp |= C.NUI_MK_CONTROL
	}
	var m uint32
	switch kind {
	case MouseDown:
		m = [...]uint32{C.NUI_WM_LBUTTONDOWN, C.NUI_WM_MBUTTONDOWN, C.NUI_WM_RBUTTONDOWN}[button-1]
	case MouseUp:
		m = [...]uint32{C.NUI_WM_LBUTTONUP, C.NUI_WM_MBUTTONUP, C.NUI_WM_RBUTTONUP}[button-1]
	case MouseMove:
		m = C.NUI_WM_MOUSEMOVE
	case MouseWheel:
		m = C.NUI_WM_MOUSEWHEEL
		wp |= uintptr(uint16(int16(-120*button))) << 16
		pt := C.POINT{C.LONG(px), C.LONG(py)}
		C.ClientToScreen(n.hwnd, &pt)
		px, py = int(pt.x), int(pt.y)
	}
	send(n.hwnd, C.UINT(m), wp, uintptr(uint16(px))|uintptr(uint16(py))<<16)
}

// PixelAt — цвет точки (x, y) виджета на экране, 0xRRGGBB.
func PixelAt(h H, x, y int) uint32 {
	n := nodes[h]
	s := float64(dpiOf(n).dpi) / 96
	dc := C.GetDC(n.hwnd)
	c := uint32(C.GetPixel(dc, C.int(float64(x)*s), C.int(float64(y)*s)))
	C.ReleaseDC(n.hwnd, dc)
	return c&0xFF<<16 | c&0xFF00 | c>>16&0xFF
}

// SimSplitDrag — перетаскивание полосы разделителя на delta логических пикселей: нажатие,
// курсор сдвигается (SetCursorPos), движение, отпускание — тот же путь, что у мыши.
func SimSplitDrag(h H, delta int) {
	n := nodes[h]
	var pt C.POINT
	C.GetCursorPos(&pt)
	send(n.hwnd, C.NUI_WM_LBUTTONDOWN, 1, 0)
	x, y, d := C.int(pt.x), C.int(pt.y), C.int(dpiOf(n).px(delta))
	if n.vertical {
		C.SetCursorPos(x, y+d)
	} else {
		C.SetCursorPos(x+d, y)
	}
	send(n.hwnd, C.NUI_WM_MOUSEMOVE, 1, 0)
	send(n.hwnd, C.NUI_WM_LBUTTONUP, 0, 0)
}

// ContextMenuOwner — чьё контекстное меню откроется правым щелчком в точке (x, y)
// клиентской области окна win (логические пиксели); Nil — ничьё. Тот же поиск, что у
// WM_CONTEXTMENU по фону окна.
func ContextMenuOwner(win H, x, y int) H {
	w := nodes[win]
	for h := nodeAt(win, w.px(x), w.px(y)); h != Nil; h = nodes[h].parent {
		if nodes[h].ctxMenu != Nil {
			return h
		}
		if nodes[h].kind == kWindow {
			break
		}
	}
	return Nil
}

// SimRadioSelect — щелчок по варианту i переключателя: BN_CLICKED, как от кнопки.
func SimRadioSelect(h H, i int) { notifyParent(nodes[nodes[h].kids[i]], C.NUI_BN_CLICKED) }

// SimSpinStep — ↑ (up) или ↓ в числовом поле.
func SimSpinStep(h H, up bool) { C.nui_sim_spin(nodes[h].hwnd, C.int(b2i(up))) }

// SimTabsSelect — пользователь щёлкнул ярлык страницы i.
func SimTabsSelect(h H, i int) {
	n := nodes[h]
	for j, k := range n.tabShown {
		if k == n.kids[i] {
			C.nui_tabs_click(n.hwnd, C.int(j))
		}
	}
}

// HwndAt — виджет, которому достанется щелчок в точке (x, y) окна win (логические пиксели
// от левого верхнего угла содержимого): с учётом регионов окон и HTTRANSPARENT. Nil — окно.
func HwndAt(win H, x, y int) H {
	w := nodes[win]
	h, _ := nodeOf(C.nui_hwnd_at(w.hwnd, C.int(w.px(x)), C.int(w.px(y))))
	return h
}
