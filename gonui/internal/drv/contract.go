package drv

// Общие для всех ОС константы и типы контракта драйвера (docs/driver-gtk3.md, §3).
// Функции контракта — в файлах *_linux.* и *_windows.*.

// Виды сообщений MessageBox.
const (
	MsgInfo  = 0 // информация, кнопка OK
	MsgAlert = 1 // ошибка или предупреждение, кнопка OK
	MsgAsk   = 2 // вопрос, кнопки Да/Нет
)

// Режимы FileDialog.
const (
	FileOpen     = 0
	FileSave     = 1
	FileFolder   = 2
	FileOpenMany = 3 // несколько существующих файлов
)

// Клавиши OnKey (KeyDecode). Буквы — 'A'…'Z', цифры — '0'…'9', пробел — ' ' (коды ASCII),
// по положению клавиши: в русской раскладке Ctrl+S — тоже 'S'. Остальные — ниже; 0 —
// клавиша, о которой nui не сообщает (модификатор, знак препинания…).
const (
	KeyEnter = 0x100 + iota
	KeyEscape
	KeyTab
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyF1 // F1…F12 — KeyF1+0…KeyF1+11
)

// Модификаторы OnKey — биты.
const (
	ModShift = 1
	ModCtrl  = 2
	ModAlt   = 4
)

// Виды событий мыши (MouseDecode); 0 — пропустить.
const (
	MouseDown  = 1
	MouseUp    = 2
	MouseMove  = 3
	MouseWheel = 4
)

// Filter — фильтр файлового диалога: имя и шаблоны вида "*.log".
type Filter struct {
	Name     string
	Patterns []string
}
