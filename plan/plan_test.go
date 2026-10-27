package plan

import (
	"testing"
	"time"

	"studytracker/model"
)

// mkDB — база с планом, старт которого задан строкой; marks — даты занятий.
func mkDB(start string, shift bool, markDates ...string) *model.DB {
	db := model.NewDB()
	db.Plan = &model.Plan{
		Start: start,
		Stages: []model.StagePlan{{Name: "A", From: 1, To: 3, Weeks: []model.WeekPlan{
			{Number: 1}, {Number: 2}, {Number: 3},
		}}},
	}
	db.Settings.ShiftPlan = shift
	for _, d := range markDates {
		m := db.MarkOrCreate(d)
		m.Minutes[model.DirVerilog] = 30 // любое положительное число минут = занятие было
	}
	return db
}

func TestWeekOfDateNoShift(t *testing.T) {
	db := mkDB("2026-03-02", false) // понедельник
	if got := WeekOfDate(db, mustDate("2026-03-01")); got != 0 {
		t.Errorf("до старта: %d, хотим 0", got)
	}
	if got := WeekOfDate(db, mustDate("2026-03-02")); got != 1 {
		t.Errorf("день старта: %d, хотим 1", got)
	}
	if got := WeekOfDate(db, mustDate("2026-03-08")); got != 1 {
		t.Errorf("последний день недели 1: %d", got)
	}
	if got := WeekOfDate(db, mustDate("2026-03-09")); got != 2 {
		t.Errorf("первый день недели 2: %d", got)
	}
	// за концом плана — прилипаем к последней неделе
	if got := WeekOfDate(db, mustDate("2027-01-01")); got != 3 {
		t.Errorf("после конца плана: %d, хотим 3", got)
	}
}

func TestWeekOfDateWithShift(t *testing.T) {
	// старт в понедельник; занятия только в календарных неделях 1 и 3
	db := mkDB("2026-03-02", true, "2026-03-03", "2026-03-17")
	// нед. плана 1 = календарная нед. 1 (занятие есть)
	if got := WeekOfDate(db, mustDate("2026-03-05")); got != 1 {
		t.Errorf("активная нед. 1: %d, хотим 1", got)
	}
	// календарная нед. 2 без занятий — пропускается, номер не расходуется
	if got := WeekOfDate(db, mustDate("2026-03-10")); got != 1 {
		t.Errorf("пустая неделя после нед.1: %d, хотим 1 (ещё не дошли до нед.2)", got)
	}
	// календарная нед. 3 с занятием — это плановая неделя 2
	if got := WeekOfDate(db, mustDate("2026-03-18")); got != 2 {
		t.Errorf("неделя после пропуска: %d, хотим 2", got)
	}
}

func TestWeekStartAndDays(t *testing.T) {
	db := mkDB("2026-03-02", false)
	ws := WeekStart(db, 2)
	if ws.Format(model.DateLayout) != "2026-03-09" {
		t.Errorf("WeekStart(2) = %s", ws.Format(model.DateLayout))
	}
	days := WeekDays(db, 1)
	if len(days) != 7 || days[6].Format(model.DateLayout) != "2026-03-08" {
		t.Errorf("WeekDays(1) неверны: %v", days)
	}
}

func TestWeekStatus(t *testing.T) {
	tasks := []*model.WeekTask{
		{Week: 1, Status: model.TaskDone},
		{Week: 1, Status: model.TaskInProgress},
		{Week: 2, Status: model.TaskNew},
		{Week: 3, Status: model.TaskDone},
		{Week: 3, Status: model.TaskSkipped},
	}
	if got := WeekStatus(tasks, 1); got != "active" {
		t.Errorf("неделя 1: %s, хотим active", got)
	}
	if got := WeekStatus(tasks, 2); got != "new" {
		t.Errorf("неделя 2: %s, хотим new", got)
	}
	if got := WeekStatus(tasks, 3); got != "done" {
		t.Errorf("неделя 3: %s, хотим done", got)
	}
	if got := WeekStatus(tasks, 9); got != "new" {
		t.Errorf("несуществующая неделя: %s, хотим new", got)
	}
}

func mark(date string, mins int) *model.DayMark {
	m := &model.DayMark{Date: date, Minutes: map[string]int{}}
	if mins > 0 {
		m.Minutes[model.DirML] = mins
	}
	return m
}

func TestCurrentStreak(t *testing.T) {
	marks := []*model.DayMark{
		mark("2026-04-10", 30), mark("2026-04-11", 30), mark("2026-04-12", 30),
		mark("2026-04-08", 30), // разрыв перед серией
		mark("2026-04-09", 0),  // нулевые минуты — не занятие
	}
	today := mustDate("2026-04-12")
	if got := CurrentStreak(marks, today); got != 3 {
		t.Errorf("серия на сегодня: %d, хотим 3", got)
	}
	// сегодня не отмечено, но вчера было — серия продолжается (вчерашняя)
	if got := CurrentStreak(marks, mustDate("2026-04-13")); got != 3 {
		t.Errorf("серия на следующий день: %d, хотим 3", got)
	}
	// пропуск дня обрывает серию
	if got := CurrentStreak(marks, mustDate("2026-04-14")); got != 0 {
		t.Errorf("серия после пропуска: %d, хотим 0", got)
	}
}

func TestLongestStreak(t *testing.T) {
	marks := []*model.DayMark{
		mark("2026-04-01", 10), mark("2026-04-02", 10), mark("2026-04-03", 10),
		mark("2026-04-05", 10), mark("2026-04-06", 10),
		mark("2026-04-04", 0), // пустой день разрывает серию
	}
	if got := LongestStreak(marks); got != 3 {
		t.Errorf("LongestStreak = %d, хотим 3", got)
	}
	if got := LongestStreak(nil); got != 0 {
		t.Errorf("пустые отметки: %d, хотим 0", got)
	}
}

func mustDate(s string) time.Time {
	d, err := model.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}
