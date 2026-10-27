package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Widget — любой элемент интерфейса nui, который можно вставить в контейнер.
// Интерфейс запечатан: его реализуют только типы nui.
type Widget interface {
	base() *widget
}

// Align — как виджет располагается в выделенной ему ячейке (layout.md, §1).
type Align uint8

const (
	AlignFill   Align = iota // занять всё доступное место (по умолчанию)
	AlignStart               // прижать к началу (влево / вверх)
	AlignCenter              // по центру
	AlignEnd                 // прижать к концу (вправо / вниз)
)

// widget — общая часть всех виджетов. Методы ниже есть у каждого виджета.
type widget struct {
	h      drv.H  // хендл ОС
	id     uint32 // id в реестре; 0 — объект уничтожен
	quiet  uint8  // >0 — идёт программное изменение, события пользователю не передаём (ADR-8)
	expand bool   // для Box: забирать свободное место по главной оси
	placed bool   // уже вставлен в контейнер (повторная вставка — паника)
}

func (w *widget) base() *widget { return w }

// dead — объект уничтожен: методы ничего не делают, геттеры возвращают нулевые значения.
func (w *widget) dead() bool {
	core.CheckThread()
	if w.id != 0 {
		return false
	}
	if core.Debug {
		println("nui: вызов метода у уничтоженного виджета")
	}
	return true
}

// destroyed — реакция на destroy: GTK уничтожил виджет, Go-объект выходит из реестра.
func (w *widget) destroyed() {
	core.Release(w.id)
	w.id = 0
	w.h = drv.Nil
}

// place отмечает вставку в контейнер: виджет вставляется ровно в один контейнер.
func (w *widget) place() {
	if w.placed {
		panic("nui: виджет уже вставлен в контейнер")
	}
	w.placed = true
}

// Alive — false после уничтожения: закрыто окно, виджет заменён через SetChild.
func (w *widget) Alive() bool { return w.id != 0 }

// SetEnabled: недоступный виджет виден, но не реагирует; в контейнере
// недоступность распространяется на детей.
func (w *widget) SetEnabled(on bool) {
	if !w.dead() {
		drv.SetEnabled(w.h, on)
	}
}

func (w *widget) Enabled() bool { return !w.dead() && drv.Enabled(w.h) }

// SetVisible: скрытый виджет не занимает места в раскладке.
func (w *widget) SetVisible(on bool) {
	if !w.dead() {
		drv.SetVisible(w.h, on)
	}
}

func (w *widget) Visible() bool { return !w.dead() && drv.Visible(w.h) }

// SetTooltip задаёт всплывающую подсказку; "" убирает её.
func (w *widget) SetTooltip(text string) {
	if !w.dead() {
		drv.SetTooltip(w.h, text)
	}
}

// SetAlign — как виджет располагается в своей ячейке; по умолчанию AlignFill, AlignFill.
func (w *widget) SetAlign(h, v Align) {
	if !w.dead() {
		drv.SetAlign(w.h, uint8(h), uint8(v))
	}
}

// Size — текущий размер виджета в логических пикселях (тех же, что у SetMinSize) — после
// последней раскладки. Виджет не на экране (скрыт сам или родитель, окно не показано) —
// 0, 0. Раскладка идёт после Show и после изменений не сразу, а в цикле событий: сразу
// после Show размер ещё 0 — спрашивайте его в обработчиках или через After.
func (w *widget) Size() (width, height int) {
	if w.dead() {
		return 0, 0
	}
	return drv.Size(w.h)
}

// SetMinSize задаёт минимальный размер в логических пикселях; -1 — не задавать по этой оси.
func (w *widget) SetMinSize(width, height int) {
	if !w.dead() {
		drv.SetMinSize(w.h, width, height)
	}
}

// Focus передаёт виджету фокус ввода.
func (w *widget) Focus() {
	if !w.dead() {
		drv.Focus(w.h)
	}
}
