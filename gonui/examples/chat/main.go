// chat — готовый виджет чата (nui.Chat). Вместо ИИ — эхо-бот, который печатает ответ
// по буквам, как модель, отвечающая частями. Настоящий ответ пишут в поток из горутины:
//
//	st := chat.Stream("ИИ")
//	go func() { defer st.Close(); io.Copy(st, ответ) }()
//
// Здесь горутин нет: буквы подаёт таймер — так пример собирается и малой сборкой TinyGo.
package main

import "nui"

func main() {
	err := nui.Run(func() {
		chat := nui.NewChat()
		chat.SetPlaceholder("Спросите что-нибудь… (Shift+Enter — новая строка)")
		chat.Add("Бот", "Здравствуйте! Я эхо-бот: отвечаю по буквам, как ИИ, который печатает ответ.")

		var typing *nui.Timer
		chat.OnSend(func(q string) {
			reply := []rune("Вы спросили: «" + q + "». Настоящий ответ пришёл бы от модели по сети — " +
				"тогда текст пишут в chat.Stream из горутины.")
			st := chat.Stream("Бот")
			i := 0
			typing = nui.Every(25, func() {
				if i == len(reply) {
					typing.Stop()
					st.Close()
					return
				}
				st.WriteString(string(reply[i]))
				i++
			})
		})
		chat.OnStop(func() { typing.Stop() }) // сообщение «Стоп» уже закончил сам

		copyBtn := nui.NewButton("Скопировать ответ", func() { nui.SetClipboard(chat.LastMessage()) })
		copyBtn.SetTooltip("Последнее сообщение — в буфер обмена")

		w := nui.NewWindow("nui · чат", nui.VBox(nui.HBox(nui.Stretch(), copyBtn), nui.Expand(chat)))
		w.SetSize(460, 420)
		w.Show()
		chat.Focus()
		selftest()
	})
	if err != nil {
		println(err.Error())
	}
}
