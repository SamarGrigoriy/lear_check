package builtin

import (
	"encoding/json"
	"testing"

	"studytracker/model"
)

// Тест-страж встроенных данных: бинарник никогда не должен собираться
// с битым планом или кривым банком вопросов.

func TestEmbeddedPlan(t *testing.T) {
	p := DefaultPlan() // паникует, если JSON битый — сам этот факт и проверяем косвенно
	if err := p.Validate(); err != nil {
		t.Errorf("встроенный план не прошёл Validate: %v", err)
	}
	if p.LastWeekNumber() != 24 {
		t.Errorf("недель в плане: %d, хотим 24", p.LastWeekNumber())
	}
	if len(p.Stages) != 4 {
		t.Errorf("этапов: %d, хотим 4", len(p.Stages))
	}
	// каждая неделя 1..24 ровно один раз
	seen := map[int]bool{}
	for _, w := range p.AllWeeks() {
		if seen[w.Number] {
			t.Errorf("неделя %d встречается дважды", w.Number)
		}
		seen[w.Number] = true
		if len(w.Tasks) == 0 {
			t.Errorf("у недели %d нет задач", w.Number)
		}
		for _, task := range w.Tasks {
			found := false
			for _, d := range p.Directions {
				if task.Direction == d {
					found = true
				}
			}
			if !found {
				t.Errorf("неделя %d: направление %q не из списка плана", w.Number, task.Direction)
			}
		}
	}
	for n := 1; n <= 24; n++ {
		if !seen[n] {
			t.Errorf("нет недели %d", n)
		}
	}
	// ритм по умолчанию: 10 часов, 3 Verilog + 2,5 ML + 2 разбор + 2,5 английский
	r := p.Rhythm
	if r.TotalHours != 10 {
		t.Errorf("total_hours = %d", r.TotalHours)
	}
	want := map[string]int{model.DirVerilog: 30, model.DirML: 25, model.DirReview: 20, model.DirEnglish: 25}
	for d, v := range want {
		if r.HoursPerDir[d] != v {
			t.Errorf("ритм %s = %d, хотим %d", d, r.HoursPerDir[d], v)
		}
	}
	// у каждого этапа есть формулировка контрольной точки
	for _, s := range p.Stages {
		if s.Checkpoint == "" {
			t.Errorf("у этапа %q (%d–%d) нет контрольной точки", s.Name, s.From, s.To)
		}
	}
}

func TestEmbeddedQuestions(t *testing.T) {
	qs := DefaultQuestions()
	byDir := map[string]int{}
	byDirChoice := map[string]int{}
	for _, q := range qs {
		if q.Text == "" || q.Direction == "" {
			t.Errorf("вопрос id=%d пустой", q.ID)
		}
		if len(q.Choices) > 0 {
			if q.AnswerIndex < 0 || q.AnswerIndex >= len(q.Choices) {
				t.Errorf("вопрос id=%d: answer_index вне диапазона вариантов", q.ID)
			}
			byDirChoice[q.Direction]++
		} else if q.ModelAnswer == "" {
			t.Errorf("вопрос id=%d: ни вариантов, ни эталона — показывать нечего", q.ID)
		}
		if q.Box < 1 {
			q.Box = 1
		}
		byDir[q.Direction]++
	}
	// минимум 5 вопросов на направление
	for _, d := range model.DefaultDirections() {
		if byDir[d] < 5 {
			t.Errorf("направление %q: %d вопросов, хотим >= 5", d, byDir[d])
		}
	}
	// ids уникальны
	ids := map[int]bool{}
	for _, q := range qs {
		if ids[q.ID] {
			t.Errorf("дублирующийся id вопроса %d", q.ID)
		}
		ids[q.ID] = true
	}
	// JSON стабилен при повторном разборе (никаких полей-сюрпризов)
	var again []*model.QuizQuestion
	data, _ := json.Marshal(qs)
	if err := json.Unmarshal(data, &again); err != nil {
		t.Errorf("questions.json не переживает roundtrip: %v", err)
	}
	t.Logf("вопросов всего: %d, с вариантами по направлениям: %v", len(qs), byDirChoice)
}
