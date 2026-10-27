//go:build !nuidebug

package core

// Debug — отладочная сборка (-tags nuidebug). В обычной сборке проверки —
// пустые функции, которые компилятор выбрасывает.
const Debug = false

func MarkUIThread(func() bool) {}

func CheckThread() {}
