package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Menu — меню в строке меню окна (Window.SetMenu, api.md, §5).
//
//	w.SetMenu(
//		nui.NewMenu("Файл",
//			nui.Item("Открыть…", open).Shortcut('O', nui.ModCtrl),
//			nui.SubMenu("Недавние", nui.Item("отчёт.txt", openReport)),
//			nui.Sep(),
//			nui.Item("Выход", nui.Quit)),
//		nui.NewMenu("Вид", nui.CheckItem("Боковая панель", showPanel)),
//	)
type Menu struct {
	title string
	items []*MenuItem
	used  bool
}

// NewMenu — меню с заголовком title в строке меню.
func NewMenu(title string, items ...*MenuItem) *Menu {
	return &Menu{title: title, items: items}
}

const (
	itemPlain = iota
	itemCheck
	itemSep
	itemSub
)

// MenuItem — пункт меню: обычный (Item), с флажком (CheckItem), подменю (SubMenu) или
// разделитель (Sep). Пункт создаётся без окна; в меню окна он попадает через SetMenu —
// один раз, как виджет в контейнер. Состояние (SetEnabled, SetChecked, SetText) можно
// задавать и до, и после.
type MenuItem struct {
	kind     uint8
	text     string
	fn       func()
	onToggle func(on bool)
	key      Key
	mod      Mod
	sub      []*MenuItem
	disabled bool
	checked  bool
	used     bool

	id    uint32 // в реестре, пока пункт в меню окна; 0 — ещё не там или уничтожен
	h     drv.H
	quiet uint8
}

// Item — пункт: fn вызывается, когда его выбирают (мышью, с клавиатуры или сочетанием).
func Item(text string, fn func()) *MenuItem {
	contextMenuHook = setContextMenu
	return &MenuItem{kind: itemPlain, text: text, fn: fn}
}

// CheckItem — пункт с флажком: выбор переключает флажок и вызывает fn с новым состоянием.
func CheckItem(text string, fn func(on bool)) *MenuItem {
	contextMenuHook = setContextMenu
	return &MenuItem{kind: itemCheck, text: text, onToggle: fn}
}

// SubMenu — пункт, открывающий вложенное меню.
func SubMenu(title string, items ...*MenuItem) *MenuItem {
	contextMenuHook = setContextMenu
	return &MenuItem{kind: itemSub, text: title, sub: items}
}

// Sep — разделитель.
func Sep() *MenuItem {
	contextMenuHook = setContextMenu
	return &MenuItem{kind: itemSep}
}

// Shortcut — сочетание клавиш: показывается в меню справа от текста и выбирает пункт,
// пока окно активно, — раньше OnKey окна и виджета с фокусом. Буквы — по положению
// клавиши, как в OnKey: Ctrl+O работает и в русской раскладке. У выключенного пункта
// сочетание не срабатывает и уходит дальше (OnKey, виджет).
func (i *MenuItem) Shortcut(k Key, m Mod) *MenuItem {
	i.key, i.mod = k, m
	return i
}

// SetEnabled — выключенный пункт виден серым и не выбирается.
func (i *MenuItem) SetEnabled(on bool) {
	core.CheckThread()
	i.disabled = !on
	if i.id != 0 && i.kind != itemSep {
		drv.MenuItemSetEnabled(i.h, on)
	}
}

// Enabled — пункт можно выбрать.
func (i *MenuItem) Enabled() bool { return !i.disabled }

// SetChecked — флажок CheckItem; fn не вызывается (ADR-8).
func (i *MenuItem) SetChecked(on bool) {
	core.CheckThread()
	i.checked = on
	if i.id != 0 && i.kind == itemCheck {
		i.quiet++
		drv.MenuItemSetChecked(i.h, on)
		i.quiet--
	}
}

// Checked — флажок CheckItem.
func (i *MenuItem) Checked() bool { return i.checked }

// SetText меняет текст пункта (или заголовок подменю).
func (i *MenuItem) SetText(s string) {
	core.CheckThread()
	i.text = s
	if i.id != 0 && i.kind != itemSep {
		drv.MenuItemSetText(i.h, s, uint32(i.key), uint8(i.mod))
	}
}

// choose — пункт выбран сочетанием клавиш: как выбор в меню.
func (i *MenuItem) choose() {
	if i.kind == itemCheck {
		i.SetChecked(!i.checked)
		if i.onToggle != nil {
			i.onToggle(i.checked)
		}
		return
	}
	if i.fn != nil {
		i.fn()
	}
}

type menuItemEv MenuItem

func (e *menuItemEv) Event(k core.Kind, _ uintptr) uintptr {
	i := (*MenuItem)(e)
	switch k {
	case core.EvClick: // выбран в меню; у флажка состояние уже переключил драйвер
		if i.quiet > 0 || i.disabled {
			return 0
		}
		if i.kind == itemCheck {
			i.checked = drv.MenuItemChecked(i.h)
			if i.onToggle != nil {
				i.onToggle(i.checked)
			}
		} else if i.fn != nil {
			i.fn()
		}
	case core.EvDestroy:
		core.Release(i.id)
		i.id = 0
		i.h = drv.Nil
	}
	return 0
}

// SetMenu ставит строку меню окна; прежняя уничтожается вместе с пунктами. Без
// аргументов — убрать меню. Меню и пункт попадают в окно один раз (повтор — паника).
// Размер окна (SetSize, Size) — область под строкой меню, на всех ОС одинаково.
func (w *Window) SetMenu(menus ...*Menu) {
	if w.dead() {
		return
	}
	for _, m := range menus {
		if m.used {
			panic("nui: меню уже в окне")
		}
		m.used = true
		markUsed(m.items)
	}
	delete(shortcuts, w)
	if len(menus) == 0 {
		drv.MenuBarRemove(w.h)
	} else {
		bar := drv.MenuBar(w.h)
		for _, m := range menus {
			_, menu := drv.MenuSub(bar, 0, m.title)
			buildMenu(menu, m.items, w)
		}
		drv.MenuDone(w.h)
	}
	if len(shortcuts[w]) > 0 {
		keyDecode = drv.KeyDecode
		menuHooks.choose = chooseShortcut
		menuHooks.has = func(w *Window) bool { return len(shortcuts[w]) > 0 }
		menuHooks.forget = func(w *Window) { delete(shortcuts, w) }
	}
	drv.WindowSetKeys(w.h, w.onKey != nil || len(shortcuts[w]) > 0)
}

func markUsed(items []*MenuItem) {
	for _, it := range items {
		if it == nil {
			panic("nui: nil вместо пункта меню")
		}
		if it.used {
			panic("nui: пункт меню уже в меню")
		}
		it.used = true
		markUsed(it.sub)
	}
}

// buildMenu создаёт пункты в меню драйвера и собирает сочетания клавиш окна w (nil —
// контекстное меню: сочетания в нём только подписываются).
func buildMenu(menu drv.H, items []*MenuItem, w *Window) {
	for _, it := range items {
		if it.kind == itemSep {
			drv.MenuSepNew(menu)
			continue
		}
		it.id = core.Register((*menuItemEv)(it))
		if it.kind == itemSub {
			var sub drv.H
			it.h, sub = drv.MenuSub(menu, it.id, it.text)
			buildMenu(sub, it.sub, w)
		} else {
			it.h = drv.MenuItemNew(menu, it.id, it.kind == itemCheck, it.text, uint32(it.key), uint8(it.mod))
			if it.key != 0 && w != nil {
				shortcuts[w] = append(shortcuts[w], it)
			}
			if it.kind == itemCheck && it.checked {
				it.quiet++
				drv.MenuItemSetChecked(it.h, true)
				it.quiet--
			}
		}
		if it.disabled {
			drv.MenuItemSetEnabled(it.h, false)
		}
	}
}

// shortcuts — пункты с сочетаниями клавиш у меню каждого окна. Не поле Window: тогда
// описание MenuItem попадало бы в программы без меню (TinyGo).
var shortcuts = map[*Window][]*MenuItem{}

// chooseShortcut выбирает пункт меню окна с этим сочетанием; false — такого нет (или
// он выключен).
func chooseShortcut(w *Window, k Key, m Mod) bool {
	for _, it := range shortcuts[w] {
		if it.key == k && it.mod == m && it.id != 0 && !it.disabled {
			it.choose()
			return true
		}
	}
	return false
}

// SetContextMenu — меню по правому щелчку по виджету, а также по клавише «Меню» или
// Shift+F10, когда виджет в фокусе. Без аргументов — убрать. Если у виджета под указателем
// своего меню нет, показывается меню ближайшего контейнера выше (Box, Grid…), а затем —
// окна (Window.SetContextMenu). У полей ввода (Entry, TextView) оно заменяет их обычное
// меню («Вырезать, Копировать, Вставить»); в списке правый щелчок сначала выбирает строку.
// Пункты — те же, что в меню окна; сочетания клавиш в контекстном меню только подписываются.
func (w *widget) SetContextMenu(items ...*MenuItem) {
	if !w.dead() && contextMenuHook != nil { // без пунктов (Item …) меню не было и нет
		contextMenuHook(w.h, items)
	}
}

// contextMenuHook — setContextMenu; его задают конструкторы пунктов. Прямой вызов из
// метода widget держал бы код и описания пунктов меню в каждой программе TinyGo: методы
// общего widget остаются у всех виджетов.
var contextMenuHook func(owner drv.H, items []*MenuItem)

// SetContextMenu — меню по правому щелчку по окну там, где у виджетов своего меню нет.
func (w *Window) SetContextMenu(items ...*MenuItem) {
	if !w.dead() {
		setContextMenu(w.h, items)
	}
}

func setContextMenu(owner drv.H, items []*MenuItem) {
	if len(items) == 0 {
		drv.ContextMenuRemove(owner)
		return
	}
	markUsed(items)
	menu := drv.ContextMenu(owner)
	buildMenu(menu, items, nil)
	drv.ContextMenuDone(owner, menu)
}
