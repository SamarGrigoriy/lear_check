// today.go — экран «Сегодня»: текущая неделя, чекбоксы задач, отметка занятия,
// серия дней и быстрая заметка.

package ui

import (
	"fmt"

	"studytracker/model"
	"studytracker/plan"
	"studytracker/storage"

	"nui"
)

type pageToday struct {
	app       *App
	root      *nui.Box
	weekL     *nui.Label
	streakL   *nui.Label
	taskBox   *nui.Box // чекбоксы задач пересобираются целиком при refresh
	quickNote *nui.Entry
}

func newPageToday(a *App) *pageToday {
	p := &pageToday{app: a}

	p.weekL = nui.NewLabel("")
	p.weekL.SetBold(true)
	p.weekL.SetFontScale(1.2)
	p.streakL = nui.NewLabel("")

	p.taskBox = nui.VBox().Spacing(4)
	tasks := nui.NewFrame("Задачи недели", nui.NewScroll(p.taskBox))

	markBtn := nui.NewButton("Отметить занятие за сегодня", p.markSession)
	p.quickNote = nui.NewEntry()
	p.quickNote.SetPlaceholder("Быстрая заметка — Enter сохранит в журнал")
	p.quickNote.OnEnter(p.saveQuickNote)

	p.root = nui.VBox(
		p.weekL, p.streakL,
		nui.Expand(tasks),
		nui.NewSeparator(),
		nui.HBox(markBtn, nui.Expand(p.quickNote)),
	).Spacing(8).Padding(12)

	p.refresh()
	return p
}

// refresh перечитывает базу и перерисовывает всё содержимое экрана.
func (p *pageToday) refresh() {
	a := p.app
	wn := a.CurrentWeek()
	w := a.DB.Plan.Week(wn)
	if w == nil {
		p.weekL.SetText(fmt.Sprintf("Неделя %d: план пуст или дата начала не задана", wn))
		p.taskBox.Clear()
		p.streakL.SetText("")
		return
	}
	goal := ""
	if len(w.Goals) > 0 {
		goal = " — " + w.Goals[0]
	}
	p.weekL.SetText(fmt.Sprintf("Неделя %d из %d%s", wn, a.DB.Plan.LastWeekNumber(), goal))

	p.taskBox.Clear()
	for _, t := range storage.TasksOfWeek(a.DB, wn) {
		task := t // локальная копия переменной цикла — на неё ссылается замыкание
		c := nui.NewCheck(fmt.Sprintf("%s  [%s, план %s ч]", task.Description, task.Direction, model.FormatHrs(task.PlannedHrs)))
		c.SetChecked(task.Status == model.TaskDone)
		c.OnChange(func() {
			if c.Checked() {
				task.Status = model.TaskDone
			} else {
				task.Status = model.TaskNew
			}
			a.Save()
		})
		p.taskBox.Add(c)
	}

	cur := plan.CurrentStreak(a.DB.Marks, a.today)
	best := plan.LongestStreak(a.DB.Marks)
	p.streakL.SetText(fmt.Sprintf("Серия: %d дн. подряд · рекорд %d дн.", cur, best))
}

// markSession добавляет по 30 минут каждому направлению с ненулевым ритмом,
// у которого сегодня ещё нет времени. Точные минуты правятся на календаре.
func (p *pageToday) markSession() {
	a := p.app
	m := a.DB.MarkOrCreate(a.TodayStr())
	added := 0
	for _, d := range a.DB.Plan.Directions {
		if m.Minutes[d] == 0 && a.DB.Plan.Rhythm.HoursPerDir[d] > 0 {
			m.Minutes[d] = 30
			added++
		}
	}
	if added == 0 {
		nui.Message(a.Win, "Уже отмечено", "Сегодня по всем направлениям уже есть время.")
		return
	}
	a.Save()
	p.refresh()
}

// saveQuickNote — Enter в поле быстрой заметки: одна запись журнала без окна.
func (p *pageToday) saveQuickNote() {
	a := p.app
	text := p.quickNote.Text()
	if text == "" {
		return
	}
	e := &model.JournalEntry{ID: a.DB.NextRecID(), Date: a.TodayStr(), Type: model.NoteEntry, Text: text, Week: a.CurrentWeek()}
	a.DB.Entries = append(a.DB.Entries, e)
	p.quickNote.SetText("")
	a.Save()
}
