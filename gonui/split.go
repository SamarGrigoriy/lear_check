package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Split — две панели с границей, которую пользователь двигает мышью (api.md, §15):
// боковое меню и чат, список и просмотр.
//
//	nui.HSplit(menu, nui.NewScroll(content)).SetPosition(220)
//
// Первая панель держит свой размер, когда окно растягивается, — растёт вторая. Сжать
// панель меньше её минимального размера нельзя ни мышью, ни SetPosition.
type Split struct{ widget }

// HSplit — панели слева и справа, граница вертикальная.
func HSplit(first, second Widget) *Split { return newSplit(false, first, second) }

// VSplit — панели сверху и снизу, граница горизонтальная.
func VSplit(first, second Widget) *Split { return newSplit(true, first, second) }

func newSplit(vertical bool, first, second Widget) *Split {
	checkCreate()
	if first == nil || second == nil {
		panic("nui: вставка nil вместо виджета")
	}
	s := &Split{}
	s.id = core.Register((*splitEv)(s))
	s.h = drv.SplitNew(s.id, vertical)
	a, b := first.base(), second.base()
	a.place()
	b.place()
	drv.SplitSet(s.h, a.h, b.h)
	return s
}

// SetPosition — размер первой панели в логических пикселях (ширина у HSplit, высота у
// VSplit). Без него первая панель получает свой естественный размер. Можно до показа.
func (s *Split) SetPosition(px int) *Split {
	if !s.dead() {
		drv.SplitSetPosition(s.h, max(px, 0))
	}
	return s
}

// Position — размер первой панели сейчас (после раскладки); например, чтобы запомнить
// его при закрытии окна и восстановить SetPosition при следующем запуске.
func (s *Split) Position() int {
	if s.dead() {
		return 0
	}
	return drv.SplitPosition(s.h)
}

type splitEv Split

func (s *splitEv) Event(k core.Kind, _ uintptr) uintptr {
	if k == core.EvDestroy {
		s.destroyed()
	}
	return 0
}
