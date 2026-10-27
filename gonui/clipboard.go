package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// SetClipboard кладёт текст в буфер обмена (Ctrl+V в любой программе).
func SetClipboard(text string) {
	core.CheckThread()
	drv.ClipboardSet(text)
}

// Clipboard — текст из буфера обмена; "" — пусто или там не текст (например, картинка).
// Синхронно: пока владелец буфера отвечает, крутится вложенный цикл событий (§0).
func Clipboard() string {
	core.CheckThread()
	return drv.ClipboardText()
}
