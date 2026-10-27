// Package plan — расписание: какая неделя плана соответствует какой дате,
// статусы недель и серии дней. Чистая логика, тестируется без GUI.
package plan

import (
	"sort"
	"time"

	"studytracker/model"
)

// WeekOfDate — номер недели плана, к которой относится дата (1-based).
// Неделя 1 начинается с plan.Start. До начала плана возвращает 0.
// Если включён сдвиг плана (Settings.ShiftPlan), целые календарные недели без
// единого занятия не получают номера: следующая активная неделя продолжает
// нумерацию плана. Нумерация задаётся проходом weekAssignments от старта,
// поэтому WeekOfDate и WeekStart всегда согласованы.
func WeekOfDate(db *model.DB, date time.Time) int {
	if db == nil || db.Plan == nil || len(db.Plan.Stages) == 0 {
		return 0
	}
	start, err := model.ParseDate(db.Plan.Start)
	if err != nil {
		return 0
	}
	diff := model.DayDiff(start, date)
	if diff < 0 {
		return 0
	}
	last := db.Plan.LastWeekNumber()
	cal := diff/7 + 1 // календарная неделя от старта (1-based)
	if !db.Settings.ShiftPlan {
		if cal > last {
			return last
		}
		return cal
	}
	num := 0
	for c := 1; c <= cal && num < last; c++ {
		if calendarWeekActive(db, start, c) {
			num++
		}
	}
	if num == 0 {
		return 1 // старт был, но занятий ещё не было: идём по календарю
	}
	if num > last {
		return last
	}
	return num
}

// calendarWeekActive — было ли хотя бы одно занятие в календарной неделе c
// (1-based от старта плана).
func calendarWeekActive(db *model.DB, start time.Time, c int) bool {
	first := model.AddDays(start, (c-1)*7)
	for _, m := range db.Marks {
		if !ActiveDay(m) {
			continue
		}
		d, err := model.ParseDate(m.Date)
		if err != nil || d.Before(first) {
			continue
		}
		if model.DayDiff(first, d) < 7 {
			return true
		}
	}
	return false
}

// ActiveDay — считается ли день занятым: сумма минут > 0.
func ActiveDay(m *model.DayMark) bool { return m.TotalMinutes() > 0 }

// WeekStart — дата начала недели плана n. Без сдвига это Start + 7*(n-1);
// со сдвигом — первая календарная неделя, на которой проход нумерации
// присваивает плановой неделе номер n.
func WeekStart(db *model.DB, n int) time.Time {
	start, _ := model.ParseDate(db.Plan.Start)
	if !db.Settings.ShiftPlan {
		return model.AddDays(start, (n-1)*7)
	}
	num := 0
	for c := 1; c <= 52*3; c++ {
		if calendarWeekActive(db, start, c) {
			num++
			if num == n {
				return model.AddDays(start, (c-1)*7)
			}
		}
	}
	return model.AddDays(start, (n-1)*7) // занятий мало — падаем на прямой счёт
}

// WeekDays — даты семи дней недели n: [0] — начало недели.
func WeekDays(db *model.DB, n int) []time.Time {
	first := WeekStart(db, n)
	days := make([]time.Time, 7)
	for i := range days {
		days[i] = model.AddDays(first, i)
	}
	return days
}

// WeekStatus — состояние недели по её задачам:
// "done" — все задачи сделаны или пропущены; "active" — есть сделанные/в процессе;
// "new" — ни одна задача не начата.
func WeekStatus(tasks []*model.WeekTask, week int) string {
	any, allDone := false, true
	for _, t := range tasks {
		if t.Week != week {
			continue
		}
		any = true
		if t.Status != model.TaskDone && t.Status != model.TaskSkipped {
			allDone = false
		}
	}
	if !any {
		return "new"
	}
	if allDone {
		return "done"
	}
	for _, t := range tasks {
		if t.Week == week && (t.Status == model.TaskDone || t.Status == model.TaskInProgress) {
			return "active"
		}
	}
	return "new"
}

// SortMarksDesc приводит отметки к порядку «свежие впереди» — так проще
// показывать серию и последние дни.
func SortMarksDesc(marks []*model.DayMark) {
	sort.Slice(marks, func(i, j int) bool { return marks[i].Date > marks[j].Date })
}

// CurrentStreak — серия подряд идущих активных дней, заканчивающаяся today
// (или yesterday: если сегодня ещё не отмечено, вчерашняя серия не обрывается).
func CurrentStreak(marks []*model.DayMark, today time.Time) int {
	set := map[string]bool{}
	for _, m := range marks {
		if ActiveDay(m) {
			set[m.Date] = true
		}
	}
	// старт: сегодня, иначе вчера
	cur := today
	if !set[ymd(cur)] {
		cur = model.AddDays(cur, -1)
		if !set[ymd(cur)] {
			return 0
		}
	}
	n := 0
	for set[ymd(cur)] {
		n++
		cur = model.AddDays(cur, -1)
	}
	return n
}

// LongestStreak — максимальная серия активных дней за всё время.
func LongestStreak(marks []*model.DayMark) int {
	dates := make([]string, 0, len(marks))
	for _, m := range marks {
		if ActiveDay(m) {
			dates = append(dates, m.Date)
		}
	}
	sort.Strings(dates)
	best, run := 0, 0
	var prev time.Time
	for _, ds := range dates {
		d, err := model.ParseDate(ds)
		if err != nil {
			continue
		}
		if run > 0 && model.DayDiff(prev, d) == 1 {
			run++
		} else {
			run = 1
		}
		if run > best {
			best = run
		}
		prev = d
	}
	return best
}

func ymd(t time.Time) string { return t.Format(model.DateLayout) }
