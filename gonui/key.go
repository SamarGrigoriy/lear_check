package nui

import "nui/internal/drv"

// Key — клавиша (Window.OnKey). Буквы — 'A'…'Z', цифры — '0'…'9' (и на цифровом блоке),
// пробел — ' ': по положению клавиши, независимо от раскладки и Shift. В русской
// раскладке Ctrl+S — это 'S' с ModCtrl, Shift+1 — '1' с ModShift. Остальные клавиши —
// константы ниже; о прочих (знаки препинания, мультимедиа) nui не сообщает.
type Key uint32

const (
	KeyEnter     Key = drv.KeyEnter // и Enter цифрового блока
	KeyEscape    Key = drv.KeyEscape
	KeyTab       Key = drv.KeyTab // Shift+Tab — KeyTab с ModShift
	KeyBackspace Key = drv.KeyBackspace
	KeyDelete    Key = drv.KeyDelete
	KeyInsert    Key = drv.KeyInsert
	KeyHome      Key = drv.KeyHome
	KeyEnd       Key = drv.KeyEnd
	KeyPageUp    Key = drv.KeyPageUp
	KeyPageDown  Key = drv.KeyPageDown
	KeyLeft      Key = drv.KeyLeft
	KeyRight     Key = drv.KeyRight
	KeyUp        Key = drv.KeyUp
	KeyDown      Key = drv.KeyDown
	KeySpace     Key = ' '
	KeyF1        Key = drv.KeyF1 // F2…F12 — KeyF1+1…KeyF1+11
)

// Mod — нажатые модификаторы (OnKey), биты.
type Mod uint8

const (
	ModShift Mod = drv.ModShift
	ModCtrl  Mod = drv.ModCtrl
	ModAlt   Mod = drv.ModAlt
)
