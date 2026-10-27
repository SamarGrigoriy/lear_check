// plan.go — экран «План»: все недели по этапам, статусы, раскрытие задач,
// переход к записям журнала недели и отметка контрольных точек.
package ui

import (
	"fmt"
	"strconv"

	"studytracker/model"
	"studytracker/plan"
	"studytracker/storage"

	"nui"
)

type pagePlan struct {
	app  *App
	root *nui.Box
	list *nui.List // строки-«карточки»: собранные заранее []row с действием по двойному щелчку

	rows []planRow
}

// planRow — одна строка таблицы плана: этап, неделя или задача.
type planRow struct {
	kind   string // "stage" | "week" | "task"
	num    int    // номер недели (для week/task)
	taskIx int    // индекс задачи в неделе (для task)
	text   [4]string
	action func() // двойной щелчок: для недели — журнал, для задачи — цикл статуса
}

func newPagePlan(a *App) *pagePlan {
	p := &pagePlan{app: a}

	p.list = nui.NewList("Раздел", "Статус", "Часы план/факт", "Комментарий")
	p.list.SetRows(0, p.cell)
	p.list.OnActivate(func(row int) {
		if row >= 0 && row < len(p.rows) && p.rows[row].action != nil {
			p.rows[row].action()
		}
	})

	hint := nui.NewLabel("Двойной щелчок по неделе — записи журнала этой недели. Кнопки ниже меняют статус выбранной задачи.")
	hint.SetFontScale(0.85)

	btns := nui.HBox(
		nui.NewButton("Не начата", func() { p.setTask(model.TaskNew) }),
		nui.NewButton("В процессе", func() { p.setTask(model.TaskInProgress) }),
		nui.NewButton("Сделана", func() { p.setTask(model.TaskDone) }),
		nui.NewButton("Пропущена", func() { p.setTask(model.TaskSkipped) }),
	)

	p.root = nui.VBox(nui.Expand(p.list), hint, btns).Spacing(6).Padding(8)
	p.refresh()
	return p
}

func (p *pagePlan) cell(row, col int) string { return p.rows[row].text[col] }

// refresh пересобирает все строки: заголовки этапов, недели и их задачи.
func (p *pagePlan) refresh() {
	a := p.app
	cur := a.CurrentWeek()
	p.rows = nil

	for si := range a.DB.Plan.Stages {
		st := &a.DB.Plan.Stages[si]
		cpDone := p.checkpointDone(st.Name)
		status := "не завершён"
		if cpDone {
			status = "✓ контрольная точка выполнена"
		} else if allStageWeeksDone(a.DB, st) {
			status = "недели закрыты — сдай точку"
		}
		p.rows = append(p.rows, planRow{
			kind: "stage",
			text: [4]string{"ЭТАП: " + st.Name + " (нед. " + strconv.Itoa(st.From) + "–" + strconv.Itoa(st.To) + ")", status, "", st.Checkpoint},
			action: func() { p.editCheckpoint(st) },
		})

		for wi := range st.Weeks {
			w := &st.Weeks[wi]
			tasks := storage.TasksOfWeek(a.DB, w.Number)
			st2 := plan.WeekStatus(tasks, w.Number)
			label := fmt.Sprintf("  Неделя %d", w.Number)
			if w.Number == cur {
				label += " ← сейчас"
			}
			planned, actual := 0, 0
			for _, t := range tasks {
				planned += t.PlannedHrs
				actual += t.ActualHrs
			}
			summary := firstLine(joinGoals(w.Goals))
			mark := map[string]string{"done": "сделана", "active": "в работе", "new": "не начата"}[st2]
			wn := w.Number
			p.rows = append(p.rows, planRow{
				kind: "week", num: wn,
				text: [4]string{label, mark, hrsCell(planned, actual), summary},
				action: func() { a.showJournalWeek(wn) },
			})

			for ti, t := range tasks {
				tk := t
				prefix := "      "
				switch tk.Status {
				case model.TaskDone:
					prefix += "✔ "
				case model.TaskSkipped:
					prefix += "✗ "
				case model.TaskInProgress:
					prefix += "▶ "
				}
				weekNum, idx := w.Number, ti
				p.rows = append(p.rows, planRow{
					kind: "task", num: weekNum, taskIx: idx,
					text: [4]string{prefix + tk.Description, tk.Direction, hrsCell(tk.PlannedHrs, tk.ActualHrs*10), ""},
					action: func() {
						cycleTask(tk)
						a.Save()
						p.refresh()
						a.RefreshAll()
					},
				})
			}
		}
	}
	p.list.SetRows(len(p.rows), p.cell)
	p.list.Update()
}

func joinGoals(g []string) string {
	out := ""
	for i, s := range g {
		if i > 0 {
			out += "; "
		}
		out += s
	}
	return out
}

// hrsCell — часы: план/факт. planned приходит в десятых долях для задач умножением на 10.
func hrsCell(plannedTenths, actualHrs int) string {
	if plannedTenths == 0 && actualHrs == 0 {
		return ""
	}
	return model.FormatHrs(plannedTenths) + " / " + model.FormatHrs(actualHrs*10) + " ч"
}

func allStageWeeksDone(db *model.DB, st *model.StagePlan) bool {
	for _, w := range st.Weeks {
		if plan.WeekStatus(storage.TasksOfWeek(db, w.Number), w.Number) != "done" {
			return false
		}
	}
	return len(st.Weeks) > 0
}

// checkpointDone — есть ли выполненная точка с таким названием этапа.
func (p *pagePlan) checkpointDone(stageName string) bool {
	for _, c := range p.app.DB.Checks {
		if c.Name == stageName && c.DoneDate != "" {
			return true
		}
	}
	return false
}

// editCheckpoint — модальное окно контрольной точки: дата выполнения и самооценка 1–5.
func (p *pagePlan) editCheckpoint(st *model.StagePlan) {
	a := p.app
	var cp *model.Checkpoint
	for _, c := range a.DB.Checks {
		if c.Name == st.Name {
			cp = c
			break
		}
	}
	if cp == nil {
		cp = &model.Checkpoint{ID: a.DB.NextRecID(), Name: st.Name, Criteria: st.Checkpoint, Week: st.To}
		a.DB.Checks = append(a.DB.Checks, cp)
	}

	title := nui.NewLabel("Контрольная точка: " + st.Name)
	title.SetBold(true)
	crit := nui.NewTextView()
	crit.SetReadOnly(true)
	crit.SetWrap(true)
	crit.SetText(st.Checkpoint)

	dateE := nui.NewEntry()
	dateE.SetText(cp.DoneDate)
	dateE.SetPlaceholder("ГГГГ-ММ-ДД — пусто, если ещё не выполнена")

	score := nui.NewCombo("—", "1", "2", "3", "4", "5")
	if cp.SelfScore >= 1 && cp.SelfScore <= 5 {
		score.SetSelected(cp.SelfScore)
	} else {
		score.SetSelected(0)
	}

	var dialog *nui.Window
	save := func() {
		d := trim(dateE.Text())
		if d != "" {
			if _, err := model.ParseDate(d); err != nil {
				nui.Alert(dialog, "Неверная дата", "Формат: ГГГГ-ММ-ДД")
				return
			}
		}
		cp.DoneDate = d
		if s := score.Selected(); s >= 1 {
			cp.SelfScore = s
		} else {
			cp.SelfScore = 0
		}
		a.Save()
		dialog.Close()
		p.refresh()
		a.RefreshAll()
	}
	ok := nui.NewButton("Сохранить", save)
	cancel := nui.NewButton("Отмена", func() { dialog.Close() })

	dialog = nui.NewWindow("Контрольная точка", nui.VBox(
		title, crit,
		nui.NewGrid().Add(0, 0, nui.NewLabel("Дата выполнения:")).Add(1, 0, dateE).
			Add(0, 1, nui.NewLabel("Самооценка 1–5:")).Add(1, 1, score).ExpandCol(1),
		nui.HBox(nui.Stretch(), cancel, ok),
	).Spacing(6).Padding(8))
	dialog.SetSize(520, 340)
	dialog.SetDefault(ok)
	dialog.Show()
}

// setTask — кнопки статуса работают по выбранной строке задачи: находим задачу
// по (неделя, индекс) и ставим статус напрямую.
func (p *pagePlan) setTask(status string) {
	i := p.list.Selected()
	if i < 0 || i >= len(p.rows) || p.rows[i].kind != "task" {
		nui.Message(p.app.Win, "План", "Сначала выберите строку задачи (строки с отступом).")
		return
	}
	row := p.rows[i]
	tasks := storage.TasksOfWeek(p.app.DB, row.num)
	if row.taskIx >= len(tasks) {
		return // список устарел — после refresh всё встанет на место
	}
	tasks[row.taskIx].Status = status
	p.app.Save()
	p.refresh()
	p.app.RefreshAll()
}

// cycleTask — двойной щелчок по задаче: не начата → в процессе → сделана → пропущена → не начата.
func cycleTask(t *model.WeekTask) {
	switch t.Status {
	case model.TaskNew:
		t.Status = model.TaskInProgress
	case model.TaskInProgress:
		t.Status = model.TaskDone
	case model.TaskDone:
		t.Status = model.TaskSkipped
	default:
		t.Status = model.TaskNew
	}
}

func trim(s string) string {
	out := []rune(s)
	for len(out) > 0 && out[0] == ' ' {
		out = out[1:]
	}
	for len(out) > 0 && out[len(out)-1] == ' ' {
		out = out[:len(out)-1]
	}
	return string(out)
}
