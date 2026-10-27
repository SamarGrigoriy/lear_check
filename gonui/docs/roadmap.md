# План работ: спайки и фазы

Сроки — ориентир для одного разработчика, который знает Go и немного C. Каждая фаза заканчивается **проверяемым** критерием готовности (DoD). Если по ходу работы реальность разошлась с документацией, сначала исправляется документ (для решений — новый ADR), потом код.

```
S1 ─► S2 ─┬─(успех)───────────────► Ф1 ─► Ф2 ─► Ф3 ─► Ф4 ─► Ф5 ─► Ф6
          └─(неудача)─► S3 ─(итог)─┘
```

---

## Спайки: сначала снимаем риски

Спайки — это **одноразовый** код в каталоге `spike/`. В основное дерево он не переносится, переносятся только выводы. Итог каждого спайка — короткая заметка `spike/<имя>/RESULT.md`: что проверено, цифры, решение.

### S1. gc + cgo без системных заголовков + GTK3 (1–2 дня)

**Итог: пройден 08.10.2026** — [spike/s1/RESULT.md](../spike/s1/RESULT.md).

**Что строим.** Один пакет `spike/s1` из четырёх файлов:
- `s1.h` — около 25 прототипов, макросы `NUI_FN` и `NUI_CONST`;
- `wrap.c` — `nui_connect`, `nui_connect_ret`, `nui_timeout`, `nui_idle`, `nui_watch`;
- `export.go` — 4 трамплина;
- `main.go` — окно, метка, кнопка, крошечный реестр и Dispatch.

**Сценарий:**
1. клик по кнопке меняет текст метки;
2. горутина через 1 с вызывает `Do` → текст метки меняется;
3. `After(2000)` меняет заголовок окна;
4. `Watch` на `os.Pipe()`: горутина пишет в канал, обработчик читает и показывает;
5. закрытие окна крестиком с подтверждением через `delete-event` (ответ 1 / 0).

**Критерии прохождения:**

| # | Проверка | Как |
|---|---|---|
| 1 | Сборка без `libgtk-3-dev` | `dpkg -s libgtk-3-dev` → не установлен; `go build` проходит |
| 2 | Линковка по soname | `readelf -d s1 \| grep NEEDED` содержит `libgtk-3.so.0` |
| 3 | Все 5 пунктов сценария работают | вручную, в сессиях Wayland и X11 (`GDK_BACKEND=x11`) |
| 4 | Нет предупреждений GTK | запуск с `G_DEBUG=fatal-criticals` не падает |
| 5 | Прототипы верны | `check-protos` на машине с `libgtk-3-dev` проходит |
| 6 | Размер | `-s -w`: записать число → основа бюджета gc |
| 7 | Флаги C-части | флаги размера передаются через `CGO_CFLAGS`/`CGO_LDFLAGS`: в `#cgo` gc их не пропускает ([driver-gtk3.md](driver-gtk3.md), §9). Записать размер с ними и без них |
| 8 | Версия glibc | сборка в `debian:bookworm`; `objdump -T` показывает максимальную версию `GLIBC_` не выше 2.36 |

**Если не прошёл:**
- `-l:` не находит библиотеки → указать полный путь к `.so.0` в LDFLAGS;
- cgo не принимает собственные прототипы → поставить `libgtk-3-dev` и подключать настоящие заголовки, а ADR-3 переписать.

### S2. Тот же код в TinyGo с `-libc=glibc` (1–3 дня)

**Итог: не понадобился.** PR #5838 не нужен: его критерии проверены в S3 на выпущенном TinyGo (ADR-17).

**Подготовка.** TinyGo с поддержкой `-libc=glibc` ([build.md](build.md), §3, варианты А/Б).

**Две сборки S1:**
- **S2-none**: `-scheduler=none`. Из сценария убраны горутины: `Do` вызывается из таймера, в pipe пишет тоже таймер;
- **S2-threads**: `-scheduler=threads`, полный сценарий.

**Критерии прохождения:**

| # | Проверка | Как |
|---|---|---|
| 1 | Сборка и линковка | `tinygo build -libc=glibc …` без ошибок; `readelf` показывает `libgtk-3.so.0` и `libc.so.6` |
| 2 | `.c`-файл пакета компилируется и линкуется | `nui_connect` вызывается (сценарий работает) |
| 3 | `//export`-трамплины вызываются из C | пункты сценария 1, 3, 4, 5 |
| 4 | Строки через память Go (ADR-13) | кириллица и эмодзи в метке отображаются верно |
| 5 | Главный поток | в обработчике `gettid() == getpid()` |
| 6 | Сборщик мусора не ломает реестр | Таймер каждые 10 мс 10 минут подряд создаёт мусор (срезы по 64 КБ) и меняет текст; затем клик по кнопке работает |
| 7 | S2-threads: `Do` из горутины | пункт сценария 2 |
| 8 | Размер S2-none | ≤ 64 КБ после `-opt=z -no-debug -panic=trap` |
| 9 | Флаги C-части | учитывает ли TinyGo `CGO_CFLAGS`/`CGO_LDFLAGS`; если нет — как передать `-ffunction-sections` и `--gc-sections`, или они не нужны, потому что C-код проходит через LLVM |

**Итог:**
- (а) Прошли все пункты → **TinyGo — релизный тулчейн** (ADR-4 получает статус «подтверждено»).
- (б) S2-none прошёл, а S2-threads нет → TinyGo используется только для приложений без горутин (`Watch`/`Every`). Это отражается в [runtime.md](runtime.md), §15.
- (в) S2-none не прошёл → переходим к S3.

### S3. Свой target-файл TinyGo без патча (не больше 1 дня)

**Итог: пройден 08.10.2026** — выпущенный TinyGo 0.42.0 + свой target-файл; все критерии S2, `-scheduler=none` — 35 952 байта ([spike/s3/RESULT.md](../spike/s3/RESULT.md), ADR-17).

**Гипотеза.** Целевую платформу TinyGo можно описать JSON-файлом (`-target=путь.json`), взяв за основу встроенную цель linux/amd64, но:
- без сборки своей libc;
- с системным `cc` в роли линковщика (тогда он подставит glibc, `crt1.o` и динамический загрузчик).

Какие поля нужны, смотрите в документации TinyGo по target-файлам и в изменениях PR #5838.

**Критерии прохождения** — те же, что в S2.

**Если не прошёл:**
- Релизный тулчейн — **gc**; бюджеты — по S1.
- TinyGo-задачи в CI остаются с пометкой «можно падать». Правила совместимости продолжают соблюдаться.
- S2 повторяется, когда PR #5838 вольют.

---

## Фаза 1. Скелет (3–5 дней)

**Состав:**
- `internal/core`: `kind.go`, `registry.go`, `dispatch.go`, `queue.go`, `debug_on/off.go` + юнит-тесты;
- `internal/drv` (linux):
  - `gtk3_linux.h` и `wrap_linux.c` (только нужное для фазы), `link_linux.go`, `export_linux.go`;
  - функции `Init`, `Main`, `Quit`, `Wake`, `IsMainThread`, `TimerAdd`, `SourceRemove`, `WatchAdd`, `IOCond`;
  - общие функции виджетов, `Window*`, `Label*`, `Button*`, `Box*`, `SpacerNew`;
  - нормализация (`adopt`);
- `nui`: `Run`, `Quit`, `SetAppID`, `ErrNoDisplay`, `Do`, `After`, `Every`, `Watch`, `Window` (`New`, `SetTitle`, `SetChild`, `Show`, `Close`, автовыход), `Label` (`SetText`/`Text`), `Button`, `VBox`/`HBox`/`Expand`/`Stretch`, общие методы;
- `examples/hello` с флагом `-selftest`;
- `Makefile` (флаги C-части через `CGO_CFLAGS`/`CGO_LDFLAGS`, цель `abi`), `tools/depcheck.sh`, `tools/sizecheck.sh`, `tools/check-protos.sh`, `size-budget.txt` (цифры из S1/S2).

**DoD:**
**Итог: готово 08.10.2026** — [docs/reports/phase-1.md](reports/phase-1.md): `hello` — 23 680 байт (TinyGo) и 1 301 352 (gc), покрытие ядра 100 %, 17 интеграционных тестов.

- [x] `make ci` зелёный, включая TinyGo-задачи (S3 пройден).
- [x] `examples/hello` ≤ 15 строк, работает под gc и TinyGo.
- [x] Покрытие `internal/core` юнит-тестами ≥ 90 %.
- [x] Интеграционные тесты (через `ui`, [testing.md](testing.md), §3): Set не вызывает On, Sim вызывает; методы уничтоженного виджета ничего не делают; уничтожение последнего видимого окна завершает `Run`, даже если есть скрытое окно; `Do` из 100 горутин сохраняет порядок внутри каждой горутины; `Every` + `Stop` изнутри обработчика.
- [x] Размеры в бюджете.

## Фаза 2. Виджеты MVP (1–2 недели)

**Состав:**
- Entry, Check, Combo/ComboEntry, Slider, Progress, TextView, Separator, Frame, Scroll, Grid;
- Label: многоточие, перенос, выделение, выравнивание текста, жирный шрифт, цвет;
- Window: `SetSize`, `SetResizable`, `SetPadding`, `SetDefault`, `SetIconName`, `OnClose`, `Hide`;
- проверки `nuidebug`;
- `Sim*` для всех виджетов;
- `examples/form` — все виджеты на одной форме.

**Итог: готово 08.10.2026** — [docs/reports/phase-2.md](reports/phase-2.md): `form` со всеми виджетами — 41 808 байт (TinyGo), 31 интеграционный тест.

**DoD:**
- [x] Каждая строка [api.md](api.md), §4–§15, реализована и покрыта интеграционным тестом: Set/Get с кириллицей, правило Set ≠ On, уничтожение.
- [x] `check-protos` проходит.
- [~] Ручной чек-лист раскладки из [testing.md](testing.md) пройден на `examples/form` — автоматизирована часть: снимки формы, тёмной темы, масштаба 2×, многоточия, Grid. Ручными остаются KDE/XFCE, дробный масштаб, IME, Orca.
- [x] Проверка DCE: `examples/hello` не вырос (TinyGo — ни на байт, gc — не больше 1 КБ).

## Фаза 3. Диалоги и оболочка (3–5 дней)

**Состав:** `Message`, `Alert`, `Ask`, `OpenFile`, `SaveFile`, `ChooseFolder`, `OpenURI`, `ShowInFolder`.

**Итог: готово 08.10.2026** — [docs/reports/phase-3.md](reports/phase-3.md). Найдена и исправлена ошибка ядра: функции `Do` из одной пачки ждали закрытия модального диалога.

**DoD:**
- [~] Ручной чек-лист диалогов ([testing.md](testing.md)) пройден в GNOME (Wayland) и KDE — автоматизирована часть: снимки диалогов, ответы «Да», «Нет», Esc и OK в тестах. GNOME/KDE, портал и клавиатура — вручную.
- [~] `ShowInFolder` выделяет файл в Nautilus и Dolphin, а без D-Bus `FileManager1` открывает папку — второе ✓ автоматически (без сессии D-Bus и с сессией без службы); выделение в Nautilus/Dolphin — вручную.
- [x] Таймеры и `Do` работают, пока открыт модальный диалог (тест: `Every` считает срабатывания, пока открыт `Message`).

## Фаза 4. Приёмка: логгер (около 1 недели)

**Состав:**
- `examples/logger` по [acceptance.md](acceptance.md);
- `.deb` для него ([build.md](build.md), §6).

**DoD:**
- [ ] Все пункты чек-листа приёмки ([acceptance.md](acceptance.md), §7) выполнены на Debian 12 и 13.
- [ ] Размер TinyGo ≤ 250 КБ. Если релиз остался на gc, размер в бюджете gc.
- [ ] Отчёт `docs/reports/size-logger.md`: `tinygo build -size full` по пакетам и сравнение с esp32_logger.

## Фаза 5. Windows (2–3 недели)

**Статус (09.10.2026):** всё, что проверяется без Windows, готово: движок раскладки, спайк S4, драйвер Win32, примеры, интеграционные тесты под Wine. Остаётся ручная проверка на Windows 10/11 (первый пункт DoD) и открытый вопрос с `TaskDialog` под Wine ([отчёт](reports/phase-5.md)).

**Состав:**
- спайк S4 ([driver-win32.md](driver-win32.md), §11) — **пройден** под Wine ([spike/s4/RESULT.md](../spike/s4/RESULT.md));
- `internal/layout` + golden-тесты ([layout.md](layout.md), §8) — **готово**: эталон снимает сам драйвер nui на GTK, `libgtk-3-dev` не нужен (ADR-18);
- драйвер `*_windows.*` — первая версия: hello и form под Wine обоими тулчейнами ([отчёт](reports/phase-5.md));
- `examples/*` на Windows — hello и form собираются gc и TinyGo, selftest и интеграционные тесты nui проходят под Wine (`make win-selftest win-test`).

**DoD:**
- [x] Golden-тесты проходят на всех случаях: 11 случаев, движок совпадает с GTK до пикселя ([отчёт](reports/phase-5.md)).
- [ ] `examples/hello`, `form` и `logger` (Windows-вариант с горутиной + `Do`) работают на Windows 10 и 11 при масштабах 100 %, 150 % и 200 %, а также при переносе окна между мониторами с разным DPI.
- [x] `check-protos` для `win32_windows.h` проходит против заголовков mingw.

## Фаза 6. Расширения (по мере надобности)

**Готово вне очереди:** `Chat` — готовый чат, например с ИИ ([api.md](api.md), §18; `examples/chat`). Собран из родных виджетов, работает на GTK и Win32, программы без чата не растут.

**Готово:**
- буфер обмена: `SetClipboard`, `Clipboard`, `Chat.LastMessage` ([api.md](api.md), §17);
- `Image` — картинка ([api.md](api.md), §19; ADR-20) и `OpenFiles` — выбор нескольких файлов (§16), [отчёт](reports/image.md);
- `List` — список и таблица со строками по требованию ([api.md](api.md), §20), [отчёт](reports/list.md);
- `Spin` — числовое поле: `NewSpin`, `SetValue`, `Value`, `OnChange` ([api.md](api.md), §11), [отчёт](reports/spin.md);
- `Radio` — переключатель: `NewRadio`, `SetSelected`, `SetHorizontal`, `OnChange` ([api.md](api.md), §9), [отчёт](reports/radio.md);
- `Tabs` — вкладки: `NewTabs`, `Add`, `Remove`, `SetCurrent`, `OnChange` ([api.md](api.md), §15), [отчёт](reports/tabs.md);
- контекстное меню: `SetContextMenu` у виджетов, окна и чата ([api.md](api.md), §4, §5), [отчёт](reports/context.md);
- `Split` — две панели с границей, которую двигают мышью: `HSplit`, `VSplit` ([api.md](api.md), §15), [отчёт](reports/split.md);
- `Canvas` — рисование: линии, прямоугольники, эллипсы, текст, мышь ([api.md](api.md), §21), [отчёт](reports/canvas.md);
- меню окна: `Window.SetMenu`, `NewMenu`, `Item`, `CheckItem`, `SubMenu`, `Sep`, сочетания клавиш ([api.md](api.md), §5), [отчёт](reports/menu.md);
- история чата в JSON: `Chat.LoadJSON`, `Chat.JSON`, `Chat.Messages`; формат API моделей тоже; `Chat.OnReply` — ответ закончен ([api.md](api.md), §18);
- `Size()` у виджетов и окна; описано, как прокручивать растущий ряд (`Scroll`, [api.md](api.md), §4, §5, §14);
- `Box.Remove`, `Box.Clear`, `Grid.Remove`, `Grid.Clear` — удаление виджетов из контейнеров ([api.md](api.md), §15);
- `Window.OnDropFiles` — перетаскивание файлов на окно, `nui.Args` — аргументы командной строки без потерь на Windows, пример `examples/viewer` ([api.md](api.md), §1, §5), [отчёт](reports/drop.md);
- `Window.OnKey` — клавиши окна, `TextView.OnEnter` и `SetPlaceholder`, многострочное поле ввода в `Chat` ([api.md](api.md), §5, §13, §18), [отчёт](reports/keys.md).

В порядке пользы:
1. ~~Canvas~~ — готово;
2. ~~меню~~ — готово;
3. ~~буфер обмена~~ — готово;
4. ~~`OnKey`~~ — готово;
5. ~~List~~ — готово;
6. ~~Image~~ — готово.

Отдельные проекты, если понадобятся:
- драйвер GTK4 — новый набор `*_linux.*` под тегом `gtk4`;
- macOS (AppKit через `objc_msgSend` с той же схемой прототипов).

Каждое расширение проходит чек-лист «Как добавить новый виджет» ([architecture.md](architecture.md)) и не увеличивает размер приложений, которые его не используют.
