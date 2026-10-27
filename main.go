// study-tracker — десктопный трекер личного обучения по 24-недельному плану.
// Единственная точка входа: разбирает аргументы и запускает ui.Run.
package main

import (
	"flag"
	"os"

	"studytracker/storage"
	"studytracker/ui"
)

func main() {
	dir := flag.String("data", storage.DefaultDataDir(), "каталог с data.json")
	flag.Parse()
	if err := ui.Run(*dir); err != nil {
		os.Stderr.WriteString("трекер: " + err.Error() + "\n")
		os.Exit(1)
	}
}
