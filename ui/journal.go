// journal.go — экран «Журнал»: лента записей с фильтрами и поиском,
// редактор записей и отдельная вкладка «Неудачи и выводы».
package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"studytracker/model"
	"studytracker/storage"

	"nui"
)

type pageJournal struct {
	app  *App
	root *nui.Box
	tabs *nui.Tabs

	// вкладка «Лента»
	fType    *nui.Combo
	fDir     *nui.Combo
	fWeek    *nui.Combo
	fTag     *nui.Combo
	search   *nui.Entry
	list     *nui.List
	viewer   *nui.TextView
	entries  []*model.JournalEntry // то, что сейчас в списке (отфильтровано, свежие впереди)
	selected int                   // строка, которую показываем в viewer

	// вкладка «Неудачи и выводы»
	failList *nui.List
	failures []*model.JournalEntry
}

func newPageJournal(a *App) *pageJournal {
	p := &pageJournal{app: a, selected: -1}

	// --- фильтры: тип / направление / неделя / тег / поиск ---
	p.fType = nui.NewCombo("Все типы", model.NoteEntry, model.NoteWin, model.NoteFailure, model.NoteQuestion)
	p.fType.SetSelected(0)
	p.fDir = nui.NewCombo("Все направления")
	for _, d := range a.DB.Plan.Directions {
		p.fDir.Add(d)
	}
	p.fDir.SetSelected(0)
	p.fWeek = nui.NewComboEntry("Все недели")
	p.fTag = nui.NewComboEntry("") // пустой список: поле со свободным вводом для тега
	p.search = nui.NewEntry()
	p.search.SetPlaceholder("Поиск по тексту")

	for _, c := range []*nui.Combo{p.fType, p.fDir, p.fWeek, p.fTag} { // NewComboEntry возвращает тот же тип *Combo
		c.OnChange(p.refreshFeed)
	}
	p.search.OnChange(p.refreshFeed)

	filterRow := nui.HBox(p.fType, p.fDir, p.fWeek, p.fTag, nui.Expand(p.search)).Spacing(6)

	// --- лента + просмотр выбранной записи ---
	p.list = nui.NewList("Дата", "Тип", "Текст")
	p.list.OnSelect(func() { p.showSelected() })
	p.list.OnActivate(func(int) { p.editSelected() })
	p.list.SetContextMenu(nui.Item("Редактировать…", p.editSelected),
		nui.Item("Удалить", p.deleteSelected))

	p.viewer = nui.NewTextView()
	p.viewer.SetReadOnly(true)
	p.viewer.SetWrap(true)

	detail := nui.VBox(
		nui.NewLabel("Запись"),
		nui.Expand(p.viewer),
		nui.HBox(nui.NewButton("Добавить…", func() { p.editNew() }),
			nui.NewButton("Редактировать…", p.editSelected),
			nui.NewButton("Удалить", p.deleteSelected)),
	).Spacing(4)

	feed := nui.VBox(filterRow,
		nui.HSplit(nui.Expand(p.list), nui.Expand(detail)).SetPosition(520),
	).Spacing(6)

	// --- вкладка «Неудачи и выводы» ---
	p.failList = nui.NewList("Дата", "Что случилось", "Почему", "Что изменю")
	p.failList.OnActivate(func(int) { p.editSelectedFailure() })
	failures := nui.VBox(
		nui.NewLabel("Для каждой неудачи: что случилось, почему и что изменю. Двойной щелчок — редактировать."),
		nui.Expand(p.failList),
		nui.HBox(nui.NewButton("Новая неудача…", p.addFailure),
			nui.NewButton("Изменить…", p.editSelectedFailure)),
	).Spacing(6)

	p.tabs = nui.NewTabs().Add("Лента", feed).Add("Неудачи и выводы", failures)
	p.root = nui.VBox(nui.Expand(p.tabs)).Padding(8)

	p.fillWeekCombo()
	p.refreshFeed()
	p.refreshFailures()
	return p
}

// fillWeekCombo — список недель плана для фильтра «Неделя».
func (p *pageJournal) fillWeekCombo() {
	last := p.app.DB.Plan.LastWeekNumber()
	p.fWeek.Clear()
	p.fWeek.Add("Все недели")
	for i := 1; i <= last; i++ {
		p.fWeek.Add(strconv.Itoa(i))
	}
	p.fWeek.SetSelected(0)
}

// filteredEntries — записи по всем фильтрам, свежие впереди.
func (p *pageJournal) filteredEntries() []*model.JournalEntry {
	a := p.app
	typ := p.fType.Text()
	dir := p.fDir.Text()
	wkStr := strings.TrimSpace(p.fWeek.Text())
	tag := strings.ToLower(strings.TrimSpace(p.fTag.Text()))
	q := strings.ToLower(p.search.Text())

	var week int
	if wkStr != "" && wkStr != "Все недели" {
		week, _ = strconv.Atoi(wkStr)
	}

	var out []*model.JournalEntry
	for _, e := range a.DB.Entries {
		if typ != "" && typ != "Все типы" && e.Type != typ {
			continue
		}
		if dir != "" && dir != "Все направления" && !entryMatchesDir(a.DB, e, dir) {
			continue
		}
		if week > 0 && e.Week != week {
			continue
		}
		if tag != "" && !hasTag(e, tag) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(e.Text+e.Cause+e.Fix), q) {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date != out[j].Date {
			return out[i].Date > out[j].Date
		}
		return out[i].ID > out[j].ID
	})
	return out
}

// entryMatchesDir — запись относится к направлению, если привязана к задаче
// этого направления либо текст содержит название направления.
func entryMatchesDir(db *model.DB, e *model.JournalEntry, dir string) bool {
	if e.Week > 0 && e.TaskIdx >= 0 {
		tasks := storage.TasksOfWeek(db, e.Week)
		if e.TaskIdx < len(tasks) && tasks[e.TaskIdx].Direction == dir {
			return true
		}
	}
	return strings.Contains(strings.ToLower(e.Text), strings.ToLower(dir))
}

func hasTag(e *model.JournalEntry, tagLower string) bool {
	for _, t := range e.Tags {
		if strings.ToLower(t) == tagLower {
			return true
		}
	}
	return false
}

// refreshFeed — пересобрать ленту после смены фильтров или данных.
// cell-функцию храним в поле: SetRows каждый раз требует её передать снова.
func (p *pageJournal) refreshFeed() {
	p.entries = p.filteredEntries()
	p.list.SetRows(len(p.entries), p.feedCell)
	p.selected = -1
	p.viewer.SetText("")
}

func (p *pageJournal) feedCell(row, col int) string {
	e := p.entries[row]
	switch col {
	case 0:
		return e.Date
	case 1:
		return e.Type
	default:
		return firstLine(e.Text)
	}
}

// showSelected — вывести выбранную запись в правую панель.
func (p *pageJournal) showSelected() {
	i := p.list.Selected()
	p.selected = i
	if i < 0 || i >= len(p.entries) {
		p.viewer.SetText("")
		return
	}
	e := p.entries[i]
	var b strings.Builder
	b.WriteString(e.Date)
	if e.Type == model.NoteFailure {
		b.WriteString(" · неудача\n\nЧто случилось:\n" + e.Text)
		if e.Cause != "" {
			b.WriteString("\n\nПочему:\n" + e.Cause)
		}
		if e.Fix != "" {
			b.WriteString("\n\nЧто изменю:\n" + e.Fix)
		}
	} else {
		b.WriteString(" · " + e.Type + "\n\n" + e.Text)
	}
	if e.Week > 0 {
		b.WriteString(fmt.Sprintf("\n\n— неделя плана %d", e.Week))
		if e.TaskIdx >= 0 && e.TaskIdx < len(storage.TasksOfWeek(p.app.DB, e.Week)) {
			b.WriteString(", задача: " + storage.TasksOfWeek(p.app.DB, e.Week)[e.TaskIdx].Description)
		}
	}
	if len(e.Tags) > 0 {
		b.WriteString("\nТеги: " + strings.Join(e.Tags, ", "))
	}
	p.viewer.SetText(b.String())
}

func (p *pageJournal) selectedEntry() *model.JournalEntry {
	if p.selected >= 0 && p.selected < len(p.entries) {
		return p.entries[p.selected]
	}
	return nil
}

// editNew — новая запись любого типа.
func (p *pageJournal) editNew() { p.editEntry(nil) }

// editSelected — открыть редактор выбранной записи.
func (p *pageJournal) editSelected() {
	e := p.selectedEntry()
	if e == nil {
		nui.Message(p.app.Win, "Журнал", "Сначала выберите запись в списке.")
		return
	}
	p.editEntry(e)
}

// deleteSelected — удалить выбранную запись после подтверждения.
func (p *pageJournal) deleteSelected() {
	e := p.selectedEntry()
	if e == nil {
		return
	}
	if !nui.Ask(p.app.Win, "Удалить запись?", firstLine(e.Text)) {
		return
	}
	a := p.app
	for i, x := range a.DB.Entries {
		if x.ID == e.ID {
			a.DB.Entries = append(a.DB.Entries[:i], a.DB.Entries[i+1:]...)
			break
		}
	}
	a.Save()
	p.refreshAll()
}

// addFailure — сразу карточка неудачи.
func (p *pageJournal) addFailure() { e := &model.JournalEntry{Type: model.NoteFailure}; p.editEntry(e) }

// editSelectedFailure — правка выбранной неудачи.
func (p *pageJournal) editSelectedFailure() {
	i := p.failList.Selected()
	if i < 0 || i >= len(p.failures) {
		nui.Message(p.app.Win, "Неудачи", "Выберите строку в списке.")
		return
	}
	p.editEntry(p.failures[i])
}

// editEntry открывает модальное окно редактора. existing == nil — запись новая.
// Окно собирается заново под каждую запись: это дешевле и проще, чем прятать
// и заполнять одно универсальное окно.
func (p *pageJournal) editEntry(existing *model.JournalEntry) {
	a := p.app
	isFailure := existing != nil && existing.Type == model.NoteFailure

	var dialog *nui.Window // окно нужно save() и кнопке «Отмена» — объявляем заранее

	typeSel := nui.NewCombo(model.NoteEntry, model.NoteWin, model.NoteFailure, model.NoteQuestion)
	if existing != nil && model.ValidNoteType(existing.Type) {
		for i, it := range []string{model.NoteEntry, model.NoteWin, model.NoteFailure, model.NoteQuestion} {
			if it == existing.Type {
				typeSel.SetSelected(i)
			}
		}
	} else {
		typeSel.SetSelected(0)
	}

	dateE := nui.NewEntry()
	if existing != nil {
		dateE.SetText(existing.Date)
	} else {
		dateE.SetText(a.TodayStr())
	}

	textV := nui.NewTextView()
	textV.SetWrap(true)
	causeV := nui.NewTextView()
	causeV.SetWrap(true)
	fixV := nui.NewTextView()
	fixV.SetWrap(true)
	if existing != nil {
		textV.SetText(existing.Text)
		causeV.SetText(existing.Cause)
		fixV.SetText(existing.Fix)
	}

	weekSp := nui.NewSpin(0, a.DB.Plan.LastWeekNumber())
	if existing != nil {
		weekSp.SetValue(existing.Week)
	} else {
		weekSp.SetValue(a.CurrentWeek())
	}

	taskSel := nui.NewCombo("(без задачи)")
	taskSel.SetSelected(0)
	rebuildTasks := func() {
		sel := taskSel.Text()
		taskSel.Clear()
		taskSel.Add("(без задачи)")
		for _, t := range storage.TasksOfWeek(a.DB, weekSp.Value()) {
			taskSel.Add(firstLine(t.Description))
		}
		taskSel.SetText(sel) // прежний выбор сохранится, если задача ещё есть
		if taskSel.Selected() < 0 {
			taskSel.SetSelected(0)
		}
	}
	weekSp.OnChange(rebuildTasks)
	rebuildTasks()
	if existing != nil && existing.TaskIdx >= 0 {
		tasks := storage.TasksOfWeek(a.DB, existing.Week)
		if existing.TaskIdx < len(tasks) {
			taskSel.SetText(firstLine(tasks[existing.TaskIdx].Description))
		}
	}

	tagsE := nui.NewEntry()
	if existing != nil {
		tagsE.SetText(strings.Join(existing.Tags, ", "))
	}
	tagsE.SetPlaceholder("теги через запятую")

	failBox := nui.VBox(
		nui.NewLabel("Почему случилось:"), nui.Expand(causeV),
		nui.NewLabel("Что изменю:"), nui.Expand(fixV),
	).Spacing(4)

	form := nui.NewGrid().
		Add(0, 0, nui.NewLabel("Тип:")).Add(1, 0, typeSel).
		Add(0, 1, nui.NewLabel("Дата:")).Add(1, 1, dateE).
		Add(0, 2, nui.NewLabel("Текст:")).Add(1, 2, nui.Expand(textV)).
		Add(0, 3, nui.NewLabel("Неделя (0 — без):")).Add(1, 3, weekSp).
		Add(0, 4, nui.NewLabel("Задача:")).Add(1, 4, taskSel).
		Add(0, 5, nui.NewLabel("Теги:")).Add(1, 5, tagsE).
		ExpandCol(1).
		Spacing(6, 6).Padding(4)

	save := func() {
		typ := typeSel.Text()
		if typ == "" {
			typ = model.NoteEntry
		}
		dt := strings.TrimSpace(dateE.Text())
		if _, err := model.ParseDate(dt); err != nil {
			nui.Alert(dialog, "Неверная дата", "Формат: ГГГГ-ММ-ДД, например 2026-10-27")
			return
		}
		if textV.Text() == "" && typ != model.NoteFailure {
			nui.Alert(dialog, "Пустая запись", "Введите текст заметки.")
			return
		}
		taskIdx := -1
		if s := taskSel.Selected(); s > 0 {
			taskIdx = s - 1 // 0-й элемент — «без задачи»
		}
		if existing == nil {
			existing = &model.JournalEntry{ID: a.DB.NextRecID()}
			a.DB.Entries = append(a.DB.Entries, existing)
		}
		existing.Date = dt
		existing.Type = typ
		existing.Text = textV.Text()
		existing.Cause = causeV.Text()
		existing.Fix = fixV.Text()
		existing.Week = weekSp.Value()
		existing.TaskIdx = taskIdx
		existing.Tags = splitTags(tagsE.Text())
		a.Save()
		dialog.Close()
		p.refreshAll()
	}

	cancel := nui.NewButton("Отмена", func() { dialog.Close() })
	okBtn := nui.NewButton("Сохранить", save)

	body := nui.VBox(form, nui.NewSeparator(), failBox, nui.HBox(nui.Stretch(), cancel, okBtn)).Spacing(6).Padding(8)
	dialog = nui.NewWindow("Заметка журнала", body)
	dialog.SetSize(640, 560)
	dialog.SetDefault(okBtn)

	// поле «почему/что изменю» показываем только для неудач
	failBox.SetVisible(isFailure)
	typeSel.OnChange(func() { failBox.SetVisible(typeSel.Text() == model.NoteFailure) })

	dialog.Show()
}

// refreshFailures — перечитать список неудач.
func (p *pageJournal) refreshFailures() {
	var out []*model.JournalEntry
	for _, e := range p.app.DB.Entries {
		if e.Type == model.NoteFailure {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date > out[j].Date })
	p.failures = out
	p.failList.SetRows(len(out), p.failCell)
}

func (p *pageJournal) failCell(row, col int) string {
	f := p.failures[row]
	switch col {
	case 0:
		return f.Date
	case 1:
		return firstLine(f.Text)
	case 2:
		return firstLine(f.Cause)
	default:
		return firstLine(f.Fix)
	}
}

// refresh — общая точка обновления экрана после изменений извне.
func (p *pageJournal) refresh() {
	p.fillWeekCombo()
	p.refreshFeed()
	p.refreshFailures()
}

// refreshAll — обновить журнал и остальные экраны (задачи могли измениться).
func (p *pageJournal) refreshAll() {
	p.app.RefreshAll()
	p.refresh()
}

// helpers -------------------------------------------------------------

// editJournalEntry — общий редактор записей для других экранов (календарь).
// existing == nil или запись без ID — создаётся новая.
func (a *App) editJournalEntry(existing *model.JournalEntry) {
	if a.pageJournal == nil {
		return
	}
	if existing != nil && existing.ID == 0 {
		existing = &model.JournalEntry{Date: existing.Date, Type: existing.Type, Week: existing.Week}
	}
	a.pageJournal.editEntry(existing)
}

// showJournalWeek — открыть вкладку «Журнал» с фильтром по неделе wn
// (переход из экрана «План»).
func (a *App) showJournalWeek(wn int) {
	p := a.pageJournal
	if p == nil {
		return
	}
	a.Tabs.SetCurrent(a.journalTab)
	p.fWeek.SetSelected(0) // сбросить прежний фильтр, чтобы SetText нашёл элемент
	p.fWeek.SetText(strconv.Itoa(wn))
	p.refreshFeed()
	for i, e := range p.entries {
		if e.Week == wn {
			p.list.Select(i) // OnSelect покажет запись в просмотрщике
			break
		}
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func splitTags(s string) []string {
	var out []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
