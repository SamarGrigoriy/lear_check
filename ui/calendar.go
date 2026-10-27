// calendar.go — экран «Календарь»: месячная сетка на Canvas, клик по дню
// открывает отметки и записи дня; сверху — текущая и максимальная серии.
package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"studytracker/model"
	"studytracker/plan"

	"nui"
)

// dirColors — цвета направлений в ячейке дня. Цвет подбирается по позиции
// направления в плане, шестое и далее — серый.
var dirColors = []nui.Color{
	nui.RGB(40, 100, 210), // Verilog/цифра
	nui.RGB(30, 150, 60),  // ML
	nui.RGB(190, 120, 10), // Разбор проекта
	nui.RGB(150, 40, 160), // Английский
	nui.RGB(200, 60, 50),  // Проект/заработок
	nui.RGB(90, 90, 90),   // шестое и далее
}

func dirColor(dirs []string, dir string) nui.Color {
	for i, d := range dirs {
		if d == dir {
			return dirColors[i%len(dirColors)]
		}
	}
	return dirColors[len(dirColors)-1]
}

type pageCalendar struct {
	app    *App
	root   *nui.Box
	canvas *nui.Canvas

	viewY, viewM int // показываемый месяц
	monthL       *nui.Label
	streakL      *nui.Label
	dayBox       *nui.Box // панель выбранного дня (пересобирается)
	selected     string   // дата выбранного дня "ГГГГ-ММ-ДД"

	cellW, cellH float64 // размеры сетки для хит-теста мыши
	gridX, gridY float64 // левый верхний угол сетки дней
}

func newPageCalendar(a *App) *pageCalendar {
	p := &pageCalendar{app: a}
	today := a.today
	p.viewY, p.viewM = today.Year(), int(today.Month())

	p.monthL = nui.NewLabel("")
	p.monthL.SetBold(true)
	p.monthL.SetFontScale(1.2)
	p.monthL.SetTextAlign(nui.AlignCenter)

	prev := nui.NewButton("←", func() { p.moveMonth(-1) })
	next := nui.NewButton("→", func() { p.moveMonth(1) })
	todayBtn := nui.NewButton("Сегодня", func() {
		p.viewY, p.viewM = a.today.Year(), int(a.today.Month())
		p.selected = a.TodayStr()
		p.refresh()
	})

	p.canvas = nui.NewCanvas(p.draw)
	p.canvas.SetMinSize(700, 380)
	p.canvas.OnMouse(func(e nui.MouseEvent) {
		if e.Kind == nui.MouseDown && e.Button == 1 {
			p.click(e.X, e.Y)
		}
	})

	p.streakL = nui.NewLabel("")

	p.dayBox = nui.VBox().Spacing(4).Padding(6)
	dayFrame := nui.NewFrame("День", nui.NewScroll(p.dayBox))
	dayFrame.SetMinSize(-1, 150)

	p.root = nui.VBox(
		nui.HBox(prev, next, nui.Expand(p.monthL), todayBtn).Spacing(6),
		p.streakL,
		nui.Expand(p.canvas),
		nui.NewSeparator(),
		nui.Expand(dayFrame),
	).Spacing(6).Padding(8)

	p.selected = a.TodayStr()
	p.refresh()
	return p
}

func (p *pageCalendar) moveMonth(d int) {
	t := time.Date(p.viewY, time.Month(p.viewM), 1, 0, 0, 0, 0, p.app.today.Location())
	t = t.AddDate(0, d, 0)
	p.viewY, p.viewM = t.Year(), int(t.Month())
	p.refresh()
}

var monthNames = []string{"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
	"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"}

const calGap = 3.0 // зазор между ячейками сетки

// layout возвращает геометрию сетки для текущего размера холста.
func (p *pageCalendar) layout(w, h float64) (cw, ch, x0, y0 float64, rows int) {
	first := time.Date(p.viewY, time.Month(p.viewM), 1, 0, 0, 0, 0, time.Local)
	daysIn := int(first.AddDate(0, 1, -1).Day())
	offset := (int(first.Weekday()) + 6) % 7 // понедельник — первый столбец
	rows = (offset + daysIn + 6) / 7

	headH := 24.0
	left, top := 8.0, 6.0
	cw = (w - 16 - calGap*6) / 7
	ch = (h - top - headH - calGap*float64(rows)) / float64(rows)
	if ch > 90 {
		ch = 90
	}
	if ch < 30 {
		ch = 30
	}
	return cw, ch, left, top + headH + calGap, rows
}

// draw рисует заголовки дней и ячейки месяца с полосами минут.
func (p *pageCalendar) draw(pt *nui.Painter) {
	w, h := pt.Size()
	pt.SetColor(nui.RGB(255, 255, 255))
	pt.FillRect(0, 0, w, h)

	cw, ch, x0, y0, rows := p.layout(w, h)
	p.cellW, p.cellH = cw, ch
	p.gridX, p.gridY = x0, y0

	first := time.Date(p.viewY, time.Month(p.viewM), 1, 0, 0, 0, 0, time.Local)
	daysIn := int(first.AddDate(0, 1, -1).Day())
	offset := (int(first.Weekday()) + 6) % 7

	// заголовки дней недели (пн..вс)
	pt.SetColor(nui.RGB(90, 90, 90))
	names := []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}
	for i, nm := range names {
		tw, _ := pt.TextSize(nm)
		pt.Text(x0+float64(i)*(cw+calGap)+(cw-tw)/2, 6, nm)
	}

	dirs := p.app.DB.Plan.Directions
	todayStr := p.app.TodayStr()
	for cell := 0; cell < rows*7; cell++ {
		col, row := cell%7, cell/7
		x := x0 + float64(col)*(cw+calGap)
		y := y0 + float64(row)*(ch+calGap)
		dnum := cell - offset + 1
		if dnum < 1 || dnum > daysIn {
			continue
		}
		ds := fmt.Sprintf("%04d-%02d-%02d", p.viewY, p.viewM, dnum)
		mark := p.app.DB.FindMark(ds)

		bg := nui.RGB(245, 245, 245)
		switch ds {
		case p.selected:
			bg = nui.RGB(255, 235, 150)
		case todayStr:
			bg = nui.RGB(210, 230, 255)
		}
		pt.SetColor(bg)
		pt.FillRect(x, y, cw, ch)

		if mark != nil && mark.TotalMinutes() > 0 {
			// полоса снизу: сегменты по направлениям пропорционально минутам
			total := mark.TotalMinutes()
			bx := x
			for _, d := range dirs {
				m := mark.Minutes[d]
				if m == 0 {
					continue
				}
				seg := cw * float64(m) / float64(total)
				pt.SetColor(dirColor(dirs, d))
				pt.FillRect(bx, y+ch-8, seg, 8)
				bx += seg
			}
			pt.SetColor(nui.RGB(60, 60, 60))
			minTxt := strconv.Itoa(mark.TotalMinutes()) + " мин"
			tw, _ := pt.TextSize(minTxt)
			pt.Text(x+cw-tw-3, y+ch-22, minTxt)
		}

		numCol := nui.RGB(40, 40, 40)
		if ds == todayStr || ds == p.selected {
			numCol = nui.RGB(0, 0, 0)
		}
		pt.SetColor(numCol)
		pt.Text(x+4, y+3, strconv.Itoa(dnum))

		pt.SetColor(nui.RGB(220, 220, 220))
		pt.Rect(x+0.5, y+0.5, cw-1, ch-1)
	}
}

// click — определить день под курсором и показать его детали.
func (p *pageCalendar) click(mx, my float64) {
	if p.cellW <= 0 {
		return
	}
	col := int((mx - p.gridX) / (p.cellW + calGap))
	row := int((my - p.gridY) / (p.cellH + calGap))
	if col < 0 || col > 6 || row < 0 {
		return
	}
	first := time.Date(p.viewY, time.Month(p.viewM), 1, 0, 0, 0, 0, time.Local)
	offset := (int(first.Weekday()) + 6) % 7
	daysIn := int(first.AddDate(0, 1, -1).Day())
	dnum := row*7 + col - offset + 1
	if dnum < 1 || dnum > daysIn {
		return
	}
	p.selected = fmt.Sprintf("%04d-%02d-%02d", p.viewY, p.viewM, dnum)
	p.canvas.Redraw()
	p.showDay()
}

// showDay — панель выбранного дня: минуты по направлениям (Spin правит сразу)
// и записи журнала этого дня.
func (p *pageCalendar) showDay() {
	a := p.app
	p.dayBox.Clear()
	if p.selected == "" {
		p.dayBox.Add(nui.NewLabel("Кликни по дню в сетке."))
		return
	}
	title := nui.NewLabel(dateHuman(p.selected))
	title.SetBold(true)
	p.dayBox.Add(title)

	mark := a.DB.FindMark(p.selected)
	shown := false
	for _, d := range a.DB.Plan.Directions {
		cur := 0
		if mark != nil {
			cur = mark.Minutes[d]
		}
		if cur == 0 && a.DB.Plan.Rhythm.HoursPerDir[d] == 0 {
			continue // направление без ритма и без времени не показываем
		}
		shown = true
		dir, val := d, cur
		sb := nui.NewSpin(0, 1440)
		sb.SetValue(val)
		lbl := nui.NewLabel(formatMin(sb.Value()))
		sb.OnChange(func() {
			m := a.DB.MarkOrCreate(p.selected)
			m.Minutes[dir] = sb.Value()
			lbl.SetText(formatMin(sb.Value()))
			a.Save()
			p.canvas.Redraw()
			p.refreshStreak()
		})
		p.dayBox.Add(nui.HBox(colorDot(a.DB.Plan.Directions, dir), nui.NewLabel(dir), nui.Expand(lbl), sb).Spacing(6))
	}
	if !shown {
		p.dayBox.Add(nui.NewLabel("Занятий в этот день не отмечено."))
	}

	found := false
	for _, e := range a.DB.Entries {
		if e.Date != p.selected {
			continue
		}
		found = true
		txt := firstLine(e.Text)
		if len(e.Text) > len(txt) {
			txt += "…"
		}
		ent := e
		p.dayBox.Add(nui.NewButton("["+e.Type+"] "+txt, func() { a.editJournalEntry(ent) }))
	}
	if !found {
		p.dayBox.Add(nui.NewLabel("Записей журнала нет."))
	}
	sel := p.selected
	p.dayBox.Add(nui.NewButton("+ Заметка за этот день…", func() {
		a.editJournalEntry(&model.JournalEntry{Date: sel, Type: model.NoteEntry})
	}))
}

// formatMin — «60 мин (1 ч)» / «90 мин (1,5 ч)».
func formatMin(m int) string {
	if m < 60 {
		return strconv.Itoa(m) + " мин"
	}
	h, rem := m/60, m%60
	if rem == 0 {
		return strconv.Itoa(m) + " мин (" + strconv.Itoa(h) + " ч)"
	}
	// десятые доли часа: h,round(rem/6)
	tenths := h*10 + rem*10/60
	return strconv.Itoa(m) + " мин (" + model.FormatHrs(tenths) + " ч)"
}

// colorDot — цветной квадратик направления.
func colorDot(dirs []string, dir string) nui.Widget {
	l := nui.NewLabel("■")
	l.SetColor(dirColor(dirs, dir))
	return l
}

func dateHuman(ds string) string {
	d, err := model.ParseDate(ds)
	if err != nil {
		return ds
	}
	return fmt.Sprintf("%d %s %d (%s)", d.Day(), strings.ToLower(monthNames[d.Month()-1]), d.Year(),
		[]string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}[(int(d.Weekday())+6)%7])
}

func (p *pageCalendar) refreshStreak() {
	cur := plan.CurrentStreak(p.app.DB.Marks, p.app.today)
	best := plan.LongestStreak(p.app.DB.Marks)
	p.streakL.SetText(fmt.Sprintf("Серия сейчас: %d дн. · рекорд: %d дн.", cur, best))
}

// refresh — перерисовать месяц, серию и выбранный день.
func (p *pageCalendar) refresh() {
	p.monthL.SetText(monthNames[p.viewM-1] + " " + strconv.Itoa(p.viewY))
	p.refreshStreak()
	p.canvas.Redraw()
	p.showDay()
}
