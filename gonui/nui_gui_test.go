//go:build gui

package nui

// Интеграционные тесты на настоящем GTK (docs/testing.md, §3). Запуск:
//   NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals xvfb-run -a go test -tags gui .
// Пакет testing запускает каждый тест в отдельной горутине, поэтому цикл GTK крутится
// в главном потоке, m.Run() — в горутине, а всё, что трогает nui, тест делает через ui().
// t.Parallel() запрещён: у тестов один UI-поток и одно состояние GTK.

import (
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"nui/internal/core"
	"nui/internal/drv"
)

func TestMain(m *testing.M) {
	if h := os.Getenv("NUI_HELPER"); h != "" {
		os.Exit(helper(h)) // сценарии, которые завершают Run, — в дочернем процессе
	}
	code := 0
	err := Run(func() {
		// Якорь: видимое окно на всё время тестов. Без него первое же закрытое тестом
		// окно завершило бы Run по ADR-15, и следующие тесты повисли бы.
		NewWindow("nui: интеграционные тесты", NewLabel("Окно-якорь: держит цикл событий")).Show()
		go func() { code = m.Run(); Do(Quit) }()
	})
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
	os.Exit(code)
}

// ui выполняет fn в UI-потоке и ждёт её завершения.
func ui(fn func()) {
	done := make(chan struct{})
	Do(func() {
		defer close(done)
		fn()
	})
	<-done
}

// waitUntil ждёт, пока cond (она выполняется в UI-потоке) не вернёт true.
func waitUntil(t *testing.T, ms int, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Duration(ms) * time.Millisecond)
	for {
		ok := false
		ui(func() { ok = cond() })
		if ok {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("условие не выполнилось за %d мс", ms)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// shot — снимок экрана для фотоотчёта, если задан каталог NUI_SHOTS. Делается из
// горутины теста: UI-поток свободен и дорисовывает окно.
func shot(t *testing.T, name string) {
	dir := os.Getenv("NUI_SHOTS")
	if dir == "" {
		return
	}
	time.Sleep(300 * time.Millisecond)
	if out, err := exec.Command("sh", "-c", "rootshot | pnmtopng > "+dir+"/"+name+".png").CombinedOutput(); err != nil {
		t.Logf("снимок %s: %v %s", name, err, out)
	}
}

// expectPanic выполняет fn и проверяет текст паники.
func expectPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if s, _ := r.(string); s != want {
			t.Errorf("паника %v, ожидалась %q", r, want)
		}
	}()
	fn()
}

func alloc(w Widget) (x, y, width, height int) { return drv.Allocation(w.base().h) }

func TestLabelText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Привет", "Привет"},
		{"эмодзи 🙂", "эмодзи 🙂"},
		{"", ""},
		{"до\x00после", "до"}, // строка C обрезается на нулевом байте
		{"a\xffb", "a�b"},     // битый UTF-8 → U+FFFD, без Gtk-CRITICAL
	}
	ui(func() {
		for _, c := range cases {
			l := NewLabel(c.in)
			if got := l.Text(); got != c.want {
				t.Errorf("NewLabel(%q).Text() = %q, ожидалось %q", c.in, got, c.want)
			}
			l.SetText(c.in)
			if got := l.Text(); got != c.want {
				t.Errorf("SetText(%q) → %q, ожидалось %q", c.in, got, c.want)
			}
		}
	})
}

func TestButtonClick(t *testing.T) {
	ui(func() {
		n := 0
		b := NewButton("Нажми", func() { n++ })
		w := NewWindow("кнопка", b)
		defer w.Close()
		b.SetText("Новая надпись") // программное изменение обработчик не вызывает (ADR-8)
		if n != 0 || b.Text() != "Новая надпись" {
			t.Errorf("SetText: n = %d, текст %q", n, b.Text())
		}
		drv.SimClick(b.h)
		if n != 1 {
			t.Errorf("SimClick: OnClick вызван %d раз, ожидался 1", n)
		}
		m := 0
		b.OnClick(func() { m++ }) // новый обработчик заменяет старый
		drv.SimClick(b.h)
		if n != 1 || m != 1 {
			t.Errorf("после замены: старый %d, новый %d", n, m)
		}
		b.OnClick(nil)
		drv.SimClick(b.h)
		if m != 1 {
			t.Error("OnClick(nil) не снял обработчик")
		}
	})
}

func TestDestroyedWidgetsAreInert(t *testing.T) {
	var w *Window
	var l *Label
	var b *Button
	var box *Box
	ui(func() {
		l, b = NewLabel("метка"), NewButton("кнопка", nil)
		box = VBox(l, b)
		w = NewWindow("будет закрыто", box)
		w.Show()
	})
	ui(func() { w.Close() }) // из кода — без OnClose
	waitUntil(t, 1000, func() bool { return !w.Alive() && !l.Alive() && !b.Alive() && !box.Alive() })
	ui(func() {
		l.SetText("x")
		b.SetText("y")
		b.SetEnabled(false)
		b.SetVisible(true)
		b.SetTooltip("z")
		b.SetAlign(AlignEnd, AlignEnd)
		b.SetMinSize(10, 10)
		b.Focus()
		box.Add(NewLabel("в мёртвый ящик")).Spacing(3).Padding(3)
		w.SetTitle("z")
		w.Show()
		w.Hide()
		w.Focus()
		w.Close()
		if l.Text() != "" || b.Text() != "" || w.Title() != "" || b.Enabled() || b.Visible() {
			t.Error("геттеры уничтоженных объектов должны возвращать нулевые значения")
		}
	})
}

func TestSetChildDestroysOld(t *testing.T) {
	ui(func() {
		old, nw := NewLabel("старый"), NewLabel("новый")
		w := NewWindow("SetChild", old)
		defer w.Close()
		w.SetChild(nw)
		if old.Alive() || !nw.Alive() {
			t.Errorf("после SetChild: старый жив = %v, новый жив = %v", old.Alive(), nw.Alive())
		}
		w.SetChild(nil)
		if nw.Alive() {
			t.Error("SetChild(nil) не уничтожил ребёнка")
		}
	})
}

func TestOnClose(t *testing.T) {
	var w *Window
	asked := 0
	ui(func() {
		w = NewWindow("крестик", NewLabel("закрой меня"))
		w.OnClose(func() bool { asked++; return asked > 1 }) // первый раз — «Нет»
		w.Show()
		drv.SimClose(w.h)
	})
	waitUntil(t, 1000, func() bool { return asked == 1 })
	ui(func() {
		if !w.Alive() {
			t.Error("OnClose вернул false, а окно уничтожено")
		}
		drv.SimClose(w.h)
	})
	waitUntil(t, 1000, func() bool { return !w.Alive() })
}

func TestInsertionRules(t *testing.T) {
	ui(func() {
		l := NewLabel("один раз")
		VBox(l)
		expectPanic(t, "nui: виджет уже вставлен в контейнер", func() { HBox(l) })
		expectPanic(t, "nui: Expand вызывается до вставки виджета в контейнер", func() { Expand(l) })
		expectPanic(t, "nui: вставка nil вместо виджета", func() { VBox(nil) })
		w := NewWindow("вставка", nil)
		defer w.Close()
		expectPanic(t, "nui: виджет уже вставлен в контейнер", func() { w.SetChild(l) })
	})
}

func TestTimers(t *testing.T) {
	var after, every int
	var ev *Timer
	var once *Timer
	ui(func() {
		once = After(10, func() { after++ })
		ev = Every(10, func() {
			every++
			if every == 3 {
				ev.Stop()
				ev.Stop() // повторный Stop изнутри безопасен
			}
		})
	})
	time.Sleep(250 * time.Millisecond)
	ui(func() {
		if after != 1 || every != 3 {
			t.Errorf("After сработал %d раз (ожидался 1), Every — %d (ожидалось 3)", after, every)
		}
		once.Stop() // Stop после срабатывания After ничего не делает
		ev.Stop()
	})
}

func TestDoIsAsyncAndOrdered(t *testing.T) {
	ran := false
	ui(func() {
		Do(func() { ran = true })
		if ran {
			t.Error("Do выполнил функцию синхронно")
		}
	})
	waitUntil(t, 1000, func() bool { return ran })

	// 100 горутин × 100 Do: порядок внутри каждой горутины сохраняется.
	const G, N = 100, 100
	last := make([]int, G) // трогается только в UI-потоке
	for i := range last {
		last[i] = -1
	}
	bad := 0
	var wg sync.WaitGroup
	for g := 0; g < G; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < N; i++ {
				i := i
				Do(func() {
					if i != last[g]+1 {
						bad++
					}
					last[g] = i
				})
			}
		}(g)
	}
	wg.Wait()
	waitUntil(t, 5000, func() bool {
		for _, v := range last {
			if v != N-1 {
				return false
			}
		}
		return true
	})
	ui(func() {
		if bad != 0 {
			t.Errorf("порядок нарушен %d раз", bad)
		}
	})
}

// Раскладка Box (layout.md, §3): окно с отступом 12, ряд шириной 400, промежуток 6.
func TestBoxLayout(t *testing.T) {
	type tcase struct {
		name  string
		build func() (row *Box, a, b, c *Label)
		check func(t *testing.T, ax, aw, bx, bw, cx, cw int)
	}
	cases := []tcase{
		{"box-expand", func() (*Box, *Label, *Label, *Label) {
			a, b, c := NewLabel("A"), NewLabel("Expand"), NewLabel("C")
			return HBox(a, Expand(b), c), a, b, c
		}, func(t *testing.T, ax, aw, bx, bw, cx, cw int) {
			if ax != 12 || bx != ax+aw+6 || cx != bx+bw+6 || cx+cw != 412 {
				t.Errorf("Expand: A[%d+%d] B[%d+%d] C[%d+%d] — B должен забрать всё свободное место", ax, aw, bx, bw, cx, cw)
			}
		}},
		{"box-stretch", func() (*Box, *Label, *Label, *Label) {
			a, b, c := NewLabel("A"), NewLabel("B"), NewLabel("C")
			return HBox(a, b, Stretch(), c), a, b, c
		}, func(t *testing.T, ax, aw, bx, bw, cx, cw int) {
			if ax != 12 || bx != ax+aw+6 || cx+cw != 412 || cx < bx+bw+100 {
				t.Errorf("Stretch: A[%d+%d] B[%d+%d] C[%d+%d] — C должен прижаться к правому краю", ax, aw, bx, bw, cx, cw)
			}
		}},
		{"box-hidden", func() (*Box, *Label, *Label, *Label) {
			a, b, c := NewLabel("A"), NewLabel("скрыт"), NewLabel("C")
			b.SetVisible(false)
			return HBox(a, b, c), a, b, c
		}, func(t *testing.T, ax, aw, bx, bw, cx, cw int) {
			if cx != ax+aw+6 {
				t.Errorf("скрытый виджет занял место: A[%d+%d] C[%d+%d]", ax, aw, cx, cw)
			}
		}},
	}
	for _, c := range cases {
		var row *Box
		var la, lb, lc *Label
		var w *Window
		ui(func() {
			row, la, lb, lc = c.build()
			row.SetMinSize(400, -1)
			w = NewWindow(c.name, VBox(NewLabel("Раскладка: "+c.name), row))
			w.Show()
		})
		waitUntil(t, 2000, func() bool { _, _, rw, _ := alloc(row); return rw == 400 })
		shot(t, c.name)
		ui(func() {
			ax, _, aw, _ := alloc(la)
			bx, _, bw, _ := alloc(lb)
			cx, _, cw, _ := alloc(lc)
			c.check(t, ax, aw, bx, bw, cx, cw)
			w.Close()
		})
	}
}

func TestDebugThreadCheck(t *testing.T) {
	if !core.Debug {
		t.Skip("только с -tags nuidebug")
	}
	done := make(chan any)
	go func() {
		defer func() { done <- recover() }()
		NewLabel("не из UI-потока")
	}()
	if r, _ := (<-done).(string); r != "nui: вызов не из UI-потока — используйте nui.Do" {
		t.Fatalf("паника %q", r)
	}
}

// Сценарии, которые завершают Run, — в дочернем процессе (тот же бинарник, NUI_HELPER).

func TestAutoQuitLastWindow(t *testing.T)           { runHelper(t, "last", 0, nil) }
func TestAutoQuitIgnoresHiddenWindows(t *testing.T) { runHelper(t, "hidden", 0, nil) }
func TestHideDoesNotQuit(t *testing.T)              { runHelper(t, "hide", 0, nil) }
func TestQuitInsideSetup(t *testing.T)              { runHelper(t, "setup-quit", 0, nil) }

func runHelper(t *testing.T, name string, want int, unset []string) {
	t.Helper()
	runHelperEnv(t, name, want, nil, unset)
}

// runHelperEnv запускает сценарий в дочернем процессе с добавленными (env) и убранными
// (unset) переменными окружения.
func runHelperEnv(t *testing.T, name string, want int, env, unset []string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	for _, kv := range os.Environ() {
		skip := false
		for _, u := range unset {
			if len(kv) > len(u) && kv[:len(u)+1] == u+"=" {
				skip = true
			}
		}
		if !skip {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	cmd.Env = append(cmd.Env, env...)
	cmd.Env = append(cmd.Env, "NUI_HELPER="+name)
	start := time.Now()
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = ee.ExitCode()
	}
	if code != want {
		t.Fatalf("сценарий %s: код %d, ожидался %d; вывод:\n%s", name, code, want, out)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("сценарий %s шёл %v", name, d)
	}
}

// helper — сценарий в дочернем процессе. Код возврата 0 — как ожидалось.
func helper(name string) int {
	code := 10 // Run вернулся раньше, чем ожидалось
	err := Run(func() {
		switch name {
		case "last": // закрытие единственного окна завершает Run
			w := NewWindow("единственное", NewLabel("единственное"))
			w.Show()
			After(200, func() { code = 0; drv.SimClose(w.h) })
		case "hidden": // ADR-15: скрытое окно программу не держит
			visible := NewWindow("видимое", NewLabel("видимое"))
			visible.Show()
			NewWindow("скрытое", NewLabel("никогда не показано"))
			After(200, func() { code = 0; drv.SimClose(visible.h) })
		case "hide": // Hide сам выход не вызывает
			w := NewWindow("прячется", NewLabel("прячется"))
			w.Show()
			After(100, w.Hide)
			After(400, func() { code = 0; Quit() })
		case "setup-quit": // Quit внутри setup: Run возвращается сразу после setup
			NewWindow("не понадобится", nil).Show()
			Quit()
			code = 0
		case "nodisplay":
			code = 4 // setup не должен вызываться
		case "args": // последний аргумент командной строки — тот, что передал родитель
			if a := Args(); len(a) > 1 && a[len(a)-1] == os.Getenv("NUI_ARG") {
				code = 0
			} else {
				code = 7
				println("Args:", len(a))
			}
			Quit()
		case "clipboard": // текст, который положил в буфер родительский процесс
			if got := Clipboard(); got == os.Getenv("NUI_CLIP") {
				code = 0
			} else {
				code = 6
				println("в буфере:", got)
			}
			Quit()
		case "openuri", "showinfolder": // подставная программа пишет в $NUI_REC, что открыли
			ok := false
			if name == "openuri" {
				ok = OpenURI("https://example.org/nui?a=1")
			} else {
				ok = ShowInFolder(os.Getenv("NUI_FILE"))
			}
			if !ok {
				code = 5
				Quit()
				break
			}
			Every(50, func() {
				if b, _ := os.ReadFile(os.Getenv("NUI_REC")); len(b) > 0 {
					code = 0
					Quit()
				}
			})
		}
		After(4000, func() { code = 11; Quit() }) // страховка: Run не завершился
	})
	if err == ErrNoDisplay {
		if code == 4 {
			return 4
		}
		return 3
	}
	if err != nil {
		return 2
	}
	return code
}
