package drv

// Узлы (driver-win32.md, §2). У Win32 нет контейнеров-окон, поэтому дерево интерфейса
// держит драйвер: Box, Grid и Stretch — виртуальные узлы без HWND, их раскладывает
// internal/layout. HWND контролов — дети окна или прокрутки; пока виджет не вставлен
// в окно, его HWND живёт у скрытого «парковочного» окна.

/*
#include "win32_windows.h"
*/
import "C"

import (
	"unsafe"

	"nui/internal/core"
	"nui/internal/layout"
)

// Виды узлов.
const (
	kWindow = iota + 1
	kLabel
	kButton
	kCheck
	kEntry
	kCombo
	kSlider
	kProgress
	kText
	kSeparator
	kBox
	kSpacer
	kGrid
	kFrame
	kScroll
	kImage
	kList
	kCanvas
	kSplit      // две панели; HWND — полоса между ними (за неё тянут)
	kTabs       // вкладки: SysTabControl32 над страницами
	kRadio      // переключатель: ряд вариантов, как Box
	kRadioItem  // вариант переключателя: BUTTON BS_RADIOBUTTON
	kSpin       // числовое поле: EDIT; ребёнок — стрелки
	kSpinArrows // стрелки поля: msctls_updown32
	kMenu       // HMENU: строка меню или подменю
	kMenuItem   // пункт меню (в hmenu — меню, где он лежит)
)

// Шрифты окна.
const (
	fontNormal = iota
	fontBold
	fontMono
)

type node struct {
	hwnd   C.HWND
	id     uint32
	kind   uint8
	lay    layout.Node
	parent H
	kids   []H
	win    H // окно, в которое вставлен узел; Nil — ещё не вставлен

	hidden     bool // SetVisible(false)
	disabled   bool // SetEnabled(false)
	expandH    bool // SetExpand: ребёнок Grid растягивает свою колонку
	expandV    bool
	minW, minH int // SetMinSize, логические пиксели; -1 — не задано
	padding    int // Box.Padding, Grid.Padding, Window.SetPadding — логические пиксели
	spacing    int // Box
	colGap     int // Grid
	rowGap     int

	text                 string   // исходный текст: Label, Button, Check, Frame, вариант Radio
	tooltip              string   // SetTooltip
	fg, bg               uint32   // SetForeground, SetBackground: 0xRRGGBB; noColor — цвет темы (ADR-19)
	font                 C.HFONT  // свой шрифт подписи (SetFontScale), удаляется вместе с узлом
	fontPct              int      // SetFontScale, проценты; 0 — обычный
	ownerDraw, isDefault bool     // Button: цветная (BS_OWNERDRAW) и кнопка по умолчанию
	items                []string // пункты Combo — для измерения
	measured             bool     // размеры листа посчитаны текущим шрифтом окна

	// Label
	ellipsis, textAlign uint8
	wrap, bold          bool
	selectable          bool   // read-only EDIT вместо STATIC
	color               uint32 // 0xRRGGBB; noColor — цвет темы
	// Entry
	widthChars int
	// Combo, Slider, Progress, Text, Separator
	editable  bool
	vertical  bool    // Separator: вертикальная линия
	last      int     // Slider: последнее значение — EvChange только при изменении
	value     float64 // Progress
	marquee   bool
	mono      bool
	wrapText  bool
	enterText bool   // TextView.OnEnter: метка nuiEnter = 2 (Enter без Shift)
	hint      string // TextView.SetPlaceholder
	scrollX   int    // Scroll: позиция прокрутки, физические пиксели
	scrollY   int
	contentW  int
	contentH  int
	viewW     int
	viewH     int
	minClient [2]int // Window: минимальный размер клиентской области, физические пиксели

	// Image
	img        []byte         // закодированная картинка: из неё перерисовка при смене dpi
	pic        unsafe.Pointer // только что декодированная (между ImageLoad и ImageShow)
	imgW, imgH int            // размер показа, логические пиксели
	bmp        C.HBITMAP      // картинка под dpi окна
	bmpW, bmpH int            // её размер, физические пиксели
	bmpDpi     uint32

	// Меню (Window.SetMenu)
	hmenu     C.HMENU  // kMenu — само меню; kMenuItem — меню, где лежит пункт
	cmd       uint16   // команда пункта (WM_COMMAND); 0 — пункт с подменю
	pos       int      // позиция пункта с подменю в hmenu
	menuBar   H        // окно: строка меню (kMenu)
	menuNodes []H      // окно: все узлы меню — уничтожаются вместе с окном или строкой
	ctxMenu   H        // контекстное меню узла (kMenu); Nil — нет
	ctxNodes  []H      // его узлы — уничтожаются вместе с узлом
	ctx       bool     // узел меню: часть контекстного меню (а не строки меню окна)
	menuKeep  [2]C.int // окно: клиентская область до смены меню (её и сохраняем)

	// Spin
	spinMin, spinMax int

	// Split
	splitPos  int    // размер первой панели по SetPosition или перетаскиванию, логические пиксели; -1 — нет
	splitCur  int    // размер первой панели после раскладки, физические пиксели
	dragStart [2]int // начало перетаскивания: splitCur и координата курсора на экране

	// Tabs
	tabCur   H      // текущая страница; Nil — нет
	tabShown []H    // страницы с ярлыками (видимые) в порядке ярлыков
	tabHole  [6]int // вырезанная область страниц: размер контрола и прямоугольник, физические пиксели
	tabTitle string // у страницы: заголовок её ярлыка

	// List
	columns   []string                  // заголовки колонок
	cell      func(row, col int) string // текст ячейки (ListSetRows)
	rows      int
	colWidths []int // ширина колонок по тексту, физические пиксели

	// Window
	dpi        uint32
	fonts      [3]C.HFONT
	dirty      bool // раскладка уже запрошена (NUI_WM_RELAYOUT в очереди)
	reorder    bool // поменялась структура — восстановить Z-порядок (он же порядок Tab)
	fixed      bool // нерастягиваемое окно
	sized      bool // окно уже получило начальный размер
	growing    bool // окно растёт под минимум содержимого (relayout)
	reqW, reqH int  // последний SetSize, логические пиксели; 0 — не задан
	defBtn     H
	tip        C.HWND // окно подсказок (tooltips_class32), создаётся при первой подсказке
	maximized  bool   // SetMaximized: развернуть при первом показе
	panels     bool   // в окне есть цветные Box или Grid: их фон рисует WM_ERASEBKGND
	fullscreen bool   // SetFullscreen
	savedStyle uintptr
	savedRect  C.RECT // окно до полноэкранного режима
}

const noColor = 1 << 24

// tabsHooks — код вкладок, нужный общему коду узлов: раскладка, фон, удаление, события.
// Задаёт TabsNew: прямые вызовы держали бы его в каждой программе, и без вкладок (TinyGo).
// Кроме key, хуки вызываются только при узле kTabs — значит, после TabsNew.
var tabsHooks struct {
	measure    func(w, n *node)
	hole       func(n *node)
	pane       func() uint32
	detach     func(h, page H) // до удаления page из детей
	sync       func(h H)
	visibility func(h, page H)
	clicked    func(h H)
	key        func(focus C.HWND, back bool) bool
}

// spinHooks — то же для числового поля; задаёт SpinNew.
var spinHooks struct {
	measure func(w, n *node)
	place   func(w, n *node) int // стрелки — к правому краю; ответ — ширина поля
	commit  func(h H)            // принять набранное: Enter, уход фокуса, стрелки
}

// radioHooks — то же для переключателя; задаёт RadioNew.
var radioHooks struct {
	clicked func(item H)                       // BN_CLICKED варианта
	key     func(focus C.HWND, back bool) bool // стрелка на варианте
}

var (
	nodes    = []*node{nil} // H — индекс; 0 — Nil
	freeList []H
)

func newNode(kind uint8, id uint32) (H, *node) {
	n := &node{kind: kind, id: id, minW: -1, minH: -1, color: noColor, fg: noColor, bg: noColor}
	if l := len(freeList); l > 0 {
		h := freeList[l-1]
		freeList = freeList[:l-1]
		nodes[h] = n
		return h, n
	}
	nodes = append(nodes, n)
	return H(len(nodes) - 1), n
}

func get(h H) *node {
	if h == Nil || h >= H(len(nodes)) {
		return nil
	}
	return nodes[h]
}

// nodeOf — узел по HWND (в GWLP_USERDATA лежит H).
func nodeOf(hwnd C.HWND) (H, *node) {
	if hwnd == nil {
		return Nil, nil
	}
	h := H(C.GetWindowLongPtrW(hwnd, C.NUI_GWLP_USERDATA))
	n := get(h)
	if n == nil || n.hwnd != hwnd {
		return Nil, nil
	}
	return h, n
}

// control создаёт HWND узла h у текущего хоста и даёт ему шрифт окна.
func control(h H, cls, text string, style, exstyle uint32) C.HWND {
	n := nodes[h]
	n.hwnd = C.nui_control(wstr(cls), wstr(text), C.DWORD(style|C.NUI_WS_VISIBLE), C.DWORD(exstyle),
		hostOf(h), C.LONG_PTR(h))
	setFont(n)
	return n.hwnd
}

// --- дерево ---

// hostOf — HWND, детьми которого должны быть HWND узла h: ближайшая прокрутка или окно
// выше по дереву; у ещё не вставленного виджета — парковочное окно.
func hostOf(h H) C.HWND {
	for p := nodes[h].parent; p != Nil; p = nodes[p].parent {
		if k := nodes[p].kind; k == kScroll || k == kWindow {
			return nodes[p].hwnd
		}
	}
	return C.nui_parking()
}

// attach делает child ребёнком parent: HWND поддерева переезжают к хосту, узлы узнают
// своё окно и его шрифт, раскладка окна запрашивается заново.
func attach(parent, child H) {
	p, c := nodes[parent], nodes[child]
	c.parent = parent
	p.kids = append(p.kids, child)
	win := p.win
	if p.kind == kWindow {
		win = parent
	}
	adoptTree(child, hostOf(child), win, effectiveDisabled(parent))
	invalidate(parent)
	if w := get(win); w != nil {
		w.reorder = true
	}
}

func adoptTree(h H, host C.HWND, win H, disabled bool) {
	n := nodes[h]
	n.win = win
	n.measured = false
	disabled = disabled || n.disabled
	if n.hwnd != nil {
		C.SetParent(n.hwnd, host)
		setFont(n)
		C.EnableWindow(n.hwnd, b2i(!disabled))
		applyTooltip(h)
		if n.kind == kScroll {
			host = n.hwnd
		}
	}
	for _, k := range n.kids {
		adoptTree(k, host, win, disabled)
	}
}

func effectiveDisabled(h H) bool {
	for ; h != Nil; h = nodes[h].parent {
		if nodes[h].disabled {
			return true
		}
	}
	return false
}

// detach убирает h из детей его родителя.
func detach(h H) {
	n := nodes[h]
	if p := get(n.parent); p != nil {
		if p.kind == kTabs {
			tabsHooks.detach(n.parent, h)
		}
		for i, k := range p.kids {
			if k == h {
				p.kids = append(p.kids[:i:i], p.kids[i+1:]...)
				break
			}
		}
		if p.kind == kTabs {
			tabsHooks.sync(n.parent)
		}
		invalidate(n.parent)
	}
	n.parent = Nil
}

// destroyTree: EvDestroy каждому узлу поддерева — сначала родителю, потом детям, как
// в GTK, — затем HWND и место в таблице. hwndToo = false — HWND уже уничтожает Windows
// (WM_DESTROY окна).
func destroyTree(h H, hwndToo bool) {
	n := get(h)
	if n == nil {
		return
	}
	if n.id != 0 {
		core.Dispatch(core.UD(n.id, core.EvDestroy), 0)
	}
	kids := n.kids
	n.kids = nil
	for _, k := range kids {
		destroyTree(k, hwndToo)
	}
	// Окно не должно ссылаться на уничтоженный узел: хендлы переиспользуются, и новая
	// кнопка с тем же H стала бы кнопкой по умолчанию.
	if w := get(n.win); w != nil && w != n {
		if w.defBtn == h {
			w.defBtn = Nil
		}
		if hwndToo && w.tip != nil && n.hwnd != nil && n.tooltip != "" {
			C.nui_tooltip_set(w.tip, n.hwnd, nil)
		}
	}
	if n.hwnd != nil {
		C.SetWindowLongPtrW(n.hwnd, C.NUI_GWLP_USERDATA, 0)
		if hwndToo {
			C.DestroyWindow(n.hwnd)
		}
		n.hwnd = nil
	}
	if n.font != nil {
		C.DeleteObject(C.HGDIOBJ(n.font))
		n.font = nil
	}
	if n.kind == kImage {
		freeImage(n)
	}
	if n.ctxMenu != Nil {
		dropContextMenu(n)
	}
	if n.kind == kWindow {
		dropMenuNodes(n)
		for i, f := range n.fonts {
			if f != nil {
				C.DeleteObject(C.HGDIOBJ(f))
				n.fonts[i] = nil
			}
		}
	}
	nodes[h] = nil
	freeList = append(freeList, h)
}

// replaceChild — содержимое окна, Frame, Scroll: старый ребёнок уничтожается.
func replaceChild(h, child H) {
	n := nodes[h]
	for len(n.kids) > 0 {
		old := n.kids[0]
		detach(old)
		destroyTree(old, true)
	}
	if child != Nil {
		attach(h, child)
	}
	invalidate(h)
}

// --- раскладка ---

// invalidate: раскладку окна пересчитать один раз на пачку изменений (layout.md, §7).
func invalidate(h H) {
	n := get(h)
	if n == nil {
		return
	}
	n.measured = false
	w := n
	if n.kind != kWindow {
		w = get(n.win)
	}
	if w == nil || w.dirty {
		return
	}
	w.dirty = true
	C.PostMessageW(w.hwnd, C.NUI_WM_RELAYOUT, 0, 0)
}

// remeasure — у листа поменялся текст или шрифт.
func remeasure(h H) { invalidate(h) }

// px — логические пиксели в физические при dpi окна.
func (w *node) px(v int) int { return (v*int(w.dpi) + 48) / 96 }

// unpx — физические пиксели в логические при dpi окна.
func (w *node) unpx(v int) int { return (v*96 + int(w.dpi)/2) / int(w.dpi) }

// relayout считает раскладку окна и расставляет HWND (layout.md, §7).
func relayout(wh H) {
	w := get(wh)
	if w == nil {
		return
	}
	w.dirty = false
	root := get(firstKid(w))
	p := w.px(w.padding)
	if root == nil {
		w.minClient = [2]int{2 * p, 2 * p}
		return
	}
	prepare(w, firstKid(w))
	layout.Measure(&root.lay)
	w.minClient = [2]int{root.lay.Min.W + 2*p, root.lay.Min.H + 2*p}
	var cs [2]C.int
	C.nui_client_size(w.hwnd, &cs[0])
	// Как в GTK: показанное окно не меньше содержимого — растёт вслед за его минимумом
	// (новые виджеты, SetMinSize, Radio в ряд); место на экране прежнее. WM_SIZE от
	// SetWindowPos раскладывает окно внутри (growing — без повторного роста).
	if gw, gh := max(int(cs[0]), w.minClient[0]), max(int(cs[1]), w.minClient[1]); w.sized && !w.growing &&
		!w.fullscreen && C.IsZoomed(w.hwnd) == 0 && (gw > int(cs[0]) || gh > int(cs[1])) {
		var r [4]C.int
		C.nui_window_frame(w.hwnd, C.int(gw), C.int(gh), 0, &r[0])
		w.growing = true
		C.SetWindowPos(w.hwnd, nil, r[0], r[1], r[2], r[3], C.NUI_SWP_NOZORDER|C.NUI_SWP_NOACTIVATE)
		w.growing = false
		C.nui_client_size(w.hwnd, &cs[0])
	}
	layout.Arrange(&root.lay, layout.Rect{X: p, Y: p, W: max(1, int(cs[0])-2*p), H: max(1, int(cs[1])-2*p)})
	place(w, firstKid(w), 0, 0, false)
	if w.panels || w.bg != noColor { // цветные зоны переехали — перерисовать фон
		C.RedrawWindow(w.hwnd, nil, nil, C.NUI_RDW_INVALIDATE|C.NUI_RDW_ERASE|C.NUI_RDW_ALLCHILDREN)
	}
	if w.reorder {
		w.reorder = false
		var prev C.HWND
		zorder(firstKid(w), &prev)
	}
}

func firstKid(n *node) H {
	if len(n.kids) == 0 {
		return Nil
	}
	return n.kids[0]
}

// prepare заполняет параметры layout.Node поддерева: размеры листьев (при dpi окна),
// отступы контейнеров, видимость, Floor из SetMinSize.
func prepare(w *node, h H) {
	n := nodes[h]
	l := &n.lay
	l.Hidden = n.hidden
	l.Floor = layout.Size{}
	if n.minW >= 0 {
		l.Floor.W = w.px(n.minW)
	}
	if n.minH >= 0 {
		l.Floor.H = w.px(n.minH)
	}
	l.Children = l.Children[:0]
	for _, k := range n.kids {
		prepare(w, k)
		l.Children = append(l.Children, &nodes[k].lay)
	}
	if n.bg != noColor && (n.kind == kBox || n.kind == kGrid) {
		w.panels = true
	}
	switch n.kind {
	case kBox:
		l.Spacing, l.Padding = w.px(n.spacing), w.px(n.padding)
	case kGrid:
		l.Kind = layout.Grid
		l.ColGap, l.RowGap, l.Padding = w.px(n.colGap), w.px(n.rowGap), w.px(n.padding)
		l.ExpandCols, l.ExpandRows = l.ExpandCols[:0], l.ExpandRows[:0]
		for _, k := range n.kids {
			c := nodes[k]
			if c.expandH && c.lay.Cols <= 1 {
				l.ExpandCols = setFlag(l.ExpandCols, c.lay.Col)
			}
			if c.expandV && c.lay.Rows <= 1 {
				l.ExpandRows = setFlag(l.ExpandRows, c.lay.Row)
			}
		}
	case kFrame:
		l.Kind = layout.Frame
		measureFrame(w, n)
	case kScroll:
		l.Kind = layout.Scroll
		l.Floor.W, l.Floor.H = max(l.Floor.W, w.px(46)), max(l.Floor.H, w.px(46))
	case kSpacer:
		l.Min, l.Nat, l.Expand = layout.Size{}, layout.Size{}, true
	case kSplit: // как ряд с промежутком-полосой; панели расставляет placeSplit
		l.Spacing, l.Padding = w.px(splitGap), 0
	case kRadio:
		l.Spacing = w.px(n.spacing)
	case kTabs:
		tabsHooks.measure(w, n)
		w.panels = true // область страниц рисует хост (eraseBackground)
	default:
		if !n.measured {
			measureLeaf(w, n)
			n.measured = true
		}
	}
}

func setFlag(f []bool, i int) []bool {
	for len(f) <= i {
		f = append(f, false)
	}
	f[i] = true
	return f
}

// place расставляет HWND по результату Arrange. ox, oy — начало клиентской области
// хоста (окна или прокрутки) в координатах окна; hide — предок скрыт.
func place(w *node, h H, ox, oy int, hide bool) {
	n := nodes[h]
	hide = hide || n.hidden
	if n.kind == kSplit {
		placeSplit(w, h, ox, oy, hide)
		return
	}
	o := n.lay.Out
	if n.hwnd != nil {
		x, y, cw, ch := o.X-ox, o.Y-oy, o.W, o.H
		if n.kind == kSpin {
			cw = spinHooks.place(w, n)
		}
		if n.kind == kCombo {
			// Высоту поля Windows задаёт сам; в ячейке — по центру, ниже — выпадающий список.
			fh := n.lay.Nat.H
			y += max(0, (ch-fh)/2)
			ch = fh + w.px(200)
		}
		C.SetWindowPos(n.hwnd, nil, C.int(x), C.int(y), C.int(max(0, cw)), C.int(max(0, ch)),
			C.NUI_SWP_NOZORDER|C.NUI_SWP_NOACTIVATE)
		show := C.int(C.NUI_SW_SHOWNA)
		if hide {
			show = C.NUI_SW_HIDE
		}
		C.ShowWindow(n.hwnd, show)
		if n.kind == kList {
			listFillLast(n)
		}
		if n.kind == kTabs {
			tabsHooks.hole(n)
		}
	}
	if n.kind == kScroll && len(n.kids) > 0 {
		placeScroll(w, h, hide)
		return
	}
	for _, k := range n.kids {
		place(w, k, ox, oy, hide || n.kind == kTabs && k != n.tabCur)
	}
}

// zorder — Z-порядок HWND как порядок обхода дерева: он же порядок Tab (IsDialogMessage
// идёт по Z-порядку), а рамка Frame оказывается над своими детьми, как в шаблонах диалогов.
func zorder(h H, prev *C.HWND) {
	n := nodes[h]
	if n.hwnd != nil {
		after := *prev // nil — HWND_TOP
		C.SetWindowPos(n.hwnd, after, 0, 0, 0, 0, C.NUI_SWP_NOMOVE|C.NUI_SWP_NOSIZE|C.NUI_SWP_NOACTIVATE)
		*prev = n.hwnd
		if n.kind == kScroll {
			var inner C.HWND
			for _, k := range n.kids {
				zorder(k, &inner)
			}
			return
		}
	}
	for _, k := range n.kids {
		zorder(k, prev)
	}
}

// --- шрифты и измерение ---

// fontsFor — шрифты окна узла; у ещё не вставленного — шрифты при системном dpi.
func fontsFor(n *node) *[3]C.HFONT {
	if w := get(n.win); w != nil {
		return &w.fonts
	}
	return &defFonts
}

func fontOf(n *node) C.HFONT {
	if n.font != nil {
		return n.font
	}
	f := fontsFor(n)
	switch {
	case n.kind == kLabel && n.bold:
		return f[fontBold]
	case n.kind == kText && n.mono:
		return f[fontMono]
	}
	return f[fontNormal]
}

// setFont даёт контролу шрифт окна; подписи с SetFontScale — свой, из шрифта окна
// (он же пересоздаётся при смене dpi и жирности). Картинка — перерисовывается под dpi.
func setFont(n *node) {
	if n.hwnd == nil {
		return
	}
	if n.kind == kImage {
		if n.bmp != nil && n.bmpDpi != dpiOf(n).dpi {
			renderImage(n)
		}
		return
	}
	old := n.font
	n.font = nil
	if n.kind == kLabel && n.fontPct > 0 && n.fontPct != 100 {
		n.font = C.nui_font_scaled(fontsFor(n)[fontNormal], C.int(n.fontPct), b2i(n.bold))
	}
	send(n.hwnd, C.NUI_WM_SETFONT, uintptr(unsafe.Pointer(fontOf(n))), 1)
	if n.kind == kList { // ширина колонок — по тексту этим шрифтом
		listAutosize(n)
	}
	if n.kind == kText { // поля внутри текста, как у TextView в GTK; WM_SETFONT их сбрасывает
		m := uintptr(dpiOf(n).px(6))
		send(n.hwnd, C.NUI_EM_SETMARGINS, C.NUI_EC_LEFTMARGIN|C.NUI_EC_RIGHTMARGIN, m|m<<16)
	}
	if old != nil {
		C.DeleteObject(C.HGDIOBJ(old))
	}
}

func dpiOf(n *node) *node {
	if w := get(n.win); w != nil {
		return w
	}
	return &defWin
}

// textSize — размер текста шрифтом узла; wrap > 0 — с переносом при этой ширине.
func textSize(n *node, s string, wrap int) (int, int) {
	b := utf16z(s)
	var out [2]C.int
	C.nui_text_size(n.hwnd, fontOf(n), wptr(b), C.int(len(b)-1), C.int(wrap), &out[0])
	return int(out[0]), int(out[1])
}

// charMetrics — средняя ширина символа и высота строки шрифта узла.
func charMetrics(n *node) (int, int) {
	var out [2]C.int
	C.nui_metrics(n.hwnd, fontOf(n), &out[0])
	return int(out[0]), int(out[1])
}

// measureLeaf — минимальный и естественный размер контрола (driver-win32.md, §6),
// в физических пикселях при dpi окна w.
func measureLeaf(w *node, n *node) {
	l := &n.lay
	set := func(minW, natW, h int) { l.Min, l.Nat = layout.Size{W: minW, H: h}, layout.Size{W: natW, H: h} }
	avg, lh := charMetrics(n)
	field := lh + w.px(8) // высота однострочного поля: строка + рамка + поля
	switch n.kind {
	case kLabel:
		tw, th := textSize(n, n.text, 0)
		switch {
		case n.wrap:
			ww := min(tw, 40*avg)
			_, wh := textSize(n, n.text, max(ww, 1))
			set(ww, ww, wh)
		case n.ellipsis != 0:
			ew, _ := textSize(n, "…", 0)
			set(min(ew, tw), tw, th)
		default:
			set(tw, tw, th)
		}
		if n.selectable { // поля EDIT по бокам
			l.Min.W += w.px(4)
			l.Nat.W += w.px(4)
		}
	case kButton, kCheck, kRadioItem:
		var out [2]C.int
		C.nui_ideal_size(n.hwnd, &out[0])
		bw, bh := int(out[0]), int(out[1])
		if bw <= 0 || bh <= 0 { // нет BCM_GETIDEALSIZE (comctl32 v5)
			tw, th := textSize(n, amp(n.text), 0)
			bw, bh = tw+w.px(16), max(th+w.px(10), field)
			if n.kind != kButton {
				bw, bh = tw+w.px(20), max(th, w.px(13))
			}
		}
		if n.kind == kButton {
			bh = max(bh, field)
			set(bw, max(bw, 75*avg/4), bh) // 75 DLU — стандартная ширина кнопки Windows
		} else {
			set(bw, bw, bh)
		}
	case kEntry:
		chars := n.widthChars
		if chars <= 0 {
			chars = 20
		}
		set(3*avg+w.px(8), chars*avg+w.px(8), field)
	case kCombo:
		iw := 0
		for _, s := range n.items {
			tw, _ := textSize(n, s, 0)
			iw = max(iw, tw)
		}
		arrow := int(C.GetSystemMetricsForDpi(C.NUI_SM_CXVSCROLL, C.UINT(w.dpi)))
		set(3*avg+arrow+w.px(8), max(iw, 3*avg)+arrow+w.px(12), field)
	case kSlider:
		set(w.px(50), w.px(150), w.px(26))
	case kProgress:
		set(w.px(50), w.px(150), w.px(15))
	case kText:
		set(w.px(46), w.px(46), w.px(46)) // как TextView в GTK: размер задают SetMinSize или Expand
	case kSpin:
		spinHooks.measure(w, n)
	case kCanvas:
		set(0, 0, 0) // своего размера нет: SetMinSize или Expand
	case kList:
		set(w.px(46), w.px(46), w.px(71)) // как GtkTreeView в прокрутке (Adwaita), с заголовками и без
	case kSeparator:
		if n.vertical {
			set(2, 2, 1)
		} else {
			set(1, 1, 2)
		}
	case kImage:
		set(n.bmpW, n.bmpW, n.bmpH)
	}
}

// measureFrame — оформление рамки: сверху заголовок, по краям поля (driver-win32.md, §6).
func measureFrame(w *node, n *node) {
	_, lh := charMetrics(n)
	side := w.px(9)
	n.lay.Inset = layout.Insets{L: side, T: lh + w.px(6), R: side, B: side}
	if n.text != "" {
		tw, _ := textSize(n, amp(n.text), 0)
		n.lay.Floor.W = max(n.lay.Floor.W, tw+2*side+w.px(8))
	}
}

// --- прокрутка ---

// placeScroll: полосы Win32 занимают место — решаем, нужны ли они, и заново раскладываем
// ребёнка в оставшейся области (layout.md, §7); дети прокрутки — относительно её
// клиентской области со сдвигом на позицию прокрутки.
func placeScroll(w *node, h H, hide bool) {
	n := nodes[h]
	c := nodes[n.kids[0]]
	o := n.lay.Out
	sbW := int(C.GetSystemMetricsForDpi(C.NUI_SM_CXVSCROLL, C.UINT(w.dpi)))
	sbH := int(C.GetSystemMetricsForDpi(C.NUI_SM_CYHSCROLL, C.UINT(w.dpi)))
	vw, vh := o.W, o.H
	needH, needV := c.lay.Min.W > vw, c.lay.Min.H > vh
	if needV {
		vw -= sbW
	}
	if needH {
		vh -= sbH
	}
	if !needH && c.lay.Min.W > vw {
		needH = true
		vh -= sbH
	}
	if !needV && c.lay.Min.H > vh {
		vw -= sbW
	}
	vw, vh = max(vw, 0), max(vh, 0)
	layout.Arrange(&c.lay, layout.Rect{X: o.X, Y: o.Y, W: max(vw, c.lay.Min.W), H: max(vh, c.lay.Min.H)})
	n.viewW, n.viewH = vw, vh
	n.contentW, n.contentH = c.lay.Out.W, c.lay.Out.H
	n.scrollX = clamp(n.scrollX, 0, max(0, n.contentW-vw))
	n.scrollY = clamp(n.scrollY, 0, max(0, n.contentH-vh))
	setBar(n.hwnd, C.NUI_SB_HORZ, n.contentW, vw, n.scrollX)
	setBar(n.hwnd, C.NUI_SB_VERT, n.contentH, vh, n.scrollY)
	place(w, n.kids[0], o.X+n.scrollX, o.Y+n.scrollY, hide)
}

func setBar(hwnd C.HWND, bar C.int, content, view, pos int) {
	si := C.SCROLLINFO{fMask: C.NUI_SIF_RANGE | C.NUI_SIF_PAGE | C.NUI_SIF_POS, nMax: C.int(max(0, content-1)),
		nPage: C.UINT(max(0, view)), nPos: C.int(pos)}
	si.cbSize = C.UINT(unsafe.Sizeof(si))
	C.SetScrollInfo(hwnd, bar, &si, 1)
}

// scrollBy — прокрутка полосой или колесом: новая позиция и перестановка детей.
func scrollBy(h H, vertical bool, code, thumb int, wheel int) {
	n := nodes[h]
	w := get(n.win)
	if w == nil || len(n.kids) == 0 {
		return
	}
	pos, content, view := &n.scrollX, n.contentW, n.viewW
	if vertical {
		pos, content, view = &n.scrollY, n.contentH, n.viewH
	}
	line := w.px(16)
	switch {
	case wheel != 0:
		*pos -= wheel * 3 * line / 120
	case code == C.NUI_SB_LINEUP:
		*pos -= line
	case code == C.NUI_SB_LINEDOWN:
		*pos += line
	case code == C.NUI_SB_PAGEUP:
		*pos -= view
	case code == C.NUI_SB_PAGEDOWN:
		*pos += view
	case code == C.NUI_SB_THUMBTRACK:
		*pos = thumb
	case code == C.NUI_SB_TOP:
		*pos = 0
	case code == C.NUI_SB_BOTTOM:
		*pos = content
	default:
		return
	}
	placeScroll(w, h, effectiveHidden(h))
}

func effectiveHidden(h H) bool {
	for ; h != Nil; h = nodes[h].parent {
		if nodes[h].hidden || offPage(h) {
			return true
		}
	}
	return false
}

func clamp(v, lo, hi int) int { return max(lo, min(v, hi)) }

// --- подсказки ---

// tooltipOf — подсказка узла или ближайшего предка: как в GTK, подсказка контейнера
// видна над его детьми, у которых своей нет.
func tooltipOf(h H) string {
	for ; h != Nil; h = nodes[h].parent {
		if t := nodes[h].tooltip; t != "" {
			return t
		}
	}
	return ""
}

// applyTooltip регистрирует подсказку контрола в окне подсказок его окна. Не вставленный
// в окно виджет регистрируется при вставке (adoptTree).
func applyTooltip(h H) {
	n := nodes[h]
	w := get(n.win)
	if n.hwnd == nil || w == nil {
		return
	}
	t := tooltipOf(h)
	if t == "" && w.tip == nil {
		return
	}
	if w.tip == nil {
		w.tip = C.nui_tooltip_new(w.hwnd)
	}
	if t == "" {
		C.nui_tooltip_set(w.tip, n.hwnd, nil)
		return
	}
	C.nui_tooltip_set(w.tip, n.hwnd, wstr(t))
}

func applyTooltipTree(h H) {
	applyTooltip(h)
	for _, k := range nodes[h].kids {
		applyTooltipTree(k)
	}
}

// recreate пересоздаёт HWND узла с другим классом или стилем (стиль ES_MULTILINE и
// замена STATIC на EDIT на лету не меняются): шрифт, доступность, подсказка, место в
// Z-порядке и раскладке — как были.
func recreate(h H, cls, text string, style, exstyle uint32) {
	n := nodes[h]
	if w := get(n.win); w != nil && w.tip != nil {
		C.nui_tooltip_set(w.tip, n.hwnd, nil)
	}
	C.SetWindowLongPtrW(n.hwnd, C.NUI_GWLP_USERDATA, 0)
	C.DestroyWindow(n.hwnd)
	control(h, cls, text, style, exstyle)
	C.EnableWindow(n.hwnd, b2i(!effectiveDisabled(h)))
	applyTooltip(h)
	if w := get(n.win); w != nil {
		w.reorder = true
	}
	invalidate(h)
}

// --- цвета (ADR-19) ---

// bgr — 0xRRGGBB в COLORREF (0x00BBGGRR).
func bgr(c uint32) C.COLORREF { return C.COLORREF(c>>16&0xFF | c&0xFF00 | (c&0xFF)<<16) }

func colorref(c uint32) C.uint {
	if c == noColor {
		return ^C.uint(0)
	}
	return C.uint(bgr(c))
}

var brushes = map[uint32]C.HBRUSH{} // кисти цветов фона: их немного, живут до конца программы

func brushFor(c uint32) C.HBRUSH {
	b := brushes[c]
	if b == nil {
		b = C.CreateSolidBrush(bgr(c))
		brushes[c] = b
	}
	return b
}

// inheritedBg — фон ближайшей цветной панели (Box, Grid), области страниц вкладок или
// окна выше узла.
func inheritedBg(n *node) uint32 {
	for p := get(n.parent); p != nil; p = get(p.parent) {
		if p.bg != noColor && (p.kind == kBox || p.kind == kGrid || p.kind == kWindow) {
			return p.bg
		}
		if p.kind == kTabs {
			return tabsHooks.pane()
		}
	}
	if w := get(n.win); w != nil {
		return w.bg
	}
	return noColor
}

// ctlColor — ответ на WM_CTLCOLOR*: цвет текста и фон контрола. Подписи, флажки,
// ползунки, рамки берут фон панели; поля ввода — только свой (как в GTK: поле белое
// и на цветной панели), а поле только для чтения тоже белое.
func ctlColor(c *node, dc C.HDC, m uint32) (uintptr, bool) {
	fg, bg := c.fg, c.bg
	if c.kind == kLabel && c.color != noColor {
		fg = c.color
	}
	field := c.kind == kEntry || c.kind == kText || c.kind == kSpin
	if bg == noColor && !field {
		bg = inheritedBg(c)
	}
	if fg != noColor {
		C.SetTextColor(dc, bgr(fg))
	}
	if bg != noColor {
		C.SetBkColor(dc, bgr(bg))
		return uintptr(unsafe.Pointer(brushFor(bg))), true
	}
	if fg == noColor && !(field && m == C.NUI_WM_CTLCOLORSTATIC) {
		return 0, false
	}
	sys := C.int(C.NUI_COLOR_BTNFACE)
	if field {
		sys = C.NUI_COLOR_WINDOW
	}
	C.SetBkColor(dc, C.COLORREF(C.GetSysColor(sys)))
	return uintptr(unsafe.Pointer(C.GetSysColorBrush(sys))), true
}

// eraseBackground — фон окна или прокрутки: цвет окна, поверх — цветные панели этого
// хоста (вместе с их отступом), родители раньше детей. false — фон по умолчанию.
func eraseBackground(n *node, dc C.HDC) bool {
	w := n
	if n.kind != kWindow {
		w = get(n.win)
	}
	if w == nil || !w.panels && w.bg == noColor && n.bg == noColor {
		return false
	}
	base := w.bg
	if n.kind == kScroll {
		base = inheritedBg(n)
	}
	var cs [2]C.int
	C.nui_client_size(n.hwnd, &cs[0])
	if base == noColor {
		C.nui_fill(dc, 0, 0, cs[0], cs[1], C.COLORREF(C.GetSysColor(C.NUI_COLOR_BTNFACE)))
	} else {
		C.nui_fill(dc, 0, 0, cs[0], cs[1], bgr(base))
	}
	ox, oy := 0, 0
	if n.kind == kScroll {
		ox, oy = n.lay.Out.X+n.scrollX, n.lay.Out.Y+n.scrollY
	}
	for _, k := range n.kids {
		paintPanels(w, k, dc, ox, oy)
	}
	return true
}

func paintPanels(w *node, h H, dc C.HDC, ox, oy int) {
	n := nodes[h]
	if n.hidden || n.kind == kScroll { // прокрутка рисует своё сама
		return
	}
	if n.bg != noColor && (n.kind == kBox || n.kind == kGrid) {
		p, o := w.px(n.padding), n.lay.Out
		C.nui_fill(dc, C.int(o.X-p-ox), C.int(o.Y-p-oy), C.int(o.W+2*p), C.int(o.H+2*p), bgr(n.bg))
	}
	if n.kind == kTabs { // область страниц (вырезана из контрола) и панели текущей страницы
		o, in := n.lay.Out, n.lay.Inset
		C.nui_fill(dc, C.int(o.X+in.L-ox), C.int(o.Y+in.T-oy), C.int(o.W-in.L-in.R), C.int(o.H-in.T-in.B),
			bgr(tabsHooks.pane()))
		if n.tabCur != Nil {
			paintPanels(w, n.tabCur, dc, ox, oy)
		}
		return
	}
	for _, k := range n.kids {
		paintPanels(w, k, dc, ox, oy)
	}
}
