//go:build tinygo && scheduler.none

package main

// background без планировщика горутин (TinyGo -scheduler=none): fn выполняется
// из таймера в UI-потоке. Do при этом остаётся асинхронным (roadmap.md, S2-none).
func background(ms int, fn func()) { After(ms, fn) }
