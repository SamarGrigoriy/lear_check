package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// OpenURI открывает "https://…" или "file:///…" программой по умолчанию.
// false — не получилось (нет программы для схемы и т. п.).
func OpenURI(uri string) bool {
	core.CheckThread()
	return drv.OpenURI(drv.Nil, uri)
}

// ShowInFolder открывает файловый менеджер с выделенным файлом: D-Bus
// org.freedesktop.FileManager1 на Linux. Запрос асинхронный — интерфейс не ждёт;
// если файловый менеджер ответил ошибкой, nui сам открывает папку.
// path — абсолютный путь в файловой системе (не URI).
// true — запрос отправлен или папка открыта; false — не удалось ни то, ни другое.
func ShowInFolder(path string) bool {
	core.CheckThread()
	return drv.ShowInFolder(path)
}
