//go:build gui

package nui

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"nui/internal/drv"
)

// OnDropFiles: пути бросенных файлов — как есть (кириллица, пробелы); nil — приём выключен.
func TestDropFiles(t *testing.T) {
	dir := tempDir(t)
	files := []string{filepath.Join(dir, "отчёт за год.pdf"), filepath.Join(dir, "фото 🙂.jpg")}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var w *Window
	var got [][]string
	ui(func() {
		w = NewWindow("перетаскивание", NewLabel("Бросьте файлы сюда"))
		w.OnDropFiles(func(paths []string) { got = append(got, paths) })
		w.SetSize(300, 120)
		w.Show()
		drv.SimDropFiles(w.h, w.id, files)
	})
	waitUntil(t, 2000, func() bool { return len(got) == 1 })
	ui(func() {
		if !slices.Equal(got[0], files) {
			t.Errorf("OnDropFiles: %q, ожидалось %q", got[0], files)
		}
		w.OnDropFiles(nil)
		drv.SimDropFiles(w.h, w.id, files[:1])
	})
	time.Sleep(300 * time.Millisecond) // на Windows сообщение идёт через очередь
	ui(func() {
		if len(got) != 1 {
			t.Errorf("после OnDropFiles(nil) обработчик вызван ещё %d раз", len(got)-1)
		}
		w.Close()
	})
}
