package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Grid — сетка: ячейки (col, row) размером cols × rows, нумерация с 0 (layout.md, §4).
//
//	nui.NewGrid().
//		Add(0, 0, nui.NewLabel("Порт:")).    Add(1, 0, port).
//		Add(0, 1, nui.NewLabel("Скорость:")).Add(1, 1, baud).
//		ExpandCol(1)
type Grid struct {
	widget
	cells   []gridCell
	expCols []bool // колонка забирает свободное место по ширине
	expRows []bool
}

type gridCell struct {
	w                    *widget
	col, row, cols, rows int
}

// NewGrid создаёт пустую сетку с промежутками 6 по колонкам и строкам.
func NewGrid() *Grid {
	checkCreate()
	g := &Grid{}
	g.id = core.Register((*gridEv)(g))
	g.h = drv.GridNew(g.id)
	return g
}

// Add кладёт виджет в ячейку 1 × 1.
func (g *Grid) Add(col, row int, w Widget) *Grid { return g.AddSpan(col, row, 1, 1, w) }

// AddSpan кладёт виджет в ячейку на cols колонок и rows строк. Если ячейки пересекаются,
// в отладочной сборке будет паника, в обычной — неопределённое наложение.
func (g *Grid) AddSpan(col, row, cols, rows int, w Widget) *Grid {
	if g.dead() {
		return g
	}
	if w == nil {
		panic("nui: вставка nil вместо виджета")
	}
	if col < 0 || row < 0 {
		panic("nui: Grid: отрицательная колонка или строка")
	}
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	b := w.base()
	if b.dead() {
		return g
	}
	c := gridCell{b, col, row, cols, rows}
	if core.Debug {
		for _, o := range g.cells {
			if o.w.id != 0 && c.col < o.col+o.cols && o.col < c.col+c.cols && c.row < o.row+o.rows && o.row < c.row+c.rows {
				panic("nui: ячейки Grid пересекаются")
			}
		}
	}
	b.place()
	drv.GridAttach(g.h, b.h, col, row, cols, rows)
	g.cells = append(g.cells, c)
	g.applyExpand(c)
	return g
}

// Remove убирает виджеты из сетки и уничтожает их (как Box.Remove); ячейки освобождаются.
// Виджет не из этой сетки — паника; уже уничтоженный пропускается.
func (g *Grid) Remove(items ...Widget) *Grid {
	if g.dead() {
		return g
	}
	removeChildren(&g.widget, items)
	g.dropDead()
	return g
}

// Clear убирает и уничтожает всех детей сетки. Растягиваемые колонки и строки (ExpandCol,
// ExpandRow) остаются.
func (g *Grid) Clear() *Grid {
	if !g.dead() {
		drv.ContainerClear(g.h)
		g.cells = nil
	}
	return g
}

// dropDead забывает ячейки уничтоженных виджетов.
func (g *Grid) dropDead() {
	live := g.cells[:0]
	for _, c := range g.cells {
		if c.w.id != 0 {
			live = append(live, c)
		}
	}
	g.cells = live
}

// ExpandCol: колонка забирает свободное место по ширине. Можно вызывать до и после Add.
func (g *Grid) ExpandCol(col int) *Grid {
	if col >= 0 && !g.dead() {
		g.expCols = setFlag(g.expCols, col)
		g.applyAll()
	}
	return g
}

// ExpandRow: строка забирает свободное место по высоте.
func (g *Grid) ExpandRow(row int) *Grid {
	if row >= 0 && !g.dead() {
		g.expRows = setFlag(g.expRows, row)
		g.applyAll()
	}
	return g
}

// Spacing — промежутки между колонками и строками, по умолчанию 6 и 6.
func (g *Grid) Spacing(colGap, rowGap int) *Grid {
	if !g.dead() {
		drv.GridSetSpacing(g.h, colGap, rowGap)
	}
	return g
}

// Background — фон сетки вместе с отступом Padding; NoColor — без фона.
func (g *Grid) Background(c Color) *Grid {
	if !g.dead() {
		drv.SetBackground(g.h, uint32(c), true)
	}
	return g
}

// SetBackground — то же, что Background, без цепочки: одно имя с окном и виджетами.
func (g *Grid) SetBackground(c Color) { g.Background(c) }

// Padding — отступ вокруг, по умолчанию 0.
func (g *Grid) Padding(px int) *Grid {
	if !g.dead() {
		drv.SetPadding(g.h, px)
	}
	return g
}

// applyExpand: в GTK колонка растягивается, если растягивается ребёнок, занимающий
// только её (layout.md, §6). Дальше Grid флаг не уходит: у самого Grid он явно FALSE.
func (g *Grid) applyExpand(c gridCell) {
	if c.w.id == 0 {
		return
	}
	hz := c.cols == 1 && flag(g.expCols, c.col)
	v := c.rows == 1 && flag(g.expRows, c.row)
	drv.SetExpand(c.w.h, hz, v)
}

func (g *Grid) applyAll() {
	for _, c := range g.cells {
		g.applyExpand(c)
	}
}

func setFlag(f []bool, i int) []bool {
	for len(f) <= i {
		f = append(f, false)
	}
	f[i] = true
	return f
}

func flag(f []bool, i int) bool { return i < len(f) && f[i] }

type gridEv Grid

func (g *gridEv) Event(k core.Kind, _ uintptr) uintptr {
	if k == core.EvDestroy {
		g.destroyed()
		g.cells = nil
	}
	return 0
}
