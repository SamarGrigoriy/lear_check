package drv

/*
#include "win32_windows.h"
*/
import "C"

import (
	"sync/atomic"
	"unsafe"

	"nui/internal/core"
)

var (
	mainTID  = uint32(C.GetCurrentThreadId()) // init пакета идёт в главном потоке (runtime.main)
	msgWnd   atomic.Uintptr                   // служебное окно: Do, таймеры, Enter
	appTitle string                           // заголовок сообщений: SetAppID или имя exe
	defFonts [3]C.HFONT                       // шрифты ещё не вставленных виджетов
	defWin   = node{dpi: 96}                  // их dpi — системный
)

// Init: Per-Monitor DPI v2, Common Controls v6, класс окон, служебное окно (driver-win32.md, §3).
func Init(appID string) bool {
	appTitle = appID
	if C.nui_init() == 0 {
		return false
	}
	msgWnd.Store(uintptr(unsafe.Pointer(C.nui_msg_window())))
	defWin.dpi = uint32(C.GetDpiForSystem())
	C.nui_fonts(C.UINT(defWin.dpi), &defFonts[0])
	if appTitle == "" {
		appTitle = exeName()
	}
	return true
}

func Main() { C.nui_loop() }
func Quit() { C.PostQuitMessage(0) }

// Wake — из любого потока: PostMessageW потокобезопасна.
func Wake() {
	if w := msgWnd.Load(); w != 0 {
		C.PostMessageW(hwndOf(w), C.NUI_WM_DO, 0, 0)
	}
}

func IsMainThread() bool { return uint32(C.GetCurrentThreadId()) == mainTID }

// TimerAdd: nIDEvent = ud целиком, WM_TIMER сразу даёт ud (driver-win32.md, §3).
func TimerAdd(ms int, ud uintptr) uint32 {
	C.SetTimer(msg(), C.UINT_PTR(ud), C.UINT(max(ms, 0)), nil)
	return uint32(ud)
}

func SourceRemove(src uint32) { C.KillTimer(msg(), C.UINT_PTR(src)) }

func msg() C.HWND { return hwndOf(msgWnd.Load()) }

// exeName — имя exe без пути и расширения: заголовок сообщений по умолчанию.
func exeName() string {
	buf := make([]uint16, 260)
	n := int(C.GetModuleFileNameW(nil, wptr(buf), C.DWORD(len(buf))))
	s := buf[:n]
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\\' || s[i] == '/' {
			s = s[i+1:]
			break
		}
	}
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			s = s[:i]
			break
		}
	}
	return fromUTF16(s)
}

// wndProc — все сообщения окон nui (nuiWndProc). false — отдать DefWindowProcW.
func wndProc(hwnd C.HWND, m uint32, wp, lp uintptr) (uintptr, bool) {
	switch m {
	case C.NUI_WM_DO:
		core.Dispatch(core.UD(0, core.KDo), 0)
		return 0, true
	case C.NUI_WM_TIMER:
		if hwnd == msg() {
			if core.Dispatch(wp, 0) == 0 {
				C.KillTimer(hwnd, C.UINT_PTR(wp))
			}
			return 0, true
		}
	case C.NUI_WM_ENTER:
		if h, n := nodeOf(hwndOf(lp)); n != nil && n.kind == kSpin {
			spinHooks.commit(h) // Enter идёт дальше: кнопка по умолчанию — с новым значением
		} else if n != nil {
			emit(n, core.EvEnter)
		}
		return 0, true
	case C.NUI_WM_KEY: // OnKey окна: ответ 1 — клавиша обработана
		if _, n := nodeOf(hwndOf(lp)); n != nil && n.kind == kWindow {
			return core.Dispatch(core.UD(n.id, core.EvKey), wp), true
		}
		return 0, true
	case C.NUI_WM_NAVKEY: // nav_key: Ctrl+Tab… во вкладках, стрелки на переключателе; 1 — обработано
		f, back := hwndOf(lp), wp&1 != 0
		if wp&2 != 0 {
			return uintptr(b2i(radioHooks.key != nil && radioHooks.key(f, back))), true
		}
		return uintptr(b2i(tabsHooks.key != nil && tabsHooks.key(f, back))), true
	case C.NUI_WM_LISTSEL: // выделение в списке снято — сообщить, если новое не появилось
		if _, n := nodeOf(hwndOf(lp)); n != nil {
			emit(n, core.EvChange)
		}
		return 0, true
	}
	// Уведомления контролов приходят их родителю — окну, прокрутке или парковочному окну
	// (у него нет узла): разбираются по HWND контрола в lParam.
	switch m {
	case C.NUI_WM_COMMAND:
		if lp != 0 {
			notify(hwndOf(lp), uint32(wp>>16))
			return 0, true
		}
		if wp>>16 == 0 && uint16(wp) >= firstMenuCmd && menuCommand(uint16(wp)) { // меню окна
			return 0, true
		}
	case C.NUI_WM_HSCROLL, C.NUI_WM_VSCROLL:
		if lp != 0 { // от трекбара
			if _, c := nodeOf(hwndOf(lp)); c != nil && c.kind == kSlider {
				if v := int(send(c.hwnd, C.NUI_TBM_GETPOS, 0, 0)); v != c.last {
					c.last = v
					emit(c, core.EvChange)
				}
			} else if c != nil && c.kind == kSpinArrows { // шаг стрелками: текст поля уже новый
				spinHooks.commit(c.parent)
			}
			return 0, true
		}
	case C.NUI_WM_CTLCOLORSTATIC, C.NUI_WM_CTLCOLOREDIT, C.NUI_WM_CTLCOLORBTN:
		if _, c := nodeOf(hwndOf(lp)); c != nil {
			if r, ok := ctlColor(c, hdcOf(wp), m); ok {
				return r, true
			}
		}
	case C.NUI_WM_CONTEXTMENU: // правый щелчок, клавиша «Меню», Shift+F10
		if contextMenu(hwndOf(wp), lp) {
			return 0, true
		}
	case C.NUI_WM_NOTIFY:
		nm := nmhdrOf(lp)
		h, c := nodeOf(nm.hwndFrom)
		if c != nil && c.kind == kList {
			return listNotify(c, lp), true
		}
		if c != nil && c.kind == kTabs && uint32(nm.code) == tcnSelChange {
			tabsHooks.clicked(h)
			return 0, true
		}
	case C.NUI_WM_DRAWITEM: // цветная кнопка (BS_OWNERDRAW) и картинка (SS_OWNERDRAW)
		d := drawItemOf(lp)
		if _, c := nodeOf(d.hwndItem); c != nil && c.kind == kButton {
			C.nui_draw_button(d, colorref(c.fg), colorref(c.bg), b2i(c.isDefault))
			return 1, true
		} else if c != nil && c.kind == kImage {
			drawImage(c, d)
			return 1, true
		}
	}
	h, n := nodeOf(hwnd)
	if n == nil {
		return 0, false
	}
	switch m {
	case C.NUI_WM_HSCROLL, C.NUI_WM_VSCROLL:
		if n.kind == kScroll { // свои полосы прокрутки
			scrollBy(h, m == C.NUI_WM_VSCROLL, int(wp&0xFFFF), int(int16(wp>>16)), 0)
			return 0, true
		}
	case C.NUI_WM_ERASEBKGND:
		if n.kind == kCanvas {
			return 1, true // фон рисует WM_PAINT холста вместе с рисунком
		}
		if (n.kind == kWindow || n.kind == kScroll) && eraseBackground(n, hdcOf(wp)) {
			return 1, true
		}
	case C.NUI_WM_MOUSEWHEEL:
		if n.kind == kScroll {
			scrollBy(h, true, 0, 0, int(int16(wp>>16)))
			return 0, true
		}
	}
	if n.kind == kWindow {
		return windowProc(h, n, m, wp, lp)
	}
	if n.kind == kCanvas {
		return canvasProc(n, m, wp, lp)
	}
	if n.kind == kSplit {
		return splitProc(h, n, m, wp, lp)
	}
	return 0, false
}

// notify — WM_COMMAND от контрола (driver-win32.md, §4).
func notify(ctl C.HWND, code uint32) {
	h, c := nodeOf(ctl)
	if c == nil {
		return
	}
	switch {
	case c.kind == kButton && code == C.NUI_BN_CLICKED:
		emit(c, core.EvClick)
	case c.kind == kCheck && code == C.NUI_BN_CLICKED:
		emit(c, core.EvChange)
	case c.kind == kRadioItem && code == C.NUI_BN_CLICKED: // и щелчок, и фокус на невыбранном
		radioHooks.clicked(h)
	case (c.kind == kEntry || c.kind == kText) && code == C.NUI_EN_CHANGE:
		emit(c, core.EvChange)
	case c.kind == kSpin && code == C.NUI_EN_KILLFOCUS:
		spinHooks.commit(h)
	case c.kind == kCombo && (code == C.NUI_CBN_SELCHANGE || code == C.NUI_CBN_EDITCHANGE):
		emit(c, core.EvChange)
	}
}
