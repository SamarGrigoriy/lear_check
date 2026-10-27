# nui — нативный GUI для Go

> **Статус (09.10.2026): Linux готов, Windows — первая версия.**
> - **Linux, фазы 1–3** ([фаза 1](docs/reports/phase-1.md), [фаза 2](docs/reports/phase-2.md), [фаза 3](docs/reports/phase-3.md)): ядро, драйвер GTK3, все виджеты MVP, диалоги и оболочка. `examples/form` со всеми виджетами — 41 808 байт в TinyGo.
> - **Windows, фаза 5** ([отчёт](docs/reports/phase-5.md)): движок раскладки, повторяющий GTK до пикселя; драйвер Win32. Примеры и интеграционные тесты проходят под Wine; на настоящей Windows их ещё не запускали.
> - **Готовый чат** `nui.NewChat()` — например, для общения с ИИ: ответ пишется частями из горутины; история загружается из JSON и сохраняется в JSON, в том числе в формате API моделей ([api.md, §18](docs/api.md), [отчёт](docs/reports/chat.md), `examples/chat`).
> - **Оформление** ([отчёт](docs/reports/styles.md)): фон у окна, Box и Grid, цвет и фон у подписей, кнопок, полей и чата, размер шрифта подписей (ADR-19). **Окно на весь экран:** `SetMaximized`, `SetFullscreen`.
> - **Буфер обмена:** `nui.SetClipboard`, `nui.Clipboard`; `Chat.LastMessage` — для кнопки «Скопировать ответ» ([api.md, §17](docs/api.md), [отчёт](docs/reports/clipboard.md)).
> - **Перетаскивание файлов** на окно `Window.OnDropFiles`; `nui.Args()` — аргументы командной строки без потерь русских букв в TinyGo на Windows. Пример `examples/viewer` — просмотрщик картинок: список, картинка, открытие нескольких файлов, перетаскивание, клавиши ([отчёт](docs/reports/drop.md)).
> - **Контекстное меню** `chat.SetContextMenu(nui.Item("Скопировать ответ", …))` — по правому щелчку у любого виджета, окна и чата ([отчёт](docs/reports/context.md)).
> - **Числовое поле** `nui.NewSpin(1, 65535)` — целое число со стрелками; набранное принимается по Enter или при уходе фокуса ([api.md, §11](docs/api.md), [отчёт](docs/reports/spin.md)).
> - **Переключатель** `nui.NewRadio("PNG", "JPEG", "WebP")` — выбран ровно один вариант; в столбик или в ряд ([api.md, §9](docs/api.md), [отчёт](docs/reports/radio.md)).
> - **Вкладки** `nui.NewTabs().Add("Общие", general).Add("Сеть", network)` — `OnChange`, `SetCurrent`, `Remove`; Ctrl+PgUp/PgDn ([api.md, §15](docs/api.md), [отчёт](docs/reports/tabs.md)).
> - **Разделитель панелей** `nui.HSplit(menu, chat)` — границу двигают мышью ([отчёт](docs/reports/split.md)).
> - **Холст** `nui.NewCanvas(func(p *nui.Painter) {…})` — своё рисование: графики, индикаторы; мышь (`OnMouse`), чёткость на HiDPI ([отчёт](docs/reports/canvas.md)).
> - **Меню окна** `w.SetMenu(nui.NewMenu("Файл", nui.Item("Открыть…", open).Shortcut('O', nui.ModCtrl), …))`: подменю, флажки, сочетания клавиш, которые показываются в меню и работают и в русской раскладке ([отчёт](docs/reports/menu.md)).
> - **Клавиши окна** `Window.OnKey` (Ctrl+S работает и в русской раскладке); в чате многострочный ввод: Enter отправляет, Shift+Enter — новая строка ([отчёт](docs/reports/keys.md)).
> - **Список и таблица** `nui.NewList` — строки по требованию, миллион строк открывается сразу ([api.md, §20](docs/api.md), [отчёт](docs/reports/list.md)).
> - **Картинки и выбор нескольких файлов:** `nui.NewImage` — PNG, JPEG, GIF, BMP, чёткие на HiDPI; `nui.OpenFiles` ([api.md, §19 и §16](docs/api.md), [отчёт](docs/reports/image.md)).
> - Фаза 4 (приёмочное приложение) отложена. Реализация ведётся на Debian.

nui — маленькая библиотека для настольных приложений на Go. Интерфейс строится из **настоящих виджетов операционной системы**: GTK3 на Linux, Win32 на Windows. Виджеты вызываются напрямую, без браузера, OpenGL и толстых обёрток.

Цели (подробно — в [docs/goals.md](docs/goals.md)):

- Размер приложения уровня «форма с полями и кнопками» — **≤ 250 КБ** при сборке TinyGo и около 1,5 МБ при сборке обычным Go.
- Во время работы нужен только системный GTK3. Он и так стоит в любом десктопном Linux.
- В `go.mod` нет внешних зависимостей.
- Hello world (`examples/hello`) укладывается в 15 строк.

## Как выглядит код

Небольшая форма: поле ввода, метка и кнопка по умолчанию — Enter в поле нажимает «OK».

```go
package main

import "nui"

func main() {
	err := nui.Run(func() {
		name := nui.NewEntry()
		hello := nui.NewLabel("Как тебя зовут?")
		ok := nui.NewButton("OK", func() {
			hello.SetText("Привет, " + name.Text())
		})
		w := nui.NewWindow("Demo", nui.VBox(hello, name, nui.Stretch(), ok))
		w.SetDefault(ok)
		w.Show()
	})
	if err != nil {
		println(err.Error())
	}
}
```

## Как подключить к своей программе

Модуль пока не опубликован, поэтому подключается локально, через `replace`:

```sh
mkdir myapp && cd myapp
go mod init myapp
go mod edit -require=nui@v0.0.0 -replace=nui=/путь/к/nui
# main.go — как в примере выше
go build -o myapp .        # Linux: нужны Go 1.22+, gcc и libgtk-3-0; -dev-пакеты не нужны
```

Windows — кросс-сборкой на Linux (пакет `gcc-mingw-w64-x86-64`):

```sh
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags "-H windowsgui" -o myapp.exe .
```

Для современных контролов и правильного масштаба на Windows положите рядом `app.manifest` и `app.rc` из [examples/hello](examples/hello) и соберите ресурс: `x86_64-w64-mingw32-windres app.rc -O coff -o rsrc_windows_amd64.syso`. Малая сборка TinyGo и всё прочее описаны в [build.md](docs/build.md).

## Как это устроено

Публичный пакет `nui` даёт простые типы (`Window`, `Button`, `Entry` и т. д.). Под ним лежат два внутренних пакета:

- `internal/core` — чистый Go без cgo. В нём реестр объектов, диспетчер событий и очередь `Do`. Тестируется на любой машине.
- `internal/drv` — драйвер ОС: GTK3 на Linux, Win32 на Windows. Он вызывает систему через cgo, но вместо системных заголовков использует **собственные прототипы функций**, сверенные с настоящими заголовками (`make check-protos`). Поэтому `-dev`-пакеты не нужны, а код собирается и обычным Go (gc), и TinyGo.
- `internal/layout` — движок раскладки на чистом Go для Windows, где нет контейнеров. Он повторяет раскладку GTK до пикселя, это проверяют golden-тесты, снятые с GTK.

События из C возвращаются в Go через общие «трамплины»: четыре на GTK, один WndProc на Windows. Объекты передаются только числовыми id, Go-указатели в C не попадают. Таймеры, ожидание файловых дескрипторов и `Do` работают на цикле событий ОС, поэтому библиотеке не нужны собственные горутины.

## Документы

Читать по порядку:

| # | Документ | О чём |
|---|---|---|
| 1 | [goals.md](docs/goals.md) | Требования, измеримые цели, эталон по размеру, сравнение с Fyne, Wails, Gio и др., нецели |
| 2 | [decisions.md](docs/decisions.md) | Архитектурные решения (ADR): что выбрано, почему, какой ценой |
| 3 | [architecture.md](docs/architecture.md) | Слои, пакеты, файлы, пути вызова и события |
| 4 | [api.md](docs/api.md) | **Контракт публичного API** — главный документ для реализации |
| 5 | [runtime.md](docs/runtime.md) | Ядро: поток UI, реестр, трамплины, `Do`, таймеры, жизненный цикл, строки |
| 6 | [layout.md](docs/layout.md) | Раскладка: семантика VBox/HBox/Grid, отображение на GTK, движок для Win32 |
| 7 | [driver-gtk3.md](docs/driver-gtk3.md) | Драйвер Linux: контракт, список C-функций, обёртки, линковка, проверка прототипов |
| 8 | [driver-win32.md](docs/driver-win32.md) | Драйвер Windows (фаза 5) |
| 9 | [size.md](docs/size.md) | Бюджет размера и правила кода, которые его держат |
| 10 | [build.md](docs/build.md) | Сборка на Debian: пакеты, Go, TinyGo, цели Makefile |
| 11 | [roadmap.md](docs/roadmap.md) | Спайки S1–S4 и фазы 1–6 с критериями готовности |
| 12 | [testing.md](docs/testing.md) | Юнит-тесты, тесты под Xvfb, selftest, контроль размера, ручной чек-лист |
| 13 | [acceptance.md](docs/acceptance.md) | Приёмочное приложение: Linux-версия «Записи логов ESP32» |

**Сборка и проверки** ([build.md](docs/build.md), §4, §7): `make dev` — примеры обычным Go, `make small` — TinyGo, `make check` — статические проверки (Linux и Win32), `make docker-ci` — тесты на GTK под Xvfb, selftest, размеры и ABI в чистом Debian 12. Для Windows: `NUI_IMAGE=nui-win tools/docker-run.sh make win win-selftest win-test` — сборка обоими тулчейнами, selftest и интеграционные тесты под Wine. Спайки S1, S3 и S4 — в [spike/](spike/).

## Чего здесь нет и не будет

Собственной отрисовки виджетов, HTML/webview, анимаций, тем оформления, data binding, декларативного DSL, мобильных платформ. Внешний вид даёт ОС. Подробнее — раздел «Нецели» в [goals.md](docs/goals.md).

## Имя модуля

В документах модуль называется `nui`, а импорт выглядит как `import "nui"`. При публикации замените имя на реальный путь репозитория, например `github.com/<owner>/nui`.
