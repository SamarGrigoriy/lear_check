# Драйвер Linux: GTK3

Пакет `internal/drv`, файлы `*_linux.*`. Требуется GTK ≥ 3.22.

## 1. Принципы

1. **Собственные прототипы вместо `<gtk/gtk.h>`** (ADR-3). В `.h` объявлены только нужные функции и непрозрачные типы.
2. **Go вызывает C только с простыми аргументами**: указатели, целые, `double`, строки (`cstr`). Всё остальное делают C-обёртки `nui_*` в `wrap_linux.c`: макросы, varargs, структуры на стеке, `GVariant`, указатели на колбэки.
3. **Подписки на сигналы делаются внутри `XxxNew`** с id объекта. Вне драйвера о сигналах не знают.
4. Каждый созданный виджет сразу проходит одинаковую «нормализацию» (§6).

## 2. Файлы

| Файл | Содержимое |
|---|---|
| `gtk3_linux.h` | типы, константы, прототипы GTK/GLib/Pango/GIO, прототипы обёрток `nui_*` и трамплинов. **Только объявления** |
| `wrap_linux.c` | определения `nui_*` |
| `link_linux.go` | `#cgo LDFLAGS` (только библиотеки, §9), `type H`, `Nil`, `cstr`, `cpath`, `gostr`, `gostrFree`, `b2i` |
| `export_linux.go` | 4 трамплина `//export` и 2 экспорта модели List ([runtime.md](runtime.md), §5) |
| `app_linux.go`, `widget_linux.go`, `controls_linux.go`, `layout_linux.go`, `dialog_linux.go` | функции контракта |
| `sim_linux.go` (тег `gui`) | `Sim*` и `Allocation` для тестов ([testing.md](testing.md), §3) |

## 3. Контракт драйвера

Это **одинаковый набор функций** для всех ОС (ADR-5). Если в каком-то драйвере функции нет, сборка под эту ОС падает с ошибкой — так контракт и проверяется. Числовые параметры `align`, `ellipsis`, `kind`, `mode` используют значения констант nui, а переводит их в значения ОС драйвер.

```go
type H = unsafe.Pointer // Linux: GtkWidget*. На Windows: type H = uintptr (индекс узла, driver-win32.md §2)
var Nil H               // нулевой хендл: nil на Linux, 0 на Windows. nui обнуляет и сравнивает хендлы только через него

// Приложение
func Init(appID string) bool            // g_set_prgname (если задан), gtk_init_check(NULL, NULL)
func Main()                             // gtk_main
func Quit()                             // gtk_main_quit
func Wake()                             // nui_idle(UD(0, KDo)); потокобезопасна
func IsMainThread() bool                // gettid() == getpid()
func TimerAdd(ms int, ud uintptr) uint32
func SourceRemove(src uint32)
func WatchAdd(fd int, cond uint8, ud uintptr) uint32 // только unix
func IOCond(gcond uintptr) uint8         // GIOCondition → биты nui.IOCond

// Общее для виджетов
func Show(h H); Hide(h H); Destroy(h H)
func SetEnabled(h H, on bool); Enabled(h H) bool
func SetVisible(h H, on bool); Visible(h H) bool
func SetTooltip(h H, s string)
func SetAlign(h H, hz, v uint8)
func SetMinSize(h H, w, ht int)
func SetExpand(h H, hz, v bool)         // для Grid.ExpandCol/Row
func SetPadding(h H, px int)            // Box, Grid, Window: border-width
func Focus(h H)

// Window
func WindowNew(id uint32, title string) H
func WindowSetTitle(h H, s string)
func WindowSetKeys(h H, on bool)          // OnKey: на GTK окно подписано с создания — ничего не делает
func KeyDecode(arg uintptr) (key uint32, mods uint8) // arg события EvKey; key 0 — о клавише не сообщаем
func WindowSetDrop(h H, id uint32, on bool)            // окно — приёмник файлов (OnDropFiles)
func Parent(h H) H                                     // контейнер виджета (Box.Remove проверяет)
func ContextMenu(owner H) H; ContextMenuDone(owner, menu H); ContextMenuRemove(owner H)
func SplitNew(id uint32, vertical bool) H; SplitSet(h, first, second H)
func SplitSetPosition(h H, px int); SplitPosition(h H) int
func TabsNew(id uint32) H; TabsAdd(h, page H, title string)     // EvChange — выбор пользователя
func TabsSetCurrent(h H, i int); TabsCurrent(h H) int; TabsSetTitle(h, page H, title string)
func SpinNew(id uint32, min, max int) H; SpinSetValue(h H, v int); SpinValue(h H) int // EvChange — значение принято
func RadioNew(id uint32) H; RadioAdd(h H, id uint32, text string)      // EvChange — выбран другой вариант
func RadioSelected(h H) int; RadioSetSelected(h H, i int); RadioSetHorizontal(h H, on bool); RadioFocus(h H)
func Size(h H) (w, ht int)                             // allocation, логические пиксели; не mapped — 0, 0
func WindowSetPadding(h H, px int)                     // у окна с меню — ряда под меню
// Canvas: draw → EvDraw (arg — cairo_t*), мышь → EvMouse (arg — GdkEvent*)
func CanvasNew(id uint32) H; CanvasRedraw(h H)
func PaintBegin(ctx uintptr, h H); PaintThemeColor(ctx uintptr, h H); PaintColor(ctx uintptr, rgb uint32)
func PaintLineWidth(ctx uintptr, w float64); PaintLine(ctx uintptr, x1, y1, x2, y2 float64)
func PaintRect(ctx uintptr, x, y, w, h float64, fill bool); PaintEllipse(ctx uintptr, x, y, w, h float64, fill bool)
func PaintText(ctx uintptr, h H, x, y float64, s string); TextSize(ctx uintptr, h H, s string) (w, ht float64)
func MouseDecode(arg uintptr) (kind uint8, x, y float64, button int, mods uint8, wheelX, wheelY float64)
// Меню окна: строка, подменю, пункты; сочетания только подписываются (срабатывают в nui)
func MenuBar(win H) H; MenuBarRemove(win H); MenuDone(win H)
func MenuSub(parent H, id uint32, title string) (item, menu H)
func MenuItemNew(menu H, id uint32, check bool, text string, key uint32, mods uint8) H
func MenuSepNew(menu H)
func MenuItemSetEnabled(h H, on bool); MenuItemSetChecked(h H, on bool); MenuItemChecked(h H) bool
func MenuItemSetText(h H, text string, key uint32, mods uint8)
func WindowSize(h H) (w, ht int)                       // gtk_window_get_size
func ContainerClear(h H)                               // уничтожить всех детей Box или Grid
func DropPaths(arg uintptr) []string                   // пути из arg события EvDrop
func Args() []string                                   // аргументы командной строки
func WindowSetChild(h, child H)         // старого ребёнка — gtk_container_remove (он уничтожится); child == Nil — только удалить
func WindowSetSize(h H, w, ht int)      // см. §7 «Window»
func WindowSetResizable(h H, on bool)
func WindowSetDefault(h, btn H)
func WindowSetIconName(h H, name string)
func WindowShow(h H)                    // gtk_widget_show + gtk_window_present
func WindowSetMaximized(h H, on bool)   // gtk_window_maximize / unmaximize (можно до показа)
func WindowSetFullscreen(h H, on bool)  // gtk_window_fullscreen / unfullscreen

// Label
func LabelNew(id uint32, text string) H
func LabelSetText(h H, s string); LabelText(h H) string
func LabelSetEllipsis(h H, e uint8)
func LabelSetWrap(h H, on bool)
func LabelSetSelectable(h H, on bool)
func LabelSetTextAlign(h H, a uint8)
func LabelSetBold(h H, on bool)         // жирность и цвет хранятся в самой метке (g_object_set_data),
func LabelSetColor(h H, color uint32)   // поэтому каждый вызов меняет только своё; 1<<24 — без цвета
func LabelSetFontScale(h H, pct int)    // размер шрифта в процентах (Pango scale), ADR-19
func SetForeground(h H, color uint32)   // цвет текста: Button, Entry, TextView (view); свой GtkCssProvider
func SetBackground(h H, color uint32, panel bool) // фон; panel — Box/Grid: border-width → CSS padding

// Button, Check
func ButtonNew(id uint32, text string) H
func CheckNew(id uint32, text string) H
func ButtonSetText(h H, s string); ButtonText(h H) string // годится и для Check
func CheckSetChecked(h H, on bool); CheckChecked(h H) bool

// Entry
func EntryNew(id uint32) H
func EntrySetText(h H, s string); EntryText(h H) string
func EntrySetPlaceholder(h H, s string)
func EntrySetPassword(h H, on bool)
func EntrySetReadOnly(h H, on bool)
func EntrySetWidthChars(h H, n int)
func EntrySetActivatesDefault(h H, on bool)

// Combo
func ComboNew(id uint32, editable bool) H
func ComboAdd(h H, s string)
func ComboClear(h H)
func ComboSetSelected(h H, i int); ComboSelected(h H) int
func ComboSetText(h H, s string); ComboText(h H) string   // для editable — текст поля

// Slider, Progress
func SliderNew(id uint32, min, max int) H
func SliderSetValue(h H, v int); SliderValue(h H) int
func ProgressNew(id uint32) H
func ImageNew(id uint32) H
func ImageLoad(h H, data []byte) (w, ht int, ok bool) // декодировать и запомнить байты; пустые — убрать
func ImageShow(h H, w, ht int)                        // показать w×ht логических пикселей
func ReadFile(path string) ([]byte, bool)              // у nui нет пакета os (ADR-7)
func ProgressSetValue(h H, f float64); ProgressValue(h H) float64
func ProgressPulse(h H)
func ProgressSetText(h H, s string)

// TextView: outer — прокрутка (для раскладки), view — сам текст
func TextNew(id uint32) (outer, view H)
func TextSetText(view H, s string); TextText(view H) string
func TextAppend(view H, s string, autoscroll bool, maxLines int)
func TextSetReadOnly(view H, on bool); TextSetMonospace(view H, on bool); TextSetWrap(view H, on bool)
func TextSetEnter(view H, id uint32, on bool) // Enter без Shift → EvEnter (TextView.OnEnter)
func TextSetPlaceholder(view H, s string)

// List: outer — прокрутка, view — сам список; строки — по требованию через cell
func ListNew(id uint32, columns []string) (outer, view H)
func ListSetRows(view H, n int, cell func(row, col int) string) // выделение снимается, прокрутка — в начало
func ListUpdate(view H)
func ListSelected(view H) int; ListSelect(view H, row int) // -1 — нет / снять

// Контейнеры
func BoxNew(id uint32, vertical bool) H
func BoxSetSpacing(h H, px int)
func BoxPack(h, child H, expand bool)
func SpacerNew(id uint32) H
func GridNew(id uint32) H
func GridAttach(h, child H, col, row, cols, rows int)
func GridSetSpacing(h H, colGap, rowGap int)
func SeparatorNew(id uint32, vertical bool) H
func FrameNew(id uint32, title string) H
func FrameSetTitle(h H, s string)
func ScrollNew(id uint32) H
func ContainerSetChild(h, child H)      // Frame, Scroll: удалить старого, добавить нового

// Диалоги и оболочка
type Filter struct{ Name string; Patterns []string }
func MessageBox(parent H, kind uint8, title, text string) bool // kind: MsgInfo, MsgAlert, MsgAsk; true = OK/Да
func FileDialog(parent H, mode uint8, title, name string, f []Filter) []string // FileOpen, FileOpenMany, FileSave, FileFolder; nil — отмена
func OpenURI(parent H, uri string) bool
func ShowInFolder(path string) bool     // асинхронно, см. §7
func ClipboardSet(s string)
func ClipboardText() string             // "" — пусто или не текст
```

Каждая функция, которая создаёт виджет (`XxxNew`), обязана:
1. создать его;
2. выполнить нормализацию (§6);
3. подписать сигналы с `UD(id, kind)`, **включая `destroy` → EvDestroy**;
4. вернуть хендл.

## 4. Заголовок `gtk3_linux.h`

Структура файла:

```c
#ifndef NUI_GTK3_H
#define NUI_GTK3_H

/* ---------- 1. Базовые типы (в режиме проверки берутся из настоящих заголовков) ---------- */
#ifndef NUI_CHECK
typedef __UINTPTR_TYPE__ uintptr_t;
typedef __SIZE_TYPE__ size_t;
typedef int gint, gboolean;           typedef unsigned int guint, guint32;
typedef unsigned short guint16;       typedef unsigned long gulong;
typedef double gdouble;               typedef float gfloat;
typedef char gchar;                   typedef void *gpointer;
/* Перечисления: в ABI это int. В режиме проверки — настоящие enum. */
typedef int GtkOrientation, GtkAlign, GtkWindowType, GtkWindowPosition, GtkPolicyType,
            GtkWrapMode, GtkFileChooserAction, GtkMessageType, GtkButtonsType,
            GtkDialogFlags, PangoEllipsizeMode, PangoWeight, GConnectFlags,
            GIOCondition, GBusType, GDBusCallFlags, GtkJustification;
typedef void (*GCallback)(void);
typedef gboolean (*GSourceFunc)(gpointer);
typedef gboolean (*GUnixFDSourceFunc)(gint, GIOCondition, gpointer);
typedef void (*GDestroyNotify)(gpointer);
typedef void (*GClosureNotify)(gpointer, void *);
/* Непрозрачные объекты */
typedef struct _GtkWidget GtkWidget;           typedef struct _GtkWindow GtkWindow;
typedef struct _GtkContainer GtkContainer;     typedef struct _GtkBin GtkBin;
typedef struct _GtkBox GtkBox;                 typedef struct _GtkGrid GtkGrid;
typedef struct _GtkLabel GtkLabel;             typedef struct _GtkButton GtkButton;
typedef struct _GtkToggleButton GtkToggleButton;
typedef struct _GtkEntry GtkEntry;             typedef struct _GtkEditable GtkEditable;
typedef struct _GtkComboBox GtkComboBox;       typedef struct _GtkComboBoxText GtkComboBoxText;
typedef struct _GtkRange GtkRange;             typedef struct _GtkScale GtkScale;
typedef struct _GtkProgressBar GtkProgressBar; typedef struct _GtkFrame GtkFrame;
typedef struct _GtkScrolledWindow GtkScrolledWindow; typedef struct _GtkAdjustment GtkAdjustment;
typedef struct _GtkTextView GtkTextView;       typedef struct _GtkTextBuffer GtkTextBuffer;
typedef struct _GtkTextIter GtkTextIter;       typedef struct _GtkTextMark GtkTextMark;
typedef struct _GtkDialog GtkDialog;           typedef struct _GtkMessageDialog GtkMessageDialog;
typedef struct _GtkFileChooser GtkFileChooser; typedef struct _GtkFileFilter GtkFileFilter;
typedef struct _GtkNativeDialog GtkNativeDialog;
typedef struct _GtkFileChooserNative GtkFileChooserNative;
typedef struct _PangoAttrList PangoAttrList;   typedef struct _PangoAttribute PangoAttribute;
typedef struct _GError GError;                 typedef struct _GVariant GVariant;
typedef struct _GVariantType GVariantType;     typedef struct _GDBusConnection GDBusConnection;
typedef struct _GCancellable GCancellable;
typedef struct _GObject GObject;               typedef struct _GAsyncResult GAsyncResult;
typedef void (*GAsyncReadyCallback)(GObject *, GAsyncResult *, gpointer);
#endif

/* Место под GtkTextIter на стеке: настоящий размер ~80 байт, берём с запасом. */
typedef struct { void *p[16]; } NuiTextIter;

/* ---------- 2. Константы ---------- */
/* NUI_CONST(GTK_ALIGN_CENTER, 3) → enum { NUI_GTK_ALIGN_CENTER = 3 };
   в режиме проверки дополнительно _Static_assert(3 == GTK_ALIGN_CENTER). */

/* ---------- 3. Прототипы ---------- */
/* NUI_FN(ret, имя, (параметры)) → объявление;
   в режиме проверки → _Static_assert(__builtin_types_compatible_p(
                           __typeof__(&имя), ret (*)(параметры)), "имя").
   NUI_FN_CONST(ret, имя) — для функций без аргументов с G_GNUC_CONST (*_get_type):
   атрибут const у GCC — часть типа функции, поэтому сверяется тип результата. */
NUI_FN(GtkWidget *, gtk_label_new, (const gchar *str))
...

/* ---------- 4. Наши обёртки и трамплины ---------- */
/* Объявляются в ОБОИХ режимах, без #ifndef NUI_CHECK:
   - в обычной сборке cgo компилирует wrap_linux.c с этим заголовком, и расхождение
     объявления с определением — ошибка компиляции;
   - в режиме проверки wrap_linux.c компилируется против настоящих заголовков (§10, шаг 2),
     и без этих объявлений трамплины nuiSignal и др. остались бы необъявленными
     (в GCC ≥ 14 это ошибка, а не предупреждение). */
gulong nui_connect(gpointer obj, const char *sig, uintptr_t ud);
/* … остальные nui_* из §8 … */
void nuiSignal(void *obj, uintptr_t ud);
/* … остальные трамплины из runtime.md, §5 … */
#endif
```

### Константы (значения GTK3)

| Константа | Значение | Константа | Значение |
|---|---|---|---|
| `GTK_WINDOW_TOPLEVEL` | 0 | `GTK_WIN_POS_CENTER` | 1 |
| `GTK_ORIENTATION_HORIZONTAL` / `VERTICAL` | 0 / 1 | `GTK_ALIGN_FILL/START/END/CENTER` | 0 / 1 / 2 / 3 |
| `GTK_POLICY_AUTOMATIC` / `NEVER` | 1 / 2 | `GTK_WRAP_NONE` / `WORD_CHAR` | 0 / 3 |
| `PANGO_ELLIPSIZE_NONE/START/MIDDLE/END` | 0 / 1 / 2 / 3 | `PANGO_WEIGHT_NORMAL` / `BOLD` | 400 / 700 |
| `GTK_FILE_CHOOSER_ACTION_OPEN/SAVE/SELECT_FOLDER` | 0 / 1 / 2 | `GTK_RESPONSE_ACCEPT` | −3 |
| `GTK_RESPONSE_OK` / `YES` | −5 / −8 | `GTK_MESSAGE_INFO/WARNING/QUESTION/ERROR` | 0 / 1 / 2 / 3 |
| `GTK_BUTTONS_OK` / `YES_NO` | 1 / 4 | `GTK_DIALOG_MODAL` / `DESTROY_WITH_PARENT` | 1 / 2 |
| `G_PRIORITY_DEFAULT` | 0 | `G_IO_IN/PRI/OUT/ERR/HUP/NVAL` | 1 / 2 / 4 / 8 / 16 / 32 |
| `G_BUS_TYPE_SESSION` | 2 | `G_DBUS_CALL_FLAGS_NONE` | 0 |
| `GTK_JUSTIFY_LEFT/RIGHT/CENTER` | 0 / 1 / 2 | | |

Все значения подтверждает `make check-protos`. Если хоть одно не совпадёт с настоящим заголовком, сборка проверки упадёт.

## 5. Нужные C-функции (≈ 100)

Сигнатуры — как в документации GTK3. Отмеченные **(w)** вызываются только из обёрток, **(t)** — только из тестов (`sim_linux.go`).

**GLib, GObject, цикл событий**
```
gboolean gtk_init_check(int *argc, char ***argv)
void     gtk_main(void)
void     gtk_main_quit(void)
void     g_set_prgname(const gchar *prgname)
gulong   g_signal_connect_data(gpointer instance, const gchar *signal, GCallback handler,
                               gpointer data, GClosureNotify destroy, GConnectFlags flags)  (w)
guint    g_timeout_add_full(gint prio, guint ms, GSourceFunc fn, gpointer data, GDestroyNotify n)   (w)
guint    g_idle_add_full(gint prio, GSourceFunc fn, gpointer data, GDestroyNotify n)               (w)
guint    g_unix_fd_add_full(gint prio, gint fd, GIOCondition cond, GUnixFDSourceFunc fn,
                            gpointer data, GDestroyNotify n)                                        (w)
gboolean g_source_remove(guint tag)
void     g_free(gpointer mem)
void     g_object_unref(gpointer object)
void     g_error_free(GError *error)                                 (w)
size_t   strlen(const char *s)
int      gettid(void)                     /* pid_t; glibc ≥ 2.30 — ниже фактического минимума 2.34, см. §9 */
int      getpid(void)
```

**GtkWidget, GtkContainer, GtkBin**
```
void gtk_widget_show(GtkWidget*)            void gtk_widget_hide(GtkWidget*)
void gtk_widget_destroy(GtkWidget*)
void gtk_widget_set_sensitive(GtkWidget*, gboolean)   gboolean gtk_widget_get_sensitive(GtkWidget*)
void gtk_widget_set_visible(GtkWidget*, gboolean)     gboolean gtk_widget_get_visible(GtkWidget*)
void gtk_widget_set_tooltip_text(GtkWidget*, const gchar*)
void gtk_widget_set_halign(GtkWidget*, GtkAlign)      void gtk_widget_set_valign(GtkWidget*, GtkAlign)
void gtk_widget_set_hexpand(GtkWidget*, gboolean)     void gtk_widget_set_vexpand(GtkWidget*, gboolean)
void gtk_widget_set_size_request(GtkWidget*, gint, gint)
void gtk_widget_grab_focus(GtkWidget*)
void gtk_widget_set_can_default(GtkWidget*, gboolean)
gboolean gtk_widget_activate(GtkWidget*)                                       (t)
void gtk_container_add(GtkContainer*, GtkWidget*)     void gtk_container_remove(GtkContainer*, GtkWidget*)
void gtk_container_set_border_width(GtkContainer*, guint)
GtkWidget *gtk_bin_get_child(GtkBin*)
```

**GtkWindow**
```
GtkWidget *gtk_window_new(GtkWindowType)
void gtk_window_set_title(GtkWindow*, const gchar*)
void gtk_window_set_default_size(GtkWindow*, gint, gint)   void gtk_window_resize(GtkWindow*, gint, gint)
void gtk_window_set_resizable(GtkWindow*, gboolean)
void gtk_window_set_position(GtkWindow*, GtkWindowPosition)
void gtk_window_set_default(GtkWindow*, GtkWidget*)
void gtk_window_set_icon_name(GtkWindow*, const gchar*)
void gtk_window_present(GtkWindow*)
void gtk_window_close(GtkWindow*)                                              (t)
```

**GtkLabel и Pango**
```
GtkWidget *gtk_label_new(const gchar*)
void gtk_label_set_text(GtkLabel*, const gchar*)      const gchar *gtk_label_get_text(GtkLabel*)
void gtk_label_set_ellipsize(GtkLabel*, PangoEllipsizeMode)
void gtk_label_set_line_wrap(GtkLabel*, gboolean)
void gtk_label_set_selectable(GtkLabel*, gboolean)
void gtk_label_set_xalign(GtkLabel*, gfloat)
void gtk_label_set_justify(GtkLabel*, GtkJustification)
void gtk_label_set_attributes(GtkLabel*, PangoAttrList*)                      (w)
PangoAttrList  *pango_attr_list_new(void)                                     (w)
void            pango_attr_list_insert(PangoAttrList*, PangoAttribute*)       (w)
void            pango_attr_list_unref(PangoAttrList*)                         (w)
PangoAttribute *pango_attr_weight_new(PangoWeight)                            (w)
PangoAttribute *pango_attr_foreground_new(guint16 r, guint16 g, guint16 b)    (w)
```

**GtkButton, GtkCheckButton, GtkToggleButton**
```
GtkWidget *gtk_button_new_with_label(const gchar*)
void gtk_button_set_label(GtkButton*, const gchar*)   const gchar *gtk_button_get_label(GtkButton*)
void gtk_button_clicked(GtkButton*)                                            (t)
GtkWidget *gtk_check_button_new_with_label(const gchar*)
void gtk_toggle_button_set_active(GtkToggleButton*, gboolean)
gboolean gtk_toggle_button_get_active(GtkToggleButton*)
```

**GtkEntry**
```
GtkWidget *gtk_entry_new(void)
void gtk_entry_set_text(GtkEntry*, const gchar*)      const gchar *gtk_entry_get_text(GtkEntry*)
void gtk_entry_set_placeholder_text(GtkEntry*, const gchar*)
void gtk_entry_set_visibility(GtkEntry*, gboolean)
void gtk_editable_set_editable(GtkEditable*, gboolean)
void gtk_entry_set_width_chars(GtkEntry*, gint)
void gtk_entry_set_activates_default(GtkEntry*, gboolean)
```

**GtkComboBoxText**
```
GtkWidget *gtk_combo_box_text_new(void)               GtkWidget *gtk_combo_box_text_new_with_entry(void)
void gtk_combo_box_text_append_text(GtkComboBoxText*, const gchar*)
void gtk_combo_box_text_remove_all(GtkComboBoxText*)
gint gtk_combo_box_get_active(GtkComboBox*)           void gtk_combo_box_set_active(GtkComboBox*, gint)
gchar *gtk_combo_box_text_get_active_text(GtkComboBoxText*)    /* g_free */
```

**GtkScale, GtkRange, GtkProgressBar**
```
GtkWidget *gtk_scale_new_with_range(GtkOrientation, gdouble min, gdouble max, gdouble step)
void gtk_scale_set_digits(GtkScale*, gint)
void gtk_range_set_value(GtkRange*, gdouble)          gdouble gtk_range_get_value(GtkRange*)
GtkWidget *gtk_progress_bar_new(void)
void gtk_progress_bar_set_fraction(GtkProgressBar*, gdouble)  gdouble gtk_progress_bar_get_fraction(GtkProgressBar*)
void gtk_progress_bar_pulse(GtkProgressBar*)
void gtk_progress_bar_set_text(GtkProgressBar*, const gchar*)
void gtk_progress_bar_set_show_text(GtkProgressBar*, gboolean)
```

**GtkTextView, GtkScrolledWindow**
```
GtkWidget *gtk_scrolled_window_new(GtkAdjustment*, GtkAdjustment*)
void gtk_scrolled_window_set_policy(GtkScrolledWindow*, GtkPolicyType, GtkPolicyType)
GtkWidget *gtk_text_view_new(void)
GtkTextBuffer *gtk_text_view_get_buffer(GtkTextView*)
void gtk_text_buffer_set_text(GtkTextBuffer*, const gchar*, gint len)
void gtk_text_view_set_editable(GtkTextView*, gboolean)
void gtk_text_view_set_cursor_visible(GtkTextView*, gboolean)
void gtk_text_view_set_monospace(GtkTextView*, gboolean)
void gtk_text_view_set_wrap_mode(GtkTextView*, GtkWrapMode)
void gtk_text_buffer_get_start_iter(GtkTextBuffer*, GtkTextIter*)            (w)
void gtk_text_buffer_get_end_iter(GtkTextBuffer*, GtkTextIter*)              (w)
void gtk_text_buffer_get_iter_at_line(GtkTextBuffer*, GtkTextIter*, gint)    (w)
gchar *gtk_text_buffer_get_text(GtkTextBuffer*, const GtkTextIter*, const GtkTextIter*, gboolean) (w)
void gtk_text_buffer_insert(GtkTextBuffer*, GtkTextIter*, const gchar*, gint) (w)
void gtk_text_buffer_delete(GtkTextBuffer*, GtkTextIter*, GtkTextIter*)      (w)
gint gtk_text_buffer_get_line_count(GtkTextBuffer*)                           (w)
gboolean gtk_text_iter_starts_line(const GtkTextIter*)                        (w)
GtkTextMark *gtk_text_buffer_get_mark(GtkTextBuffer*, const gchar *name)     (w)
GtkTextMark *gtk_text_buffer_create_mark(GtkTextBuffer*, const gchar *name,
                                         const GtkTextIter*, gboolean left_gravity) (w)
void gtk_text_buffer_move_mark(GtkTextBuffer*, GtkTextMark*, const GtkTextIter*) (w)
void gtk_text_view_scroll_mark_onscreen(GtkTextView*, GtkTextMark*)          (w)
```

**Контейнеры**
```
GtkWidget *gtk_box_new(GtkOrientation, gint spacing)
void gtk_box_set_spacing(GtkBox*, gint)
void gtk_box_pack_start(GtkBox*, GtkWidget*, gboolean expand, gboolean fill, guint padding)
GtkWidget *gtk_grid_new(void)
void gtk_grid_attach(GtkGrid*, GtkWidget*, gint left, gint top, gint width, gint height)
void gtk_grid_set_row_spacing(GtkGrid*, guint)        void gtk_grid_set_column_spacing(GtkGrid*, guint)
GtkWidget *gtk_separator_new(GtkOrientation)
GtkWidget *gtk_frame_new(const gchar*)                void gtk_frame_set_label(GtkFrame*, const gchar*)
```

**Диалоги и оболочка**
```
GtkWidget *gtk_message_dialog_new(GtkWindow*, GtkDialogFlags, GtkMessageType, GtkButtonsType,
                                  const gchar *fmt, ...)                                  (w)
void gtk_message_dialog_format_secondary_text(GtkMessageDialog*, const gchar *fmt, ...)    (w)
gint gtk_dialog_run(GtkDialog*)                                                            (w)
GtkFileChooserNative *gtk_file_chooser_native_new(const gchar *title, GtkWindow *parent,
                      GtkFileChooserAction, const gchar *accept, const gchar *cancel)
gint   gtk_native_dialog_run(GtkNativeDialog*)
gchar *gtk_file_chooser_get_filename(GtkFileChooser*)                  /* g_free */
void   gtk_file_chooser_set_current_name(GtkFileChooser*, const gchar*)
void   gtk_file_chooser_set_do_overwrite_confirmation(GtkFileChooser*, gboolean)
GtkFileFilter *gtk_file_filter_new(void)
void   gtk_file_filter_set_name(GtkFileFilter*, const gchar*)
void   gtk_file_filter_add_pattern(GtkFileFilter*, const gchar*)
void   gtk_file_chooser_add_filter(GtkFileChooser*, GtkFileFilter*)
gboolean gtk_show_uri_on_window(GtkWindow*, const char *uri, guint32 timestamp, GError **error)
gchar *g_filename_to_uri(const gchar *filename, const gchar *hostname, GError **error)  /* g_free */
gchar *g_strdup(const gchar *str)                                                          (w)
GDBusConnection *g_bus_get_sync(GBusType, GCancellable*, GError**)                         (w)
GVariant *g_variant_new(const gchar *format, ...)                                          (w)
void g_dbus_connection_call(GDBusConnection*, const gchar *bus, const gchar *path,
          const gchar *iface, const gchar *method, GVariant *params, const GVariantType *reply,
          GDBusCallFlags, gint timeout_ms, GCancellable*, GAsyncReadyCallback, gpointer)    (w)
GVariant *g_dbus_connection_call_finish(GDBusConnection*, GAsyncResult*, GError**)         (w)
void g_variant_unref(GVariant*)                                                            (w)
```

## 6. Нормализация каждого созданного виджета

Это делает внутренняя функция `adopt(h, id)` в `widget_linux.go`:

1. `gtk_widget_set_hexpand(h, FALSE)` и `gtk_widget_set_vexpand(h, FALSE)` — чтобы флаг expand не всплывал к родителям ([layout.md](layout.md), §6);
2. `nui_connect(h, "destroy", UD(id, EvDestroy))`;
3. `gtk_widget_show(h)` — для всех, **кроме окон** (ADR-14).

## 7. Таблица соответствия nui → GTK

| nui | Создание | Сигналы → Kind | Особенности |
|---|---|---|---|
| Window | `gtk_window_new(TOPLEVEL)`, `set_title`, `set_position(CENTER)`, `border_width 12` | `delete-event` → EvClose (через `nui_connect_ret`), `key-press-event` → EvKey (ответ 1 — клавиша обработана, окно не передаёт её виджету с фокусом), `destroy` → EvDestroy | Окно не показывается при создании. `WindowSetSize`: у растягиваемого окна до показа — `set_default_size`, после — `gtk_window_resize`; у нерастягиваемого — `gtk_widget_set_size_request` (окно = max(запрос, естественный размер)). `SetDefault`: `gtk_widget_set_can_default(btn, TRUE)` + `gtk_window_set_default` |
| Label | `gtk_label_new`, `gtk_label_set_xalign(0.0)` | — | Жирность и цвет: `nui_label_style` хранит упакованное состояние в самой метке (`g_object_set_data`, число, не указатель Go) — у `nui.Label` нет лишних полей, и `hello` от них не растёт. `SetTextAlign` → `gtk_label_set_xalign` (0 / 0.5 / 1) **и** `gtk_label_set_justify` (LEFT / CENTER / RIGHT): `xalign` двигает блок текста целиком, а `justify` выравнивает строки внутри многострочного текста. `SetStyle` → `nui_label_style` (атрибуты Pango: вес 700/400, цвет `r*257, g*257, b*257`). Многоточие: `gtk_label_set_ellipsize`; с `EllipsisNone` минимальная ширина = ширина текста |
| Button | `gtk_button_new_with_label` | `clicked` → EvClick | Подчёркивания в тексте не превращаются в мнемоники — это нормально |
| Entry | `gtk_entry_new`, `set_activates_default(TRUE)` | `changed` → EvChange, `activate` → EvEnter | `SetReadOnly` → `gtk_editable_set_editable(!on)`. nui выключает `activates_default`, когда задан `OnEnter` |
| Check | `gtk_check_button_new_with_label` | `toggled` → EvChange | — |
| Combo | `gtk_combo_box_text_new` / `_new_with_entry` | `changed` → EvChange | У ComboEntry текст поля: `gtk_bin_get_child` → `GtkEntry`; его `changed` уже приходит в `changed` комбобокса, отдельно подписываться не нужно. Полю ComboEntry ставится `gtk_entry_set_activates_default(TRUE)`: Enter в нём нажимает кнопку окна по умолчанию, как в Entry без `OnEnter`. `Clear` (`remove_all`) тоже присылает `changed` — гасится `quiet` |
| Slider | `gtk_scale_new_with_range(HORIZONTAL, min, max, 1)`, `set_digits(0)` | `value-changed` → EvChange | Значение округляется до целого на стороне Go. При `min >= max` GTK вернул бы `NULL`, поэтому nui проверяет это до вызова и паникует |
| Progress | `gtk_progress_bar_new` | — | `SetText("")` → `set_show_text(FALSE)` |
| List | `nui_list_new`: прокрутка + `gtk_tree_view_new` (`fixed-height-mode`, без поиска, выбор SINGLE); колонка — `nui_list_column`: `GtkCellRendererText` с многоточием, `FIXED`, `resizable`, `expand` у последней | `changed` **у GtkTreeSelection** → EvChange; `row-activated` (свой C-обработчик: у сигнала лишние аргументы) → EvEnter | Своя модель `NuiListModel` (GObject + интерфейс `GtkTreeModel`, раскладки сверяет check-protos): строка — номер в `iter->user_data`, `get_value` спрашивает текст у Go (`nuiListCell`). `SetRows` — новая модель; старую отпускает GtkTreeView, её `finalize` вызывает `nuiListDrop`. Ширина колонок — по заголовку и первым 100 строкам (`PangoLayout`), не больше 400. После `SetRows` курсор ставится на первую строку и выделение снимается: иначе в режиме SINGLE GtkTreeView сам выделит строку при получении фокуса. Повторы `changed` (движение курсора) отбрасывает nui |
| Контекстное меню | `gtk_menu_new` в данных виджета `"nui-ctx"`, `gtk_menu_attach_to_widget` — уничтожится с ним (`nui_ctx_set`) | `button-press-event` (кнопка 3) и `popup-menu` (клавиша «Меню», Shift+F10) на виджете — у GtkScrolledWindow и на его ребёнке (TextView, список); `button-press-event` на окне верхнего уровня (подключается при появлении в нём виджета с меню: `hierarchy-changed`) | Виджеты со своим окном событий (поля, список, холст, кнопки) ловят щелчок сами, раньше своего обычного меню. У подписей и рядов окна нет — щелчок доходит до окна, и `nui_ctx_owner_at` ищет самый глубокий видимый виджет под указателем (`gtk_container_forall`, `translate_coordinates`), меню — у него или ближайшего выше. Список: перед меню `gtk_tree_view_get_path_at_pos` → `set_cursor`. Показ — `gtk_menu_popup_at_pointer`, по клавише — `gtk_menu_popup_at_widget` |
| Split | `gtk_paned_new(orientation)`; первая — `gtk_paned_pack1(resize=FALSE, shrink=FALSE)`, вторая — `pack2(resize=TRUE, shrink=FALSE)` | — | `gtk_paned_set_position`, `get_position`. Без `set_position` GtkPaned даёт первой панели её естественный размер и не меняет его при растягивании окна; упоры — минимумы панелей. Это правило повторяет драйвер Win32 |
| Spin | `gtk_spin_button_new_with_range(min, max, 1)`, `set_numeric(TRUE)` (только цифры и знак), `gtk_entry_set_activates_default(TRUE)` | `value-changed` → EvChange | Значение GtkSpinButton принимает сама: стрелки и ↑ ↓ — сразу, набранное — по `activate` (Enter: сначала принять, потом кнопка по умолчанию) и `focus-out`; вне диапазона — край, пустое — 0 и тоже в диапазон (замер). `get_value_as_int` — принятое значение. `SetValue` — `set_value` (её `value-changed` гасит `quiet`) |
| Radio | `gtk_box_new(VERTICAL, 6)`; вариант — `nui_radio_add`: `gtk_radio_button_new_with_label_from_widget` (группа — от первой кнопки ряда, первая включена), `pack_start` | `toggled` у каждой кнопки (свой C-обработчик) → EvChange, только от включившейся: выключившаяся тоже присылает `toggled` | Выбранный — первая включённая кнопка ряда (`nui_radio_selected`), `SetSelected` — `set_active` кнопки (её `toggled` гасит `quiet`). Ряд — `gtk_orientable_set_orientation` и промежуток 12. Стрелки обрабатывает сама GtkRadioButton: соседняя кнопка получает фокус и включается, у края группы — ничего (замер: «вниз» на последней остаётся на ней) |
| Tabs | `nui_tabs_new`: `gtk_notebook_new`, `set_scrollable(TRUE)` — ярлыки, которые не помещаются, прокручиваются стрелками; ярлык — `GtkLabel` (`append_page`) | `switch-page` дважды: до класс-обработчика (`tabs_before`) запоминается, видна ли текущая страница, после (`G_CONNECT_AFTER`) — EvChange, если была видна | `switch-page` приходит и на программные переключения: GTK сам уходит со страницы, которую прячут или убирают (следующая видимая, иначе предыдущая). С видимой переключает только пользователь или nui (`Add`, `Remove`, `SetCurrent` — их гасит `quiet`). После класс-обработчика `get_current_page` уже новая. Спрятанная страница прячет свой ярлык, `set_current_page` на неё не переходит; не текущие страницы не mapped — `Size` 0, 0. Заголовок — `set_tab_label_text` |
| Canvas | `gtk_drawing_area_new` + `gtk_widget_add_events(POINTER_MOTION, BUTTON_PRESS, BUTTON_RELEASE, SCROLL)` | `draw` → EvDraw (arg — `cairo_t*`, через `nui_connect_ret`), `button-press-event`, `button-release-event`, `motion-notify-event`, `scroll-event` → EvMouse (ответ 1 — обработано) | Рисование — `nui_paint_*` (cairo; логические пиксели, масштаб экрана cairo учитывает сам; каждая фигура — с нового пути, эллипс — дуга в масштабе, толщина — после `restore`); начальный цвет — `gtk_style_context_get_color`. Текст — `gtk_widget_create_pango_layout` + `pango_cairo_show_layout`. Мышь — функциями доступа GDK (`gdk_event_get_event_type`, `_coords`, `_button`, `_state`, `_scroll_direction`); `GDK_2BUTTON_PRESS` пропускается: GTK присылает его лишним событием после второго обычного нажатия. Раскладки `GdkEventButton`, `Motion`, `Scroll` нужны только тестам (`nui_sim_mouse`), их сверяет check-protos |
| Меню окна | `nui_window_menubar`: при первом вызове окно перестраивается — `GtkWindow` (без отступа) → вертикальный ряд [`GtkMenuBar`, inner], inner — ряд с отступом окна, в нём содержимое (`WindowSetChild`, `WindowSetPadding` работают с ним). Без меню окно прежнее. Меню — `GtkMenu` под `GtkMenuItem` строки; пункты — `GtkMenuItem`, `GtkCheckMenuItem`, `GtkSeparatorMenuItem` | `activate` → EvClick (у флажка состояние уже переключено: класс-обработчик идёт первым), `destroy` → EvDestroy | Сочетание — `gtk_accel_label_set_accel` у `GtkAccelLabel` пункта: только подпись, без `GtkAccelGroup`; срабатывает в nui через EvKey окна. `gtk_check_menu_item_set_active` тоже шлёт `activate` — nui гасит его `quiet`. `nui_window_menubar_resize`: строка меню появилась, сменилась или убрана — окно (размер по умолчанию, текущий размер, size request) меняется на разницу её высоты: область под меню прежняя, как на Windows. `WindowSize` вычитает высоту строки, `WindowSetSize` прибавляет |
| OnDropFiles | `nui_drop_enable`: `gtk_drag_dest_set(окно, GTK_DEST_DEFAULT_ALL, {"text/uri-list"}, COPY)` | `drag-data-received` (свой C-обработчик) → EvDrop | `GTK_DEST_DEFAULT_ALL`: GTK сам подсвечивает, запрашивает данные и вызывает `gtk_drag_finish`. `gtk_selection_data_get_uris` → `g_filename_from_uri` (не локальные URI пропускаются) → arg события — строка «путь\0путь\0\0». Поля ввода — сами приёмники текста: брошенное на них остаётся им |
| Args | — | — | `/proc/self/cmdline` через `ReadFile` (`g_file_get_contents`), разделитель — `\0` |
| OnKey (`KeyDecode`) | — | — | Из `GdkEventKey`: `keyval` → код nui (буквы — в верхний регистр, KP_*, F1…F12, стрелки…), `state` → Shift (`GDK_SHIFT_MASK`), Ctrl (`GDK_CONTROL_MASK`), Alt (`GDK_MOD1_MASK`). Если `keyval` не латиница и не цифра (русская раскладка: Ctrl+S — «ы»), `nui_key_by_code` ищет латинскую букву или цифру той же клавиши (`hardware_keycode`) во всех раскладках через `gdk_keymap_get_entries_for_keycode` |
| TextView.OnEnter, SetPlaceholder | `nui_text_enter`, `nui_text_placeholder` | `key-press-event` поля (свой C-обработчик) → EvEnter | Enter (Return, KP_Enter, ISO_Enter) без Shift: сначала `gtk_text_view_im_context_filter_keypress` — Enter, которым метод ввода заканчивает набор, его; иначе EvEnter, и перевода строки нет. Подсказка — обработчик `draw` после текста (`G_CONNECT_AFTER`): пустой буфер → `PangoLayout` с переносом от начала первой строки (`get_iter_location` → `buffer_to_window_coords`), цвет текста с половинной прозрачностью (`pango_cairo_show_layout`); `changed` буфера перерисовывает поле целиком |
| TextView | `gtk_scrolled_window_new(NULL, NULL)` + `set_policy(AUTOMATIC, AUTOMATIC)` + `gtk_text_view_new` внутри | `changed` **у буфера** → EvChange (тот же ud) | Хендлов два: outer — для раскладки, view — для текста. `SetWrap` → `WORD_CHAR`/`NONE`. `SetReadOnly` → `editable` и `cursor_visible`. `Append`, `Clear` и `SetText` присылают `changed` синхронно — гасится `quiet`. `SetText` на непустом буфере и вставка поверх выделения присылают **два** `changed` (удаление и вставка) |
| Separator | `gtk_separator_new(orientation)` | — | — |
| Frame | `gtk_frame_new(title)` | — | Ребёнок — `gtk_container_add` |
| Scroll | `gtk_scrolled_window_new(NULL, NULL)` + `AUTOMATIC` | — | — |
| Box | `gtk_box_new(o, 6)` | — | Ребёнок — `gtk_box_pack_start(box, c, expand, TRUE, 0)` |
| Stretch | `gtk_box_new(HORIZONTAL, 0)` | — | Упаковывается с `expand = TRUE` |
| Grid | `gtk_grid_new`, spacing 6/6 | — | См. [layout.md](layout.md), §6 |
| Message/Alert/Ask | `nui_message` | — | Типы: INFO+OK, ERROR+OK, QUESTION+YES_NO; ответ YES/OK → `true` |
| OpenFile/OpenFiles/SaveFile/ChooseFolder | `gtk_file_chooser_native_new` | — | `accept_label` и `cancel_label` = `NULL`: подписи по умолчанию GTK переведёт на язык системы сам. OpenFiles → `set_select_multiple(TRUE)`. Save → `do_overwrite_confirmation(TRUE)`, `set_current_name`. Ответ `ACCEPT` → `get_filenames` (GSList: каждую строку — `gostrFree`, список — `g_slist_free`); затем `g_object_unref(native)`. В тестах несколько файлов выбирает `nui_sim_file_chooser_add`: `select_filename` ставит курсор и снимает прежнее выделение (GTK 3.24), поэтому выделенные строки запоминаются и выделяются снова |
| Image | `nui_image_new` = `gtk_image_new` + данные `NuiImage` + `notify::scale-factor` | — | `GdkPixbufLoader` (`write`, `close` — и после ошибки) → `gdk_pixbuf_apply_embedded_orientation` (EXIF). Показ: `gdk_pixbuf_scale_simple(BILINEAR)` до w·k×h·k точек (k — `gtk_widget_get_scale_factor`) → `gdk_cairo_surface_create_from_pixbuf(pb, k)` → `gtk_image_set_from_surface`: на HiDPI картинка чёткая. Байты файла хранятся в `NuiImage`; при смене масштаба C-обработчик перерисовывает картинку из них сам, без Go. Декодированная копия живёт только между `ImageLoad` и `ImageShow` |
| OpenURI | `gtk_show_uri_on_window(parent, uri, 0, NULL)` | — | — |
| ShowInFolder | `g_filename_to_uri(path)` и `g_filename_to_uri(dir(path))` → `nui_show_items(uri, dir_uri)` | — | URI строит GLib (`%20`, не-UTF-8 байты имени); путь передаётся через `cpath`, без замены байтов. Вызов D-Bus **асинхронный**: интерфейс не ждёт файловый менеджер. Если тот ответил ошибкой (например, `ServiceUnknown`), C-колбэк сам открывает папку через `gtk_show_uri_on_window`. Нет D-Bus-сессии → папка открывается сразу |
| ClipboardSet / ClipboardText | `nui_clipboard()` = `gtk_clipboard_get(GDK_SELECTION_CLIPBOARD)` (атом 69, без libgdk); `gtk_clipboard_set_text(cb, s, -1)` + `gtk_clipboard_set_can_store(cb, NULL, 0)`; `gtk_clipboard_wait_for_text` → `g_free` | — | Буфер CLIPBOARD, не PRIMARY (выделение мышью). `set_can_store` — при выходе GTK отдаёт текст менеджеру буфера обмена. `wait_for_text` крутит вложенный цикл, пока владелец буфера не ответит |

## 8. C-обёртки (`wrap_linux.c`)

```c
gulong nui_connect(gpointer obj, const char *sig, uintptr_t ud);
    /* g_signal_connect_data(obj, sig, (GCallback)nuiSignal, (gpointer)ud, NULL, 0) */
gulong nui_connect_ret(gpointer obj, const char *sig, uintptr_t ud);
    /* то же с nuiSignalRet */
guint  nui_timeout(guint ms, uintptr_t ud);
    /* g_timeout_add_full(G_PRIORITY_DEFAULT, ms, (GSourceFunc)nuiSource, (gpointer)ud, NULL) */
guint  nui_idle(uintptr_t ud);
    /* g_idle_add_full(G_PRIORITY_DEFAULT, (GSourceFunc)nuiSource, (gpointer)ud, NULL) */
guint  nui_watch(int fd, unsigned cond, uintptr_t ud);
    /* g_unix_fd_add_full(G_PRIORITY_DEFAULT, fd, cond | HUP | ERR | NVAL, nuiWatch, (gpointer)ud, NULL) */
void   nui_label_style(GtkLabel *l, int bold, int has_color, unsigned rgb);
    /* новый PangoAttrList: weight (700/400), при has_color — foreground; set_attributes; unref */
char  *nui_text_get(GtkTextBuffer *b);
    /* start/end iter (NuiTextIter на стеке) → gtk_text_buffer_get_text(…, FALSE); освобождать через g_free */
void   nui_text_append(GtkTextView *v, const char *s, int len, int autoscroll, int max_lines);
    /* end iter → insert;
       если max_lines > 0: n = get_line_count − (end на начале строки ? 1 : 0)
         (GTK считает пустую строку после завершающего \n отдельной строкой);
         если n > max_lines — delete [начало, начало строки (n − max_lines)];
       если autoscroll — метка "nui-end" (создаётся при первом вызове: create_mark(…, left_gravity = FALSE))
         переносится в конец (move_mark), затем scroll_mark_onscreen(метка).
       Курсор и выделение пользователя не трогаются: place_cursor не используется. */
int    nui_message(GtkWindow *parent, int type, int buttons, const char *title, const char *text);
    /* gtk_message_dialog_new(parent, MODAL|DESTROY_WITH_PARENT, type, buttons, "%s", title);
       format_secondary_text("%s", text); gtk_dialog_run; gtk_widget_destroy; вернуть response */
int    nui_show_items(const char *uri, const char *dir_uri);
    /* g_bus_get_sync(SESSION); нет сессии → вернуть gtk_show_uri_on_window(NULL, dir_uri, 0, NULL).
       Иначе g_dbus_connection_call("org.freedesktop.FileManager1", "/org/freedesktop/FileManager1",
       "org.freedesktop.FileManager1", "ShowItems", g_variant_new("(^ass)", {uri, NULL}, ""),
       NULL, NONE, -1 (таймаут D-Bus по умолчанию), NULL, nui_show_items_done, g_strdup(dir_uri));
       вернуть 1: запрос отправлен.
       nui_show_items_done (static, только в C): call_finish; при ошибке — gtk_show_uri_on_window(NULL, dir_uri, …);
       освободить GError, GVariant, dir_uri (g_free) и соединение (g_object_unref).
       Таймаут намеренно не укорочен: если файловый менеджер запускается дольше таймаута, он всё равно
       выполнит ShowItems, и вместе с запасным путём откроются два окна. */
```

Трамплины объявлены в `.h` как `extern` с сигнатурами из [runtime.md](runtime.md), §5. Обёртки приводят их к `GCallback`/`GSourceFunc` и т. п. В Go трамплины принимают `ud` как `C.uintptr_t`, а не указатель, чтобы в Go не появлялись «поддельные» указатели. По ABI `gpointer` и `uintptr_t` передаются одинаково на всех поддерживаемых платформах (amd64, arm64). На том же приёме держится и сам макрос `G_CALLBACK` в GLib.

## 9. Линковка

```go
// link_linux.go
/*
#cgo LDFLAGS: -l:libgtk-3.so.0 -l:libgobject-2.0.so.0 -l:libglib-2.0.so.0 -l:libgio-2.0.so.0 -l:libpango-1.0.so.0
#cgo LDFLAGS: -l:libgdk-3.so.0 -l:libgdk_pixbuf-2.0.so.0 -l:libcairo.so.2 -l:libpangocairo-1.0.so.0
#include "gtk3_linux.h"
*/
import "C"
```

**Флаги размера в `#cgo` писать нельзя.** gc пропускает в `#cgo CFLAGS/LDFLAGS` только флаги из белого списка (`cmd/go/internal/work/security.go`). `-ffunction-sections`, `-fdata-sections` и `-Wl,--gc-sections` в него не входят, и сборка падает с ошибкой `invalid flag in #cgo CFLAGS: -ffunction-sections` (проверено на Go 1.26). Переменные окружения белым списком не проверяются, поэтому Makefile передаёт эти флаги через них ([build.md](build.md), §4):

```make
export CGO_CFLAGS  := -Os -ffunction-sections -fdata-sections
export CGO_LDFLAGS := -Wl,--gc-sections -Wl,--as-needed
```

- go ставит флаги из `CGO_LDFLAGS` **перед** флагами из `#cgo`, поэтому `--as-needed` действует на все `-l:` из `link_linux.go`.
- Обычный `go build` без Makefile тоже работает, только C-часть не вычищается.

- `-l:имя.so.N` линкует по точному имени файла. Поэтому хватает runtime-пакетов: `libgtk-3-0` (в Debian 13 — `libgtk-3-0t64`), `libglib2.0-0`, `libpango-1.0-0`. Символьные ссылки `libgtk-3.so` из `-dev` не нужны.
- `--gc-sections` и `-ffunction-sections` выбрасывают неиспользуемые обёртки, а `--as-needed` — неиспользуемые библиотеки. Например, без `ShowInFolder` не нужна libgio, а без `Image` — libgdk-3, libgdk_pixbuf-2.0 и libcairo, без подсказки в `TextView` (её использует и `Chat`) — libpangocairo (вторая строка `#cgo LDFLAGS`; всё это — зависимости самой GTK, отдельно ставить не нужно).
- gc при cgo в стороннем пакете линкует внешним линковщиком (gcc), поэтому флаги работают.
- TinyGo (ADR-17) линкует встроенным `ld.lld`. Путь `-L/usr/lib/<multiarch>`, `crt*.o`, `-lc` и `--dynamic-linker` задаёт target-файл nui, поэтому `-l:имя.so.N` из `#cgo LDFLAGS` находятся без дополнительных переменных. `CGO_CFLAGS`/`CGO_LDFLAGS` TinyGo не нужны: релизная сборка спайка S3 весит 35 952 байта без них.

Проверка после сборки: `ldd bin/hello` показывает `libgtk-3.so.0` и её зависимости. `readelf -d bin/hello | grep NEEDED` показывает только то, что мы указали явно (плюс libc).

**Версия glibc.** Бинарник требует ту glibc, на которой собран, или новее: загрузчик сверяет версии символов (`GLIBC_2.34` и т. п.). Поэтому релизные бинарники собираются на старейшей поддерживаемой системе — Debian 12 (glibc 2.36). Отсюда фактический минимум для запуска — glibc 2.34: бинарник, собранный с glibc 2.34 и новее, ссылается на `__libc_start_main@GLIBC_2.34` ([goals.md](goals.md)). `gettid()` (glibc 2.30) эту границу не поднимает. `make abi` проверяет, что максимальная версия `GLIBC_` в релизных бинарниках не выше 2.36:

```sh
objdump -T bin/small/logger | grep -o 'GLIBC_[0-9.]*' | sort -uV | tail -1
```

## 10. Проверка прототипов (`make check-protos`)

Ошибка в собственном прототипе — это неопределённое поведение без единого предупреждения компилятора. Поэтому в CI (на машине с `libgtk-3-dev`) выполняется:

```sh
# tools/check-protos.sh
CFLAGS="-std=c11 -D_GNU_SOURCE -DNUI_CHECK -Werror $(pkg-config --cflags gtk+-3.0 gio-unix-2.0)"
# 1. Все NUI_FN и NUI_CONST сверяются с настоящими объявлениями через _Static_assert
printf '#include <gtk/gtk.h>\n#include <glib-unix.h>\n#include <gio/gio.h>\n#include <stdint.h>\n#include <string.h>\n#include <unistd.h>\n#include "gtk3_linux.h"\n' \
  | cc $CFLAGS -I internal/drv -x c -fsyntax-only -
# 2. Обёртки компилируются против настоящих заголовков (их объявления и трамплины — в разделе 4 gtk3_linux.h)
cc $CFLAGS -I internal/drv -include gtk/gtk.h -include glib-unix.h -include gio/gio.h \
   -include stdint.h -include string.h \
   -fsyntax-only internal/drv/wrap_linux.c
# 3. Размер NuiTextIter не меньше настоящего GtkTextIter
#    (_Static_assert(sizeof(NuiTextIter) >= sizeof(GtkTextIter)) — внутри gtk3_linux.h в режиме NUI_CHECK)
```

- Несовпадение количества или типа параметров, типа результата или значения константы → `static assertion failed: "имя"`.
- Перечисления в режиме проверки берутся настоящие, поэтому проверка не ругается на то, что в обычном режиме `enum` объявлен как `int`. На ABI это не влияет.
- Если компилятор не примет имена параметров в `ret (*)(params)`, уберите имена из прототипов.
- Раздел 4 заголовка (обёртки и трамплины) в режиме проверки **не** выключается: иначе шаг 2 упадёт на необъявленных `nuiSignal` и др.
- В режиме проверки `NUI_FN` превращается в `_Static_assert` и ничего не объявляет. Поэтому всё, что вызывает `wrap_linux.c`, в шаге 2 должно приходить из настоящих заголовков — отсюда `-include`.

## 11. Подводные камни GTK3 (собраны заранее)

1. `g_signal_connect` — это **макрос**, в `.so` такого символа нет. Используйте `nui_connect`.
2. Функции с переменным числом аргументов (`gtk_message_dialog_new`, `g_variant_new`) cgo вызвать не может — только через обёртку.
3. `gtk_widget_show_all` **запрещён** (ADR-14).
4. Программные изменения порождают `changed`: не только `set_text` и `set_active`, но и `remove_all` у Combo, вставка и удаление в `GtkTextBuffer` (`Append`, `Clear`). Это гасится счётчиком `quiet`, которым обёрнут **каждый** изменяющий метод ([runtime.md](runtime.md), §7).
5. У `GtkLabel` по умолчанию `xalign = 0.5` (текст по центру). Ставим `0.0` при создании.
6. Ячейка с многоточием сжимается, только если у метки `halign = FILL` (это значение по умолчанию) и родитель отдаёт ей меньше места.
7. Внутри `gtk_dialog_run` и `gtk_native_dialog_run` крутится вложенный цикл, и в это время срабатывают таймеры, `Watch` и `Do` (это записано в контракте).
8. На Wayland `gtk_window_set_position` и любые координаты игнорируются. Это нормально.
9. В тестах ставьте `NO_AT_BRIDGE=1`, чтобы не было предупреждений at-spi, и `G_DEBUG=fatal-criticals`, чтобы любое `Gtk-CRITICAL` роняло тест: критическое предупреждение GTK — это всегда наш баг.
10. `gtk_init_check(NULL, NULL)` допустим: аргументы командной строки GTK (`--display`) nui не разбирает.
11. `GtkTextIter` — структура на стеке; её место выделяет только C-обёртка (`NuiTextIter`). Go к ней не прикасается.
12. `gtk_combo_box_text_get_active_text`, `gtk_file_chooser_get_filename`, `gtk_text_buffer_get_text` и `g_filename_to_uri` возвращают строки, которые **освобождаем мы** (`g_free`).
13. `gtk_text_buffer_set_text` на непустом буфере и вставка поверх выделения присылают **два** `changed`: удаление и вставка. Поэтому `OnChange` у TextView может сработать дважды за одно действие.
14. GTK проверяет UTF-8: `gtk_text_buffer_insert` и другие функции на невалидной строке пишут `Gtk-CRITICAL` и **ничего не делают**, а с `G_DEBUG=fatal-criticals` процесс падает. Поэтому `cstr` заменяет битые последовательности на U+FFFD ([runtime.md](runtime.md), §12). Пути файловой системы не обязаны быть UTF-8 и идут через `cpath` без замены.
15. `gtk_text_buffer_get_line_count` считает пустую строку после завершающего `\n` отдельной строкой: у `"a\nb\n"` три строки. `nui_text_append` это учитывает (§8).
16. `gtk_window_close` (в тестах — `SimClose`) присылает `delete-event` не сразу, а через idle-источник. Результат проверяется после того, как цикл событий отработает ([testing.md](testing.md), §3).
