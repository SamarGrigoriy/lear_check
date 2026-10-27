// Package nui — маленькая библиотека нативного GUI для Go: настоящие виджеты ОС
// (GTK3 на Linux) без браузера и OpenGL. Контракт API — docs/api.md.
//
// Все функции и методы вызываются только в UI-потоке: внутри setup у Run,
// в обработчиках On*, таймерах, Watch и функциях из Do. Из любого потока
// можно вызывать единственную функцию — Do (и запись в ChatStream).
package nui

import (
	"errors"
	"runtime"

	"nui/internal/core"
	"nui/internal/drv"
)

// init выполняется в главном потоке: главная горутина остаётся на нём (runtime.md, §1).
func init() { runtime.LockOSThread() }

// ErrNoDisplay — Run не смог инициализировать GTK: нет DISPLAY/WAYLAND_DISPLAY и т. п.
var ErrNoDisplay = errors.New("nui: нет дисплея или GTK не инициализировался")

const (
	stIdle = iota
	stSetup
	stRunning
	stDone
)

var (
	state     uint8
	quitAsked bool
	appID     string
	wins      []*Window // живые окна — для автовыхода (ADR-15)
)

// SetAppID задаёт имя приложения до Run, например "esp32-logger". Оно должно совпадать
// с именем .desktop-файла: от него зависят WM_CLASS и app_id на Wayland.
// По умолчанию берётся имя исполняемого файла.
func SetAppID(id string) { appID = id }

// Run инициализирует GTK, вызывает setup и крутит цикл событий до Quit или до тех пор,
// пока не будет уничтожено окно, а среди оставшихся не останется видимых (ADR-15).
// Вызывается один раз и только из main. Если GTK не инициализировался, setup
// не вызывается и Run возвращает ErrNoDisplay.
func Run(setup func()) error {
	if state != stIdle || !drv.IsMainThread() {
		panic("nui: Run вызывается один раз и из main")
	}
	if !drv.Init(appID) {
		state = stDone
		return ErrNoDisplay
	}
	core.MarkUIThread(drv.IsMainThread)
	core.SetWaker(drv.Wake)
	state = stSetup
	setup()
	if !quitAsked {
		state = stRunning
		drv.Main()
	}
	state = stDone
	core.Close()
	return nil
}

// Quit просит цикл событий завершиться; Run возвращает nil. Внутри setup Run вернётся
// сразу после setup. Если открыт модальный диалог, Quit подействует после его закрытия.
func Quit() {
	core.CheckThread()
	switch state {
	case stSetup:
		quitAsked = true // gtk_main ещё не запущен — gtk_main_quit вызывать нельзя
	case stRunning:
		if !quitAsked {
			quitAsked = true
			drv.Quit()
		}
	}
}

// Do ставит fn в очередь на выполнение в UI-потоке. Единственная функция nui,
// которую можно вызывать из любого потока или горутины. Всегда асинхронна,
// порядок вызовов сохраняется; после завершения Run функция не выполнится.
//
// Do срабатывает раньше перерисовки окна: вызывайте его пачками, не чаще
// 20–50 раз в секунду, иначе окно перестанет перерисовываться (runtime.md, §8).
func Do(fn func()) {
	if core.Post(fn) {
		drv.Wake()
	}
}

// checkCreate — общие проверки конструкторов в отладочной сборке.
func checkCreate() {
	core.CheckThread()
	if core.Debug && state == stIdle {
		panic("nui: виджеты создаются только после Run — внутри setup")
	}
}
