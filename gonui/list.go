package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// List — список или таблица (api.md, §20) с полосами прокрутки. Строки не хранятся:
// текст ячейки спрашивается у функции cell, когда строка видна, — так список на миллион
// строк открывается сразу. Выбирается одна строка; двойной щелчок или Enter — OnActivate.
//
//	files := []string{"a.txt", "b.txt"}
//	l := nui.NewList("Имя")
//	l.SetRows(len(files), func(row, col int) string { return files[row] })
type List struct {
	widget           // h — прокрутка: раскладка и общие методы
	view       drv.H // сам список
	rows       int
	sel        int // строка, о которой OnSelect уже сообщил
	onSelect   func()
	onActivate func(row int)
}

// NewList создаёт пустой список с колонками columns — это заголовки. Без колонок или
// с одним пустым заголовком — простой список без заголовка.
func NewList(columns ...string) *List {
	checkCreate()
	l := &List{sel: -1}
	l.id = core.Register((*listEv)(l))
	l.h, l.view = drv.ListNew(l.id, columns)
	return l
}

// SetRows задаёт данные: n строк, текст ячейки (row, col) возвращает cell — её список
// вызывает в UI-потоке, когда строка видна. Выделение снимается, прокрутка — в начало;
// OnSelect не вызывается. Данные поменялись, а число строк то же — Update.
func (l *List) SetRows(n int, cell func(row, col int) string) {
	if l.dead() {
		return
	}
	if n < 0 || cell == nil {
		n = 0
	}
	if cell == nil {
		cell = func(int, int) string { return "" }
	}
	l.quiet++
	drv.ListSetRows(l.view, n, cell)
	l.quiet--
	l.rows, l.sel = n, -1
}

// Update перерисовывает видимые строки: cell вызовется заново.
func (l *List) Update() {
	if !l.dead() {
		drv.ListUpdate(l.view)
	}
}

// Rows — число строк.
func (l *List) Rows() int {
	if l.dead() {
		return 0
	}
	return l.rows
}

// Selected — выбранная строка; -1 — ничего не выбрано.
func (l *List) Selected() int {
	if l.dead() {
		return -1
	}
	return drv.ListSelected(l.view)
}

// Select выбирает строку и прокручивает к ней; -1 (или строка вне списка) снимает
// выделение. OnSelect не вызывается (ADR-8).
func (l *List) Select(row int) {
	if l.dead() {
		return
	}
	if row >= l.rows {
		row = -1
	}
	l.quiet++
	drv.ListSelect(l.view, row)
	l.quiet--
	l.sel = drv.ListSelected(l.view)
}

// OnSelect: пользователь выбрал другую строку или снял выделение. Номер — Selected().
func (l *List) OnSelect(fn func()) { l.onSelect = fn }

// OnActivate: двойной щелчок по строке или Enter на ней; row — эта строка.
func (l *List) OnActivate(fn func(row int)) { l.onActivate = fn }

type listEv List

func (l *listEv) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvChange:
		// Драйвер сообщает о смене выделения с повторами (GTK — и при движении курсора,
		// Win32 — отдельно о снятии и о выборе): OnSelect — только если строка другая.
		row := drv.ListSelected(l.view)
		if l.quiet == 0 && row != l.sel {
			l.sel = row
			if l.onSelect != nil {
				l.onSelect()
			}
		}
	case core.EvEnter:
		if row := drv.ListSelected(l.view); row >= 0 && l.onActivate != nil {
			l.onActivate(row)
		}
	case core.EvDestroy:
		l.destroyed()
		l.view = drv.Nil
	}
	return 0
}
