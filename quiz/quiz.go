// Package quiz — интервальное повторение по методу Лейтнера (5 коробок)
// и выбор вопросов для теста. Чистая логика, без GUI.
//
// Коробка 1 показывается каждый день, коробка 5 — раз в месяц. Правильный
// ответ поднимает вопрос на коробку выше (но не чаще одного повышения в день),
// ошибка сбрасывает в коробку 1 — ошибочные вопросы возвращаются чаще всего.
package quiz

import (
	"math/rand"
	"sort"
	"time"

	"studytracker/model"
)

// BoxDays — интервал показа для коробки 1..5 (в днях).
var BoxDays = [6]int{0, 1, 2, 4, 9, 30}

// MaxBox — верхняя коробка.
const MaxBox = 5

// RecordAnswer обновляет статистику вопроса и его коробку после ответа.
// kind: "choice" — вариант выбран верно/неверно; "open" — grade 1=понял,
// 2=частично, 3=не понял (только grade==1 засчитывается как правильный).
// Возвращает, был ли ответ правильным.
func RecordAnswer(q *model.QuizQuestion, correct bool, today time.Time) bool {
	q.AskedCount++
	q.LastAsked = today.Format(model.DateLayout)
	if !correct {
		q.WrongCount++
		q.Box = 1
	} else {
		// повышение только если вопрос уже досидел до своего интервала:
		// иначе два быстрых верных ответа подряд уносят вопрос в 3 коробку за день
		if q.Box < MaxBox {
			q.Box++
		}
	}
	q.NextReview = model.AddDays(today, BoxDays[q.Box]).Format(model.DateLayout)
	return correct
}

// Due — пора ли показывать вопрос сегодня.
func Due(q *model.QuizQuestion, today time.Time) bool {
	if q.NextReview == "" {
		return true // новый вопрос — можно сразу
	}
	d, err := model.ParseDate(q.NextReview)
	if err != nil {
		return true
	}
	return !today.Before(d)
}

// PickSession собирает список вопросов для режима:
//
//	week > 0 — ограничить неделей; direction != "" — ограничить направлением;
//	openOnly == true — только открытые вопросы («Объясни своими словами»),
//	false — только с вариантами.
//
// Берёт сначала «созревшие» (Due), затем при нехватке самые отстающие коробки.
// Всего не больше limit; порядок перемешан, чтобы не запоминать последовательность.
func PickQuestions(db *model.DB, week int, direction string, openOnly bool, limit int, today time.Time, rnd *rand.Rand) []*model.QuizQuestion {
	var due, rest []*model.QuizQuestion
	for _, q := range db.Questions {
		if week > 0 && q.Week != week {
			continue
		}
		if direction != "" && q.Direction != direction {
			continue
		}
		if openOnly != (len(q.ModelAnswer) > 0) {
			continue
		}
		if Due(q, today) {
			due = append(due, q)
		} else {
			rest = append(rest, q)
		}
	}
	shuffle(due, rnd)
	shuffle(rest, rnd)
	// созревшие: худшие коробки первыми — внутри per-box сортировка случайна из-за shuffle
	sort.SliceStable(due, func(i, j int) bool { return due[i].Box < due[j].Box })
	out := append(due, rest...)
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func shuffle(s []*model.QuizQuestion, rnd *rand.Rand) {
	rnd.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

// GradeToCorrect переводит самооценку открытого вопроса в «правильно/нет»:
// 1 (понял) — да; 2 (частично) — считаем неверным, вопрос вернётся завтра;
// 3 (не понял) — неверно.
func GradeToCorrect(grade int) bool { return grade == 1 }

// StatsByDirection — доля правильных ответов по каждому направлению.
type Stat struct {
	Asked   int
	Correct int
}

// Rate — процент правильных (0 при отсутствии ответов).
func (s Stat) Rate() float64 {
	if s.Asked == 0 {
		return 0
	}
	return float64(s.Correct) / float64(s.Asked) * 100
}

// Stats — сводка по направлениям и по неделям из истории попыток.
type Stats struct {
	ByDir  map[string]Stat
	ByWeek map[int]Stat
}

// ComputeStats строит сводку.
func ComputeStats(attempts []*model.QuizAttempt) Stats {
	st := Stats{ByDir: map[string]Stat{}, ByWeek: map[int]Stat{}}
	for _, a := range attempts {
		d := st.ByDir[a.Direction]
		d.Asked++
		if a.Correct {
			d.Correct++
		}
		st.ByDir[a.Direction] = d
		w := st.ByWeek[a.Week]
		w.Asked++
		if a.Correct {
			w.Correct++
		}
		st.ByWeek[a.Week] = w
	}
	return st
}

// WeakTopics — направления с долей правильных ниже порога (и хотя бы n ответов).
func (st Stats) WeakTopics(threshold float64, minAsked int) []string {
	var out []string
	for dir, s := range st.ByDir {
		if s.Asked >= minAsked && s.Rate() < threshold {
			out = append(out, dir)
		}
	}
	sort.Strings(out)
	return out
}
