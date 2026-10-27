package main

/*
#include "s1.h"
*/
import "C"

import (
	"os"
	"runtime"
	"syscall"
)

// init выполняется в главном потоке: главная горутина остаётся на нём (runtime.md, §1).
func init() { runtime.LockOSThread() }

// Проверки сценария S1 (docs/roadmap.md).
const (
	chkClick    = iota // 1. клик меняет текст метки
	chkDo              // 2. Do из горутины (без планировщика — из таймера)
	chkAfter           // 3. After меняет заголовок окна
	chkWatch           // 4. Watch на pipe
	chkCloseNo         // 5а. крестик → «Нет» → окно осталось
	chkCloseYes        // 5б. крестик → «Да» → окно закрыто, цикл завершён
	chkThread          // обработчики выполняются в главном потоке
	chkUTF8            // кириллица и эмодзи проходят Go → GTK → Go без потерь
	nChecks
)

var names = [nChecks]string{
	"1. клик меняет текст метки",
	"2. Do из горутины",
	"3. After меняет заголовок окна",
	"4. Watch на pipe",
	"5а. крестик → «Нет» → окно осталось",
	"5б. крестик → «Да» → окно закрыто, цикл завершён",
	"обработчики в главном потоке",
	"кириллица и эмодзи без потерь",
}

var passed [nChecks]bool

var (
	shotDir   string // -shots DIR: каталог снимков; сами снимки — только со сборкой -tags shots
	stressSec int    // -stress SEC: проверка сборщика мусора вместо автопилота
)

func main() {
	auto := false
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-auto":
			auto = true
		case "-shots": // каталог для снимков фотоотчёта
			if i+1 < len(os.Args) {
				i++
				shotDir = os.Args[i]
			}
		case "-stress": // roadmap.md, S2, п. 6
			if i+1 < len(os.Args) {
				i++
				stressSec = atoi(os.Args[i])
			}
		}
	}
	if C.gtk_init_check(nil, nil) == 0 {
		println("s1: нет дисплея или GTK не инициализировался")
		os.Exit(2)
	}

	status := NewLabel("Нажмите кнопку")
	doLabel := NewLabel("Do: ждём горутину…")
	pipeLabel := NewLabel("Watch: ждём данные из pipe…")

	// 1. Клик меняет текст метки.
	clicks := 0
	button := NewButton("Нажми меня", func() {
		clicks++
		want := "Нажато " + itoa(clicks) + " · Привет, мир 🙂"
		status.SetText(want)
		passed[chkClick] = true
		passed[chkUTF8] = status.Text() == want
		passed[chkThread] = isMainThread()
	})

	win := NewWindow("nui · спайк S1", VBox(status, button, doLabel, pipeLabel))

	// 5. Закрытие крестиком с подтверждением: delete-event с ответом 1 / 0.
	closedByUser := false
	win.onClose = func() bool {
		yes := Ask(win, "Закрыть окно?")
		if yes {
			closedByUser = true
		} else {
			After(0, func() { passed[chkCloseNo] = win.id != 0 })
		}
		return yes
	}

	// 2. Do из горутины (в сборке без планировщика — из таймера).
	background(500, func() {
		Do(func() {
			doLabel.SetText("Do: выполнено в UI-потоке")
			passed[chkDo] = isMainThread()
		})
	})

	// 3. After меняет заголовок окна.
	const title = "nui · спайк S1 · After сработал"
	After(800, func() {
		win.SetTitle(title)
		passed[chkAfter] = win.Title() == title
	})

	// 4. Watch на pipe: «фон» пишет, обработчик в UI-потоке читает.
	var p [2]int
	if err := syscall.Pipe2(p[:], syscall.O_NONBLOCK|syscall.O_CLOEXEC); err != nil {
		panic(err)
	}
	const msg = "данные из горутины"
	var watcher *Watcher
	watcher = Watch(p[0], ioRead, func(got uint) {
		var buf [256]byte
		n, _ := syscall.Read(p[0], buf[:])
		if n > 0 {
			pipeLabel.SetText("Watch: " + string(buf[:n]))
			passed[chkWatch] = string(buf[:n]) == msg && isMainThread()
		}
		watcher.Stop() // до close(fd) — по контракту
		syscall.Close(p[0])
	})
	background(600, func() {
		syscall.Write(p[1], []byte(msg))
		syscall.Close(p[1])
	})

	win.Show()
	switch {
	case stressSec > 0:
		stress(win, button, status)
	case auto:
		autopilot(win, button)
	}
	C.gtk_main()

	if stressSec > 0 {
		stressReport()
		return
	}
	passed[chkCloseYes] = closedByUser && win.id == 0
	report()
}

// autopilot имитирует пользователя. Ответы в диалоге приходят из таймеров,
// которые срабатывают внутри вложенного цикла gtk_dialog_run.
// Снимки (тег shots) — с паузой после действия, чтобы GTK успел перерисовать окно;
// новому окну диалога нужно больше времени, чем перерисовке метки.
func autopilot(win *Window, button *Button) {
	After(250, func() { shot("start") })
	After(300, func() { C.gtk_button_clicked((*C.GtkButton)(button.h)) })
	After(450, func() { shot("click") })
	After(1000, func() { shot("do-watch-after") })
	After(1200, func() { C.gtk_window_close((*C.GtkWindow)(win.h)) }) // откроет «Закрыть окно?»
	After(1900, func() { shot("ask") })
	After(2000, func() { C.nui_test_respond(0) }) // «Нет»
	After(2200, func() { shot("after-no") })
	After(2400, func() { C.gtk_window_close((*C.GtkWindow)(win.h)) })
	After(2700, func() { C.nui_test_respond(1) }) // «Да» → окно уничтожено → gtk_main_quit
	After(10000, func() {
		println("s1: автопилот не дождался завершения")
		os.Exit(3)
	})
}

// Проверка сборщика мусора (roadmap.md, S2, п. 6): каждые 10 мс — 64 КБ мусора,
// смена текста метки и одноразовый таймер (реестр выдаёт и освобождает id).
// В конце клик должен дойти до своего обработчика: Go-объекты из реестра живы.
var stressSteps, stressRSS0, stressRSS1 int
var stressClickOK bool

func stress(win *Window, button *Button, status *Label) {
	stressRSS0 = rssKB()
	var keep [][]byte
	t := Every(10, func() {
		stressSteps++
		b := make([]byte, 64*1024)
		b[0], b[len(b)-1] = byte(stressSteps), byte(stressSteps>>8)
		if len(keep) == 16 {
			keep = keep[:0] // часть мусора живёт дольше одного шага
		}
		keep = append(keep, b)
		status.SetText("Стресс: шаг " + itoa(stressSteps))
		After(0, func() {})
	})
	After(stressSec*1000, func() {
		t.Stop()
		passed[chkClick] = false
		C.gtk_button_clicked((*C.GtkButton)(button.h))
		stressClickOK = passed[chkClick] && passed[chkUTF8]
		stressRSS1 = rssKB()
		shot("stress-end")
		win.onClose = nil
		C.gtk_window_close((*C.GtkWindow)(win.h))
	})
}

func stressReport() {
	mark := "✓"
	if !stressClickOK {
		mark = "✗"
	}
	println(mark, "сборщик мусора:", stressSteps, "шагов по 64 КБ за", stressSec, "с; после — клик работает:", stressClickOK)
	println("  RSS: в начале", stressRSS0, "КБ, в конце", stressRSS1, "КБ")
	if !stressClickOK {
		os.Exit(1)
	}
}

// rssKB — резидентная память процесса из /proc/self/statm (страницы по 4 КБ).
func rssKB() int {
	b, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return -1
	}
	i := 0
	for i < len(b) && b[i] != ' ' {
		i++
	}
	return atoi(string(b[i+1:])) * 4
}

func report() {
	ok := true
	for i, name := range names {
		mark := "✓"
		if !passed[i] {
			mark, ok = "✗", false
		}
		println(mark, name)
	}
	if !ok {
		println("s1: есть непройденные проверки")
		os.Exit(1)
	}
	println("s1: все проверки пройдены")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for ; n > 0; n /= 10 {
		i--
		b[i] = byte('0' + n%10)
	}
	return string(b[i:])
}

// atoi читает число до первого нецифрового символа.
func atoi(s string) int {
	n := 0
	for i := 0; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
	}
	return n
}
