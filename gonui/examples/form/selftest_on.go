//go:build selftest

package main

import "nui"

// selftest: окно показано, цикл крутится, выход с кодом 0 (docs/testing.md, §4).
func selftest() { nui.After(300, nui.Quit) }
