// settings.go — экран «Настройки»: путь к данным, экспорт/импорт одним файлом,
// дата начала плана, сдвиг плана при пропусках, недельный ритм (часы по
// направлениям) и редактирование плана через формы интерфейса (без правки JSON).
package ui

import (
	"fmt"
	"strconv"
	"strings"

	"studytracker/builtin"
	"studytracker/model"
	"studytracker/storage"

	"nui"
)

type pageSettings struct {
	app  *App
	root *nui.Box

	dirL   *nui.Label // путь к data.json
	status *nui.Label // результат последнего действия
}

func newPageSettings(a *App) *pageSettings {
	p := &pageSettings{app: a}

	p.dirL = nui.NewLabel("Данные: " + storage.Path(a.Dir))
	p.dirL.SetWrap(true)

	openBtn := nui.NewButton("Открыть папку данных…", p.openDataDir)
	saveBtn := nui.NewButton("Сохранить в эту папку", p.saveHere)
	expBtn := nui.NewButton("Экспорт в файл…", p.exportData)
	impBtn := nui.NewButton("Импорт из файла…", p.importData)
	resetBtn := nui.NewButton("Сбросить план к встроенному", p.resetPlan)

	shift := nui.NewCheck("Сдвигать недели при отставании")
	shift.SetChecked(a.DB.Settings.ShiftPlan)
	shift.OnChange(func() {
		a.DB.Settings.ShiftPlan = shift.Checked()
		a.Save()
		a.RefreshAll()
		p.set("Сдвиг плана: " + onOff(shift.Checked()))
	})

	p.root = nui.VBox(
		nui.NewFrame("Хранение данных", nui.VBox(
			p.dirL,
			nui.HBox(openBtn, saveBtn).Spacing(6),
			nui.HBox(expBtn, impBtn, resetBtn).Spacing(6),
		).Spacing(6).Padding(8)),
		nui.NewFrame("План", nui.VBox(
			p.buildStartRow(),
			shift,
			p.buildRhythm(),
			nui.NewButton("Редактор недель и задач…", p.openWeekEditor),
		).Spacing(8).Padding(8)),
		p.status,
	).Spacing(10).Padding(10)
	return p
}

func onOff(b bool) string {
	if b {
		return "включён"
	}
	return "выключен"
}

func (p *pageSettings) set(s string) { p.status.SetText(s) }

// --- дата начала ------------------------------------------------------

func (p *pageSettings) buildStartRow() nui.Widget {
	a := p.app
	ent := nui.NewEntry()
	ent.SetText(a.DB.Plan.Start)
	ent.SetWidthChars(12)
	btn := nui.NewButton("Применить", func() {
		s := strings.TrimSpace(ent.Text())
		t, err := model.ParseDate(s)
		if err != nil {
			nui.Alert(a.Win, "Неверная дата", "Нужен формат ГГГГ-ММ-ДД, например 2026-01-05.")
			return
		}
		a.DB.Plan.Start = t.Format(model.DateLayout)
		a.Save()
		a.RefreshAll()
		p.set("Дата начала: " + a.DB.Plan.Start)
	})
	return nui.HBox(nui.NewLabel("Дата начала плана:"), ent, btn).Spacing(6)
}

// --- ритм (часы по направлениям) --------------------------------------

func (p *pageSettings) buildRhythm() nui.Widget {
	a := p.app
	box := nui.VBox(nui.NewLabel("Недельный ритм, часы по направлениям (можно дробные: 2,5):"))
	grid := nui.NewGrid()
	for i, d := range a.DB.Plan.Directions {
		sp := nui.NewSpin(0, 80) // десятки часов: 25 = 2,5 ч
		sp.SetValue(a.DB.Plan.Rhythm.HoursPerDir[d])
		idx := i
		sp.OnChange(func() {
			a.DB.Plan.Rhythm.HoursPerDir[a.DB.Plan.Directions[idx]] = sp.Value()
			a.DB.Plan.Rhythm.TotalHours = totalTenths(a.DB.Plan.Rhythm) / 10
			a.Save()
			a.RefreshAll()
		})
		grid.Add(0, i, nui.NewLabel(d)).Add(1, i, sp)
	}
	total := nui.NewLabel(fmt.Sprintf("Всего: %s ч/нед.",
		model.FormatHrs(totalTenths(a.DB.Plan.Rhythm))))
	a.DB.Plan.Rhythm.HoursPerDir[a.DB.Plan.Directions[0]].GetType() // no-op guard removed below
	box.Add(grid)
	box.Add(total)
	return box
}

func totalTenths(r model.Rhythm) int {
	s := 0
	for _, v := range r.HoursPerDir {
		s += v
	}
	return s
}

// --- экспорт / импорт / выбор папки ------------------------------------

func (p *pageSettings) openDataDir() {
	a := p.app
	dir, ok := nui.ChooseFolder(a.Win, "Каталог с данными (data.json)")
	if !ok {
		return
	}
	db, err := storage.Load(dir)
	if err != nil {
		nui.Alert(a.Win, "Ошибка", err.Error())
		return
	}
	if db == nil {
		if !nui.Ask(a.Win, "Нет данных", "В выбранной папке нет data.json. Создать новую базу?") {
			return
		}
		db = freshDB(dir)
	}
	a.DB = db
	a.Dir = dir
	p.dirL.SetText("Данные: " + storage.Path(dir))
	a.RefreshAll()
	p.set("Загружена база из " + dir)
}

func (p *pageSettings) saveHere() {
	a := p.app
	if err := storage.Save(a.Dir, a.DB); err != nil {
		nui.Alert(a.Win, "Ошибка", err.Error())
		return
	}
	p.set("Сохранено: " + storage.Path(a.Dir))
}

func (p *pageSettings) exportData() {
	a := p.app
	path, ok := nui.SaveFile(a.Win, "Экспорт всех данных", "study-backup.json",
		nui.Filter{Name: "JSON", Patterns: []string{"*.json"}})
	if !ok {
		return
	}
	if err := storage.Export(path, a.DB); err != nil {
		nui.Alert(a.Win, "Экспорт не удался", err.Error())
		return
	}
	p.set("Экспортировано в " + path)
}

func (p *pageSettings) importData() {
	a := p.app
	path, ok := nui.OpenFile(a.Win, "Импорт из файла бэкапа",
		nui.Filter{Name: "JSON", Patterns: []string{"*.json"}})
	if !ok {
		return
	}
	if !nui.Ask(a.Win, "Импорт", "Импорт заменит ВСЕ текущие данные. Продолжить?") {
		return
	}
	db, err := storage.Import(path)
	if err != nil {
		nui.Alert(a.Win, "Импорт не удался", err.Error())
		return
	}
	a.DB = db
	if err := storage.Save(a.Dir, db); err != nil {
		nui.Alert(a.Win, "Ошибка записи", err.Error())
	}
	a.RefreshAll()
	p.set("Импортировано из " + path)
}

func (p *pageSettings) resetPlan() {
	a := p.app
	if !nui.Ask(a.Win, "Сброс плана",
		"Вернуть встроенный план (24 недели)? Журнал, отметки и тесты останутся; задачи будут пересозданы из плана.") {
		return
	}
	pl := *builtin.DefaultPlan()
	pl.Start = a.DB.Plan.Start // дату начала бережём
	a.DB.Plan = &pl
	storage.SyncTasks(a.DB)
	a.Save()
	a.RefreshAll()
	p.set("План сброшен к встроенному.")
}

// --- редактор недель и задач -------------------------------------------

// openWeekEditor — окно со списком недель; кнопки добавляют/правят/удаляют
// неделю, внутри недели — список задач с теми же операциями. Всё через формы,
// JSON пользователь руками не трогает.
func (p *pageSettings) openWeekEditor() {
	a := p.app

	var weekList *nui.List
	wk := func() *model.WeekPlan { // выбранная неделя
		i := weekList.Selected()
		all := a.DB.Plan.AllWeeks()
		if i < 0 || i >= len(all) {
			return nil
		}
		return a.DB.Plan.Week(all[i].Number)
	}
	redrawWeeks := func() {
		all := a.DB.Plan.AllWeeks()
		weekList.SetRows(len(all), func(r, c int) string {
			w := all[r]
			switch c {
			case 0:
				return strconv.Itoa(w.Number)
			case 1:
				st := a.DB.Plan.StageByWeek(w.Number)
				if st != nil {
					return st.Name
				}
				return ""
			default:
				return strings.Join(mapS(w.Goals), "; ")
			}
		})
	}
	weekList = nui.NewList("Нед.", "Этап", "Цели")
	redrawWeeks()

	addW := nui.NewButton("Добавить неделю…", func() { p.editWeek(nil, redrawWeeks) })
	modW := nui.NewButton("Изменить неделю…", func() {
		if w := wk(); w != nil {
			p.editWeek(w, redrawWeeks)
		}
	})
	delW := nui.NewButton("Удалить неделю", func() {
		w := wk()
		if w == nil {
			return
		}
		if !nui.Ask(a.Win, "Удалить неделю?", "Неделя "+strconv.Itoa(w.Number)+
			" будет убрана из плана; её задачи — из списка задач.") {
			return
		}
		removeWeek(a.DB, w.Number)
		a.Save()
		redrawWeeks()
		a.RefreshAll()
	})

	dlg := nui.NewWindow("Редактор плана", nui.VBox(
		nui.NewLabel("Недели плана:"),
		nui.Expand(weekList),
		nui.HBox(addW, modW, delW).Spacing(6),
	).Spacing(6).Padding(8))
	dlg.SetSize(720, 460)
	dlg.Show()
}

// editWeek — форма одной недели: номер, этап (для новой — выбор), цели (текст
// построчно), список задач. existing == nil — новая неделя.
func (p *pageSettings) editWeek(existing *model.WeekPlan, done func()) {
	a := p.app

	numSp := nui.NewSpin(1, 99)
	goalsV := nui.NewTextView()
	goalsV.SetWrap(true)

	// список задач этой недели
	var taskList *nui.List
	type row struct{ t *model.WeekTask }
	src := func() []*model.WeekTask {
		if existing == nil {
			return nil
		}
		out := make([]*model.WeekTask, len(existing.Tasks))
		for i := range existing.Tasks {
			out[i] = &existing.Tasks[i]
		}
		return out
	}
	redrawTasks := func() {
		ts := src()
		taskList.SetRows(len(ts), func(r, c int) string {
			t := ts[r]
			switch c {
			case 0:
				return t.Direction
			default:
				return firstLine(t.Description)
			}
		})
	}
	taskList = nui.NewList("Направление", "Задача")

	num := 0
	if existing != nil {
		num = existing.Number
		numSp.SetValue(num)
		goalsV.SetText(strings.Join(existing.Goals, "\n"))
	} else {
		numSp.SetValue(a.DB.Plan.LastWeekNumber() + 1)
	}
	redrawTasks()

	var dlg *nui.Window

	addT := nui.NewButton("Добавить задачу…", func() {
		p.editTask(existing, -1, redrawTasks)
	})
	modT := nui.NewButton("Изменить задачу…", func() {
		i := taskList.Selected()
		ts := src()
		if i >= 0 && i < len(ts) {
			p.editTask(existing, i, redrawTasks)
		}
	})
	delT := nui.NewButton("Удалить задачу", func() {
		i := taskList.Selected()
		if existing == nil || i < 0 || i >= len(existing.Tasks) {
			return
		}
		existing.Tasks = append(existing.Tasks[:i], existing.Tasks[i+1:]...)
		redrawTasks()
	})

	save := nui.NewButton("Сохранить неделю", func() {
		n := numSp.Value()
		if n < 1 {
			nui.Alert(dlg, "Номер", "Неделя должна быть не меньше 1.")
			return
		}
		if existing == nil && a.DB.Plan.Week(n) != nil {
			nui.Alert(dlg, "Дубликат", "Неделя "+strconv.Itoa(n)+" уже есть в плане.")
			return
		}
		goals := splitLines(goalsV.Text())
		if existing == nil {
			existing = &model.WeekPlan{Number: n}
			appendWeek(a.DB, existing)
		} else if existing.Number != n {
			renumberWeek(a.DB, existing.Number, n)
			existing.Number = n
		}
		existing.Goals = goals
		if err := a.DB.Plan.Validate(); err != nil {
			nui.Alert(dlg, "План некорректен", err.Error())
			return
		}
		storage.SyncTasks(a.DB)
		a.Save()
		a.RefreshAll()
		if done != nil {
			done()
		}
		dlg.Close()
	})

	dlg = nui.NewWindow("Неделя плана", nui.VBox(
		nui.HBox(nui.NewLabel("Номер недели:"), numSp).Spacing(6),
		nui.NewLabel("Цели (по одной на строку):"),
		nui.Expand(goalsV),
		nui.NewLabel("Задачи недели:"),
		nui.Expand(taskList),
		nui.HBox(addT, modT, delT).Spacing(6),
		nui.HBox(nui.Stretch(), nui.NewButton("Отмена", func() { dlg.Close() }), save).Spacing(6),
	).Spacing(6).Padding(8))
	dlg.SetSize(680, 620)
	dlg.SetDefault(save)
	dlg.Show()
}

// editTask — форма задачи: направление, описание, статус, часы план/факт.
// week == nil — задача ещё не привязана (редактор сохраняет её в existing.Tasks
// только если неделя уже существует).
func (p *pageSettings) editTask(week *model.WeekPlan, idx int, done func()) {
	a := p.app

	dirSel := nui.NewCombo()
	for _, d := range a.DB.Plan.Directions {
		dirSel.Add(d)
	}
	dirSel.SetSelected(0)

	descV := nui.NewTextView()
	descV.SetWrap(true)
	descV.SetMinSize(-1, 60)

	statSel := nui.NewCombo(model.TaskNew, model.TaskInProgress, model.TaskDone, model.TaskSkipped)
	statSel.SetSelected(0)

	planned := nui.NewSpin(0, 100)
	actual := nui.NewSpin(0, 200)

	var dlg *nui.Window
	save := nui.NewButton("Сохранить", func() {
		txt := strings.TrimSpace(descV.Text())
		if txt == "" {
			nui.Alert(dlg, "Описание", "Опишите задачу.")
			return
		}
		if week == nil {
			nui.Alert(dlg, "Неделя не сохранена", "Сначала сохраните неделю, затем добавьте задачи.")
			return
		}
		t := model.WeekTask{
			Description: txt, Direction: dirSel.Text(), Week: week.Number,
			Status: statSel.Text(), PlannedHrs: planned.Value(), ActualHrs: actual.Value(),
		}
		if idx >= 0 && idx < len(week.Tasks) {
			week.Tasks[idx] = t
		} else {
			week.Tasks = append(week.Tasks, t)
		}
		storage.SyncTasks(a.DB)
		a.Save()
		a.RefreshAll()
		if done != nil {
			done()
		}
		dlg.Close()
	})

	if week != nil && idx >= 0 && idx < len(week.Tasks) {
		t := week.Tasks[idx]
		dirSel.SetText(t.Direction)
		descV.SetText(t.Description)
		statSel.SetText(t.Status)
		planned.SetValue(t.PlannedHrs)
		actual.SetValue(t.ActualHrs)
	}

	dlg = nui.NewWindow("Задача недели", nui.VBox(
		nui.NewGrid().
			Add(0, 0, nui.NewLabel("Направление:")).Add(1, 0, dirSel).
			Add(0, 1, nui.NewLabel("Описание:")).Add(1, 1, descV).
			Add(0, 2, nui.NewLabel("Статус:")).Add(1, 2, statSel).
			Add(0, 3, nui.NewLabel("Часы план:")).Add(1, 3, planned).
			Add(0, 4, nui.NewLabel("Часы факт:")).Add(1, 4, actual).
			ExpandCol(1).Spacing(6, 6),
		nui.HBox(nui.Stretch(), nui.NewButton("Отмена", func() { dlg.Close() }), save).Spacing(6),
	).Spacing(6).Padding(8))
	dlg.SetSize(520, 360)
	dlg.SetDefault(save)
	dlg.Show()
}

// --- правки структуры плана (только для редактируемой копии DB.Plan) ----

func mapS(ss []string) []string {
	out := make([]string, len(ss))
	copy(out, ss)
	return out
}

func splitLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// appendWeek добавляет неделю в первый этап, куда она попадает по номеру
// (границы этапов не двигаем: это редкая операция, а не правка каждого дня).
func appendWeek(db *model.DB, w *model.WeekPlan) {
	for i := range db.Plan.Stages {
		st := &db.Plan.Stages[i]
		if w.Number >= st.From && w.Number <= st.To {
			st.Weeks = append(st.Weeks, *w)
			sortWeeks(st)
			return
		}
	}
	// не попала ни в один этап — добавляем в последний и расширяем его
	last := &db.Plan.Stages[len(db.Plan.Stages)-1]
	if w.Number > last.To {
		last.To = w.Number
	}
	last.Weeks = append(last.Weeks, *w)
	sortWeeks(last)
}

func sortWeeks(st *model.StagePlan) {
	for i := 1; i < len(st.Weeks); i++ {
		for j := i; j > 0 && st.Weeks[j].Number < st.Weeks[j-1].Number; j-- {
			st.Weeks[j], st.Weeks[j-1] = st.Weeks[j-1], st.Weeks[j]
		}
	}
}

func removeWeek(db *model.DB, n int) {
	for i := range db.Plan.Stages {
		st := &db.Plan.Stages[i]
		for j, w := range st.Weeks {
			if w.Number == n {
				st.Weeks = append(st.Weeks[:j], st.Weeks[j+1:]...)
				// удаляем задачи этой недели из плоского списка
				var keep []*model.WeekTask
				for _, t := range db.Tasks {
					if t.Week != n {
						keep = append(keep, t)
					}
				}
				db.Tasks = keep
				return
			}
		}
	}
}

func renumberWeek(db *model.DB, old, new int) {
	if db.Plan.Week(new) != nil {
		return // защита от дубликата: переименование не выполняется
	}
	for i := range db.Plan.Stages {
		st := &db.Plan.Stages[i]
		for j := range st.Weeks {
			if st.Weeks[j].Number == old {
				st.Weeks[j].Number = new
				for k := range st.Weeks[j].Tasks {
					st.Weeks[j].Tasks[k].Week = new
				}
				sortWeeks(st)
				for _, t := range db.Tasks {
					if t.Week == old {
						t.Week = new
					}
				}
				return
			}
		}
	}
}
