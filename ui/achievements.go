// achievements.go — экран «Достижения»: список побед из журнала,
// контрольные точек с самооценкой и прогресс по этапам.
package ui

import (
	"fmt"
	"strings"

	"studytracker/model"
	"studytracker/plan"
	"studytracker/storage"

	"nui"
)

type pageAchievements struct {
	app    *App
	root   *nui.Box
	wins   *nui.List // достижения из журнала
	checks *nui.List // контрольные точки
	prog   *nui.Box  // прогресс по этапам (строки-метки)

	winRows []*model.JournalEntry
	cpRows  []cpRow
}

// cpRow — строка таблицы контрольных точек (один этап = одна строка).
type cpRow struct {
	name  string
	week  int
	done  string
	score int
	st    *model.StagePlan
}

func newPageAchievements(a *App) *pageAchievements {
	p := &pageAchievements{app: a}

	p.wins = nui.NewList("Дата", "Достижение")
	p.wins.OnActivate(func(int) { p.editSelectedWin() })

	p.checks = nui.NewList("Этап", "Неделя", "Выполнено", "Самооценка")
	p.checks.OnActivate(func(int) { p.editSelectedCheck() })

	p.prog = nui.VBox()

	left := nui.VBox(
		nui.NewLabel("Победы из журнала (двойной щелчок — открыть)."),
		nui.Expand(p.wins),
		nui.HBox(nui.NewButton("Новое достижение…", p.addWin)),
	).Spacing(4)

	right := nui.VBox(
		nui.NewLabel("Контрольные точки этапов (двойной щелчок — отметить)."),
		nui.Expand(p.checks),
		nui.NewSeparator(),
		nui.NewLabel("Прогресс по этапам:"),
		nui.Expand(nui.NewScroll(p.prog)),
	).Spacing(4)

	p.root = nui.VBox(nui.Expand(nui.HSplit(nui.Expand(left), nui.Expand(right)).SetPosition(460))).Padding(8)
	p.refresh()
	return p
}

// refresh перечитывает все три блока.
func (p *pageAchievements) refresh() {
	a := p.app

	// победы
	var wins []*model.JournalEntry
	for _, e := range a.DB.Entries {
		if e.Type == model.NoteWin {
			wins = append(wins, e)
		}
	}
	sortEntriesDesc(wins)
	p.winRows = wins
	p.wins.SetRows(len(wins), func(row, col int) string {
		if col == 0 {
			return wins[row].Date
		}
		return firstLine(wins[row].Text)
	})

	// контрольные точки: одна строка на этап плана
	var rows []cpRow
	for i := range a.DB.Plan.Stages {
		st := &a.DB.Plan.Stages[i]
		r := cpRow{name: st.Name, week: st.To, st: st}
		for _, c := range a.DB.Checks {
			if c.Name == st.Name {
				r.done = c.DoneDate
				r.score = c.SelfScore
			}
		}
		rows = append(rows, r)
	}
	p.cpRows = rows
	p.checks.SetRows(len(rows), func(row, col int) string {
		r := rows[row]
		switch col {
		case 0:
			return r.name
		case 1:
			return fmt.Sprintf("нед. %d", r.week)
		case 2:
			if r.done != "" {
				return "✓ " + r.done
			}
			return "не выполнена"
		default:
			if r.score >= 1 {
				return fmt.Sprintf("%d/5", r.score)
			}
			return ""
		}
	})

	// прогресс этапов: доля закрытых недель
	p.prog.Clear()
	for i := range a.DB.Plan.Stages {
		st := &a.DB.Plan.Stages[i]
		done, total := 0, len(st.Weeks)
		for _, w := range st.Weeks {
			if plan.WeekStatus(storage.TasksOfWeek(a.DB, w.Number), w.Number) == "done" {
				done++
			}
		}
		p.prog.Add(nui.NewLabel(fmt.Sprintf("%s (нед. %d–%d): %d/%d недель%s",
			st.Name, st.From, st.To, done, total, bar(done, total))))
	}
}

// bar — текстовая полоска «▮▮▮▯▯▯».
func bar(done, total int) string {
	if total == 0 {
		return ""
	}
	const width = 12
	filled := done * width / total
	return " [" + strings.Repeat("▮", filled) + strings.Repeat("▯", width-filled) + "]"
}

func sortEntriesDesc(es []*model.JournalEntry) {
	for i := 1; i < len(es); i++ {
		for j := i; j > 0 && es[j].Date > es[j-1].Date; j-- {
			es[j], es[j-1] = es[j-1], es[j]
		}
	}
}

func (p *pageAchievements) addWin() {
	e := &model.JournalEntry{Type: model.NoteWin, Date: p.app.TodayStr(), Week: p.app.CurrentWeek()}
	p.app.editJournalEntry(e)
}

func (p *pageAchievements) editSelectedWin() {
	i := p.wins.Selected()
	if i < 0 || i >= len(p.winRows) {
		return
	}
	p.app.editJournalEntry(p.winRows[i])
}

func (p *pageAchievements) editSelectedCheck() {
	i := p.checks.Selected()
	if i < 0 || i >= len(p.cpRows) {
		return
	}
	p.editCheck(p.cpRows[i].st)
}

// editCheck — модалка контрольной точки (как в плане, но от экрана достижений).
func (p *pageAchievements) editCheck(st *model.StagePlan) {
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
	ok := nui.NewButton("Сохранить", func() {
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
	})
	dialog = nui.NewWindow("Контрольная точка: "+st.Name, nui.VBox(
		crit,
		nui.NewGrid().Add(0, 0, nui.NewLabel("Дата выполнения:")).Add(1, 0, dateE).
			Add(0, 1, nui.NewLabel("Самооценка 1–5:")).Add(1, 1, score).ExpandCol(1),
		nui.HBox(nui.Stretch(), nui.NewButton("Отмена", func() { dialog.Close() }), ok),
	).Spacing(6).Padding(8))
	dialog.SetSize(520, 340)
	dialog.SetDefault(ok)
	dialog.Show()
}
