package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Window — окно верхнего уровня. Не Widget: в контейнер его не вставить.
type Window struct {
	h       drv.H
	id      uint32
	onClose func() bool
	onKey   func(k Key, m Mod) bool
	onDrop  func(paths []string)
}

// NewWindow создаёт окно — скрытым, с отступом 12, по центру экрана (на Wayland
// положение выбирает композитор), с размером по содержимому. child может быть nil.
func NewWindow(title string, child Widget) *Window {
	checkCreate()
	w := &Window{}
	w.id = core.Register((*windowEv)(w))
	w.h = drv.WindowNew(w.id, title)
	wins = append(wins, w)
	if child != nil {
		w.SetChild(child)
	}
	return w
}

func (w *Window) dead() bool {
	core.CheckThread()
	if w.id != 0 {
		return false
	}
	if core.Debug {
		println("nui: вызов метода у уничтоженного окна")
	}
	return true
}

// Alive — false после уничтожения окна.
func (w *Window) Alive() bool { return w.id != 0 }

// SetTitle задаёт заголовок окна.
func (w *Window) SetTitle(s string) {
	if !w.dead() {
		drv.WindowSetTitle(w.h, s)
	}
}

// Title — заголовок окна.
func (w *Window) Title() string {
	if w.dead() {
		return ""
	}
	return drv.WindowTitle(w.h)
}

// SetChild заменяет содержимое окна; старый ребёнок уничтожается. nil — только убрать.
func (w *Window) SetChild(c Widget) {
	if w.dead() {
		return
	}
	ch := drv.Nil
	if c != nil {
		b := c.base()
		if b.dead() {
			return
		}
		b.place()
		ch = b.h
	}
	drv.WindowSetChild(w.h, ch)
}

// Show показывает окно, а если оно уже показано — поднимает наверх.
func (w *Window) Show() {
	if !w.dead() {
		drv.WindowShow(w.h)
	}
}

// Hide скрывает окно, не уничтожая его. Скрытое окно не удерживает программу:
// когда закроется последнее видимое окно, Run завершится (ADR-15).
func (w *Window) Hide() {
	if !w.dead() {
		drv.Hide(w.h)
	}
}

// Close закрывает и уничтожает окно из кода, без вызова OnClose (ADR-14).
func (w *Window) Close() {
	if !w.dead() {
		drv.Destroy(w.h)
	}
}

// Size — текущий размер окна без рамки и заголовка (вместе с отступом окна), логические
// пиксели. У скрытого окна — размер, с которым оно появится.
func (w *Window) Size() (width, height int) {
	if w.dead() {
		return 0, 0
	}
	return drv.WindowSize(w.h)
}

// SetSize: растягиваемое окно — начальный размер, пользователь может его менять;
// нерастягиваемое — фиксированный размер, но не меньше естественного размера содержимого.
func (w *Window) SetSize(width, height int) {
	if !w.dead() {
		drv.WindowSetSize(w.h, width, height)
	}
}

// SetResizable — можно ли менять размер окна мышью (по умолчанию можно).
func (w *Window) SetResizable(on bool) {
	if !w.dead() {
		drv.WindowSetResizable(w.h, on) // заново применит последний SetSize: его смысл от этого зависит
	}
}

// SetMaximized разворачивает окно на весь экран — с рамкой и заголовком, пользователь
// может его восстановить; false — восстановить. Можно вызывать до Show: окно сразу
// откроется развёрнутым, а SetSize задаст его размер после восстановления.
func (w *Window) SetMaximized(on bool) {
	if !w.dead() {
		drv.WindowSetMaximized(w.h, on)
	}
}

// SetFullscreen — полноэкранный режим: без рамки и заголовка, поверх панелей; false —
// вернуть обычный. Можно вызывать до Show.
func (w *Window) SetFullscreen(on bool) {
	if !w.dead() {
		drv.WindowSetFullscreen(w.h, on)
	}
}

// SetBackground — фон окна; NoColor — фон темы.
func (w *Window) SetBackground(c Color) {
	if !w.dead() {
		drv.SetBackground(w.h, uint32(c), false)
	}
}

// SetPadding — отступ между краем окна и содержимым, по умолчанию 12.
func (w *Window) SetPadding(px int) {
	if !w.dead() {
		drv.WindowSetPadding(w.h, px)
	}
}

// SetDefault — кнопка по умолчанию: её нажимает Enter в Entry без OnEnter и в поле
// ComboEntry. nil — убрать.
func (w *Window) SetDefault(b *Button) {
	if w.dead() {
		return
	}
	btn := drv.Nil
	if b != nil && b.id != 0 {
		btn = b.h
	}
	drv.WindowSetDefault(w.h, btn)
}

// SetIconName — иконка из темы иконок по имени, например "utilities-terminal"
// или имя из вашего .desktop-файла.
func (w *Window) SetIconName(name string) {
	if !w.dead() {
		drv.WindowSetIconName(w.h, name)
	}
}

// OnClose: пользователь закрывает окно (крестик, Alt+F4). fn возвращает false —
// окно остаётся. Без обработчика окно закрывается. nil снимает обработчик.
func (w *Window) OnClose(fn func() bool) { w.onClose = fn }

// OnKey: нажата клавиша, пока окно активно, — раньше, чем её получит виджет с фокусом.
// fn возвращает true, если клавиша обработана: тогда виджет её не получит (Enter не
// нажмёт кнопку по умолчанию, буква не попадёт в поле). Повтор при удержании клавиши
// тоже приходит сюда. Об отпускании клавиш и о самих Shift, Ctrl, Alt nui не сообщает.
// nil снимает обработчик.
//
//	w.OnKey(func(k nui.Key, m nui.Mod) bool {
//		if k == 'S' && m == nui.ModCtrl {
//			save()
//			return true
//		}
//		return false
//	})
func (w *Window) OnKey(fn func(k Key, m Mod) bool) {
	core.CheckThread()
	keyDecode = drv.KeyDecode
	w.onKey = fn
	if w.id != 0 {
		drv.WindowSetKeys(w.h, fn != nil || menuHooks.has != nil && menuHooks.has(w))
	}
}

// OnDropFiles: на окно перетащили файлы или папки из файлового менеджера; paths —
// абсолютные пути (на Linux — байты имён как есть). Бросить можно на любое место окна,
// кроме полей ввода: туда, как и везде, перетаскивается текст. nil снимает обработчик.
func (w *Window) OnDropFiles(fn func(paths []string)) {
	core.CheckThread()
	dropDecode = drv.DropPaths
	w.onDrop = fn
	if w.id != 0 {
		drv.WindowSetDrop(w.h, w.id, fn != nil)
	}
}

// Focus передаёт фокус окну.
func (w *Window) Focus() {
	if !w.dead() {
		drv.Focus(w.h)
	}
}

// keyDecode, dropDecode и menuHooks — drv.KeyDecode, drv.DropPaths и сочетания меню, но
// только после первого OnKey, OnDropFiles и SetMenu: прямой вызов (и даже поле с типом
// пункта меню в Window) держал бы их код и описания типов в каждой программе TinyGo.
var (
	keyDecode  func(arg uintptr) (key uint32, mods uint8)
	dropDecode func(arg uintptr) []string
	menuHooks  struct {
		choose func(w *Window, k Key, m Mod) bool // выбрать пункт меню окна по сочетанию
		has    func(w *Window) bool               // у меню окна есть сочетания
		forget func(w *Window)                    // окно уничтожено
	}
)

type windowEv Window

func (w *windowEv) Event(k core.Kind, arg uintptr) uintptr {
	switch k {
	case core.EvClose:
		if w.onClose != nil && !w.onClose() {
			return 1 // окно остаётся
		}
	case core.EvDrop:
		if w.onDrop != nil {
			if paths := dropDecode(arg); len(paths) > 0 {
				w.onDrop(paths)
			}
		}
	case core.EvKey: // сначала сочетания меню, потом OnKey
		if w.onKey == nil && (menuHooks.has == nil || !menuHooks.has((*Window)(w))) {
			return 0
		}
		key, m := keyDecode(arg)
		if key == 0 {
			return 0
		}
		if menuHooks.choose != nil && menuHooks.choose((*Window)(w), Key(key), Mod(m)) {
			return 1
		}
		if w.onKey != nil && w.onKey(Key(key), Mod(m)) {
			return 1 // клавиша обработана
		}
	case core.EvDestroy:
		core.Release(w.id)
		w.id = 0
		w.h = drv.Nil
		windowGone((*Window)(w))
	}
	return 0
}

// windowGone — автовыход (ADR-15): окно уничтожено, и видимых окон не осталось.
// Скрытые окна программу не держат; Hide сам выход не вызывает.
func windowGone(w *Window) {
	if menuHooks.forget != nil {
		menuHooks.forget(w)
	}
	for i, o := range wins {
		if o == w {
			wins = append(wins[:i], wins[i+1:]...)
			break
		}
	}
	for _, o := range wins {
		if drv.Visible(o.h) {
			return
		}
	}
	Quit()
}
