# Архитектура

## Слои

```
┌───────────────────────────────────────────────────────────────┐
│  Приложение                         examples/hello, logger …  │
└───────────────┬───────────────────────────────────────────────┘
                │  публичный API (api.md)
┌───────────────▼───────────────────────────────────────────────┐
│  nui                                                          │
│  типы Window/Button/…, раскладка VBox/HBox/Grid, диалоги,     │
│  Run/Quit/Do/After/Every/Watch, правила ADR-8, ADR-14, ADR-15 │
└───────┬──────────────────────────────────────┬────────────────┘
        │ Register/Release/Post               │ drv.LabelSetText(h, s) …
┌───────▼──────────────────┐   Dispatch   ┌───▼────────────────────────────┐
│  internal/core           │◄─────────────┤  internal/drv                  │
│  чистый Go, без cgo:     │              │  cgo; файлы *_linux.* (GTK3),  │
│  реестр, kind/ud,        │              │  *_windows.* (Win32)           │
│  диспетчер, очередь Do   │              │  трамплины //export, обёртки C │
└──────────────────────────┘              └───┬────────────────────────────┘
                                              │ вызовы C, сигналы GTK
                                  ┌───────────▼─────────────────────────┐
                                  │ libgtk-3 · libgobject · libglib ·   │
                                  │ libgio · libpango (системные .so)   │
                                  └─────────────────────────────────────┘
```

## Пакеты и правило зависимостей

| Пакет | Что внутри | Может импортировать |
|---|---|---|
| `nui` | публичный API, состояние виджетов на стороне Go, правила семантики | `internal/core`, `internal/drv`, белый список stdlib |
| `internal/core` | реестр объектов, кодирование `user_data`, `Dispatch`, очередь `Do`, отладочные проверки | только белый список stdlib, **без cgo** |
| `internal/drv` | вызовы ОС; одинаковый набор функций для каждой ОС (контракт) | `internal/core`, `C`, белый список stdlib |
| `internal/layout` (фаза 5) | движок раскладки для Win32, чистый Go | только белый список stdlib |

Правила:

- `core` не знает ни о GTK, ни о nui. Значит, его можно тестировать на любой машине командой `go test` без cgo.
- `drv` не знает о типах nui. Он оперирует хендлами `drv.H`, числами и строками. Вызывать nui он может только через `core.Dispatch`.
- В `nui` нет ни одной строки C, весь cgo — в `drv`.
- Циклических зависимостей нет: `nui → drv → core`, `nui → core`.

## Дерево репозитория

```
nui/
├── go.mod                   module nui; go 1.22; без require
├── Makefile                 dev, small, test, selftest, size, deps, abi, check-protos, windows, ci
├── size-budget.txt          бюджеты размера (size.md)
├── README.md
├── docs/                    эта документация
│
├── app.go                   Run, Quit, SetAppID, ErrNoDisplay, init(){LockOSThread}
├── do.go                    Do
├── timer.go                 After, Every, Timer
├── watch_unix.go            Watch, Watcher, IOCond          //go:build unix
├── widget.go                Widget, widget (база), Align, общие методы
├── color.go                 Color, RGB, NoColor
├── window.go                Window
├── label.go  button.go  entry.go  check.go  combo.go
├── slider.go  progress.go  textview.go  image.go  list.go
├── chat.go                  Chat, ChatStream — готовый чат из виджетов выше
├── misc.go                  Separator, Frame, Scroll, spacer (Stretch)
├── box.go                   Box, VBox, HBox, Expand, Stretch
├── grid.go                  Grid
├── dialog.go                Message, Alert, Ask, OpenFile, OpenFiles, SaveFile, ChooseFolder, Filter
├── shell.go                 OpenURI, ShowInFolder
├── clipboard.go             SetClipboard, Clipboard
├── key.go                   Key, Mod — для Window.OnKey
├── menu.go                  Menu, MenuItem, Window.SetMenu, SetContextMenu
├── canvas.go                Canvas, Painter, MouseEvent
├── split.go                 Split, HSplit, VSplit
├── tabs.go                  Tabs — вкладки
├── radio.go                 Radio — переключатель
├── spin.go                  Spin — числовое поле
├── args.go                  Args — аргументы командной строки
├── nui_gui_test.go          интеграционные тесты            //go:build gui
│
├── internal/core/
│   ├── kind.go              Kind, UD, Decode
│   ├── registry.go          Receiver, Register, Release, Lookup
│   ├── dispatch.go          Dispatch
│   ├── queue.go             Post, Drain
│   ├── text.go              itoa, FixUTF8, dir — вместо fmt/strings/path (size.md)
│   ├── image.go             Fit (SetFit), Orientation (EXIF-поворот JPEG)
│   ├── json.go              ParseJSON, QuoteJSON — история чата без encoding/json
│   ├── debug_on.go          проверка потока и прочее         //go:build nuidebug
│   ├── debug_off.go         пустые заглушки                  //go:build !nuidebug
│   └── *_test.go
│
├── internal/drv/
│   ├── gtk3_linux.h         непрозрачные типы, константы, ПРОТОТИПЫ (только объявления)
│   ├── wrap_linux.c         C-обёртки nui_* (определения)
│   ├── link_linux.go        #cgo LDFLAGS (только библиотеки), тип H, Nil, cstr/cpath/gostr
│   ├── export_linux.go      4 трамплина //export → core.Dispatch; 2 экспорта модели List
│   ├── app_linux.go         Init, Main, Quit, Wake, TimerAdd, WatchAdd, SourceRemove
│   ├── window_linux.go      WindowNew, WindowSetChild, WindowShow …
│   ├── widget_linux.go      общие функции: Show, Destroy, SetEnabled, SetAlign …
│   ├── controls_linux.go    Label*, Button*, Entry*, Check*, Combo*, Slider*, Progress*, Text*
│   ├── layout_linux.go      Box*, Grid*, Frame*, Scroll*, Separator*
│   ├── dialog_linux.go      Message, FileDialog, OpenURI, ShowInFolder
│   ├── image_linux.go       ImageNew, ImageLoad, ImageShow, ReadFile (GdkPixbuf)
│   ├── list_linux.go        ListNew, ListSetRows … — модель GtkTreeModel в wrap_linux.c
│   ├── clipboard_linux.go   ClipboardSet, ClipboardText
│   ├── key_linux.go         KeyDecode, WindowSetKeys (OnKey)
│   ├── drop_linux.go        WindowSetDrop, DropPaths (OnDropFiles)
│   ├── menu_linux.go        MenuBar, MenuSub, MenuItemNew … (окно перестраивается в wrap_linux.c)
│   ├── canvas_linux.go      CanvasNew, Paint*, MouseDecode (cairo)
│   ├── split_linux.go       SplitNew, SplitSet, SplitSetPosition (GtkPaned)
│   ├── tabs_linux.go        TabsNew, TabsAdd, TabsSetCurrent … (GtkNotebook)
│   ├── radio_linux.go       RadioNew, RadioAdd, RadioSelected … (GtkRadioButton)
│   ├── spin_linux.go        SpinNew, SpinSetValue, SpinValue (GtkSpinButton)
│   ├── args_linux.go        Args (/proc/self/cmdline)
│   ├── sim_linux.go         Sim* — имитация действий пользователя  //go:build gui
│   ├── contract.go          общие константы контракта (MsgInfo, FileOpen, Filter)
│   ├── win32_windows.h      собственные прототипы Win32
│   ├── wrap_windows.c       класс окна, цикл, шрифты под DPI, измерение текста, диалоги
│   ├── export_windows.go    трамплин nuiWndProc //export
│   ├── node_windows.go      узлы (Box, Grid — без HWND), раскладка, DPI, прокрутка
│   ├── *_windows.go         остальное — как у Linux: app, window, widget, controls, layout, dialog,
│   │                        image (GDI+), list (ListView), canvas (GDI+), split, key, keyname, menu, drop, args,
│   │                        clipboard, sim
│   └── link_tinygo_windows.go  путь к библиотекам импорта mingw для TinyGo
│
├── internal/layout/         движок раскладки для Win32: Measure, Arrange (layout.md, §7)
│   ├── golden/              язык случаев и golden-файлы (только для тестов)
│   └── testdata/            cases/*.txt, golden/*.txt — снимаются с GTK (make golden)
│
├── examples/
│   ├── hello/               минимальный пример (≤ 15 строк)
│   ├── form/                все виджеты MVP, Grid, диалоги
│   ├── chat/                Chat с эхо-ботом, «Скопировать ответ»
│   ├── viewer/              просмотр картинок: List, Image, OpenFiles, перетаскивание, OnKey
│   └── logger/              приёмочное приложение (acceptance.md; фаза 4 отложена)
│
├── targets/
│   └── linux-glibc-amd64.json   цель TinyGo: системная glibc вместо musl (ADR-17)
│
├── docs/reports/            отчёты фаз с фотоотчётами (снимки — make shots)
│
└── tools/
    ├── docker-run.sh        команда в чистом Debian 12 (образ nui-bookworm)
    ├── bookworm.Dockerfile  образ: runtime GTK, Xvfb, openbox, rootshot, make
    ├── rootshot.c           снимок экрана одним XGetImage(root) — не xwd (spike/s1)
    ├── abicheck.sh          glibc не новее Debian 12
    ├── depcheck.sh          белый список импортов
    ├── sizecheck.sh         сверка с size-budget.txt
    └── check-protos.sh      сверка прототипов с <gtk/gtk.h>
```

## Объектная модель

```go
// internal/core
type Kind uint8                      // тип события, 5 бит
type Receiver interface {
	Event(k Kind, arg uintptr) uintptr // результат нужен только событиям «с ответом»
}

// nui
type widget struct {                 // встраивается в каждый виджет
	h      drv.H  // хендл ОС (GtkWidget* / HWND)
	id     uint32 // id в реестре; 0 — объект уничтожен
	quiet  uint8  // >0 — идёт программный Set*, события пользователю не передаём
	expand bool   // для Box: забирать свободное место
	placed bool   // уже вставлен в контейнер (повторная вставка — паника)
}

type Button struct {
	widget
	onClick func()
}

// Обработчик — на неэкспортируемом типе с тем же устройством: метод Event
// не попадает в публичный API, а преобразование указателя ничего не стоит.
type buttonEv Button

func (b *buttonEv) Event(k core.Kind, _ uintptr) uintptr { … } // реализует core.Receiver

// в конструкторе: b.id = core.Register((*buttonEv)(b))
```

- Каждый тип nui регистрирует в реестре свой `*xxxEv` — он реализует `core.Receiver`. Реестр хранит интерфейсы, поэтому Go-объект жив, пока жив виджет, даже если приложение не держит на него ссылку.
- `Window` не встраивает `widget`: у окна свой набор методов, и `Widget` оно не реализует — в контейнер его не вставить.
- Таймеры (`*Timer`) и наблюдатели (`*Watcher`) тоже реализуют `Receiver` и лежат в том же реестре: один механизм на всё.
- Методы публичного API **встраиваются** из `widget` (`SetEnabled`, `SetVisible` …). Каждый метод начинается с `if w.id == 0 { return }`: вызов у уничтоженного объекта ничего не делает.

## Путь вызова сверху вниз: `label.SetText("Привет")`

```
nui.(*Label).SetText(s)
  ├─ if l.id == 0 → return                     // уничтожен
  ├─ l.quiet++                                 // ADR-8: не будить On*
  ├─ drv.LabelSetText(l.h, s)
  │    └─ C.gtk_label_set_text((*C.GtkLabel)(h), cstr(s))
  │         cstr(s): []byte(s)+"\x00" в памяти Go (битый UTF-8 → U+FFFD), живёт до конца вызова (ADR-13)
  └─ l.quiet--
```

## Путь события снизу вверх: пользователь нажал кнопку

```
GTK: сигнал "clicked" на GtkButton
  └─ C: nuiSignal(obj, user_data)              // подключено через nui_connect() в wrap_linux.c
       └─ Go //export nuiSignal (export_linux.go)
            └─ core.Dispatch(ud, 0)
                 ├─ id, kind := Decode(ud)     // kind = EvClick
                 ├─ r := objs[id]              // *nui.Button
                 └─ r.Event(EvClick, 0)
                      └─ if b.quiet == 0 && b.onClick != nil → b.onClick()
```

Событие с ответом, например закрытие окна крестиком:

```
GTK: "delete-event" → nuiSignalRet(obj, event, ud) → core.Dispatch(ud, event)
  → (*Window).Event(EvClose) → onClose() == false → вернуть 1 → GTK не закрывает окно
```

## Путь `Do` из другой горутины

```
горутина:   nui.Do(fn)
              └─ core.Post(fn) → wake?            // под mutex: добавить в очередь;
                                                   // wake=true, только если очередь была пуста
              └─ if wake → drv.Wake()              // g_idle_add_full(..., nuiSource, UD(0, KDo))
главный поток (цикл GLib):
            nuiSource(ud) → core.Dispatch(ud, 0) → kind=KDo → core.Drain()
              └─ забрать всю очередь под mutex, выполнить функции по порядку
```

## Путь таймера и `Watch`

```
nui.Every(1000, fn) → t := &Timer{fn} → id := core.Register(t)
                    → t.src = drv.TimerAdd(1000, core.UD(id, KTimer))   // g_timeout_add
раз в секунду: nuiSource(ud) → Dispatch → (*Timer).Event(KTimer) → fn() → 1 (продолжать)
t.Stop() → drv.SourceRemove(t.src); core.Release(id)

nui.Watch(fd, IORead, fn) → то же, через g_unix_fd_add; arg = набор GIOCondition
```

## Кто за что отвечает

| Вопрос | Кто решает |
|---|---|
| Когда вызывать обработчик пользователя | `nui` (флаги `quiet`, `id`) |
| Куда доставить событие | `core.Dispatch` |
| Как подписаться на сигнал ОС | `drv` (внутри `XxxNew`, по id объекта) |
| Кто владеет памятью виджета | ОС: GTK уничтожает детей вместе с окном. nui только узнаёт об этом по `destroy` |
| Кто владеет Go-объектом | реестр `core`, до события `EvDestroy` |
| Когда завершать программу | `nui`: окно уничтожено, и видимых окон не осталось (ADR-15) |
| Как разместить виджеты | GTK-контейнеры; на Win32 — `internal/layout` |
| В каком потоке что выполняется | всё в UI-потоке; исключения — только `Do` и `core.Post` |

## Как добавить новый виджет (чек-лист)

1. Добавить в [api.md](api.md) тип, конструктор, методы и события с семантикой.
2. Добавить в `gtk3_linux.h` нужные прототипы (с точными типами из документации GTK), запустить `make check-protos`.
3. Написать в `drv`: `XxxNew(id uint32, …) H` (создание, подписка на сигналы через `nui_connect(h, "signal", UD(id, kind))`, `gtk_widget_show`) и сеттеры/геттеры.
4. Написать в `nui`: `type Xxx struct{ widget; on… func() }`, конструктор (`core.Register`, `drv.XxxNew`), методы с `quiet` — **каждый**, который меняет виджет из кода, а не только `Set*` ([runtime.md](runtime.md), §7), `Event`.
5. Добавить `Sim*` для тестов и интеграционный тест: программные методы (`Set*`, `Add`, `Clear`…) не вызывают On, Sim вызывает, после destroy методы ничего не делают.
6. Добавить в таблицу [driver-gtk3.md](driver-gtk3.md) и в бюджет функций.
7. Прогнать `make size`: прирост для приложений, которые **не используют** новый виджет, должен быть нулевым для TinyGo и не больше 1 КБ для gc (проверка DCE, [size.md](size.md), §3).
