//go:build gui

package nui

import (
	"strconv"
	"testing"

	"nui/internal/drv"
)

// Список: данные по требованию, выбор пользователем и программой, активация, Update,
// новые данные, миллион строк, уничтожение.
func TestList(t *testing.T) {
	names := []string{"отчёт.pdf", "фото 🙂.jpg", "заметки.txt"}
	sizes := []string{"120 КБ", "2,4 МБ", "3 КБ"}
	cell := func(row, col int) string {
		if col == 0 {
			return names[row]
		}
		return sizes[row]
	}
	var w *Window
	var l *List
	var selects, activated []int
	ui(func() {
		l = NewList("Имя", "Размер")
		l.OnSelect(func() { selects = append(selects, l.Selected()) })
		l.OnActivate(func(row int) { activated = append(activated, row) })
		l.SetRows(len(names), cell)
		w = NewWindow("список", VBox(Expand(l), NewLabel("Двойной щелчок — открыть")))
		w.SetSize(360, 240)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, _, ww, hh := alloc(l); return ww > 200 && hh > 100 })
	ui(func() {
		if l.Rows() != 3 || l.Selected() != -1 {
			t.Errorf("после SetRows: Rows = %d, Selected = %d", l.Rows(), l.Selected())
		}
		if got := drv.ListCellText(l.view, 1, 0); got != "фото 🙂.jpg" {
			t.Errorf("ячейка (1, 0) = %q", got)
		}
		if got := drv.ListCellText(l.view, 2, 1); got != "3 КБ" {
			t.Errorf("ячейка (2, 1) = %q", got)
		}
		l.Select(2) // программно: OnSelect молчит
		if l.Selected() != 2 || len(selects) != 0 {
			t.Errorf("Select(2): Selected = %d, OnSelect %v", l.Selected(), selects)
		}
		drv.SimListSelect(l.view, 0) // пользователь
	})
	waitUntil(t, 2000, func() bool { return len(selects) == 1 })
	shot(t, "list")
	ui(func() {
		if selects[0] != 0 {
			t.Errorf("OnSelect: %v, ожидалось [0]", selects)
		}
		drv.SimListActivate(l.view, 1)
	})
	waitUntil(t, 2000, func() bool { return len(activated) == 1 })
	ui(func() {
		if activated[0] != 1 {
			t.Errorf("OnActivate(%d), ожидалось 1", activated[0])
		}
		names[1] = "фото (2).jpg"
		l.Update()
		if got := drv.ListCellText(l.view, 1, 0); got != "фото (2).jpg" {
			t.Errorf("после Update ячейка (1, 0) = %q", got)
		}
		l.Select(-1)
		l.Select(99) // вне списка — снять выделение
		if l.Selected() != -1 {
			t.Errorf("Select(99): Selected = %d", l.Selected())
		}
		// Миллион строк: открывается сразу, cell вызывается только для видимых.
		calls := 0
		l.SetRows(1_000_000, func(row, col int) string { calls++; return "строка " + strconv.Itoa(row) })
		if l.Rows() != 1_000_000 || l.Selected() != -1 {
			t.Errorf("миллион: Rows = %d, Selected = %d", l.Rows(), l.Selected())
		}
		l.Select(999_999)
		if l.Selected() != 999_999 {
			t.Errorf("Select(999999): Selected = %d", l.Selected())
		}
		if got := drv.ListCellText(l.view, 999_999, 0); got != "строка 999999" {
			t.Errorf("ячейка последней строки = %q", got)
		}
		if calls > 10_000 {
			t.Errorf("cell вызвана %d раз — строки должны спрашиваться по требованию", calls)
		}
		l.SetRows(0, nil)
	})
	waitUntil(t, 2000, func() bool { return true })
	ui(func() {
		if l.Rows() != 0 || l.Selected() != -1 {
			t.Errorf("пустой список: Rows = %d, Selected = %d", l.Rows(), l.Selected())
		}
		// 0 — щелчок, 1 — выбор перед Enter; программные Select и SetRows молчат.
		if len(selects) != 2 || selects[1] != 1 {
			t.Errorf("OnSelect: %v, ожидалось [0 1]", selects)
		}
		w.Close()
		l.SetRows(5, cell) // уничтоженный список — ничего не делает
		if l.Rows() != 0 || l.Selected() != -1 {
			t.Errorf("уничтоженный список: Rows = %d, Selected = %d", l.Rows(), l.Selected())
		}
	})
}

// Список без заголовков и без колонок: одна колонка.
func TestListNoHeader(t *testing.T) {
	ui(func() {
		l := NewList()
		l.SetRows(2, func(row, _ int) string { return []string{"а", "б"}[row] })
		w := NewWindow("без заголовка", l)
		w.Show()
		if got := drv.ListCellText(l.view, 1, 0); got != "б" {
			t.Errorf("ячейка = %q", got)
		}
		w.Close()
	})
}
