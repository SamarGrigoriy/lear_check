//go:build gui

package nui

// Интеграционные тесты диалогов и оболочки (фаза 3). Диалог модальный и синхронный,
// поэтому он открывается через Do и ждёт ответа во вложенном цикле, а тест отвечает
// через ui() — Do работает и внутри вложенного цикла.

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"testing"
	"time"

	"nui/internal/drv"
)

// openDialog запускает диалог в UI-потоке и ждёт, пока он появится на экране.
func openDialog[T any](t *testing.T, chooser bool, show func() T) <-chan T {
	t.Helper()
	res := make(chan T, 1)
	Do(func() { res <- show() })
	waitUntil(t, 5000, func() bool { return drv.DialogOpen(chooser) })
	return res
}

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("диалог не закрылся")
	}
	var zero T
	return zero
}

func TestMessageDialogs(t *testing.T) {
	var w *Window
	ticks := 0
	var tm *Timer
	ui(func() {
		w = NewWindow("диалоги", NewLabel("Родитель диалогов"))
		w.Show()
		tm = Every(20, func() { ticks++ })
	})

	// Ask: «Да», «Нет», Esc.
	for _, c := range []struct {
		name   string
		answer func() bool
		want   bool
	}{
		{"ask", func() bool { return drv.SimAnswer(true) }, true},
		{"ask-no", func() bool { return drv.SimAnswer(false) }, false},
		{"ask-esc", drv.SimEsc, false},
	} {
		res := openDialog(t, false, func() bool {
			return Ask(w, "Выход", "Закрыть программу?\n\nВсё сохранено. 🙂")
		})
		if c.name == "ask" {
			shot(t, "ask")
		}
		ui(func() {
			if !c.answer() {
				t.Error("диалог не найден")
			}
		})
		if got := receive(t, res); got != c.want {
			t.Fatalf("%s: Ask вернул %v, ожидалось %v", c.name, got, c.want)
		}
	}

	// Message и Alert; пока открыт Message, таймер продолжает тикать.
	before := 0
	ui(func() { before = ticks })
	done := openDialog(t, false, func() bool { Message(w, "Готово", "Файл сохранён в ~/esp32_logs"); return true })
	time.Sleep(200 * time.Millisecond)
	shot(t, "message")
	during := 0
	ui(func() {
		during = ticks
		drv.SimOK()
	})
	receive(t, done)
	if during-before < 3 {
		t.Fatalf("пока открыт Message, Every тикнул %d раз — таймеры должны работать", during-before)
	}
	done = openDialog(t, false, func() bool {
		Alert(nil, "Нет доступа к порту", "Добавьте пользователя в группу dialout.")
		return true
	})
	shot(t, "alert")
	ui(func() { drv.SimOK() })
	receive(t, done)
	ui(func() {
		tm.Stop()
		w.Close()
	})
}

// Родитель уничтожается, пока открыт диалог: DESTROY_WITH_PARENT закрывает диалог, Ask
// возвращает false, и ничего не падает (своя ссылка в nui_message).
func TestDialogParentDestroyed(t *testing.T) {
	var w *Window
	ui(func() {
		w = NewWindow("родитель", NewLabel("будет закрыт"))
		w.Show()
		After(300, w.Close)
	})
	res := openDialog(t, false, func() bool { return Ask(w, "Вопрос", "Родитель закроется сам") })
	if receive(t, res) {
		t.Fatal("Ask после уничтожения родителя должен вернуть false")
	}
}

func TestFileDialogs(t *testing.T) {
	dir := tempDir(t)
	for _, f := range []string{"журнал.log", "заметки.txt"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var w *Window
	ui(func() {
		w = NewWindow("файлы", NewLabel("Родитель файловых диалогов"))
		w.Show()
	})
	defer ui(func() { w.Close() })
	type result struct {
		path string
		ok   bool
	}
	pick := func(name string, show func() (string, bool), choose string, accept bool) result {
		t.Helper()
		res := openDialog(t, true, func() result { p, ok := show(); return result{p, ok} })
		if choose != "" {
			ui(func() {
				if !drv.SimChooserSet(choose) {
					t.Errorf("%s: не удалось выбрать %s", name, choose)
				}
			})
			time.Sleep(700 * time.Millisecond) // каталог загружается асинхронно
		}
		shot(t, name)
		ui(func() { drv.SimChooserDone(accept) })
		return receive(t, res)
	}

	logs := Filter{Name: "Логи", Patterns: []string{"*.log", "*.txt"}}
	want := filepath.Join(dir, "журнал.log")
	if r := pick("open-file", func() (string, bool) { return OpenFile(w, "Открыть журнал", logs) }, want, true); !r.ok || r.path != want {
		t.Fatalf("OpenFile: %q, %v; ожидалось %q", r.path, r.ok, want)
	}
	if r := pick("open-cancel", func() (string, bool) { return OpenFile(w, "Открыть журнал") }, "", false); r.ok || r.path != "" {
		t.Fatalf("отмена OpenFile: %q, %v", r.path, r.ok)
	}
	want = filepath.Join(dir, "новый.log")
	if r := pick("save-file", func() (string, bool) { return SaveFile(w, "Сохранить журнал", "новый.log", logs) }, want, true); !r.ok || r.path != want {
		t.Fatalf("SaveFile: %q, %v; ожидалось %q", r.path, r.ok, want)
	}
	if r := pick("choose-folder", func() (string, bool) { return ChooseFolder(w, "Папка для логов") }, dir, true); !r.ok || r.path != dir {
		t.Fatalf("ChooseFolder: %q, %v; ожидалось %q", r.path, r.ok, dir)
	}

	// OpenFiles: два файла, один файл (на Windows ответ диалога тогда другого вида), отмена.
	type many struct {
		paths []string
		ok    bool
	}
	pickMany := func(name string, choose []string, accept bool) many {
		t.Helper()
		res := openDialog(t, true, func() many { p, ok := OpenFiles(w, "Открыть журналы", logs); return many{p, ok} })
		for i, f := range choose { // первый файл — как OpenFile, остальные — Ctrl+щелчок
			ui(func() {
				if i == 0 && !drv.SimChooserSet(f) || i > 0 && !drv.SimChooserAdd(f) {
					t.Errorf("%s: не удалось выбрать %s", name, f)
				}
			})
			time.Sleep(700 * time.Millisecond) // каталог загружается асинхронно
		}
		shot(t, name)
		ui(func() { drv.SimChooserDone(accept) })
		r := receive(t, res)
		sort.Strings(r.paths)
		return r
	}
	two := []string{filepath.Join(dir, "журнал.log"), filepath.Join(dir, "заметки.txt")}
	sort.Strings(two)
	if r := pickMany("open-files", two, true); !r.ok || !slices.Equal(r.paths, two) {
		t.Fatalf("OpenFiles: %q, %v; ожидалось %q", r.paths, r.ok, two)
	}
	if r := pickMany("open-files-one", two[:1], true); !r.ok || !slices.Equal(r.paths, two[:1]) {
		t.Fatalf("OpenFiles, один файл: %q, %v; ожидалось %q", r.paths, r.ok, two[:1])
	}
	if r := pickMany("open-files-cancel", nil, false); r.ok || r.paths != nil {
		t.Fatalf("отмена OpenFiles: %q, %v", r.paths, r.ok)
	}
}

// tempDir — как t.TempDir, но ошибка уборки не валит тест: под Wine 8 os.RemoveAll не
// работает (Go удаляет файлы через FileDispositionInfoEx, а Wine его не реализует).
func tempDir(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		return t.TempDir()
	}
	dir, err := os.MkdirTemp("", "nui-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}
