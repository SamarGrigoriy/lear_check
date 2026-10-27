# Ядро: потоки, реестр, события, жизненный цикл

Документ описывает, **как** nui выполняет контракт из [api.md](api.md). Код здесь — наброски, показывающие логику. Имена в них обязательны, детали — нет.

## 1. UI-поток

GTK и Win32 работают с окнами только из главного потока процесса.

```go
// app.go
func init() { runtime.LockOSThread() } // init выполняется в главном потоке → главная горутина к нему прибита
```

- **gc.** После `LockOSThread` в `init` функция `main` и всё, что она вызывает синхронно, выполняется в главном потоке ОС. Обработчики вызываются из C в том же потоке (ведь `gtk_main` вызвали мы), поэтому ограничение действует и для них.
- **TinyGo.** Главная горутина выполняется в главном потоке; при `-scheduler=threads` это **проверяется в спайке S2**. `LockOSThread` в TinyGo может ничего не делать, и это не мешает.
- **Как проверить.** `drv.IsMainThread()` на Linux сравнивает `gettid()` и `getpid()` (glibc ≥ 2.30). `Run` проверяет это всегда, остальные методы — только при сборке с тегом `nuidebug`.

## 2. Run и Quit

```go
var state uint8 // stIdle → stSetup → stRunning → stDone
var quitAsked bool
var wins []*Window // живые окна (для автовыхода, ADR-15)

func Run(setup func()) error {
	if state != stIdle || !drv.IsMainThread() { panic("nui: Run вызывается один раз и из main") }
	if !drv.Init(appID) { state = stDone; return ErrNoDisplay } // gtk_init_check
	core.MarkUIThread(drv.IsMainThread)                        // для проверок nuidebug: ядро без cgo
	state = stSetup
	setup()
	if !quitAsked {
		state = stRunning
		drv.Main() // gtk_main; возвращается после drv.Quit()
	}
	state = stDone
	core.Close() // дальнейшие Post игнорируются
	return nil
}

func Quit() {
	switch state {
	case stSetup:   quitAsked = true      // gtk_main ещё не запущен — вызывать gtk_main_quit нельзя
	case stRunning: quitAsked = true; drv.Quit()
	}
}
```

Автоматический выход (ADR-15): `NewWindow` добавляет окно в `wins`, а обработка `EvDestroy` окна удаляет его оттуда. Затем nui проверяет оставшиеся окна: если ни одно не видимо (`drv.Visible`), вызывается `Quit()`. Скрытые окна — ещё не показанные или после `Hide()` — программу не держат. `Hide()` сам выход не вызывает: проверка делается только при уничтожении окна. Если `setup` не создал ни одного окна, цикл работает до явного `Quit`: в такой программе есть только таймеры или `Watch`.

## 3. Реестр (`internal/core/registry.go`)

```go
type Receiver interface{ Event(k Kind, arg uintptr) uintptr }

var (
	objs []Receiver // objs[0] == nil навсегда: id 0 означает «нет объекта»
	free []uint32   // стек освободившихся id
)

func Register(r Receiver) uint32 // взять id из free или дописать в конец objs
func Release(id uint32)          // objs[id] = nil; положить id в free
func Lookup(id uint32) Receiver  // nil, если id вне диапазона или освобождён
```

Инварианты:

- Реестр трогают только из UI-потока, поэтому мьютекс не нужен. При `nuidebug` каждая функция проверяет поток.
- Пока объект в реестре, Go-объект достижим и сборщик его не тронет. Так соблюдается ADR-6.
- `Release` вызывается ровно один раз:
  - у виджетов — по `EvDestroy`;
  - у таймеров — по `Stop` или после срабатывания `After`;
  - у `Watcher` — по `Stop`.
- После `Release` ОС больше не вызывает колбэков с этим id:
  - виджет уничтожен, и GTK не шлёт сигналы мёртвому объекту;
  - источник таймера или fd удалён через `g_source_remove`.

  Поэтому повторно выданный id не получит чужих событий, и счётчик поколений не нужен.

## 4. Тип события и `user_data` (`internal/core/kind.go`)

```go
type Kind uint8
const (
	EvClick   Kind = 1  // clicked
	EvChange  Kind = 2  // changed, toggled, value-changed
	EvEnter   Kind = 3  // activate (Enter в Entry)
	EvClose   Kind = 4  // delete-event — С ОТВЕТОМ: 1 = не закрывать
	EvDestroy Kind = 5  // destroy
	// 6–15 — для фазы 6: EvDraw, EvKey, EvMouse …
	KTimer    Kind = 16 // источник GLib: таймер
	KWatch    Kind = 17 // источник GLib: fd
	KDo       Kind = 18 // источник GLib: idle для очереди Do (id всегда 0)
)

func UD(id uint32, k Kind) uintptr   { return uintptr(id)<<5 | uintptr(k) }
func Decode(ud uintptr) (uint32, Kind) { return uint32(ud >> 5), Kind(ud & 31) }
```

В GTK значение `ud` передаётся как `user_data` (`gpointer`). Это просто число, а не указатель на память Go.

## 5. Трамплины (`internal/drv/export_linux.go`)

На любой колбэк C → Go есть ровно четыре экспортированные функции:

| Трамплин (C-сигнатура) | Кто вызывает | Ответ |
|---|---|---|
| `void nuiSignal(void *obj, uintptr_t ud)` | сигналы без результата: `clicked`, `changed`, `toggled`, `activate`, `value-changed`, `destroy` | — |
| `int nuiSignalRet(void *obj, void *arg, uintptr_t ud)` | сигналы с результатом: `delete-event` (позже `draw`, `key-press-event`) | `1` — событие обработано / не закрывать |
| `int nuiSource(uintptr_t ud)` | `g_timeout_add`, `g_idle_add` | `1` — повторять, `0` — удалить источник |
| `int nuiWatch(int fd, unsigned cond, uintptr_t ud)` | `g_unix_fd_add` | `1` — продолжать, `0` — удалить |

`ud` объявлен как `uintptr_t`, а не `gpointer`. В C это одно и то же по ABI, а в Go так не появляются «поддельные» указатели: числа, которые сборщик мусора мог бы принять за адрес.

```go
//export nuiSignal
func nuiSignal(obj unsafe.Pointer, ud C.uintptr_t) { core.Dispatch(uintptr(ud), 0) }

//export nuiSignalRet
func nuiSignalRet(obj, arg unsafe.Pointer, ud C.uintptr_t) C.int {
	return C.int(core.Dispatch(uintptr(ud), uintptr(arg))) // arg — указатель C (GdkEvent* и т. п.)
}

//export nuiSource
func nuiSource(ud C.uintptr_t) C.int { return C.int(core.Dispatch(uintptr(ud), 0)) }

//export nuiWatch
func nuiWatch(fd C.int, cond C.uint, ud C.uintptr_t) C.int {
	return C.int(core.Dispatch(uintptr(ud), uintptr(cond)))
}
```

Ещё два экспорта — не события, а вопросы модели `List` к данным (driver-gtk3.md, §7): `gchar *nuiListCell(uintptr_t key, int row, int col)` — текст ячейки (копия в `g_malloc`, её забирает `GValue`) и `void nuiListDrop(uintptr_t key)` — модель уничтожена. `key` — номер модели; функции ячеек лежат в Go в `map[uintptr]func(row, col int) string`, в C — только число, как требует ADR-6.

Go-код указатели на трамплины не получает. Их подставляют C-обёртки `nui_connect`, `nui_connect_ret`, `nui_timeout`, `nui_idle` и `nui_watch` из `wrap_linux.c` ([driver-gtk3.md](driver-gtk3.md)). Это работает одинаково в gc и TinyGo.

> Правило cgo для gc: если в файле есть `//export`, в его преамбуле могут быть только **объявления**. Поэтому определения обёрток лежат в отдельном `wrap_linux.c`, а `gtk3_linux.h` содержит только `typedef` и прототипы.

## 6. Диспетчер (`internal/core/dispatch.go`)

```go
func Dispatch(ud, arg uintptr) uintptr {
	id, k := Decode(ud)
	if k == KDo { Drain(); return 0 } // idle-источник одноразовый
	r := Lookup(id)
	if r == nil { return 0 }          // гонка с уничтожением — тихо игнорируем
	return r.Event(k, arg)
}
```

Пример `Event` у виджета в `nui`:

```go
func (b *Button) Event(k core.Kind, _ uintptr) uintptr {
	switch k {
	case core.EvClick:
		if b.quiet == 0 && b.onClick != nil { b.onClick() }
	case core.EvDestroy:
		b.destroyed() // core.Release(b.id); b.id = 0; b.h = drv.Nil
	}
	return 0
}
```

## 7. Тихий режим: Set* не вызывает On* (ADR-8)

GTK шлёт сигналы и на программные изменения: `gtk_entry_set_text` вызывает `changed`, `gtk_combo_box_set_active` — тоже `changed`. Сигнал приходит **синхронно, внутри** вызова, поэтому достаточно счётчика:

```go
func (e *Entry) SetText(s string) {
	if e.id == 0 { return }
	e.quiet++
	drv.EntrySetText(e.h, s)
	e.quiet--
}
```

Счётчик (`uint8`), а не `bool`, нужен на случай вложенности: один `Set*` внутри другого.

**`quiet` оборачивает каждый метод, который меняет виджет из кода**, а не только `Set*`. GTK присылает сигналы и на `Combo.Add`/`Clear` (`remove_all` → `changed`), и на `TextView.Append`/`Clear` (вставка и удаление в буфере → `changed`). Правило для реализации: если метод вызывает функцию `drv`, которая меняет содержимое или состояние виджета, вызов обёрнут в `quiet++`/`quiet--`. Интеграционные тесты проверяют это для каждого такого метода ([testing.md](testing.md), §3).

**Сколько раз вызывается `On*`.** Одно действие пользователя может прислать два сигнала: вставка поверх выделения в TextView — это удаление и вставка, и `OnChange` сработает дважды. Поэтому контракт обещает «обработчик вызван», а не «вызван ровно один раз»: обработчик читает текущее значение геттером и не считает вызовы.

На Win32 работает так же: `WM_COMMAND/EN_CHANGE` приходит синхронно внутри `SetWindowTextW`.

## 8. Очередь `Do` (`internal/core/queue.go`)

```go
var (
	mu      sync.Mutex
	queue   []func()
	pending bool // idle-источник уже запланирован
	closed  bool // Run завершился
)

// Post можно вызывать из любого потока. true — нужно «разбудить» цикл (drv.Wake).
func Post(fn func()) bool {
	mu.Lock(); defer mu.Unlock()
	if closed { return false }
	queue = append(queue, fn)
	if pending { return false }
	pending = true
	return true
}

// Drain — только в UI-потоке, из Dispatch(KDo). Функции — по одной (см. ниже).
func Drain() {
	mu.Lock(); n := len(queue); pending = false; mu.Unlock()
	for ; n > 0; n-- {
		mu.Lock()
		if len(queue) == 0 { mu.Unlock(); return } // остальное выполнил вложенный Drain
		fn := queue[0]; queue = queue[1:]
		wake := len(queue) > 0 && !pending        // в очереди есть ещё — разбудить заранее
		if wake { pending = true }
		mu.Unlock()
		if wake { waker() }                        // waker = drv.Wake, задаёт Run
		fn()
	}
}

// nui:
func Do(fn func()) { if core.Post(fn) { drv.Wake() } }
```

- `drv.Wake()` вызывает `g_idle_add_full(G_PRIORITY_DEFAULT, nuiSource, UD(0, KDo), NULL)`. Функцию `g_idle_add_full` GLib разрешает вызывать из любого потока.
- Приоритет `G_PRIORITY_DEFAULT` (0), а не `DEFAULT_IDLE` (200). Это компромисс:
  - плюс: `Do` не «голодает», пока интерфейс постоянно перерисовывается. Зацикливания внутри одного `Drain` тоже нет: он обрабатывает только пачку, накопленную к моменту вызова;
  - минус: приоритет выше, чем у пересчёта размеров (110) и перерисовки (`GDK_PRIORITY_REDRAW`, 120). Если горутина вызывает `Do` непрерывно, быстрее, чем UI-поток выполняет пачки, idle-источник готов на каждой итерации цикла, и окно **перестаёт перерисовываться**;
  - поэтому в контракте ([api.md](api.md), §1) записано: копить данные и вызывать `Do` пачками, не чаще 20–50 раз в секунду. Нагрузочный чек-лист проверяет, что при таком темпе окно перерисовывается ([testing.md](testing.md), §7).
- `Drain` выполняет функции, накопленные к моменту вызова, **по одной**. Функции, добавленные во время `Drain`, уйдут в следующий idle — бесконечного цикла нет.
- **Почему по одной и почему «будить заранее».** Функция может открыть модальный диалог: тогда она не вернётся, пока диалог открыт, а внутри крутится вложенный цикл. Если бы `Drain` забирал пачку целиком, остальные функции той же пачки ждали бы закрытия диалога — а контракт обещает, что `Do` работает, пока диалог открыт (найдено тестом фазы 3, [reports/phase-3.md](reports/phase-3.md)). Поэтому перед вызовом каждой функции, если в очереди есть ещё, `Drain` будит цикл — и остаток выполнит вложенный `Drain`. Каждую функцию забирает ровно один `Drain`, поэтому порядок сохраняется.

## 9. Таймеры

```go
type Timer struct {
	id, src uint32
	fn      func()
	once    bool
}

func every(ms int, fn func(), once bool) *Timer {
	t := &Timer{fn: fn, once: once}
	t.id = core.Register(t)
	t.src = drv.TimerAdd(ms, core.UD(t.id, core.KTimer)) // g_timeout_add(ms ≤ 0 → 0)
	return t
}

func (t *Timer) Event(core.Kind, uintptr) uintptr {
	t.fn()           // внутри fn может быть вызван t.Stop()
	if t.id == 0 { return 0 } // остановлен изнутри
	if t.once { core.Release(t.id); t.id = 0; return 0 } // источник удалит сам GLib
	return 1
}

func (t *Timer) Stop() {
	if t.id == 0 { return }
	drv.SourceRemove(t.src) // GLib разрешает удалять источник и изнутри его колбэка
	core.Release(t.id); t.id = 0
}
```

## 10. `Watch` (unix)

Тот же шаблон, что у `Timer`, с двумя отличиями:

- `drv.WatchAdd(fd, cond, ud)` вызывает `g_unix_fd_add_full(G_PRIORITY_DEFAULT, fd, cond, nuiWatch, ud, NULL)`.
- `arg` в `Event` — это набор `GIOCondition`, и драйвер переводит его в `IOCond`.

| `IOCond` | `GIOCondition` |
|---|---|
| `IORead` | `G_IO_IN` (1) \| `G_IO_PRI` (2) |
| `IOWrite` | `G_IO_OUT` (4) |
| `IOErr` | `G_IO_ERR` (8) \| `G_IO_NVAL` (32) |
| `IOHup` | `G_IO_HUP` (16) |

`G_IO_HUP`, `G_IO_ERR` и `G_IO_NVAL` подписываются всегда.

## 11. Жизненный цикл виджета

**Создание** (на примере `NewButton`):

```go
b := &Button{onClick: fn}
b.id = core.Register(b)
b.h  = drv.ButtonNew(b.id, text)
// drv внутри:
//   gtk_button_new_with_label
//   nui_connect(h, "clicked", UD(id, EvClick))
//   nui_connect(h, "destroy", UD(id, EvDestroy))
//   gtk_widget_show(h)   — виджет видим с рождения (ADR-14), окно — нет
```

**Вставка в контейнер:**
- если `placed == true` — паника;
- `placed = true`;
- `drv.BoxPack` / `GridAttach` / `WindowSetChild`. Контейнер GTK «поглощает» плавающую ссылку (floating reference) и становится владельцем.

**Уничтожение.** Своих ссылок (`g_object_ref`) nui не держит, владеет GTK. Виджет уничтожается в трёх случаях:
1. закрыто или уничтожено окно — GTK уничтожает всех потомков;
2. вызван `SetChild`: старый ребёнок удаляется из контейнера, ссылок на него не остаётся, и он уничтожается;
3. программа завершается.

На каждое уничтожение приходит `destroy` → `EvDestroy` → `destroyed()`:

```go
func (w *widget) destroyed() { core.Release(w.id); w.id = 0; w.h = drv.Nil } // не nil: на Windows H — uintptr
```

С этого момента все методы проверяют `id == 0` и ничего не делают, а `Alive()` возвращает `false`.

**Закрытие окна:**

| Кто закрывает | Последовательность |
|---|---|
| Пользователь (крестик) | `delete-event` → `EvClose` → `OnClose()`; `false` → вернуть `1`, конец. `true` или нет обработчика → вернуть `0` → GTK уничтожает окно → `destroy` у окна и всех детей → окно удаляется из `wins` → если видимых окон не осталось, вызывается `Quit()` (ADR-15) |
| Код: `w.Close()` | `gtk_widget_destroy` → `destroy` … (`OnClose` не вызывается, ADR-14) |

## 12. Строки (ADR-13)

```go
// Go → C: память Go, действует только на время вызова; GTK копирует всё, что хранит.
// Невалидный UTF-8 заменяется на U+FFFD: GTK проверяет UTF-8 и на битой строке
// пишет Gtk-CRITICAL и ничего не делает (driver-gtk3.md, §11, п. 14).
func cstr(s string) *C.char {
	if !utf8.ValidString(s) {
		s = core.FixUTF8(s) // core/text.go: каждая битая последовательность → "�"
	}
	return cpath(s)
}

// Без проверки UTF-8 — для путей файловой системы: они не обязаны быть UTF-8.
func cpath(s string) *C.char {
	b := make([]byte, len(s)+1)
	copy(b, s)                // нулевой байт внутри s обрежет строку в C — это в контракте
	return (*C.char)(unsafe.Pointer(&b[0]))
}

// C → Go: копируем сразу.
func gostr(p *C.char) string // длина по strlen; nil → ""
func gostrFree(p *C.char) string { s := gostr(p); C.g_free(unsafe.Pointer(p)); return s }
```

Кто владеет возвращённой строкой:

| Функция | Владелец | Что делаем |
|---|---|---|
| `gtk_label_get_text`, `gtk_entry_get_text`, `gtk_button_get_label` | виджет | `gostr` |
| `gtk_combo_box_text_get_active_text` | вызывающий | `gostrFree` |
| `gtk_file_chooser_get_filename` | вызывающий | `gostrFree` |
| `nui_text_get` (обёртка над `gtk_text_buffer_get_text`) | вызывающий | `gostrFree` |
| `g_filename_to_uri` | вызывающий | `gostrFree` |

`gostr` реализуется без `C.GoString`: длина считается через `strlen` (прототип в нашем `.h`), затем `unsafe.Slice` и `string(...)`. Так код одинаково работает в gc и TinyGo.

`cpath` используется только для путей, которые уходят в GLib как имена файлов (`ShowInFolder` → `g_filename_to_uri`). Всё остальное идёт через `cstr`. Пути, которые возвращают диалоги, приходят в кодировке файловой системы и копируются как есть.

## 13. Паники (ADR-9)

`recover` нигде нет. Что происходит при панике:

| Сборка | Результат |
|---|---|
| gc | паника внутри колбэка из C завершает процесс и печатает стек |
| TinyGo `-panic=print` | печатает сообщение и завершает процесс |
| TinyGo `-panic=trap` (релиз) | SIGILL без сообщения |

Сообщения собственных паник nui начинаются с `nui: `. Это строковые константы без форматирования.

## 14. Отладочная сборка (`-tags nuidebug`)

| Проверка | Где |
|---|---|
| Вызов не из UI-потока → паника | `core.CheckThread()` в начале каждого публичного метода (в обычной сборке это пустая функция, которую выбрасывает компилятор) |
| Виджет создан до `Run` → паника | конструкторы |
| Повторная вставка → паника (есть всегда, не только при `nuidebug`) | `Box.Add`, `Grid.Add`, `SetChild` |
| Пересечение ячеек Grid → паника | `Grid.Add` |
| Вызов у уничтоженного объекта → предупреждение в stderr: `nui: Label.SetText у уничтоженного виджета` | базовая проверка `id == 0` |

## 15. Режимы планировщика TinyGo (ADR-7)

| Сборка | Горутины в приложении | `Do` из горутины | Таймеры и `Watch` | Статус |
|---|---|---|---|---|
| gc | да | да | да | поддерживается |
| TinyGo `-scheduler=none` | **нет** (ошибка компиляции) | не нужен | да | поддерживается; самый маленький вариант |
| TinyGo `-scheduler=threads` | да | да | да | поддерживается: `Do` из горутины проверен в S3. В TinyGo 0.42 это планировщик по умолчанию на Linux |
| TinyGo `-scheduler=tasks` | **не работают**: пока поток в `gtk_main`, горутины не получают управления | — | да | **не поддерживается** |

Отсюда рекомендация для приложений: ввод-вывод — через `Watch`, периодика — через `Every`. Тогда приложение собирается в любом режиме. Приёмочный логгер ([acceptance.md](acceptance.md)) устроен именно так.

## 16. Производительность — чего ожидать

- Событие: трамплин, `Decode`, поиск в срезе и вызов через интерфейс — это сотни наносекунд (в gc больше всего стоит переход C → Go).
- `Set*`: одна аллокация `cstr` и один вызов C.
- Тяжёлое место — `TextView.Append` большими объёмами. `nui_text_append` вставляет текст одним вызовом и подрезает буфер по `SetMaxLines`. Приложению стоит копить данные и дописывать их не чаще раза в 50–100 мс, а не на каждый байт.
