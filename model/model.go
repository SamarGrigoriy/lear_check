// Package model — типы данных трекера обучения. Чистый Go, без GUI и без cgo:
// этот пакет можно тестировать на любой машине.
//
// Все даты в модели хранятся строками "ГГГГ-ММ-ДД": так JSON читается глазами,
// файл стабилен между запусками и не зависит от часового пояса.
package model

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// DateLayout — формат дат в JSON и в интерфейсе.
const DateLayout = "2006-01-02"

// ParseDate разбирает строку "ГГГГ-ММ-ДД".
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateLayout, s)
}

// Today возвращает сегодняшнюю дату, усечённую до дня (в локальном поясе).
func Today() time.Time {
	return StartOfDay(time.Now())
}

// StartOfDay отбрасывает время, оставляя начало суток.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// AddDays прибавляет (или отнимает) n дней к дате.
func AddDays(t time.Time, n int) time.Time {
	return StartOfDay(t.AddDate(0, 0, n))
}

// DayDiff — целое число дней от a до b (b − a).
func DayDiff(a, b time.Time) int {
	return int(StartOfDay(b).Sub(StartOfDay(a)).Hours() / 24)
}

// Направления обучения. constant'ы — чтобы в коде не опечататься в русском слове.
const (
	DirVerilog = "Verilog/цифра"
	DirML      = "ML"
	DirReview  = "Разбор проекта"
	DirEnglish = "Английский"
	DirProject = "Проект/заработок"
)

// DefaultDirections — базовый набор направлений. Пользователь может расширить
// его полем directions в файле плана.
func DefaultDirections() []string {
	return []string{DirVerilog, DirML, DirReview, DirEnglish, DirProject}
}

// Статусы задачи.
const (
	TaskNew        = "не начата"
	TaskInProgress = "в процессе"
	TaskDone       = "сделана"
	TaskSkipped    = "пропущена"
)

// ValidTaskStatus проверяет строку статуса.
func ValidTaskStatus(s string) bool {
	switch s {
	case TaskNew, TaskInProgress, TaskDone, TaskSkipped:
		return true
	}
	return false
}

// Типы записей журнала.
const (
	NoteEntry    = "заметка"
	NoteWin      = "достижение"
	NoteFailure  = "неудача"
	NoteQuestion = "вопрос"
)

// ValidNoteType проверяет тип записи.
func ValidNoteType(s string) bool {
	switch s {
	case NoteEntry, NoteWin, NoteFailure, NoteQuestion:
		return true
	}
	return false
}

// WeekTask — задача недели в одном направлении.
type WeekTask struct {
	Description string `json:"description"`
	Direction   string `json:"direction"`
	Week        int    `json:"week"`
	Status      string `json:"status"`
	PlannedHrs  int    `json:"planned_hours"` // запланированные часы
	ActualHrs   int    `json:"actual_hours"`  // фактические часы
}

// WeekPlan — неделя плана из файла plan.json.
type WeekPlan struct {
	Number int        `json:"number"` // 1..24
	Goals  []string   `json:"goals"`  // цели недели
	Tasks  []WeekTask `json:"tasks"`  // задачи по направлениям
}

// StagePlan — этап: несколько недель + контрольная точка на выходе.
type StagePlan struct {
	Name       string     `json:"name"`
	From       int        `json:"from"`       // номер первой недели этапа
	To         int        `json:"to"`         // номер последней недели этапа
	Checkpoint string     `json:"checkpoint"` // формулировка контрольной точки
	Weeks      []WeekPlan `json:"weeks"`
}

// Rhythm — недельный ритм по умолчанию: сколько часов на направление.
// Часы дробные (2,5 часа), поэтому храним десятые доли: value = часы*10.
type Rhythm struct {
	TotalHours  int            `json:"total_hours"`         // всего часов в неделю, напр. 10
	HoursPerDir map[string]int `json:"hours_per_direction"` // десятые доли: 30 = 3 ч, 25 = 2,5 ч
}

// Hrs — человекочитаемое число часов по направлению ("3", "2,5").
func (r Rhythm) Hrs(dir string) string { return FormatHrs(r.HoursPerDir[dir]) }

// FormatHrs — форматирование десятых долей часа по-русски: 25 -> "2,5".
func FormatHrs(tenths int) string {
	whole := tenths / 10
	frac := tenths % 10
	if frac == 0 {
		return strconv.Itoa(whole)
	}
	return strconv.Itoa(whole) + "," + strconv.Itoa(frac)
}

// Plan — весь план: этапы, ритм, направления и дата начала.
// Дата начала живёт здесь же, чтобы экспорт/импорт оставались одним файлом;
// если план отредактирован вручную без неё, приложение подставит текущую.
type Plan struct {
	Title      string      `json:"title"`
	Start      string      `json:"start_date"`
	Directions []string    `json:"directions"`
	Rhythm     Rhythm      `json:"rhythm"`
	Stages     []StagePlan `json:"stages"`
}

// AllWeeks — все недели плана по порядку.
func (p *Plan) AllWeeks() []WeekPlan {
	var out []WeekPlan
	for _, st := range p.Stages {
		out = append(out, st.Weeks...)
	}
	return out
}

// Week — неделя по номеру (1-based), nil если нет.
func (p *Plan) Week(num int) *WeekPlan {
	for si := range p.Stages {
		ws := p.Stages[si].Weeks
		for wi := range ws {
			if ws[wi].Number == num {
				return &ws[wi]
			}
		}
	}
	return nil
}

// StageByWeek — этап, которому принадлежит неделя.
func (p *Plan) StageByWeek(num int) *StagePlan {
	for i := range p.Stages {
		s := &p.Stages[i]
		for _, w := range s.Weeks {
			if w.Number == num {
				return s
			}
		}
	}
	return nil
}

// LastWeekNumber — максимальный номер недели в плане.
func (p *Plan) LastWeekNumber() int {
	max := 0
	for _, s := range p.Stages {
		for _, w := range s.Weeks {
			if w.Number > max {
				max = w.Number
			}
		}
	}
	return max
}

// Validate — проверки целостности плана; возвращает первую найденную ошибку.
func (p *Plan) Validate() error {
	if len(p.Stages) == 0 {
		return errors.New("план пуст: нет ни одного этапа")
	}
	seen := map[int]bool{}
	for _, s := range p.Stages {
		if s.From < 1 || s.To < s.From {
			return errors.New("этап «" + s.Name + "»: неверные границы недель")
		}
		for _, w := range s.Weeks {
			if seen[w.Number] {
				return errors.New("дублируется неделя " + strconv.Itoa(w.Number))
			}
			seen[w.Number] = true
			if w.Number < s.From || w.Number > s.To {
				return errors.New("неделя " + strconv.Itoa(w.Number) + " вне границ этапа «" + s.Name + "»")
			}
			for _, t := range w.Tasks {
				if strings.TrimSpace(t.Description) == "" {
					return errors.New("у задачи недели " + strconv.Itoa(w.Number) + " пустое описание")
				}
			}
		}
	}
	return nil
}

// JournalEntry — запись журнала (заметка/достижение/неудача/вопрос).
type JournalEntry struct {
	ID      int      `json:"id"`
	Date    string   `json:"date"`
	Type    string   `json:"type"`
	Text    string   `json:"text"`
	Week    int      `json:"week,omitempty"`     // привязка к неделе плана, 0 — без привязки
	TaskIdx int      `json:"task_idx,omitempty"` // индекс задачи в неделе, 0 — без привязки
	Tags    []string `json:"tags,omitempty"`
	Cause   string   `json:"cause,omitempty"` // «почему случилось» (для неудач)
	Fix     string   `json:"fix,omitempty"`   // «что изменю» (для неудач)
}

// DayMark — отметка дня: по каким направлениям было занятие и сколько минут.
// Направление без ключа или с 0 минут означает «не было».
type DayMark struct {
	Date    string         `json:"date"`
	Minutes map[string]int `json:"minutes"`
}

// Has — было ли занятие по направлению.
func (d *DayMark) Has(dir string) bool { return d.Minutes[dir] > 0 }

// TotalMinutes — сумма минут по всем направлениям.
func (d *DayMark) TotalMinutes() int {
	sum := 0
	for _, m := range d.Minutes {
		sum += m
	}
	return sum
}

// Checkpoint — контрольная точка этапа: критерии, дата выполнения, самооценка.
type Checkpoint struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Criteria  string `json:"criteria"`
	Week      int    `json:"week"`                 // неделя этапа, к которой привязана точка
	DoneDate  string `json:"done_date,omitempty"`  // пусто — не выполнена
	SelfScore int    `json:"self_score,omitempty"` // самооценка 1..5
}

// QuizQuestion — вопрос теста. Два режима задаются наличием полей:
// Choices непусто — вопрос с вариантами; ModelAnswer непусто — «объясни своими словами».
type QuizQuestion struct {
	ID          int      `json:"id"`
	Direction   string   `json:"direction"`
	Week        int      `json:"week"`
	Text        string   `json:"text"`
	Choices     []string `json:"choices,omitempty"`
	AnswerIndex int      `json:"answer_index"`
	ModelAnswer string   `json:"model_answer,omitempty"` // эталон для самопроверки
	Explanation string   `json:"explanation,omitempty"`
	Box         int      `json:"box"`                   // коробка Leitner 1..5 (см. пакет quiz)
	NextReview  string   `json:"next_review,omitempty"` // когда снова показать
	AskedCount  int      `json:"asked_count"`
	WrongCount  int      `json:"wrong_count"`
	LastAsked   string   `json:"last_asked,omitempty"`
}

// QuizAttempt — история одного ответа: из неё считается статистика.
type QuizAttempt struct {
	Date       string `json:"date"`
	QuestionID int    `json:"question_id"`
	Direction  string `json:"direction"`
	Week       int    `json:"week"`
	Correct    bool   `json:"correct"`
}

// Settings — настройки приложения.
type Settings struct {
	ShiftPlan bool `json:"shift_plan_on_skip"` // сдвигать нумерацию недель при отставании
}

// DB — всё состояние приложения; сериализуется в data.json одним файлом.
type DB struct {
	Version   int             `json:"version"`
	Plan      *Plan           `json:"plan"`  // рабочий план (копия встроенного, правится пользователем)
	Tasks     []*WeekTask     `json:"tasks"` // плоский список задач со статусами и часами
	Entries   []*JournalEntry `json:"journal"`
	Marks     []*DayMark      `json:"marks"`
	Checks    []*Checkpoint   `json:"checkpoints"`
	Questions []*QuizQuestion `json:"questions"`
	Attempts  []*QuizAttempt  `json:"attempts"`
	Settings  Settings        `json:"settings"`
	NextID    int             `json:"next_id"` // счётчик id для записей/точек/вопросов
}

// NewDB создаёт пустую базу.
func NewDB() *DB {
	return &DB{Version: 1, NextID: 1}
}

// NextRecID выдаёт следующий уникальный id записи.
func (db *DB) NextRecID() int {
	id := db.NextID
	db.NextID++
	return id
}

// FindMark — отметка дня по дате или nil.
func (db *DB) FindMark(date string) *DayMark {
	for _, m := range db.Marks {
		if m.Date == date {
			return m
		}
	}
	return nil
}

// MarkOrCreate — отметка дня; если её нет, создаётся пустая и добавляется в базу.
func (db *DB) MarkOrCreate(date string) *DayMark {
	if m := db.FindMark(date); m != nil {
		return m
	}
	m := &DayMark{Date: date, Minutes: map[string]int{}}
	db.Marks = append(db.Marks, m)
	return m
}

// QuestionByID — вопрос по id или nil.
func (db *DB) QuestionByID(id int) *QuizQuestion {
	for _, q := range db.Questions {
		if q.ID == id {
			return q
		}
	}
	return nil
}
