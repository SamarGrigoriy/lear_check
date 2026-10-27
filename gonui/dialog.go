package nui

import (
	"nui/internal/core"
	"nui/internal/drv"
)

// Все диалоги модальные и синхронные: функция возвращается, когда диалог закрыт.
// Пока он открыт, внутри крутится вложенный цикл событий — таймеры, Watch и Do
// продолжают работать, и может сработать другой обработчик (api.md, §0).
// parent может быть nil. title — крупный заголовок сообщения, text — пояснение под ним
// (может содержать \n).

// Message — информация, кнопка OK.
func Message(parent *Window, title, text string) {
	drv.MessageBox(parentH(parent), drv.MsgInfo, title, text)
}

// Alert — ошибка или предупреждение, кнопка OK.
func Alert(parent *Window, title, text string) {
	drv.MessageBox(parentH(parent), drv.MsgAlert, title, text)
}

// Ask — вопрос «Да/Нет»: true — «Да»; Esc — false.
func Ask(parent *Window, title, text string) bool {
	return drv.MessageBox(parentH(parent), drv.MsgAsk, title, text)
}

// Filter — фильтр файлового диалога.
type Filter struct {
	Name     string   // "Логи"
	Patterns []string // {"*.log", "*.txt"}
}

// OpenFile — выбор существующего файла. ok == false — пользователь отменил выбор.
func OpenFile(parent *Window, title string, filters ...Filter) (path string, ok bool) {
	return first(fileDialog(parent, drv.FileOpen, title, "", filters))
}

// OpenFiles — выбор одного или нескольких существующих файлов (Ctrl или Shift + щелчок).
// ok == false — пользователь отменил выбор; иначе paths не пуст.
func OpenFiles(parent *Window, title string, filters ...Filter) (paths []string, ok bool) {
	return fileDialog(parent, drv.FileOpenMany, title, "", filters)
}

// SaveFile — выбор имени для сохранения; name — предлагаемое имя. Если файл уже есть,
// диалог сам спрашивает о перезаписи.
func SaveFile(parent *Window, title, name string, filters ...Filter) (path string, ok bool) {
	return first(fileDialog(parent, drv.FileSave, title, name, filters))
}

// ChooseFolder — выбор каталога.
func ChooseFolder(parent *Window, title string) (path string, ok bool) {
	return first(fileDialog(parent, drv.FileFolder, title, "", nil))
}

func fileDialog(parent *Window, mode uint8, title, name string, filters []Filter) ([]string, bool) {
	var fs []drv.Filter
	for _, f := range filters {
		fs = append(fs, drv.Filter{Name: f.Name, Patterns: f.Patterns})
	}
	paths := drv.FileDialog(parentH(parent), mode, title, name, fs)
	return paths, len(paths) > 0
}

func first(paths []string, ok bool) (string, bool) {
	if !ok {
		return "", false
	}
	return paths[0], true
}

// parentH — хендл окна-родителя или Nil; заодно проверка UI-потока.
func parentH(w *Window) drv.H {
	core.CheckThread()
	if w == nil || w.id == 0 {
		return drv.Nil
	}
	return w.h
}
