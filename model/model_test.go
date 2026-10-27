package model

import "testing"

func TestParseDateAndArith(t *testing.T) {
	d, err := ParseDate("2026-02-27")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if got := AddDays(d, 1).Format(DateLayout); got != "2026-02-28" {
		t.Errorf("AddDays(+1) = %s", got)
	}
	// февраль 2026 не високосный: 27+2 = 1 марта
	if got := AddDays(d, 2).Format(DateLayout); got != "2026-03-01" {
		t.Errorf("AddDays через границу месяца = %s", got)
	}
	if got := DayDiff(d, AddDays(d, 5)); got != 5 {
		t.Errorf("DayDiff = %d, хотим 5", got)
	}
	if got := DayDiff(AddDays(d, 5), d); got != -5 {
		t.Errorf("DayDiff обратный = %d, хотим -5", got)
	}
	// время в течение суток не влияет на разницу дней: DayDiff усекает до дня
	if got := DayDiff(d, AddDays(d, 3).Add(20*timeHour)); got != 3 {
		t.Errorf("DayDiff с хвостом времени = %d, хотим 3", got)
	}
}

// timeHour — один час; отдельная константа, чтобы не тянуть пакет time в тесты арифметики.
const timeHour = 3600e9 // nanoseconds, равно time.Hour

func TestFormatHrs(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{{30, "3"}, {25, "2,5"}, {0, "0"}, {10, "1"}, {45, "4,5"}}
	for _, c := range cases {
		if got := FormatHrs(c.in); got != c.want {
			t.Errorf("FormatHrs(%d) = %q, хотим %q", c.in, got, c.want)
		}
	}
}

func TestValidStatusAndType(t *testing.T) {
	if !ValidTaskStatus(TaskDone) || ValidTaskStatus("сделанно") {
		t.Error("ValidTaskStatus работает неверно")
	}
	if !ValidNoteType(NoteFailure) || ValidNoteType("меморандум") {
		t.Error("ValidNoteType работает неверно")
	}
}

func TestPlanWeekLookup(t *testing.T) {
	p := &Plan{Stages: []StagePlan{
		{Name: "A", From: 1, To: 2, Weeks: []WeekPlan{
			{Number: 1, Tasks: []WeekTask{{Description: "t1"}}},
			{Number: 2},
		}},
		{Name: "B", From: 3, To: 3, Weeks: []WeekPlan{{Number: 3}}},
	}}
	if w := p.Week(3); w == nil || w.Number != 3 {
		t.Error("Week(3) не нашёл неделю")
	}
	if s := p.StageByWeek(2); s == nil || s.Name != "A" {
		t.Error("StageByWeek(2) вернул не тот этап")
	}
	if p.LastWeekNumber() != 3 {
		t.Errorf("LastWeekNumber = %d", p.LastWeekNumber())
	}
	if len(p.AllWeeks()) != 3 {
		t.Errorf("AllWeeks = %d", len(p.AllWeeks()))
	}
}

func TestPlanValidate(t *testing.T) {
	base := func() *Plan {
		return &Plan{Stages: []StagePlan{{Name: "A", From: 1, To: 2, Weeks: []WeekPlan{
			{Number: 1, Tasks: []WeekTask{{Description: "ok"}}},
			{Number: 2},
		}}}}
	}
	if err := base().Validate(); err != nil {
		t.Errorf("корректный план отвергнут: %v", err)
	}
	p := base()
	p.Stages[0].Weeks = append(p.Stages[0].Weeks, WeekPlan{Number: 1}) // дубль недели
	if err := p.Validate(); err == nil {
		t.Error("дублирование недели не поймано")
	}
	p = base()
	p.Stages[0].Weeks[1].Number = 9 // вне границ этапа
	if err := p.Validate(); err == nil {
		t.Error("неделя вне границ этапа не поймана")
	}
	p = base()
	p.Stages[0].Weeks[0].Tasks[0].Description = "  "
	if err := p.Validate(); err == nil {
		t.Error("пустое описание задачи не поймано")
	}
}

func TestMarkOrCreateAndTotals(t *testing.T) {
	db := NewDB()
	m := db.MarkOrCreate("2026-03-01")
	if len(db.Marks) != 1 {
		t.Fatalf("Marks = %d", len(db.Marks))
	}
	m.Minutes[DirVerilog] = 50
	again := db.MarkOrCreate("2026-03-01")
	if again != m {
		t.Error("MarkOrCreate создал дубликат вместо возврата существующей")
	}
	if got := again.TotalMinutes(); got != 50 {
		t.Errorf("TotalMinutes = %d", got)
	}
	if again.Has(DirML) {
		t.Error("Has по направлению без минут должен быть false")
	}
	if db.FindMark("2026-03-02") != nil {
		t.Error("FindMark нашёл несуществующую дату")
	}
}

func TestNextRecID(t *testing.T) {
	db := NewDB()
	if db.NextRecID() != 1 || db.NextRecID() != 2 {
		t.Error("NextRecID не инкрементируется")
	}
}
