//go:build gui

package nui

import (
	"testing"

	"nui/internal/drv"
)

// Tabs: первая страница — текущая; размер — по самой большой странице, все страницы
// на одном месте; SetCurrent без OnChange; щелчок по ярлыку и Ctrl+PgDn — с OnChange;
// скрытая и удалённая текущая страница уступает место следующей (или предыдущей);
// заголовки; контекстное меню страницы; паники.
func TestTabs(t *testing.T) {
	var w *Window
	var tabs *Tabs
	var name *Entry
	var general *Box
	var network, about *Label
	changes := 0
	ui(func() {
		name = NewEntry()
		general = VBox(NewLabel("Имя:"), name)
		network = NewLabel("Прокси не используется")
		network.SetMinSize(320, 160)
		about = NewLabel("nui — маленькая библиотека")
		tabs = NewTabs().
			Add("Общие", general).
			Add("Сеть & прокси", network).
			Add("О программе", about)
		tabs.OnChange(func() { changes++ })
		w = NewWindow("вкладки", tabs)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { gw, _ := general.Size(); return gw > 0 })
	shot(t, "tabs-general")
	ui(func() {
		if tabs.Current() != 0 || tabs.Len() != 3 || tabs.Title(1) != "Сеть & прокси" || tabs.Page(2) != Widget(about) {
			t.Errorf("начало: Current %d, Len %d, Title(1) %q", tabs.Current(), tabs.Len(), tabs.Title(1))
		}
		if tabs.Page(3) != nil || tabs.Title(-1) != "" {
			t.Error("Page и Title вне диапазона")
		}
		gw, gh := general.Size()
		if gw < 320 || gh < 160 {
			t.Errorf("страница %d×%d — меньше самой большой (320×160)", gw, gh)
		}
		gx, gy, _, _ := alloc(general)
		nx, ny, nw, nh := alloc(network)
		if gx != nx || gy != ny || gw != nw || gh != nh {
			t.Errorf("страницы не на одном месте: %d,%d %d×%d и %d,%d %d×%d", gx, gy, gw, gh, nx, ny, nw, nh)
		}
		if sw, sh := network.Size(); sw != 0 || sh != 0 {
			t.Errorf("не текущая страница видна: %d×%d", sw, sh)
		}
		tabs.SetCurrent(2)
		tabs.SetCurrent(7) // вне диапазона — ничего
	})
	waitUntil(t, 2000, func() bool { aw, _ := about.Size(); return aw > 0 })
	ui(func() {
		if gw, _ := general.Size(); gw != 0 || tabs.Current() != 2 || changes != 0 {
			t.Errorf("SetCurrent(2): Current %d, первая шириной %d, OnChange ×%d", tabs.Current(), gw, changes)
		}
		drv.SimTabsSelect(tabs.h, 1) // щелчок по ярлыку
		if tabs.Current() != 1 || changes != 1 {
			t.Errorf("щелчок по ярлыку: Current %d, OnChange ×%d", tabs.Current(), changes)
		}
	})
	waitUntil(t, 2000, func() bool { nw, _ := network.Size(); return nw > 0 })
	shot(t, "tabs-network")

	ui(func() { // Ctrl+PgDn, когда фокус внутри вкладок
		tabs.SetCurrent(0)
		name.Focus()
	})
	waitUntil(t, 2000, func() bool { gw, _ := general.Size(); return gw > 0 })
	ui(func() {
		drv.SimKey(w.h, drv.KeyPageDown, drv.ModCtrl)
		if tabs.Current() != 1 || changes != 2 {
			t.Errorf("Ctrl+PgDn: Current %d, OnChange ×%d", tabs.Current(), changes)
		}
		drv.SimKey(w.h, drv.KeyPageUp, drv.ModCtrl)
		if tabs.Current() != 0 || changes != 3 {
			t.Errorf("Ctrl+PgUp: Current %d, OnChange ×%d", tabs.Current(), changes)
		}

		tabs.SetCurrent(1)
		network.SetVisible(false) // скрытая текущая — дальше следующая
		if tabs.Current() != 2 {
			t.Errorf("скрыта текущая: Current %d, ожидалась 2", tabs.Current())
		}
		tabs.SetCurrent(1) // скрытую выбрать нельзя
		network.SetVisible(true)
		if tabs.Current() != 2 {
			t.Errorf("SetCurrent скрытой или показ: Current %d, ожидалась 2", tabs.Current())
		}

		tabs.SetCurrent(1)
		tabs.Remove(network) // удалена текущая — следующая, теперь под номером 1
		if tabs.Current() != 1 || tabs.Len() != 2 || network.Alive() || tabs.Page(1) != Widget(about) ||
			tabs.Title(1) != "О программе" {
			t.Errorf("удалена текущая: Current %d, Len %d", tabs.Current(), tabs.Len())
		}
		tabs.Remove(about) // последняя — предыдущая
		if tabs.Current() != 0 || tabs.Len() != 1 {
			t.Errorf("удалена последняя: Current %d, Len %d", tabs.Current(), tabs.Len())
		}
		if changes != 3 {
			t.Errorf("OnChange на программные изменения: ×%d", changes)
		}
		tabs.SetTitle(0, "Главная")
		if tabs.Title(0) != "Главная" {
			t.Errorf("SetTitle: %q", tabs.Title(0))
		}
		tabs.Remove(about) // уже уничтожена — ничего
		expectPanic(t, "nui: Remove: виджет не в этом контейнере", func() { tabs.Remove(NewLabel("чужая")) })
		expectPanic(t, "nui: вставка nil вместо виджета", func() { tabs.Add("пусто", nil) })
		expectPanic(t, "nui: виджет уже вставлен в контейнер", func() { tabs.Add("снова", general) })
	})

	// Контекстное меню страницы: правый щелчок по странице вкладок.
	var page *Box
	ui(func() {
		page = VBox(NewLabel("ещё страница"))
		page.SetContextMenu(Item("Закрыть вкладку", nil))
		tabs.Add("Ещё", page)
		tabs.SetCurrent(1)
	})
	waitUntil(t, 2000, func() bool { pw, _ := page.Size(); return pw > 1 }) // 1×1 — ещё до раскладки
	ui(func() {
		px, py, pw, ph := alloc(page)
		if got := drv.ContextMenuOwner(w.h, px+pw/2, py+ph/2); got != page.h {
			t.Errorf("меню у страницы: %v, ожидалось %v", got, page.h)
		}
		w.Close()
	})
	waitUntil(t, 2000, func() bool { return !tabs.Alive() && !page.Alive() })
}

// Вкладки без страниц, с одной страницей и с длинными заголовками: ширина — не меньше
// самого широкого ярлыка, Current без страниц — -1.
func TestTabsSizes(t *testing.T) {
	var w *Window
	var empty, long *Tabs
	var short *Label
	ui(func() {
		empty = NewTabs()
		short = NewLabel("x")
		long = NewTabs().Add("Очень длинный заголовок вкладки", short)
		w = NewWindow("размеры вкладок", VBox(empty, long))
		w.Show()
	})
	waitUntil(t, 2000, func() bool { sw, _ := short.Size(); return sw > 0 })
	ui(func() {
		if empty.Current() != -1 || empty.Len() != 0 {
			t.Errorf("пустые: Current %d, Len %d", empty.Current(), empty.Len())
		}
		lw, _ := long.Size()
		sw, _ := short.Size()
		if lw < 150 || sw < lw-10 {
			t.Errorf("длинный ярлык: вкладки шириной %d, страница %d", lw, sw)
		}
		w.Close()
	})
}
