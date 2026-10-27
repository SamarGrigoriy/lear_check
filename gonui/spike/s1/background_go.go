//go:build !(tinygo && scheduler.none)

package main

import "time"

// background выполняет fn в отдельной горутине через ms миллисекунд:
// так сценарий проверяет Do и запись в pipe из другого потока.
func background(ms int, fn func()) {
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		fn()
	}()
}
