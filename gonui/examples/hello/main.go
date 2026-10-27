package main

import "nui"

func main() {
	err := nui.Run(func() {
		selftest()
		nui.NewWindow("nui", nui.VBox(nui.NewLabel("Привет, мир!"), nui.NewButton("Закрыть", nui.Quit))).Show()
	})
	if err != nil {
		println(err.Error())
	}
}
