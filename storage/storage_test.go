package storage

import (
	"os"
	"path/filepath"
	"testing"

	"studytracker/model"
)

func mkDB() *model.DB {
	db := model.NewDB()
	db.Plan = &model.Plan{
		Title: "Тест", Start: "2026-03-02",
		Stages: []model.StagePlan{{Name: "A", From: 1, To: 1, Weeks: []model.WeekPlan{
			{Number: 1, Tasks: []model.WeekTask{
				{Description: "задача 1", Direction: model.DirVerilog, Week: 1, PlannedHrs: 30},
				{Description: "задача 2", Direction: model.DirML, Week: 1, PlannedHrs: 25},
			}},
		}}},
	}
	SyncTasks(db) // наполним db.Tasks из плана
	e := &model.JournalEntry{ID: db.NextRecID(), Date: "2026-03-02", Type: model.NoteWin, Text: "победа"}
	db.Entries = append(db.Entries, e)
	m := db.MarkOrCreate("2026-03-02")
	m.Minutes[model.DirVerilog] = 45
	q := &model.QuizQuestion{ID: db.NextRecID(), Direction: model.DirML, Week: 1, Text: "?", Choices: []string{"а", "б"}, AnswerIndex: 1}
	db.Questions = append(db.Questions, q)
	return db
}

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	db := mkDB()
	if err := Save(dir, db); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got == nil {
		t.Fatal("Load вернул nil для существующего файла")
	}
	if len(got.Tasks) != 2 || got.Tasks[0].Status != model.TaskNew {
		t.Errorf("tasks после загрузки: %+v", got.Tasks)
	}
	if got.Entries[0].Text != "победа" || got.Entries[0].Type != model.NoteWin {
		t.Errorf("journal: %+v", got.Entries[0])
	}
	if mm := got.FindMark("2026-03-02"); mm == nil || mm.Minutes[model.DirVerilog] != 45 {
		t.Errorf("marks: %+v", mm)
	}
	// id в mkDB: запись=1, вопрос=2, NextID стал 3. После перезагрузки fix()
	// должен оставить следующий id не меньше максимала+1 — без дублей.
	id := got.NextRecID()
	if id < 3 {
		t.Errorf("NextRecID после загрузки = %d, хотим >= 3", id)
	}
	if id == got.Questions[0].ID || id == got.Entries[0].ID {
		t.Error("NextRecID выдал уже занятый id")
	}
}

func TestLoadMissingFileReturnsNil(t *testing.T) {
	db, err := Load(filepath.Join(t.TempDir(), "нет-такой-папки"))
	if err != nil || db != nil {
		t.Errorf("Load несуществующего каталога: db=%v err=%v, хотим nil,nil", db, err)
	}
}

func TestLoadBrokenFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(Path(dir), []byte("{битый json"), 0o644)
	if _, err := Load(dir); err == nil {
		t.Error("битый файл должен давать ошибку, а не молча пустую базу")
	}
}

func TestFixFillsDefaults(t *testing.T) {
	dir := t.TempDir()
	// файл без next_id, без типов, с отметкой без minutes-мапы
	raw := `{"version":1,"plan":{"stages":[{"name":"A","from":1,"to":1,"weeks":[{"number":1,"tasks":[]}]}]},
	 "journal":[{"id":7,"date":"2026-01-01","text":"x"}],
	 "marks":[{"date":"2026-01-01"}],
	 "questions":[{"id":3,"direction":"ML","week":1,"text":"?","box":0}]}`
	os.WriteFile(Path(dir), []byte(raw), 0o644)
	db, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if db.NextID != 8 { // max id = 7 -> следующий 8
		t.Errorf("NextID = %d, хотим 8", db.NextID)
	}
	if db.Entries[0].Type != model.NoteEntry {
		t.Errorf("тип записи по умолчанию = %q", db.Entries[0].Type)
	}
	if db.Marks[0].Minutes == nil {
		t.Error("minutes не превращён в пустую карту")
	}
	if db.Questions[0].Box != 1 {
		t.Errorf("box = %d, хотим 1", db.Questions[0].Box)
	}
	if db.Plan.Start == "" {
		t.Error("пустая дата старта не заполнена сегодняшним днём")
	}
	if len(db.Plan.Directions) != 5 {
		t.Errorf("направления не проставлены: %v", db.Plan.Directions)
	}
}

func TestSyncTasksPreservesStatus(t *testing.T) {
	db := mkDB()
	if len(db.Tasks) != 2 {
		t.Fatalf("после первого SyncTasks задач %d", len(db.Tasks))
	}
	db.Tasks[0].Status = model.TaskDone
	db.Tasks[0].ActualHrs = 3

	// пользователь отредактировал план: текст задачи 1 изменился, добавлена задача 3
	w := db.Plan.Week(1)
	w.Tasks[0].Description = "задача 1 — уточнённая"
	w.Tasks = append(w.Tasks, model.WeekTask{Description: "задача 3", Direction: model.DirEnglish, Week: 1})

	SyncTasks(db)
	if len(db.Tasks) != 3 {
		t.Fatalf("после второго SyncTasks задач %d, хотим 3", len(db.Tasks))
	}
	if db.Tasks[0].Description != "задача 1 — уточнённая" {
		t.Errorf("текст не обновился из плана: %q", db.Tasks[0].Description)
	}
	if db.Tasks[0].Status != model.TaskDone || db.Tasks[0].ActualHrs != 3 {
		t.Errorf("статус/часы потеряны: %+v", db.Tasks[0])
	}
	if db.Tasks[2].Status != model.TaskNew {
		t.Errorf("новая задача должна быть «не начата», получилась %q", db.Tasks[2].Status)
	}

	// удаление задачи из плана убирает её из db.Tasks
	w = db.Plan.Week(1)
	w.Tasks = w.Tasks[:2]
	SyncTasks(db)
	if len(db.Tasks) != 2 {
		t.Errorf("после удаления задач в плане: %d, хотим 2", len(db.Tasks))
	}
}

func TestTasksOfWeekOrder(t *testing.T) {
	db := mkDB()
	tasks := TasksOfWeek(db, 1)
	if len(tasks) != 2 || tasks[0].Direction != model.DirVerilog {
		t.Errorf("TasksOfWeek: %+v", tasks)
	}
	if TasksOfWeek(db, 5) != nil {
		t.Error("для несуществующей недели хотим nil")
	}
}

func TestExportImport(t *testing.T) {
	dir := t.TempDir()
	db := mkDB()
	dst := filepath.Join(dir, "backup.json")
	if err := Export(dst, db); err != nil {
		t.Fatalf("Export: %v", err)
	}
	got, err := Import(dst)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if got.Entries[0].Text != "победа" || len(got.Tasks) != 2 {
		t.Error("импорт не вернул исходные данные")
	}
	// импорт файла без плана — отказ
	bad := filepath.Join(dir, "bad.json")
	os.WriteFile(bad, []byte(`{"version":1}`), 0o644)
	if _, err := Import(bad); err == nil {
		t.Error("импорт без плана должен отклоняться")
	}
	// импорт чужого файла
	os.WriteFile(bad, []byte(`просто текст`), 0o644)
	if _, err := Import(bad); err == nil {
		t.Error("импорт не-JSON должен отклоняться")
	}
}
