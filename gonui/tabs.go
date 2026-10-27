package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Tabs — вкладки (api.md, §15): несколько страниц на одном месте, сверху — ярлыки
// с заголовками; видна страница выбранного ярлыка.
//
//	tabs := nui.NewTabs().
//		Add("Общие", general).
//		Add("Сеть", network)
//
// Размер вкладок — по самой большой странице, поэтому окно не прыгает при переключении.
type Tabs struct {
	widget
	pages    []Widget
	titles   []string
	onChange func()
}

// NewTabs — вкладки без страниц.
func NewTabs() *Tabs {
	checkCreate()
	t := &Tabs{}
	t.id = core.Register((*tabsEv)(t))
	t.h = drv.TabsNew(t.id)
	return t
}

// Add добавляет страницу page с ярлыком title в конец. Первая страница становится
// текущей. Страница — любой виджет, обычно VBox или Grid.
func (t *Tabs) Add(title string, page Widget) *Tabs {
	if t.dead() {
		return t
	}
	if page == nil {
		panic("nui: вставка nil вместо виджета")
	}
	w := page.base()
	if w.dead() {
		return t
	}
	w.place()
	t.pages = append(t.pages, page)
	t.titles = append(t.titles, title)
	t.quiet++
	drv.TabsAdd(t.h, w.h, title)
	t.quiet--
	return t
}

// Remove убирает страницы вместе с ярлыками и уничтожает их, как Box.Remove. Если убрана
// текущая, текущей становится следующая (или предыдущая, если убрана последняя).
// Страница не из этих вкладок — паника; уже уничтоженная пропускается.
func (t *Tabs) Remove(pages ...Widget) *Tabs {
	if t.dead() {
		return t
	}
	t.quiet++
	removeChildren(&t.widget, pages)
	t.quiet--
	n := 0
	for i, p := range t.pages {
		if p.base().id != 0 {
			t.pages[n], t.titles[n] = p, t.titles[i]
			n++
		}
	}
	for i := n; i < len(t.pages); i++ {
		t.pages[i] = nil // уничтоженные страницы не держим
	}
	t.pages, t.titles = t.pages[:n], t.titles[:n]
	return t
}

// Len — число страниц.
func (t *Tabs) Len() int { return len(t.pages) }

// Page — страница i; вне диапазона — nil.
func (t *Tabs) Page(i int) Widget {
	if i < 0 || i >= len(t.pages) {
		return nil
	}
	return t.pages[i]
}

// Current — номер текущей страницы; -1 — страниц нет.
func (t *Tabs) Current() int {
	if t.dead() {
		return -1
	}
	return drv.TabsCurrent(t.h)
}

// SetCurrent делает текущей страницу i. Номер вне диапазона и скрытая страница
// (SetVisible(false)) — ничего не меняется.
func (t *Tabs) SetCurrent(i int) {
	if t.dead() || i < 0 || i >= len(t.pages) || !t.pages[i].base().Visible() {
		return
	}
	t.quiet++
	drv.TabsSetCurrent(t.h, i)
	t.quiet--
}

// Title — заголовок ярлыка страницы i; вне диапазона — "".
func (t *Tabs) Title(i int) string {
	if i < 0 || i >= len(t.titles) {
		return ""
	}
	return t.titles[i]
}

// SetTitle меняет заголовок ярлыка страницы i.
func (t *Tabs) SetTitle(i int, title string) {
	if t.dead() || i < 0 || i >= len(t.pages) {
		return
	}
	t.titles[i] = title
	drv.TabsSetTitle(t.h, t.pages[i].base().h, title)
}

// OnChange — пользователь выбрал другую вкладку: щелчком по ярлыку или с клавиатуры
// (стрелки на ярлыках, Ctrl+PgUp и Ctrl+PgDn). Новый номер — Current().
func (t *Tabs) OnChange(fn func()) { t.onChange = fn }

type tabsEv Tabs

func (t *tabsEv) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvChange:
		if t.quiet == 0 && t.onChange != nil {
			t.onChange()
		}
	case core.EvDestroy:
		t.destroyed()
	}
	return 0
}
