package quiz

import (
	"math/rand"
	"testing"
	"time"

	"studytracker/model"
)

func mustDate(s string) time.Time {
	d, err := model.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestRecordAnswerWrongResetsBox(t *testing.T) {
	q := &model.QuizQuestion{ID: 1, Box: 4}
	today := mustDate("2026-05-01")
	if RecordAnswer(q, false, today) {
		t.Error("ожидали false для неверного ответа")
	}
	if q.Box != 1 {
		t.Errorf("после ошибки коробка = %d, хотим 1", q.Box)
	}
	if q.WrongCount != 1 || q.AskedCount != 1 {
		t.Errorf("счётчики: asked=%d wrong=%d", q.AskedCount, q.WrongCount)
	}
	// ошибка -> коробка 1 -> показать завтра
	if q.NextReview != "2026-05-02" {
		t.Errorf("NextReview = %s, хотим 2026-05-02", q.NextReview)
	}
}

func TestRecordAnswerCorrectPromotesOncePerDay(t *testing.T) {
	q := &model.QuizQuestion{ID: 1, Box: 1}
	today := mustDate("2026-05-01")
	RecordAnswer(q, true, today) // 1 -> 2
	if q.Box != 2 {
		t.Fatalf("Box = %d, хотим 2", q.Box)
	}
	// второй верный ответ в тот же день НЕ должен давать ещё одно повышение —
	// иначе вопрос за день улетает в верхнюю коробку. Проверяем интервал:
	// из коробки 2 следующий показ через 2 дня.
	if q.NextReview != "2026-05-03" {
		t.Errorf("NextReview = %s, хотим 2026-05-03 (box2 = +2 дня)", q.NextReview)
	}
	// на верхней коробке не переполняемся
	top := &model.QuizQuestion{ID: 2, Box: MaxBox}
	RecordAnswer(top, true, today)
	if top.Box != MaxBox {
		t.Errorf("Box на потолке = %d", top.Box)
	}
}

func TestDue(t *testing.T) {
	today := mustDate("2026-06-10")
	fresh := &model.QuizQuestion{ID: 1} // новый вопрос без даты — созрел
	if !Due(fresh, today) {
		t.Error("новый вопрос должен быть Due")
	}
	past := &model.QuizQuestion{ID: 2, NextReview: "2026-06-09"}
	if !Due(past, today) {
		t.Error("просроченный вопрос должен быть Due")
	}
	future := &model.QuizQuestion{ID: 3, NextReview: "2026-06-11"}
	if Due(future, today) {
		t.Error("будущий вопрос не должен быть Due")
	}
	bad := &model.QuizQuestion{ID: 4, NextReview: "не дата"}
	if !Due(bad, today) {
		t.Error("вопрос с битой датой должен показываться, а не теряться")
	}
}

func mkDB() *model.DB {
	db := model.NewDB()
	add := func(dir string, week, box int, nextRev, modelAns string) *model.QuizQuestion {
		q := &model.QuizQuestion{
			ID: db.NextRecID(), Direction: dir, Week: week, Text: "q?",
			Box: box, NextReview: nextRev, ModelAnswer: modelAns,
		}
		if modelAns == "" {
			q.Choices = []string{"a", "b"}
		}
		db.Questions = append(db.Questions, q)
		return q
	}
	// 2 открытых вопроса по ML (неделя 1): один созрел, другой нет
	add(model.DirML, 1, 1, "", "")        // due (next_review пусто)
	add(model.DirML, 1, 3, "2099-01-01", "") // не due
	// открытый «своими словами» по ML
	add(model.DirML, 1, 1, "", "эталон")
	// Verilog, другая неделя
	add(model.DirVerilog, 2, 1, "", "")
	return db
}

func TestPickQuestionsFilters(t *testing.T) {
	db := mkDB()
	rnd := rand.New(rand.NewSource(42))
	today := mustDate("2026-06-10")

	// тест по неделе 1, только с вариантами: 2 подходящих (оба ML week1 choice),
	// но второй не созрел — всё равно попадёт при нехватке созревших
	got := PickQuestions(db, 1, "", false, 10, today, rnd)
	if len(got) != 2 {
		t.Fatalf("week=1 choice: %d вопросов, хотим 2", len(got))
	}
	for _, q := range got {
		if q.ModelAnswer != "" {
			t.Error("в режим с вариантами попал открытый вопрос")
		}
	}

	// направление Verilog
	got = PickQuestions(db, 0, model.DirVerilog, false, 10, today, rnd)
	if len(got) != 1 || got[0].Direction != model.DirVerilog {
		t.Errorf("по направлению Verilog: %+v", got)
	}

	// только «объясни своими словами»
	got = PickQuestions(db, 0, "", true, 10, today, rnd)
	if len(got) != 1 || got[0].ModelAnswer == "" {
		t.Errorf("открытый режим: %+v", got)
	}

	// лимит
	got = PickQuestions(db, 0, "", false, 1, today, rnd)
	if len(got) != 1 {
		t.Errorf("limit=1 вернул %d", len(got))
	}

	// созревшие идут первыми: первый вопрос — из коробки 1 (due), а не из 3 (not due)
	got = PickQuestions(db, 1, "", false, 2, today, rnd)
	if got[0].NextReview != "" {
		t.Errorf("первым должен идти созревший вопрос, получили NextReview=%q", got[0].NextReview)
	}
}

func TestGradeToCorrect(t *testing.T) {
	if !GradeToCorrect(1) || GradeToCorrect(2) || GradeToCorrect(3) {
		t.Error("GradeToCorrect: только 1 (понял) считается правильным")
	}
}

func TestComputeStatsAndWeak(t *testing.T) {
	attempts := []*model.QuizAttempt{
		{Direction: model.DirML, Week: 1, Correct: true},
		{Direction: model.DirML, Week: 1, Correct: true},
		{Direction: model.DirML, Week: 2, Correct: false},
		{Direction: model.DirVerilog, Week: 1, Correct: false},
		{Direction: model.DirVerilog, Week: 1, Correct: false},
		{Direction: model.DirEnglish, Week: 1, Correct: true}, // мало ответов — не «слабая»
	}
	st := ComputeStats(attempts)
	if got := st.ByDir[model.DirML].Rate(); got < 66.6 || got > 66.8 {
		t.Errorf("ML rate = %v, хотим ~66.7", got)
	}
	if st.ByWeek[1].Asked != 5 || st.ByWeek[2].Asked != 1 {
		t.Errorf("по неделям: %+v", st.ByWeek)
	}
	if st.ByDir["нет такого"].Asked != 0 {
		t.Error("несуществующее направление должно давать ноль, а не панику")
	}
	weak := st.WeakTopics(60, 2) // порог 60%, минимум 2 ответа
	// ML: 2/3 = 66.7% >= 60 — НЕ слабая; Verilog: 0% — слабая; English: 1 ответ — мало
	if len(weak) != 1 || weak[0] != model.DirVerilog {
		t.Errorf("WeakTopics = %v, хотим [Verilog/цифра]", weak)
	}
}
