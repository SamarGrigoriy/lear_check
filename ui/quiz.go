// quiz.go — экран «Тесты»: режимы «Тест по неделе», «Случайный по направлению»
// и «Объясни своими словами», интервальное повторение (Лейтнер), редактор
// вопросов и сводка статистики.
package ui

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"studytracker/model"
	"studytracker/quiz"

	"nui"
)

type pageQuiz struct {
	app  *App
	root *nui.Box

	// панель запуска
	modeSel   *nui.Combo // выбор режима
	pickWeek  *nui.Combo // неделя для теста по неделе
	pickDir   *nui.Combo // направление для случайного
	openCombo *nui.Combo // открытый вопрос по ... (направление или неделя)
	area      *nui.Box   // сюда встраивается зона прогона / статистики

	// состояние текущей сессии
	session []*model.QuizQuestion
	pos     int
	rnd     *rand.Rand
}

func newPageQuiz(a *App) *pageQuiz {
	p := &pageQuiz{app: a, rnd: rand.New(rand.NewSource(a.today.UnixNano()))}

	p.modeSel = nui.NewCombo("Тест по неделе", "Случайный по направлению", "Объясни своими словами", "Повторение (созревшие)")
	p.modeSel.SetSelected(0)

	p.pickWeek = nui.NewCombo("Неделя")
	p.pickWeek.Clear()
	p.pickWeek.Add("Все недели")
	for i := 1; i <= a.DB.Plan.LastWeekNumber(); i++ {
		p.pickWeek.Add(strconv.Itoa(i))
	}
	p.pickWeek.SetSelected(0)

	p.pickDir = nui.NewCombo("Направление")
	p.pickDir.Add("Все направления")
	for _, d := range a.DB.Plan.Directions {
		p.pickDir.Add(d)
	}
	p.pickDir.SetSelected(0)

	p.openCombo = nui.NewComboEntry("")
	start := nui.NewButton("Начать", p.startSession)
	statsBtn := nui.NewButton("Статистика", p.showStats)
	editBtn := nui.NewButton("Редактор вопросов…", p.openEditor)

	top := nui.VBox(
		nui.HBox(nui.NewLabel("Режим:"), nui.Expand(p.modeSel)).Spacing(6),
		nui.HBox(nui.NewLabel("Неделя:"), p.pickWeek, nui.NewLabel("Направление:"), p.pickDir,
			nui.NewLabel("Открытые по:"), nui.Expand(p.openCombo)).Spacing(6),
		nui.HBox(start, statsBtn, editBtn).Spacing(6),
	).Spacing(6)

	p.area = nui.VBox(nui.NewLabel("Выберите режим и нажмите «Начать».")).Spacing(6).Padding(4)

	p.root = nui.VBox(top, nui.NewSeparator(), nui.Expand(nui.NewScroll(p.area))).Padding(8)
	return p
}

// startSession собирает список вопросов по режиму и начинает прогон.
func (p *pageQuiz) startSession() {
	a := p.app
	mode := p.modeSel.Text()
	var qs []*model.QuizQuestion

	week := 0
	if s := strings.TrimSpace(p.pickWeek.Text()); s != "" && s != "Все недели" {
		week, _ = strconv.Atoi(s)
	}
	dir := p.pickDir.Text()
	if dir == "Все направления" {
		dir = ""
	}

	switch mode {
	case "Тест по неделе":
		qs = quiz.PickQuestions(a.DB, week, "", false, 10, a.today, p.rnd)
	case "Случайный по направлению":
		qs = quiz.PickQuestions(a.DB, 0, dir, false, 10, a.today, p.rnd)
	case "Объясни своими словами":
		oc := strings.TrimSpace(p.openCombo.Text())
		ow, od := 0, ""
		if n, err := strconv.Atoi(oc); err == nil {
			ow = n
		} else if oc != "" {
			od = oc
		}
		qs = quiz.PickQuestions(a.DB, ow, od, true, 5, a.today, p.rnd)
	case "Повторение (созревшие)":
		// все созревшие вопросы, независимо от режима
		for _, q := range a.DB.Questions {
			if quiz.Due(q, a.today) {
				qs = append(qs, q)
			}
		}
		p.rnd.Shuffle(len(qs), func(i, j int) { qs[i], qs[j] = qs[j], qs[i] })
		if len(qs) > 12 {
			qs = qs[:12]
		}
	}

	if len(qs) == 0 {
		nui.Message(a.Win, "Тесты", "Под фильтр нет ни одного вопроса. Добавьте вопросы в редакторе.")
		return
	}
	p.session = qs
	p.pos = 0
	p.askOne()
}

// askOne рисует один вопрос текущего прогона.
func (p *pageQuiz) askOne() {
	a := p.app
	p.area.Clear()
	if p.pos >= len(p.session) {
		p.area.Add(nui.NewLabel(fmt.Sprintf("Готово: %d вопросов за сессию.", len(p.session))))
		p.area.Add(nui.NewButton("Ещё раз", p.startSession))
		return
	}
	q := p.session[p.pos]
	head := nui.NewLabel(fmt.Sprintf("Вопрос %d/%d · %s · нед. %d",
		p.pos+1, len(p.session), q.Direction, q.Week))
	head.SetFontScale(0.9)
	text := nui.NewTextView()
	text.SetReadOnly(true)
	text.SetWrap(true)
	text.SetText(q.Text)
	text.SetMinSize(-1, 70)
	p.area.Add(head)
	p.area.Add(text)

	if len(q.Choices) > 0 {
		// вопрос с вариантами: одна радиогруппа на все варианты
		radio := nui.NewRadio(q.Choices...)
		p.area.Add(radio)
		check := nui.NewButton("Ответить", func() {
			sel := radio.Selected()
			if sel < 0 {
				nui.Message(a.Win, "Тесты", "Выберите вариант ответа.")
				return
			}
			correct := sel == q.AnswerIndex
			p.record(q, correct)
			p.showFeedback(correct, q)
		})
		p.area.Add(nui.HBox(check, nui.NewButton("Пропустить", p.next)))
		return
	}

	// открытый вопрос: «объясни своими словами»
	answerV := nui.NewTextView()
	answerV.SetWrap(true)
	answerV.SetMinSize(-1, 90)
	p.area.Add(nui.NewLabel("Ваш ответ:"))
	p.area.Add(answerV)
	show := nui.NewButton("Показать эталон", func() {
		m := nui.NewTextView()
		m.SetReadOnly(true)
		m.SetWrap(true)
		m.SetText("Эталон:\n" + q.ModelAnswer)
		p.area.Add(m)
		row := nui.NewLabel("Понял?")
		g1 := nui.NewButton("Понял", func() { p.record(q, true); p.next() })
		g2 := nui.NewButton("Частично", func() { p.record(q, false); p.next() })
		g3 := nui.NewButton("Не понял", func() { p.record(q, false); p.next() })
		p.area.Add(nui.HBox(row, g1, g2, g3).Spacing(6))
	})
	p.area.Add(nui.HBox(show, nui.NewButton("Пропустить", p.next)).Spacing(6))
}

// record заносит ответ: коробка Лейтнера + история попыток + сохранение.
func (p *pageQuiz) record(q *model.QuizQuestion, correct bool) {
	a := p.app
	quiz.RecordAnswer(q, correct, a.today)
	a.DB.Attempts = append(a.DB.Attempts, &model.QuizAttempt{
		Date: a.TodayStr(), QuestionID: q.ID, Direction: q.Direction, Week: q.Week, Correct: correct,
	})
	a.Save()
}

// showFeedback — правильный ответ и пояснение, кнопка «Дальше».
func (p *pageQuiz) showFeedback(correct bool, q *model.QuizQuestion) {
	res := nui.NewLabel("Верно!")
	if !correct {
		res = nui.NewLabel("Неверно.")
		res.SetColor(nui.RGB(200, 60, 50))
	}
	p.area.Add(res)
	right := nui.NewTextView()
	right.SetReadOnly(true)
	right.SetWrap(true)
	txt := "Правильный ответ: " + q.Choices[q.AnswerIndex]
	if q.Explanation != "" {
		txt += "\n\n" + q.Explanation
	}
	right.SetText(txt)
	p.area.Add(right)
	p.area.Add(nui.NewButton("Дальше →", p.next))
}

func (p *pageQuiz) next() {
	p.pos++
	p.askOne()
}

// showStats — процент правильных по направлениям и неделям + слабые темы.
func (p *pageQuiz) showStats() {
	st := quiz.ComputeStats(p.app.DB.Attempts)
	p.area.Clear()
	p.area.Add(nui.NewLabel("Статистика тестов"))
	dirs := []string{}
	for _, d := range p.app.DB.Plan.Directions {
		if s, ok := st.ByDir[d]; ok {
			dirs = append(dirs, d)
			p.area.Add(nui.NewLabel(fmt.Sprintf("%s: %d%% (%d из %d)", d, int(s.Rate()+0.5), s.Correct, s.Asked)))
		}
	}
	for _, d := range dirs {
		delete(st.ByDir, d) // чтобы не печатать дважды
	}
	for d, s := range st.ByDir { // направления, добавленные позже
		p.area.Add(nui.NewLabel(fmt.Sprintf("%s: %d%% (%d из %d)", d, int(s.Rate()+0.5), s.Correct, s.Asked)))
	}
	if weak := st.WeakTopics(70, 3); len(weak) > 0 {
		p.area.Add(nui.NewLabel("Слабые темы: " + strings.Join(weak, ", ")))
	}
	p.area.Add(nui.NewButton("← К тестам", p.startSession))
}

// openEditor — таблица всех вопросов с добавлением/правкой/удалением.
func (p *pageQuiz) openEditor() {
	a := p.app
	var list *nui.List
	rows := func() []*model.QuizQuestion { return a.DB.Questions }
	list = nui.NewList("ID", "Направление", "Нед.", "Вопрос", "Коробка")
	redraw := func() {
		qs := rows()
		list.SetRows(len(qs), func(r, c int) string {
			q := qs[r]
			switch c {
			case 0:
				return strconv.Itoa(q.ID)
			case 1:
				return q.Direction
			case 2:
				return strconv.Itoa(q.Week)
			case 3:
				return firstLine(q.Text)
			default:
				return strconv.Itoa(q.Box)
			}
		})
	}
	redraw()
	sel := func() *model.QuizQuestion {
		i := list.Selected()
		qs := rows()
		if i < 0 || i >= len(qs) {
			return nil
		}
		return qs[i]
	}
	add := nui.NewButton("Добавить…", func() { p.editQuestion(nil, redraw) })
	mod := nui.NewButton("Изменить…", func() {
		if q := sel(); q != nil {
			p.editQuestion(q, redraw)
		}
	})
		q := sel()
		if q == nil {
			return
		}
		if !nui.Ask(a.Win, "Удалить вопрос?", firstLine(q.Text)) {
			return
		}
		for i, x := range a.DB.Questions {
			if x.ID == q.ID {
				a.DB.Questions = append(a.DB.Questions[:i], a.DB.Questions[i+1:]...)
				break
			}
		}
		a.Save()
		redraw()
	})
	list.OnActivate(func(int) { mod.Click() })

	dlg := nui.NewWindow("Банк вопросов", nui.VBox(
		nui.Expand(list), nui.HBox(add, mod, del).Spacing(6),
	).Spacing(6).Padding(8))
	dlg.SetSize(760, 520)
	dlg.Show()
}

// editQuestion — модалка одного вопроса. existing == nil — новый.
func (p *pageQuiz) editQuestion(existing *model.QuizQuestion, done func()) {
	a := p.app
	isChoice := existing == nil || len(existing.Choices) > 0

	kindSel := nui.NewCombo("С вариантами", "Объясни своими словами")
	if !isChoice {
		kindSel.SetSelected(1)
	} else {
		kindSel.SetSelected(0)
	}

	dirSel := nui.NewCombo()
	for _, d := range a.DB.Plan.Directions {
		dirSel.Add(d)
	}
	if existing != nil {
		dirSel.SetText(existing.Direction)
	} else {
		dirSel.SetSelected(0)
	}

	weekSp := nui.NewSpin(1, a.DB.Plan.LastWeekNumber())
	if existing != nil && existing.Week > 0 {
		weekSp.SetValue(existing.Week)
	} else {
		weekSp.SetValue(a.CurrentWeek())
	}

	textV := nui.NewTextView()
	textV.SetWrap(true)
	explV := nui.NewTextView()
	explV.SetWrap(true)
	modelV := nui.NewTextView() // эталон для открытого вопроса
	modelV.SetWrap(true)
	if existing != nil {
		textV.SetText(existing.Text)
		explV.SetText(existing.Explanation)
		modelV.SetText(existing.ModelAnswer)
	}

	// варианты ответов: 4 поля ввода + галочка «верный» у каждого.
	// Галочки в gonui независимы, поэтому держим ровно одну включённой сами.
	choiceGrid := nui.NewGrid()
	var entries [4]*nui.Entry
	var marks [4]*nui.Check
	for i := 0; i < 4; i++ {
		e := nui.NewEntry()
		c := nui.NewCheck("верный")
		entries[i] = e
		marks[i] = c
		idx := i
		c.OnChange(func() {
			if !c.Checked() {
				return
			}
			for j, m := range marks {
				if j != idx {
					m.SetChecked(false)
				}
			}
		})
		choiceGrid.Add(0, i, e).Add(1, i, c)
	}
	if existing != nil {
		for i, ch := range existing.Choices {
			if i < 4 {
				entries[i].SetText(ch)
			}
		}
		if existing.AnswerIndex >= 0 && existing.AnswerIndex < 4 {
			marks[existing.AnswerIndex].SetChecked(true)
		}
	}

	openBox := nui.VBox(nui.NewLabel("Эталонный ответ:"), nui.Expand(modelV)).Spacing(4)
	setMode := func() {
		is := kindSel.Selected() == 0
		choiceGrid.SetVisible(is)
		openBox.SetVisible(!is)
	}
	kindSel.OnChange(setMode)
	setMode()

	var dialog *nui.Window
	ok := nui.NewButton("Сохранить", func() {
		if textV.Text() == "" {
			nui.Alert(dialog, "Пустой вопрос", "Введите текст вопроса.")
			return
		}
		if existing == nil {
			existing = &model.QuizQuestion{ID: a.DB.NextRecID(), Box: 1}
			a.DB.Questions = append(a.DB.Questions, existing)
		}
		existing.Direction = dirSel.Text()
		existing.Week = weekSp.Value()
		existing.Text = textV.Text()
		existing.Explanation = explV.Text()
		if kindSel.Selected() == 0 {
			var ch []string
			ans := -1
			for i, e := range entries {
				t := strings.TrimSpace(e.Text())
				if t == "" {
					continue
				}
				if marks[i].Checked() {
					ans = len(ch)
				}
				ch = append(ch, t)
			}
			if len(ch) < 2 {
				nui.Alert(dialog, "Мало вариантов", "Заполните минимум два варианта ответа.")
				return
			}
			if ans < 0 {
				nui.Alert(dialog, "Нет верного", "Отметьте галочкой правильный вариант.")
				return
			}
			existing.Choices = ch
			existing.AnswerIndex = ans
			existing.ModelAnswer = ""
		} else {
			if modelV.Text() == "" {
				nui.Alert(dialog, "Нет эталона", "Для открытого вопроса нужен эталонный ответ.")
				return
			}
			existing.Choices = nil
			existing.ModelAnswer = modelV.Text()
		}
		a.Save()
		dialog.Close()
		if done != nil {
			done()
		}
	})
	dialog = nui.NewWindow("Вопрос теста", nui.VBox(
		nui.NewGrid().
			Add(0, 0, nui.NewLabel("Тип:")).Add(1, 0, kindSel).
			Add(0, 1, nui.NewLabel("Направление:")).Add(1, 1, dirSel).
			Add(0, 2, nui.NewLabel("Неделя:")).Add(1, 2, weekSp).
			Add(0, 3, nui.NewLabel("Вопрос:")).Add(1, 3, nui.Expand(textV)).
			Add(0, 4, nui.NewLabel("Пояснение:")).Add(1, 4, nui.Expand(explV)).
			ExpandCol(1).Spacing(6, 6),
		nui.NewLabel("Варианты (галочка — правильный):"), choiceGrid, openBox,
		nui.HBox(nui.Stretch(), nui.NewButton("Отмена", func() { dialog.Close() }), ok),
	).Spacing(6).Padding(8))
	dialog.SetSize(640, 640)
	dialog.SetDefault(ok)
	dialog.Show()
}

// refresh — после смены плана/импорта перечитать фильтры.
func (p *pageQuiz) refresh() {
	last := p.app.DB.Plan.LastWeekNumber()
	p.pickWeek.SetSelected(0)
	_ = last
}
