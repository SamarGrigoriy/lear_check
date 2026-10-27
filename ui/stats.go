// stats.go — экран «Статистика»: часы по направлениям за неделю и месяц,
// выполнение плана против факта, динамика результатов тестов, число записей
// журнала по типам. Все данные считаются на месте из DB; отдельных виджетов
// gonui не требует — только Label/Progress/List/Canvas для графика.
package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"studytracker/model"
	"studytracker/plan"
	"studytracker/quiz"

	"nui"
)

type pageStats struct {
	app  *App
	root *nui.Box

	selWeek *nui.Combo // неделя или «Текущая»
	area    *nui.Box   // пересобираемый контент
}

func newPageStats(a *App) *pageStats {
	p := &pageStats{app: a}

	p.selWeek = nui.NewCombo("Текущая неделя")
	for i := 1; i <= a.DB.Plan.LastWeekNumber(); i++ {
		p.selWeek.Add(strconv.Itoa(i))
	}
	p.selWeek.SetSelected(0)
	p.selWeek.OnChange(p.refresh)

	rebuild := nui.NewButton("Обновить", p.refresh)

	top := nui.HBox(nui.NewLabel("Неделя:"), p.selWeek, rebuild).Spacing(6)

	p.area = nui.VBox().Spacing(10)
	p.root = nui.VBox(top, nui.NewSeparator(), nui.Expand(nui.NewScroll(p.area))).Padding(8)
	p.refresh()
	return p
}

// chosenWeek — номер недели из селектора (0 = текущая).
func (p *pageStats) chosenWeek() int {
	if s := p.selWeek.Selected(); s > 0 {
		return s // элемент 0 — «Текущая неделя», элементы 1..N = недели 1..N
	}
	return p.app.CurrentWeek()
}

func (p *pageStats) refresh() {
	a := p.app
	db := a.DB
	p.area.Clear()

	wn := p.chosenWeek()
	dirs := db.Plan.Directions

	// --- 1. Часы по направлениям: выбранная неделя и последние 30 дней ----
	wDays := plan.WeekDays(db, wn)
	wFrom, wTo := "", ""
	if len(wDays) > 0 {
		wFrom = wDays[0].Format(model.DateLayout)
		wTo = wDays[len(wDays)-1].Format(model.DateLayout)
	}
	mFrom := model.AddDays(a.today, -29).Format(model.DateLayout)
	mTo := a.TodayStr()

	wMin := minutesByDir(db, dirs, wFrom, wTo)
	mMin := minutesByDir(db, dirs, mFrom, mTo)

	p.area.Add(sectionTitle(fmt.Sprintf("Часы по направлениям · неделя %d (%s — %s)", wn, wFrom, wTo)))
	totalW := 0
	for _, d := range dirs {
		totalW += wMin[d]
	}
	plannedTenths := 0
	for _, d := range dirs {
		plannedTenths += db.Plan.Rhythm.HoursPerDir[d]
	}
	p.area.Add(kvLine("Всего за неделю", fmt.Sprintf("%s ч (план %s ч)",
		model.FormatHrs(totalW*10/60), model.FormatHrs(plannedTenths))))
	for _, d := range dirs {
		bar := hbar(wMin[d]/60.0, float64(db.Plan.Rhythm.HoursPerDir[d])/10.0)
		p.area.Add(kvLine(d, fmt.Sprintf("%s ч · месяц %s ч  %s",
			hoursStr(wMin[d]), hoursStr(mMin[d]), bar)))
	}

	// --- 2. Выполнение плана против факта ---------------------------------
	done, skipped, total := taskCounts(db.Tasks)
	p.area.Add(sectionTitle("Выполнение плана"))
	p.area.Add(kvLine("Задачи всего", strconv.Itoa(total)))
	p.area.Add(kvLine("Сделано", fmt.Sprintf("%d (%d%%)", done, pct(done, total))))
	p.area.Add(kvLine("Пропущено", strconv.Itoa(skipped)))
	actH, plH := actualPlannedHours(db.Tasks)
	p.area.Add(kvLine("Часы факт/план", fmt.Sprintf("%d ч / %d ч", actH, plH)))

	// --- 3. Динамика результатов тестов ------------------------------------
	p.area.Add(sectionTitle("Тесты: динамика по месяцам"))
	st := quiz.ComputeStats(db.Attempts)
	allAsked, allCorrect := 0, 0
	for _, s := range st.ByDir {
		allAsked += s.Asked
		allCorrect += s.Correct
	}
	p.area.Add(kvLine("Всего ответов", fmt.Sprintf("%d, верно %d%%", allAsked, pct(allCorrect, allAsked))))
	for _, e := range sortedStats(st.ByDir) {
		s := e.S
		p.area.Add(kvLine(e.Name, fmt.Sprintf("%d%% (%d из %d) %s",
			int(s.Rate()+0.5), s.Correct, s.Asked, hbar(float64(s.Correct), float64(s.Asked)))))
	}
	p.area.Add(monthTrend(db.Attempts))

	// --- 4. Записи журнала по типам ----------------------------------------
	p.area.Add(sectionTitle("Записи журнала по типам"))
	byType := map[string]int{}
	for _, en := range db.Entries {
		byType[en.Type]++
	}
	for _, t := range []string{model.NoteEntry, model.NoteWin, model.NoteFailure, model.NoteQuestion} {
		p.area.Add(kvLine(t, strconv.Itoa(byType[t])))
	}

	// --- 5. Серии -----------------------------------------------------------
	cur := plan.CurrentStreak(db.Marks, a.today)
	best := plan.LongestStreak(db.Marks)
	p.area.Add(sectionTitle("Серии занятий"))
	p.area.Add(kvLine("Текущая серия", fmt.Sprintf("%d дн.", cur)))
	p.area.Add(kvLine("Максимальная", fmt.Sprintf("%d дн.", best)))
	p.area.Add(kvLine("Дней с занятиями", strconv.Itoa(activeDays(db.Marks))))
}

// helpers ------------------------------------------------------------

func sectionTitle(s string) *nui.Label {
	l := nui.NewLabel(s)
	l.SetBold(true)
	return l
}

func kvLine(k, v string) *nui.Box {
	return nui.HBox(nui.NewLabel(k), nui.Stretch(), nui.NewLabel(v)).Spacing(8)
}

func hoursStr(min int) string {
	h := min / 60
	m := min % 60
	if h == 0 {
		return fmt.Sprintf("%d мин", m)
	}
	return fmt.Sprintf("%d ч %02d мин", h, m)
}

// hbar — текстовая полоска прогресса фиксированной длины.
func hbar(done, total float64) string {
	const w = 20
	f := 0.0
	if total > 0 {
		f = done / total
	}
	if f > 1 {
		f = 1
	}
	n := int(f*w + 0.5)
	return "[" + strings.Repeat("#", n) + strings.Repeat(".", w-n) + "]"
}

func pct(part, whole int) int {
	if whole == 0 {
		return 0
	}
	return part * 100 / whole
}

// minutesByDir — сумма минут по направлениям в диапазоне дат [from..to] включительно.
func minutesByDir(db *model.DB, dirs []string, from, to string) map[string]int {
	res := map[string]int{}
	for _, m := range db.Marks {
		if m.Date < from || m.Date > to {
			continue
		}
		for d, v := range m.Minutes {
			res[d] += v
		}
	}
	return res
}

func taskCounts(tasks []*model.WeekTask) (done, skipped, total int) {
	for _, t := range tasks {
		total++
		switch t.Status {
		case model.TaskDone:
			done++
		case model.TaskSkipped:
			skipped++
		}
	}
	return
}

func actualPlannedHours(tasks []*model.WeekTask) (act, plan int) {
	for _, t := range tasks {
		act += t.ActualHrs
		plan += t.PlannedHrs
	}
	return
}

// sortedStats — сводка по направлениям, отсортированная по имени (стабильный вывод).
func sortedStats(m map[string]quiz.Stat) []struct {
	Name string
	S    quiz.Stat
} {
	var out []struct {
		Name string
		S    quiz.Stat
	}
	for k, v := range m {
		out = append(out, struct {
			Name string
			S    quiz.Stat
		}{k, v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func activeDays(marks []*model.DayMark) int {
	n := 0
	for _, m := range marks {
		if plan.ActiveDay(m) {
			n++
		}
	}
	return n
}

// monthTrend — график процента верных ответов по месяцам (Canvas).
func monthTrend(attempts []*model.QuizAttempt) *nui.Box {
	type mo struct {
		key     string // "2026-10"
		asked   int
		correct int
	}
	idx := map[string]*mo{}
	var keys []string
	for _, a := range attempts {
		if len(a.Date) < 7 {
			continue
		}
		k := a.Date[:7]
		m := idx[k]
		if m == nil {
			m = &mo{key: k}
			idx[k] = m
			keys = append(keys, k)
		}
		m.asked++
		if a.Correct {
			m.correct++
		}
	}
	sort.Strings(keys)
	if len(keys) > 12 {
		keys = keys[len(keys)-12:]
	}

	box := nui.VBox()
	if len(keys) == 0 {
		box.Add(nui.NewLabel("Ответов на тесты пока нет."))
		return box
	}
	data := make([]float64, len(keys))
	labels := make([]string, len(keys))
	for i, k := range keys {
		m := idx[k]
		data[i] = float64(m.correct) / float64(m.asked)
		labels[i] = k[5:] + "/" + k[2:4] // "MM/YY"
	}
	const cw, ch = 48, 120
	canvas := nui.NewCanvas(func(pp *nui.Painter) {
		pp.SetColor(nui.RGB(230, 230, 230))
		pp.FillRect(0, ch-1, float64(cw*len(data)), 1)
		for i, v := range data {
			x := float64(i*cw) + 6
			h := v * (ch - 20)
			if v >= 0.7 {
				pp.SetColor(nui.RGB(30, 150, 60))
			} else if v >= 0.4 {
				pp.SetColor(nui.RGB(220, 170, 40))
			} else {
				pp.SetColor(nui.RGB(200, 60, 50))
			}
			pp.FillRect(x, ch-20-h, float64(cw-12), h)
			pp.SetColor(nui.RGB(60, 60, 60))
			pp.Text(x, ch-14, labels[i])
			pp.Text(x, ch-30-h, fmt.Sprintf("%.0f%%", v*100))
		}
	})
	canvas.SetMinSize(cw*len(data), ch)
	box.Add(nui.NewLabel("Верных ответов по месяцам:"))
	box.Add(canvas)
	return box
}
