// Package ui — экраны приложения на gonui.
// app.go: общее состояние (App), каркас окна с вкладками и точки обновления.
package ui

import (
	"os"
	"time"

	"studytracker/builtin"
	"studytracker/model"
	"studytracker/plan"
	"studytracker/storage"

	"nui"
)

// App — всё, что нужно экранам: база, каталог данных, окно и вкладки.
type App struct {
	DB   *model.DB
	Dir  string // каталог, где лежит data.json
	Win  *nui.Window
	Tabs *nui.Tabs

	today time.Time // «сегодня» на момент запуска; перезапуск обновляет

	// ссылки на страницы — чтобы один экран мог попросить другой перерисоваться
	pageToday *pageToday
}

// Save пишет базу на диск и показывает ошибку диалогом.
// Вызывается после каждого изменения данных — файл всегда актуален.
func (a *App) Save() {
	if err := storage.Save(a.Dir, a.DB); err != nil {
		nui.Alert(a.Win, "Ошибка записи", err.Error())
	}
}

// TodayStr — сегодняшняя дата в формате модели ("2026-10-27").
func (a *App) TodayStr() string { return a.today.Format(model.DateLayout) }

// CurrentWeek — номер недели плана, которой соответствует сегодня.
func (a *App) CurrentWeek() int { return plan.WeekOfDate(a.DB, a.today) }

// RefreshAll перерисовывает экраны, зависящие от данных. Дёргается после
// импорта, правки плана и других «оптовых» изменений.
func (a *App) RefreshAll() {
	if a.pageToday != nil {
		a.pageToday.refresh()
	}
}

// NewApp собирает главное окно со всеми вкладками.
// Экраны добавляются по мере готовности; вместо готового экрана пока заглушка.
func NewApp(dir string, db *model.DB) *App {
	a := &App{DB: db, Dir: dir, today: model.Today()}

	a.Tabs = nui.NewTabs()
	a.pageToday = newPageToday(a)
	a.Tabs.Add("Сегодня", a.pageToday.root)
	a.Tabs.Add("План", placeholder("Раздел «План» — следующий срез"))
	a.Tabs.Add("Журнал", placeholder("Раздел «Журнал» — следующий срез"))
	a.Tabs.Add("Достижения", placeholder("Раздел «Достижения» — следующий срез"))
	a.Tabs.Add("Календарь", placeholder("Раздел «Календарь» — следующий срез"))
	a.Tabs.Add("Тесты", placeholder("Раздел «Тесты» — следующий срез"))
	a.Tabs.Add("Статистика", placeholder("Раздел «Статистика» — следующий срез"))
	a.Tabs.Add("Настройки", placeholder("Раздел «Настройки» — следующий срез"))

	a.Win = nui.NewWindow("Трекер обучения", a.Tabs)
	a.Win.SetSize(980, 680)
	return a
}

// Run поднимает приложение: загружает (или создаёт) базу и открывает окно.
func Run(dataDir string) error {
	db, err := storage.Load(dataDir)
	if err != nil {
		// битый файл не выбрасываем: переименовываем в .broken и начинаем заново
		broken := storage.Path(dataDir) + ".broken"
		_ = os.Rename(storage.Path(dataDir), broken)
		println("трекер: повреждённый data.json сохранён как", broken)
		db = nil
	}
	if db == nil {
		db = freshDB(dataDir) // встроенный план + банк вопросов, задачи развёрнуты
	}
	return nui.Run(func() {
		app := NewApp(dataDir, db)
		app.Win.Show()
	})
}

// freshDB — первая база: встроенный план, встроенные вопросы, дата старта — сегодня.
func freshDB(dir string) *model.DB {
	db := model.NewDB()
	p := *builtin.DefaultPlan() // копия верхнего уровня
	db.Plan = &p                 // этапы/недели внутри читаются через методы Plan;
	// правки плана экран настроек делает через явное присваивание db.Plan
	db.Plan.Start = model.Today().Format(model.DateLayout)
	db.Questions = builtin.DefaultQuestions()
	storage.SyncTasks(db)
	if err := storage.Save(dir, db); err != nil {
		println("трекер: не удалось сохранить новую базу:", err.Error())
	}
	return db
}

// placeholder — временная страница для вкладок, которые ещё не написаны.
func placeholder(text string) nui.Widget {
	return nui.NewScroll(nui.VBox(nui.NewLabel(text)).Padding(20))
}
