package layout

import "testing"

func leaf(w, h int) *Node { return &Node{Min: Size{w, h}, Nat: Size{w, h}} }

func flex(minW, natW, h int) *Node { return &Node{Min: Size{minW, h}, Nat: Size{natW, h}} }

func box(kind uint8, spacing int, kids ...*Node) *Node {
	return &Node{Kind: kind, Spacing: spacing, Children: kids}
}

func run(n *Node, r Rect) {
	Measure(n)
	Arrange(n, r)
}

func want(t *testing.T, what string, got, exp Rect) {
	t.Helper()
	if got != exp {
		t.Errorf("%s: %v, ожидалось %v", what, got, exp)
	}
}

func TestBoxNaturalLeavesRestAtEnd(t *testing.T) {
	a, b := leaf(100, 20), leaf(50, 30)
	row := box(HBox, 6, a, b)
	run(row, Rect{0, 0, 300, 40})
	if row.Min != (Size{156, 30}) || row.Nat != (Size{156, 30}) {
		t.Fatalf("размер HBox: min %v nat %v", row.Min, row.Nat)
	}
	want(t, "a", a.Out, Rect{0, 0, 100, 40})
	want(t, "b", b.Out, Rect{106, 0, 50, 40})
}

func TestBoxExpandTakesFreeSpace(t *testing.T) {
	a, b, c := leaf(100, 20), leaf(50, 20), leaf(30, 20)
	b.Expand = true
	run(box(HBox, 6, a, b, c), Rect{0, 0, 300, 20})
	want(t, "a", a.Out, Rect{0, 0, 100, 20})
	want(t, "b", b.Out, Rect{106, 0, 158, 20})
	want(t, "c", c.Out, Rect{270, 0, 30, 20})
}

// Остаток от деления свободного места достаётся первым растягиваемым.
func TestBoxExpandRemainderGoesFirst(t *testing.T) {
	a, b, c := leaf(10, 10), leaf(10, 10), leaf(10, 10)
	a.Expand, c.Expand = true, true
	run(box(VBox, 0, a, b, c), Rect{0, 0, 10, 37})
	want(t, "a", a.Out, Rect{0, 0, 10, 14})
	want(t, "b", b.Out, Rect{0, 14, 10, 10})
	want(t, "c", c.Out, Rect{0, 24, 10, 13})
}

// Нехватка места (layout.md, §3, п. 4): сначала полностью получают своё дети с наименьшим
// разрывом «естественный − минимальный», доля — с округлением вверх.
func TestBoxShortfall(t *testing.T) {
	a, b, c := flex(10, 50, 10), flex(10, 20, 10), flex(10, 100, 10)
	run(box(HBox, 0, a, b, c), Rect{0, 0, 90, 10}) // 60 сверх минимумов
	if a.Out.W != 35 || b.Out.W != 20 || c.Out.W != 35 {
		t.Fatalf("ширины %d %d %d, ожидалось 35 20 35", a.Out.W, b.Out.W, c.Out.W)
	}
	// Меньше суммы минимумов: все получают минимум и вылезают за край.
	run(box(HBox, 0, a, b, c), Rect{0, 0, 20, 10})
	if a.Out.W != 10 || b.Out.W != 10 || c.Out != (Rect{20, 0, 10, 10}) {
		t.Fatalf("при нехватке минимумов: %v %v %v", a.Out, b.Out, c.Out)
	}
}

func TestDistributeTiesFavourFirst(t *testing.T) {
	sizes, rest := distribute(10, []int{0, 0, 0}, []int{10, 10, 10})
	if sizes[0] != 4 || sizes[1] != 3 || sizes[2] != 3 || rest != 0 {
		t.Fatalf("равные разрывы: %v (остаток %d), ожидалось [4 3 3]", sizes, rest)
	}
	sizes, rest = distribute(100, []int{5, 5}, []int{10, 20})
	if sizes[0] != 10 || sizes[1] != 20 || rest != 80 {
		t.Fatalf("место сверх естественных: %v, остаток %d", sizes, rest)
	}
}

// Как в GTK: если естественный размер линии меньше минимального (бывает в Grid после
// ячеек на несколько линий), распределение сжимает её до естественного.
func TestDistributeNegativeGapShrinks(t *testing.T) {
	sizes, rest := distribute(10, []int{75, 75}, []int{120, 30})
	if sizes[0] != 120 || sizes[1] != 30 || rest != 10 {
		t.Fatalf("%v, остаток %d; ожидалось [120 30], 10", sizes, rest)
	}
}

func TestAlign(t *testing.T) {
	n := leaf(50, 20)
	n.HAlign, n.VAlign = AlignEnd, AlignCenter
	run(n, Rect{10, 10, 200, 41})
	want(t, "End,Center", n.Out, Rect{160, 20, 50, 20})
	n.HAlign, n.VAlign = AlignStart, AlignStart
	run(n, Rect{0, 0, 30, 10}) // ячейка меньше естественного размера
	want(t, "Start в тесной ячейке", n.Out, Rect{0, 0, 30, 10})
	n.HAlign, n.VAlign = AlignCenter, AlignEnd
	run(n, Rect{0, 0, 30, 10})
	want(t, "Center, End в тесной ячейке", n.Out, Rect{0, 0, 30, 10})
}

func TestHiddenTakesNoSpace(t *testing.T) {
	a, b, c := leaf(10, 10), leaf(10, 10), leaf(10, 10)
	b.Hidden = true
	row := box(HBox, 6, a, b, c)
	run(row, Rect{0, 0, 100, 10})
	if row.Min.W != 26 {
		t.Fatalf("ширина с невидимым ребёнком %d, ожидалось 26", row.Min.W)
	}
	want(t, "c", c.Out, Rect{16, 0, 10, 10})
	if b.Out != (Rect{}) {
		t.Fatalf("скрытый узел получил место: %v", b.Out)
	}
	run(box(HBox, 6, b), Rect{0, 0, 10, 10}) // одни скрытые — ничего не падает
}

// Padding входит в размер узла, но не в Out — как border-width в GTK.
func TestPaddingAndFloor(t *testing.T) {
	a := leaf(50, 20)
	col := box(VBox, 6, a)
	col.Padding = 10
	run(col, Rect{0, 0, 100, 100})
	if col.Min != (Size{70, 40}) {
		t.Fatalf("размер с отступом %v", col.Min)
	}
	want(t, "Out без отступа", col.Out, Rect{10, 10, 80, 80})
	want(t, "ребёнок", a.Out, Rect{10, 10, 80, 20})

	col.Floor = Size{200, 0} // SetMinSize(200, -1)
	Measure(col)
	if col.Min != (Size{200, 40}) || col.Nat != (Size{200, 40}) {
		t.Fatalf("Floor: min %v nat %v", col.Min, col.Nat)
	}
	b := flex(10, 30, 10)
	b.Floor = Size{20, 15}
	Measure(b)
	if b.Min != (Size{20, 15}) || b.Nat != (Size{30, 15}) {
		t.Fatalf("Floor у листа: min %v nat %v", b.Min, b.Nat)
	}
}

// Ячейка не бывает меньше 1 px (так делает GTK), а позиция сдвигается на настоящий размер.
func TestZeroSizeCellIsOnePixel(t *testing.T) {
	a, s, b := leaf(10, 10), leaf(0, 0), leaf(10, 10)
	s.Expand = true
	run(box(HBox, 6, a, s, b), Rect{0, 0, 32, 10})
	want(t, "Stretch без места", s.Out, Rect{16, 0, 1, 10})
	want(t, "b", b.Out, Rect{22, 0, 10, 10})
}

func cell(n *Node, col, row, cols, rows int) *Node {
	n.Col, n.Row, n.Cols, n.Rows = col, row, cols, rows
	return n
}

// Форма «подпись: поле», растягивается колонка 1.
func TestGridExpandCol(t *testing.T) {
	l1, e1 := cell(leaf(60, 20), 0, 0, 1, 1), cell(flex(30, 150, 20), 1, 0, 1, 1)
	l2, e2 := cell(leaf(40, 20), 0, 1, 1, 1), cell(flex(30, 150, 20), 1, 1, 1, 1)
	g := &Node{Kind: Grid, ColGap: 6, RowGap: 6, ExpandCols: []bool{false, true}, Children: []*Node{l1, e1, l2, e2}}
	run(g, Rect{0, 0, 400, 46})
	if g.Min != (Size{96, 46}) || g.Nat != (Size{216, 46}) {
		t.Fatalf("размер Grid: min %v nat %v", g.Min, g.Nat)
	}
	want(t, "l1", l1.Out, Rect{0, 0, 60, 20})
	want(t, "l2", l2.Out, Rect{0, 26, 60, 20})
	want(t, "e1", e1.Out, Rect{66, 0, 334, 20})
	want(t, "e2", e2.Out, Rect{66, 26, 334, 20})
}

// Ячейка на две колонки добирает ширину из растягиваемой колонки (layout.md, §4).
func TestGridSpanFavoursExpandCol(t *testing.T) {
	l := cell(leaf(60, 20), 0, 0, 1, 1)
	e := cell(flex(30, 150, 20), 1, 0, 1, 1)
	wide := cell(leaf(300, 20), 0, 1, 2, 1)
	g := &Node{Kind: Grid, ColGap: 6, ExpandCols: []bool{false, true}, Children: []*Node{l, e, wide}}
	run(g, Rect{0, 0, 300, 40})
	if g.Min.W != 300 || g.Nat.W != 300 {
		t.Fatalf("Grid: min %v nat %v", g.Min, g.Nat)
	}
	want(t, "колонка 0 не выросла", l.Out, Rect{0, 0, 60, 20})
	want(t, "колонка 1 забрала всё", e.Out, Rect{66, 0, 234, 20})
	want(t, "широкая ячейка", wide.Out, Rect{0, 20, 300, 20})
}

// Без растягиваемых колонок недостающее делится между всеми: поровну с округлением вниз,
// остаток — последней. Естественная ширина колонки 0 оказывается меньше минимальной,
// и при лишнем месте GTK сжимает её до естественной.
func TestGridSpanWithoutExpandCols(t *testing.T) {
	l := cell(leaf(60, 20), 0, 0, 1, 1)
	e := cell(flex(30, 150, 20), 1, 0, 1, 1)
	wide := cell(leaf(301, 20), 0, 1, 2, 1)
	g := &Node{Kind: Grid, ColGap: 6, Children: []*Node{l, e, wide}}
	run(g, Rect{0, 0, 301, 40})
	// min: 60+30+6 = 96, не хватает 205 → 102 и 103: колонки 162 и 133.
	want(t, "колонка 0 (минимум)", l.Out, Rect{0, 0, 162, 20})
	want(t, "колонка 1 (минимум)", e.Out, Rect{168, 0, 133, 20})
	// nat: 60+150+6 = 216, не хватает 85 → 42 и 43: естественные 102 и 193.
	run(g, Rect{0, 0, 311, 40})
	want(t, "колонка 0 сжата до естественной", l.Out, Rect{0, 0, 102, 20})
	want(t, "колонка 1 естественная", e.Out, Rect{108, 0, 193, 20})
}

// Промежутки — только между непустыми линиями; пустые линии нулевые.
func TestGridEmptyLinesAndHidden(t *testing.T) {
	a := cell(leaf(10, 10), 0, 0, 1, 1)
	b := cell(leaf(10, 10), 3, 0, 1, 1)
	h := cell(leaf(50, 50), 1, 0, 1, 1)
	h.Hidden = true
	g := &Node{Kind: Grid, ColGap: 6, RowGap: 6, Children: []*Node{a, h, b}}
	run(g, Rect{0, 0, 26, 10})
	if g.Min != (Size{26, 10}) {
		t.Fatalf("Grid с пустыми колонками: %v", g.Min)
	}
	want(t, "b", b.Out, Rect{16, 0, 10, 10})
	run(&Node{Kind: Grid, Children: []*Node{h}}, Rect{0, 0, 10, 10}) // только скрытые
}

func TestGridExpandRowAndRemainder(t *testing.T) {
	a := cell(leaf(10, 10), 0, 0, 1, 1)
	b := cell(leaf(10, 10), 0, 1, 1, 1)
	c := cell(leaf(10, 10), 0, 2, 1, 1)
	g := &Node{Kind: Grid, RowGap: 0, ExpandRows: []bool{true, false, true}, Children: []*Node{a, b, c}}
	run(g, Rect{0, 0, 10, 35})
	want(t, "a", a.Out, Rect{0, 0, 10, 13})
	want(t, "b", b.Out, Rect{0, 13, 10, 10})
	want(t, "c", c.Out, Rect{0, 23, 10, 12})
}

func TestFrame(t *testing.T) {
	a := flex(40, 80, 20)
	f := &Node{Kind: Frame, Inset: Insets{1, 18, 1, 1}, Floor: Size{100, 0}, Children: []*Node{a}}
	run(f, Rect{0, 0, 120, 60})
	if f.Min != (Size{100, 39}) || f.Nat != (Size{100, 39}) {
		t.Fatalf("Frame с длинным заголовком: min %v nat %v", f.Min, f.Nat)
	}
	want(t, "ребёнок внутри оформления", a.Out, Rect{1, 18, 118, 41})
	f.Floor = Size{}
	Measure(f)
	if f.Min != (Size{42, 39}) || f.Nat != (Size{82, 39}) {
		t.Fatalf("Frame: min %v nat %v", f.Min, f.Nat)
	}
	empty := &Node{Kind: Frame, Inset: Insets{1, 18, 1, 1}}
	Measure(empty)
	if empty.Min != (Size{2, 19}) {
		t.Fatalf("пустой Frame: %v", empty.Min)
	}
}

func TestStack(t *testing.T) {
	a, b, c := flex(40, 80, 20), flex(60, 70, 50), flex(300, 300, 300)
	c.Hidden = true
	s := &Node{Kind: Stack, Inset: Insets{1, 38, 1, 1}, Children: []*Node{a, b, c}}
	run(s, Rect{0, 0, 120, 100})
	if s.Min != (Size{62, 89}) || s.Nat != (Size{82, 89}) {
		t.Fatalf("Stack: min %v nat %v (скрытая страница не в счёт)", s.Min, s.Nat)
	}
	want(t, "первая страница", a.Out, Rect{1, 38, 118, 61})
	want(t, "вторая — там же", b.Out, Rect{1, 38, 118, 61})
	s.Floor = Size{150, 0} // ярлыки шире страниц
	Measure(s)
	if s.Min != (Size{150, 89}) || s.Nat != (Size{150, 89}) {
		t.Fatalf("Stack с широкими ярлыками: min %v nat %v", s.Min, s.Nat)
	}
	empty := &Node{Kind: Stack, Inset: Insets{1, 1, 1, 1}}
	Measure(empty)
	if empty.Min != (Size{2, 2}) {
		t.Fatalf("пустой Stack: %v", empty.Min)
	}
}

func TestScroll(t *testing.T) {
	a := flex(50, 200, 300)
	s := &Node{Kind: Scroll, Floor: Size{40, 40}, Children: []*Node{a}}
	run(s, Rect{0, 0, 100, 100})
	if s.Min != (Size{40, 40}) || s.Nat != (Size{200, 300}) {
		t.Fatalf("Scroll: min %v nat %v", s.Min, s.Nat)
	}
	// Ребёнок — не меньше своего минимума, остальное по видимой области.
	want(t, "ребёнок", a.Out, Rect{0, 0, 100, 300})
	run(s, Rect{0, 0, 30, 500})
	want(t, "ребёнок в узкой прокрутке", a.Out, Rect{0, 0, 50, 500})
	a.Hidden = true
	Measure(s)
	if s.Nat != (Size{40, 40}) {
		t.Fatalf("Scroll без ребёнка: nat %v", s.Nat)
	}
}

// Вложенность: VBox(Grid, Frame(HBox(...))), раскладка от корня.
func TestNested(t *testing.T) {
	x, y := leaf(30, 10), leaf(30, 10)
	y.Expand = true
	inner := box(HBox, 6, x, y)
	fr := &Node{Kind: Frame, Inset: Insets{2, 2, 2, 2}, Children: []*Node{inner}}
	fr.Expand = true
	top := leaf(100, 20)
	root := box(VBox, 6, top, fr)
	run(root, Rect{12, 12, 200, 100})
	want(t, "top", top.Out, Rect{12, 12, 200, 20})
	want(t, "frame", fr.Out, Rect{12, 38, 200, 74})
	want(t, "x", x.Out, Rect{14, 40, 30, 70})
	want(t, "y", y.Out, Rect{50, 40, 160, 70})
}
