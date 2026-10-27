//go:build nuidebug

package core

// Debug — отладочная сборка (-tags nuidebug): проверки потока и прочие паники.
const Debug = true

var isUI func() bool

// MarkUIThread запоминает, как проверить UI-поток. Вызывает Run после инициализации.
func MarkUIThread(f func() bool) { isUI = f }

// CheckThread паникует, если вызов пришёл не из UI-потока.
func CheckThread() {
	if isUI != nil && !isUI() {
		panic("nui: вызов не из UI-потока — используйте nui.Do")
	}
}
