// form — все виджеты MVP на одной форме: Grid «подпись: поле», журнал в рамке,
// строка состояния, кнопка по умолчанию (Enter в поле нажимает «Отправить»).
package main

import "nui"

func main() {
	err := nui.Run(func() {
		status := nui.NewLabel("Заполните анкету")
		status.SetBold(true)

		name := nui.NewEntry()
		name.SetPlaceholder("Иван Петров")
		pass := nui.NewEntry()
		pass.SetPassword(true)
		city := nui.NewCombo("Москва", "Санкт-Петербург", "Новосибирск")
		city.SetSelected(0)
		speed := nui.NewComboEntry("9600", "115200", "921600")
		speed.SetSelected(1)
		volume := nui.NewSlider(0, 100)
		volume.SetValue(40)
		level := nui.NewProgress()
		level.SetValue(0.4)
		level.SetText("40 %")
		agree := nui.NewCheck("Согласен с условиями")

		journal := nui.NewTextView()
		journal.SetMonospace(true)
		journal.SetReadOnly(true)
		journal.SetAutoScroll(true)
		journal.SetMaxLines(100)
		journal.SetMinSize(-1, 90)
		say := func(s string) { journal.Append(s + "\n") }

		volume.OnChange(func() {
			v := volume.Value()
			level.SetValue(float64(v) / 100)
			level.SetText(itoa(v) + " %")
		})
		agree.OnChange(func() { say("Согласие: " + yesNo(agree.Checked())) })
		city.OnChange(func() { say("Город: " + city.Text()) })

		send := nui.NewButton("Отправить", func() {
			if !agree.Checked() {
				status.SetColor(nui.RGB(200, 0, 0))
				status.SetText("Нужно согласие с условиями")
				return
			}
			status.SetColor(nui.RGB(0, 120, 40))
			status.SetText("Отправлено: " + name.Text())
			say("Отправлено: " + name.Text() + ", " + city.Text() + ", " + speed.Text() + " бод")
		})
		reset := nui.NewButton("Очистить", func() {
			name.SetText("")
			pass.SetText("")
			journal.Clear()
			status.SetColor(nui.NoColor)
			status.SetText("Заполните анкету")
		})

		form := nui.NewGrid().
			Add(0, 0, nui.NewLabel("Имя:")).AddSpan(1, 0, 2, 1, name).
			Add(0, 1, nui.NewLabel("Пароль:")).AddSpan(1, 1, 2, 1, pass).
			Add(0, 2, nui.NewLabel("Город:")).AddSpan(1, 2, 2, 1, city).
			Add(0, 3, nui.NewLabel("Скорость:")).Add(1, 3, speed).Add(2, 3, nui.NewLabel("бод")).
			Add(0, 4, nui.NewLabel("Громкость:")).AddSpan(1, 4, 2, 1, volume).
			AddSpan(1, 5, 2, 1, level).
			AddSpan(1, 6, 2, 1, agree).
			ExpandCol(1)

		w := nui.NewWindow("nui · форма", nui.VBox(
			form,
			nui.NewSeparator(),
			nui.Expand(nui.NewFrame("Журнал", journal)),
			status,
			nui.HBox(nui.Stretch(), reset, send),
		))
		w.SetSize(440, 480)
		w.SetDefault(send)
		w.Show()
		say("Форма открыта")
		selftest()
	})
	if err != nil {
		println(err.Error())
	}
}

func yesNo(b bool) string {
	if b {
		return "да"
	}
	return "нет"
}

// itoa — без strconv: размер приложения тоже в бюджете (docs/size.md).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for ; n > 0; n /= 10 {
		i--
		b[i] = byte('0' + n%10)
	}
	return string(b[i:])
}
