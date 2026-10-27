// viewer — просмотр картинок: список файлов слева, картинка справа. Файлы добавляются
// кнопкой «Открыть…» (можно несколько сразу), перетаскиванием на окно или в командной
// строке. Клавиши: Ctrl+O — открыть, Delete — убрать из списка, Ctrl+C — путь
// выбранного файла в буфер обмена, стрелки — по списку.
package main

import (
	"path/filepath"
	"strconv"

	"nui"
)

func main() {
	var files []string
	err := nui.Run(func() {
		var w *nui.Window
		list := nui.NewList("Файл")
		list.SetMinSize(220, -1)
		img := nui.NewImage()
		img.SetFit(480, 360)
		info := nui.NewLabel("Откройте картинки или перетащите их на окно")
		info.SetEllipsis(nui.EllipsisMiddle)

		show := func() {
			i := list.Selected()
			switch {
			case i < 0:
				img.Clear()
				info.SetText("Выберите файл в списке")
			case img.SetFile(files[i]) != nil:
				img.Clear()
				info.SetText("Не картинка: " + files[i])
			default:
				iw, ih := img.ImageSize()
				info.SetText(strconv.Itoa(iw) + "×" + strconv.Itoa(ih) + " — " + files[i])
			}
		}
		setRows := func() {
			list.SetRows(len(files), func(row, _ int) string { return filepath.Base(files[row]) })
		}
		add := func(paths []string) {
			first := len(files)
			files = append(files, paths...)
			setRows()
			list.Select(first)
			show()
		}
		open := func() {
			pics := nui.Filter{Name: "Картинки", Patterns: []string{"*.png", "*.jpg", "*.jpeg", "*.gif", "*.bmp"}}
			if paths, ok := nui.OpenFiles(w, "Открыть картинки", pics); ok {
				add(paths)
			}
		}
		remove := func() {
			i := list.Selected()
			if i < 0 {
				return
			}
			files = append(files[:i], files[i+1:]...)
			setRows()
			list.Select(min(i, len(files)-1))
			show()
		}
		copyPath := func() {
			if i := list.Selected(); i >= 0 {
				nui.SetClipboard(files[i])
				info.SetText("Путь скопирован: " + files[i])
			}
		}

		list.OnSelect(show)
		list.OnActivate(func(row int) { nui.ShowInFolder(files[row]) })
		bar := nui.HBox(nui.NewButton("Открыть…", open), nui.NewButton("Убрать", remove),
			nui.NewButton("Скопировать путь", copyPath))
		w = nui.NewWindow("nui · картинки", nui.VBox(bar, nui.Expand(nui.HBox(list, nui.Expand(img))), info))
		w.OnDropFiles(add)
		w.OnKey(func(k nui.Key, m nui.Mod) bool {
			switch {
			case k == 'O' && m == nui.ModCtrl:
				open()
			case k == 'C' && m == nui.ModCtrl && list.Selected() >= 0:
				copyPath()
			case k == nui.KeyDelete && m == 0:
				remove()
			default:
				return false
			}
			return true
		})
		w.SetSize(760, 480)
		w.Show()
		if args := nui.Args(); len(args) > 1 { // не os.Args: в TinyGo на Windows он портит русские имена
			var paths []string
			for _, a := range args[1:] {
				if p, err := filepath.Abs(a); err == nil {
					paths = append(paths, p)
				}
			}
			add(paths)
		}
		list.Focus()
		selftest()
	})
	if err != nil {
		println(err.Error())
	}
}
