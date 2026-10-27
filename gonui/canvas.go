package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Canvas — холст (api.md, §21): своё рисование — графики, индикаторы, схемы. Рисует
// функция draw, когда холст надо показать (первый показ, изменение размера, Redraw).
// Координаты — логические пиксели от левого верхнего угла, как у SetMinSize; сглаживание
// включено. Своего размера у холста нет — задайте SetMinSize или вставьте через Expand.
//
//	c := nui.NewCanvas(func(p *nui.Painter) {
//		w, h := p.Size()
//		p.SetColor(nui.RGB(40, 90, 200))
//		p.FillEllipse(0, 0, w, h)
//	})
//	c.SetMinSize(200, 120)
type Canvas struct {
	widget
	draw    func(p *Painter)
	onMouse func(e MouseEvent)
}

// NewCanvas создаёт холст; draw может быть nil (пустой холст, например, только мышь).
func NewCanvas(draw func(p *Painter)) *Canvas {
	checkCreate()
	c := &Canvas{draw: draw}
	c.id = core.Register((*canvasEv)(c))
	c.h = drv.CanvasNew(c.id)
	return c
}

// Redraw просит нарисовать холст заново (draw вызовется при ближайшей перерисовке окна,
// несколько Redraw подряд — одна перерисовка).
func (c *Canvas) Redraw() {
	if !c.dead() {
		drv.CanvasRedraw(c.h)
	}
}

// OnMouse: нажатие, отпускание кнопки, движение мыши над холстом и колесо. Пока кнопка
// нажата, движение и отпускание приходят и за пределами холста (координаты — тоже от его
// угла, могут быть отрицательными).
func (c *Canvas) OnMouse(fn func(e MouseEvent)) { c.onMouse = fn }

// MouseKind — вид события мыши.
type MouseKind uint8

const (
	MouseDown  MouseKind = drv.MouseDown  // кнопка нажата
	MouseUp    MouseKind = drv.MouseUp    // отпущена
	MouseMove  MouseKind = drv.MouseMove  // мышь сдвинулась
	MouseWheel MouseKind = drv.MouseWheel // колесо: WheelX, WheelY
)

// MouseEvent — событие мыши над холстом.
type MouseEvent struct {
	Kind   MouseKind
	X, Y   float64 // логические пиксели от левого верхнего угла холста
	Button int     // MouseDown, MouseUp: 1 — левая, 2 — средняя, 3 — правая
	Mod    Mod     // нажатые Shift, Ctrl, Alt
	WheelX float64 // MouseWheel: −1 — влево, +1 — вправо (на щелчок колеса)
	WheelY float64 // MouseWheel: −1 — вверх (от себя), +1 — вниз (к себе)
}

// Painter рисует на холсте; действует только внутри draw. Начальный цвет — цвет текста
// темы, толщина линии — 1. Чёткая линия толщиной 1 — по координате с половиной (10.5):
// иначе она ляжет на два ряда точек, как во всех векторных библиотеках.
type Painter struct {
	ctx uintptr
	h   drv.H
}

// Size — размер холста, логические пиксели.
func (p *Painter) Size() (width, height float64) {
	w, h := drv.Size(p.h)
	return float64(w), float64(h)
}

// SetColor — цвет линий, заливки и текста; NoColor — цвет текста темы.
func (p *Painter) SetColor(c Color) {
	if p.ctx == 0 {
		return
	}
	if c == NoColor {
		drv.PaintThemeColor(p.ctx, p.h)
		return
	}
	drv.PaintColor(p.ctx, uint32(c))
}

// SetLineWidth — толщина линий (Line, Rect, Ellipse), логические пиксели.
func (p *Painter) SetLineWidth(w float64) {
	if p.ctx != 0 {
		drv.PaintLineWidth(p.ctx, w)
	}
}

// Line — отрезок.
func (p *Painter) Line(x1, y1, x2, y2 float64) {
	if p.ctx != 0 {
		drv.PaintLine(p.ctx, x1, y1, x2, y2)
	}
}

// Rect — контур прямоугольника.
func (p *Painter) Rect(x, y, w, h float64) {
	if p.ctx != 0 {
		drv.PaintRect(p.ctx, x, y, w, h, false)
	}
}

// FillRect — залитый прямоугольник.
func (p *Painter) FillRect(x, y, w, h float64) {
	if p.ctx != 0 {
		drv.PaintRect(p.ctx, x, y, w, h, true)
	}
}

// Ellipse — контур эллипса, вписанного в прямоугольник (круг — при w == h).
func (p *Painter) Ellipse(x, y, w, h float64) {
	if p.ctx != 0 {
		drv.PaintEllipse(p.ctx, x, y, w, h, false)
	}
}

// FillEllipse — залитый эллипс.
func (p *Painter) FillEllipse(x, y, w, h float64) {
	if p.ctx != 0 {
		drv.PaintEllipse(p.ctx, x, y, w, h, true)
	}
}

// Text — текст шрифтом интерфейса; (x, y) — левый верхний угол. \n — перевод строки.
func (p *Painter) Text(x, y float64, s string) {
	if p.ctx != 0 {
		drv.PaintText(p.ctx, p.h, x, y, s)
	}
}

// TextSize — размер текста шрифтом интерфейса: чтобы выровнять или вписать его.
func (p *Painter) TextSize(s string) (width, height float64) {
	if p.ctx == 0 {
		return 0, 0
	}
	return drv.TextSize(p.ctx, p.h, s)
}

type canvasEv Canvas

func (c *canvasEv) Event(k core.Kind, arg uintptr) uintptr {
	switch k {
	case core.EvDraw:
		if c.draw != nil {
			p := &Painter{ctx: arg, h: c.h}
			drv.PaintBegin(arg, c.h)
			c.draw(p)
			p.ctx = 0 // Painter, сохранённый на потом, ничего не рисует
		}
	case core.EvMouse:
		if c.onMouse == nil {
			return 0
		}
		kind, x, y, button, mods, wx, wy := drv.MouseDecode(arg)
		if kind == 0 {
			return 0
		}
		c.onMouse(MouseEvent{Kind: MouseKind(kind), X: x, Y: y, Button: button, Mod: Mod(mods), WheelX: wx, WheelY: wy})
		return 1
	case core.EvDestroy:
		c.destroyed()
	}
	return 0
}
