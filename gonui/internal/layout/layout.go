// Package layout — движок раскладки на чистом Go для драйверов без собственных
// контейнеров (Win32). Поведение повторяет GTK3 — эталон ADR-10: алгоритмы взяты
// из gtkbox.c, gtkgrid.c, gtkwidget.c и gtksizerequest.c (GTK 3.24) и сверяются
// golden-тестами, снятыми с настоящего GTK (layout.md, §7–8).
//
// Модель: дерево Node. Измерение (Measure) идёт снизу вверх: размеры листьев задаёт
// драйвер, размеры контейнеров считает движок. Раскладка (Arrange) идёт сверху вниз:
// каждый узел получает ячейку и размещается в ней по своему выравниванию.
package layout

// Size — размер в логических пикселях.
type Size struct{ W, H int }

// Rect — прямоугольник в координатах окна.
type Rect struct{ X, Y, W, H int }

// Insets — оформление между краем узла и его ребёнком: рамка и заголовок Frame,
// рамка Scroll. Его измеряет драйвер.
type Insets struct{ L, T, R, B int }

// Виды узлов.
const (
	Leaf   = iota // виджет: Min и Nat задаёт драйвер
	VBox          // дети сверху вниз
	HBox          // дети слева направо
	Grid          // сетка: Col, Row, Cols, Rows у детей
	Frame         // один ребёнок внутри оформления Inset
	Scroll        // один ребёнок; если не помещается — прокручивается
	Stack         // дети в одном месте внутри Inset (страницы Tabs); размер — по самому большому
)

// Выравнивание внутри ячейки — значения совпадают с nui.Align.
const (
	AlignFill = iota
	AlignStart
	AlignCenter
	AlignEnd
)

// Node — узел дерева раскладки.
type Node struct {
	Kind           uint8
	Hidden         bool  // SetVisible(false): не занимает места, вокруг нет промежутков
	Expand         bool  // ребёнок Box забирает свободное место по главной оси
	HAlign, VAlign uint8 // выравнивание в ячейке

	// Min, Nat — минимальный и естественный размер: у листьев их задаёт драйвер,
	// у контейнеров считает Measure. Floor поднимает оба (SetMinSize, size request
	// в GTK); у Frame драйвер кладёт сюда и ширину заголовка, у Scroll — свой минимум
	// (полосы прокрутки).
	Min, Nat, Floor Size

	Inset   Insets // Frame, Scroll, Stack: оформление вокруг детей
	Padding int    // отступ вокруг содержимого: Box.Padding, Grid.Padding

	Spacing        int // Box: промежуток между детьми
	ColGap, RowGap int // Grid: промежутки между непустыми колонками и строками

	Col, Row, Cols, Rows   int    // положение ребёнка Grid; Cols, Rows < 1 считаются 1
	ExpandCols, ExpandRows []bool // Grid: какие колонки и строки забирают свободное место

	Children []*Node

	// Out — результат Arrange: место узла без Padding, как allocation в GTK.
	// Координаты — от начала окна; у потомков Scroll — при нулевой прокрутке.
	Out Rect
}

// Measure считает Min и Nat контейнеров снизу вверх; у листьев только применяет Floor.
func Measure(n *Node) {
	switch n.Kind {
	case VBox, HBox:
		measureBox(n)
	case Grid:
		measureGrid(n)
	case Frame, Scroll:
		measureBin(n)
	case Stack:
		measureStack(n)
	}
	// gtk_widget_real_adjust_size_request: size request поднимает минимум,
	// а естественный размер не бывает меньше минимального.
	n.Min.W = maxInt(n.Min.W, n.Floor.W)
	n.Min.H = maxInt(n.Min.H, n.Floor.H)
	n.Nat.W = maxInt(n.Nat.W, n.Min.W)
	n.Nat.H = maxInt(n.Nat.H, n.Min.H)
}

// Arrange размещает n в ячейке cell по его выравниванию, затем — его детей.
// Перед Arrange дерево должно быть измерено (Measure).
func Arrange(n *Node, cell Rect) {
	// gtk_container_adjust_size_allocation: отступ контейнера вычитается раньше
	// выравнивания — результат тот же, что выровнять целиком и затем вычесть.
	r := align(cell, n)
	p := n.Padding
	n.Out = Rect{r.X + p, r.Y + p, r.W - 2*p, r.H - 2*p}
	switch n.Kind {
	case VBox, HBox:
		arrangeBox(n)
	case Grid:
		arrangeGrid(n)
	case Frame:
		if c := child(n); c != nil {
			Arrange(c, inner(n))
		}
	case Stack: // GtkNotebook: все видимые страницы получают одно место
		for _, c := range n.Children {
			if !c.Hidden {
				Arrange(c, inner(n))
			}
		}
	case Scroll:
		// GtkViewport: ребёнок получает видимую область, но не меньше своего минимума.
		if c := child(n); c != nil {
			v := inner(n)
			Arrange(c, Rect{v.X, v.Y, maxInt(v.W, c.Min.W), maxInt(v.H, c.Min.H)})
		}
	}
}

// align — adjust_for_align из gtkwidget.c: Fill занимает всю ячейку; Start, Center,
// End получают естественный размер (не больше ячейки) и прижимаются.
func align(cell Rect, n *Node) Rect {
	cell.X, cell.W = alignAxis(n.HAlign, cell.X, cell.W, n.Nat.W)
	cell.Y, cell.H = alignAxis(n.VAlign, cell.Y, cell.H, n.Nat.H)
	return cell
}

func alignAxis(a uint8, pos, size, nat int) (int, int) {
	switch a {
	case AlignStart:
		return pos, minInt(size, nat)
	case AlignEnd:
		if size > nat {
			return pos + size - nat, nat
		}
	case AlignCenter:
		if size > nat {
			return pos + (size-nat)/2, nat
		}
	}
	return pos, size
}

// --- Frame, Scroll ---

// child — единственный ребёнок Frame и Scroll, если он есть и виден.
func child(n *Node) *Node {
	if len(n.Children) == 0 || n.Children[0].Hidden {
		return nil
	}
	return n.Children[0]
}

func inner(n *Node) Rect {
	o, in := n.Out, n.Inset
	return Rect{o.X + in.L, o.Y + in.T, o.W - in.L - in.R, o.H - in.T - in.B}
}

// measureBin: Frame — ребёнок плюс оформление. Scroll — минимум только из оформления
// (остальное драйвер кладёт в Floor), естественный — по ребёнку, как GtkScrolledWindow
// с propagate-natural-width/height.
func measureBin(n *Node) {
	dx := n.Inset.L + n.Inset.R + 2*n.Padding
	dy := n.Inset.T + n.Inset.B + 2*n.Padding
	n.Min, n.Nat = Size{dx, dy}, Size{dx, dy}
	c := child(n)
	if c == nil {
		return
	}
	Measure(c)
	n.Nat = Size{c.Nat.W + dx, c.Nat.H + dy}
	if n.Kind == Frame {
		n.Min = Size{c.Min.W + dx, c.Min.H + dy}
	}
}

// measureStack — GtkNotebook: оформление плюс самая большая из видимых страниц (по
// каждой оси отдельно); скрытые страницы не в счёт.
func measureStack(n *Node) {
	dx := n.Inset.L + n.Inset.R + 2*n.Padding
	dy := n.Inset.T + n.Inset.B + 2*n.Padding
	n.Min, n.Nat = Size{dx, dy}, Size{dx, dy}
	for _, c := range n.Children {
		if c.Hidden {
			continue
		}
		Measure(c)
		n.Min = Size{maxInt(n.Min.W, c.Min.W+dx), maxInt(n.Min.H, c.Min.H+dy)}
		n.Nat = Size{maxInt(n.Nat.W, c.Nat.W+dx), maxInt(n.Nat.H, c.Nat.H+dy)}
	}
}

// --- Box (gtkbox.c) ---

func measureBox(n *Node) {
	vertical := n.Kind == VBox
	var minMain, natMain, minCross, natCross, vis int
	for _, c := range n.Children {
		if c.Hidden {
			continue
		}
		Measure(c)
		vis++
		minMain += mainAxis(c.Min, vertical)
		natMain += mainAxis(c.Nat, vertical)
		minCross = maxInt(minCross, crossAxis(c.Min, vertical))
		natCross = maxInt(natCross, crossAxis(c.Nat, vertical))
	}
	if vis > 1 {
		minMain += n.Spacing * (vis - 1)
		natMain += n.Spacing * (vis - 1)
	}
	p := 2 * n.Padding
	n.Min = make2(minMain+p, minCross+p, vertical)
	n.Nat = make2(natMain+p, natCross+p, vertical)
}

// arrangeBox — gtk_box_size_allocate_no_center: всем минимум, остаток сверх минимумов
// раздаёт distribute, а то, что осталось и после естественных размеров, — поровну детям
// с Expand (первым — по лишнему пикселю).
func arrangeBox(n *Node) {
	vertical := n.Kind == VBox
	r := n.Out
	var vis []*Node
	nexp := 0
	for _, c := range n.Children {
		if c.Hidden {
			continue
		}
		vis = append(vis, c)
		if c.Expand {
			nexp++
		}
	}
	if len(vis) == 0 {
		return
	}
	size := mainAxis(Size{r.W, r.H}, vertical) - (len(vis)-1)*n.Spacing
	mins := make([]int, len(vis))
	nats := make([]int, len(vis))
	for i, c := range vis {
		mins[i], nats[i] = mainAxis(c.Min, vertical), mainAxis(c.Nat, vertical)
		size -= mins[i]
	}
	sizes, size := distribute(maxInt(0, size), mins, nats)
	extra, rest := 0, 0
	if nexp > 0 {
		extra, rest = size/nexp, size%nexp
	}
	pos := r.X
	if vertical {
		pos = r.Y
	}
	for i, c := range vis {
		s := sizes[i]
		if c.Expand {
			s += extra
			if rest > 0 {
				s++
				rest--
			}
		}
		// Ячейка не бывает меньше 1 px, а позиция сдвигается на настоящий размер (GTK).
		if vertical {
			Arrange(c, Rect{r.X, pos, maxInt(1, r.W), maxInt(1, s)})
		} else {
			Arrange(c, Rect{pos, r.Y, maxInt(1, s), maxInt(1, r.H)})
		}
		pos += s + n.Spacing
	}
}

// distribute — gtk_distribute_natural_allocation: раздаёт space сверх минимумов, сначала
// полностью удовлетворяя детей с наименьшим разрывом «естественный − минимальный»;
// доля — оставшееся место на оставшихся детей с округлением вверх. Возвращает размеры
// и нераспределённый остаток.
func distribute(space int, mins, nats []int) ([]int, int) {
	n := len(mins)
	sizes := append([]int(nil), mins...)
	// Порядок как у GTK: разрыв по убыванию, при равенстве — индекс по убыванию;
	// обход с конца — от наименьшего разрыва.
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	gap := func(i int) int { return maxInt(nats[i]-mins[i], 0) }
	for i := 1; i < n; i++ { // вставками: n мало, без sort (ADR-11)
		for j := i; j > 0 && less(order[j], order[j-1], gap); j-- {
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	for i := n - 1; space > 0 && i >= 0; i-- {
		k := order[i]
		glue := (space + i) / (i + 1)
		// Без обрезки снизу, как в GTK: у линии Grid естественный размер может оказаться
		// меньше минимального (после ячеек на несколько линий), и тогда она сжимается.
		e := minInt(glue, nats[k]-mins[k])
		sizes[k] += e
		space -= e
	}
	return sizes, space
}

// less: a идёт раньше b — у a больший разрыв, при равенстве — больший индекс.
func less(a, b int, gap func(int) int) bool {
	if ga, gb := gap(a), gap(b); ga != gb {
		return ga > gb
	}
	return a > b
}

// --- Grid (gtkgrid.c) ---

type line struct {
	min, nat   int
	expand     bool
	empty      bool
	alloc, pos int
}

// gridLines — линии одной оси: колонки (cols) или строки. lo — номер первой линии.
func gridLines(n *Node, cols bool) (lines []line, lo int) {
	lo, hi := 1<<30, -1
	for _, c := range n.Children {
		if c.Hidden {
			continue
		}
		p, s := span(c, cols)
		lo, hi = minInt(lo, p), maxInt(hi, p+s)
	}
	if hi < 0 {
		return nil, 0
	}
	lines = make([]line, hi-lo)
	for i := range lines {
		lines[i].empty = true
	}
	flags, gapSize := n.ExpandRows, n.RowGap
	if cols {
		flags, gapSize = n.ExpandCols, n.ColGap
	}
	// Ячейки на одну линию задают её min/nat; растягивается линия, если растягивается
	// такой ребёнок (в GTK — hexpand/vexpand, layout.md, §6). Линии под ячейками
	// на несколько линий тоже непустые.
	for _, c := range n.Children {
		if c.Hidden {
			continue
		}
		p, s := span(c, cols)
		for i := p; i < p+s; i++ {
			lines[i-lo].empty = false
		}
		if s != 1 {
			continue
		}
		l := &lines[p-lo]
		l.min = maxInt(l.min, axis(c.Min, cols))
		l.nat = maxInt(l.nat, axis(c.Nat, cols))
		if p < len(flags) && flags[p] {
			l.expand = true
		}
	}
	// gtk_grid_request_spanning: ячейка на несколько линий добирает недостающее —
	// из растягиваемых линий, а если таких нет, из всех; поровну с округлением вниз,
	// остаток — последней. Так и для минимума, и для естественного размера.
	for _, c := range n.Children {
		if c.Hidden {
			continue
		}
		p, s := span(c, cols)
		if s == 1 {
			continue
		}
		spanMin, spanNat, nexp := (s-1)*gapSize, (s-1)*gapSize, 0
		for i := p; i < p+s; i++ {
			spanMin += lines[i-lo].min
			spanNat += lines[i-lo].nat
			if lines[i-lo].expand {
				nexp++
			}
		}
		force := nexp == 0
		if force {
			nexp = s
		}
		if m := axis(c.Min, cols); spanMin < m {
			spread(lines[p-lo:p-lo+s], m-spanMin, nexp, force, true)
		}
		if nt := axis(c.Nat, cols); spanNat < nt {
			spread(lines[p-lo:p-lo+s], nt-spanNat, nexp, force, false)
		}
	}
	return lines, lo
}

// spread добавляет extra к минимуму (toMin) или к естественному размеру линий ls.
func spread(ls []line, extra, nexp int, force, toMin bool) {
	for i := range ls {
		l := &ls[i]
		if !force && !l.expand {
			continue
		}
		add := extra / nexp
		if toMin {
			l.min += add
		} else {
			l.nat += add
		}
		extra -= add
		nexp--
	}
}

// sumLines — размер оси: линии плюс промежутки между непустыми (gtk_grid_request_sum).
func sumLines(lines []line, gapSize int) (min, nat int) {
	nonempty := 0
	for _, l := range lines {
		if l.empty {
			continue
		}
		nonempty++
		min += l.min
		nat += l.nat
	}
	if nonempty > 1 {
		min += (nonempty - 1) * gapSize
		nat += (nonempty - 1) * gapSize
	}
	return min, nat
}

func measureGrid(n *Node) {
	for _, c := range n.Children {
		if !c.Hidden {
			Measure(c)
		}
	}
	cl, _ := gridLines(n, true)
	rl, _ := gridLines(n, false)
	cmin, cnat := sumLines(cl, n.ColGap)
	rmin, rnat := sumLines(rl, n.RowGap)
	p := 2 * n.Padding
	n.Min = Size{cmin + p, rmin + p}
	n.Nat = Size{cnat + p, rnat + p}
}

// allocLines — gtk_grid_request_allocate и _position: размер и положение каждой линии.
func allocLines(lines []line, gapSize, total, start int) {
	nonempty, nexp := 0, 0
	for _, l := range lines {
		if !l.empty {
			nonempty++
			if l.expand {
				nexp++
			}
		}
	}
	if nonempty == 0 {
		return
	}
	size := total - (nonempty-1)*gapSize
	var mins, nats []int
	for _, l := range lines {
		if !l.empty {
			size -= l.min
			mins = append(mins, l.min)
			nats = append(nats, l.nat)
		}
	}
	sizes, size := distribute(maxInt(0, size), mins, nats)
	extra, rest := 0, 0
	if nexp > 0 {
		extra, rest = size/nexp, size%nexp
	}
	j, pos := 0, start
	for i := range lines {
		l := &lines[i]
		if l.empty {
			continue
		}
		l.alloc = sizes[j]
		if l.expand {
			l.alloc += extra
			if rest > 0 {
				l.alloc++
				rest--
			}
		}
		l.pos = pos
		pos += l.alloc + gapSize
		j++
	}
}

func arrangeGrid(n *Node) {
	r := n.Out
	cl, clo := gridLines(n, true)
	rl, rlo := gridLines(n, false)
	allocLines(cl, n.ColGap, r.W, r.X)
	allocLines(rl, n.RowGap, r.H, r.Y)
	for _, c := range n.Children {
		if c.Hidden {
			continue
		}
		cp, cs := span(c, true)
		rp, rs := span(c, false)
		x, w := cellOf(cl, clo, cp, cs, n.ColGap)
		y, h := cellOf(rl, rlo, rp, rs, n.RowGap)
		Arrange(c, Rect{x, y, maxInt(1, w), maxInt(1, h)})
	}
}

// cellOf — положение и размер ячейки на s линий, начиная с p (gtk_grid_request_alloc).
func cellOf(lines []line, lo, p, s, gapSize int) (pos, size int) {
	pos = lines[p-lo].pos
	size = (s - 1) * gapSize
	for i := p; i < p+s; i++ {
		size += lines[i-lo].alloc
	}
	return pos, size
}

func span(c *Node, cols bool) (pos, size int) {
	if cols {
		return c.Col, maxInt(c.Cols, 1)
	}
	return c.Row, maxInt(c.Rows, 1)
}

// --- мелочи ---

func mainAxis(s Size, vertical bool) int {
	if vertical {
		return s.H
	}
	return s.W
}

func crossAxis(s Size, vertical bool) int {
	if vertical {
		return s.W
	}
	return s.H
}

func make2(mainV, crossV int, vertical bool) Size {
	if vertical {
		return Size{crossV, mainV}
	}
	return Size{mainV, crossV}
}

func axis(s Size, cols bool) int {
	if cols {
		return s.W
	}
	return s.H
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
