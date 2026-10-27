// Package storage — хранение всех данных приложения в одном JSON-файле.
//
// Почему JSON, а не SQLite: данных мало (тысячи строк за 24 недели), весь файл
// всегда лежит в памяти, фильтрация простая; формат виден глазами и правится
// текстовым редактором; не нужна внешняя библиотека (SQLite потребовал бы cgo
// или чистый порт на сотни КБ). Бэкап = копия одного файла.
//
// Запись атомарная: пишем в data.json.tmp и переименовываем поверх. Если процесс
// умрёт посреди записи, старый файл останется целым.
package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"studytracker/model"
)

// DefaultDataDir — каталог данных по умолчанию: ~/.local/share/study-tracker
// (на Windows — %AppData%\study-tracker).
func DefaultDataDir() string {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "study-tracker")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".study-tracker"
	}
	if p := os.Getenv("APPDATA"); p != "" { // Windows
		return filepath.Join(p, "study-tracker")
	}
	return filepath.Join(home, ".local", "share", "study-tracker")
}

// Path — полный путь к файлу базы в каталоге dir.
func Path(dir string) string { return filepath.Join(dir, "data.json") }

// Load читает базу из dir. Если файла нет, возвращает nil, nil — решение
// «создать новую базу из встроенного плана» принимает вызывающий.
func Load(dir string) (*model.DB, error) {
	data, err := os.ReadFile(Path(dir))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	db := model.NewDB()
	if err := json.Unmarshal(data, db); err != nil {
		return nil, errors.New("файл " + Path(dir) + " повреждён: " + err.Error())
	}
	fix(db)
	return db, nil
}

// Save пишет базу атомарно (tmp + rename).
func Save(dir string, db *model.DB) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := marshal(db)
	if err != nil {
		return err
	}
	tmp := Path(dir) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, Path(dir))
}

// Export пишет копию базы в произвольный файл (для бэкапа одним файлом).
func Export(dst string, db *model.DB) error {
	data, err := marshal(db)
	if err != nil {
		return err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

// Import читает файл бэкапа и возвращает базу; существующие данные не трогает.
// Импорт заменяет всё состояние целиком — это осознанно: частичное слияние
// двух баз породило бы конфликты id, а для восстановления из бэкапа оно не нужно.
func Import(src string) (*model.DB, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, err
	}
	db := model.NewDB()
	if err := json.Unmarshal(data, db); err != nil {
		return nil, errors.New("файл не является базой трекера: " + err.Error())
	}
	if db.Plan == nil || len(db.Plan.Stages) == 0 {
		return nil, errors.New("в файле нет плана обучения — импорт отменён")
	}
	fix(db)
	return db, nil
}

// marshal — единое место сериализации: с отступом, чтобы файл можно было читать
// и править глазами.
func marshal(db *model.DB) ([]byte, error) {
	return json.MarshalIndent(db, "", "  ")
}

// fix доводит загруженную базу до консистентного состояния: пустые maps,
// недостающие id, значения по умолчанию. Вызывается после каждого чтения.
func fix(db *model.DB) {
	if db.NextID < 1 {
		db.NextID = 1
	}
	maxID := 0
	for _, e := range db.Entries {
		if e.ID > maxID {
			maxID = e.ID
		}
		if e.Type == "" {
			e.Type = model.NoteEntry
		}
	}
	for _, c := range db.Checks {
		if c.ID > maxID {
			maxID = c.ID
		}
	}
	for _, q := range db.Questions {
		if q.ID > maxID {
			maxID = q.ID
		}
		if q.Box < 1 {
			q.Box = 1
		}
	}
	if maxID >= db.NextID {
		db.NextID = maxID + 1
	}
	for _, m := range db.Marks {
		if m.Minutes == nil {
			m.Minutes = map[string]int{}
		}
	}
	if db.Plan != nil {
		if len(db.Plan.Directions) == 0 {
			db.Plan.Directions = model.DefaultDirections()
		}
		if db.Plan.Start == "" {
			db.Plan.Start = time.Now().Format(model.DateLayout)
		}
		if db.Plan.Rhythm.HoursPerDir == nil {
			db.Plan.Rhythm.HoursPerDir = map[string]int{
				model.DirVerilog: 30, model.DirML: 25, model.DirReview: 20, model.DirEnglish: 25, model.DirProject: 0,
			}
			db.Plan.Rhythm.TotalHours = 10
		}
	}
}

// TasksOfWeek — задачи недели n в порядке плана. Индекс задачи в этом срезе —
// её TaskIdx (к нему привязаны записи журнала). Порядок db.Tasks поддерживается
// SyncTasks: сначала все задачи недели 1, потом недели 2 и т.д.
func TasksOfWeek(db *model.DB, n int) []*model.WeekTask {
	var out []*model.WeekTask
	for _, t := range db.Tasks {
		if t.Week == n {
			out = append(out, t)
		}
	}
	return out
}

// SyncTasks сверяет плоский список db.Tasks с планом: добавляет задачи новых
// недель (после правки плана), удаляет задачи исчезнувших, сохраняет статусы и
// часы существующих. Ключ соответствия — (неделя, позиция задачи в неделе):
// если пользователь переставил задачи местами в файле плана, статусы могут
// съехать — это принятая простота формата.
func SyncTasks(db *model.DB) {
	// старые задачи группируем по (неделя, позиция)
	type key struct{ week, idx int }
	old := map[key]*model.WeekTask{}
	perWeek := map[int]int{}
	for _, t := range db.Tasks {
		i := perWeek[t.Week]
		old[key{t.Week, i}] = t
		perWeek[t.Week] = i + 1
	}
	var out []*model.WeekTask
	for _, w := range db.Plan.AllWeeks() {
		for i := range w.Tasks {
			src := &w.Tasks[i]
			k := key{w.Number, i}
			if prev, ok := old[k]; ok {
				prev.Description = src.Description // текст/направление/час-план берём из плана
				prev.Direction = src.Direction
				prev.PlannedHrs = src.PlannedHrs
				out = append(out, prev)
			} else {
				nt := *src
				nt.Status = model.TaskNew
				nt.ActualHrs = 0
				out = append(out, &nt)
			}
		}
	}
	db.Tasks = out
}
