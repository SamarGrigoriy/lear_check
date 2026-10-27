// Package core — ядро nui на чистом Go, без cgo: реестр объектов, кодирование
// user_data, диспетчер событий и очередь Do (docs/runtime.md).
// Пакет не знает ни о GTK, ни о nui, поэтому тестируется на любой машине.
package core

// Kind — тип события. В user_data он занимает младшие kindBits бит.
type Kind uint8

const (
	EvClick   Kind = 1 // clicked
	EvChange  Kind = 2 // changed, toggled, value-changed
	EvEnter   Kind = 3 // activate (Enter в Entry)
	EvClose   Kind = 4 // delete-event — с ответом: 1 = не закрывать
	EvDestroy Kind = 5 // destroy
	EvKey     Kind = 6 // нажатие клавиши в окне (OnKey) — с ответом: 1 = обработано
	EvDrop    Kind = 7 // на окно перетащили файлы (OnDropFiles); arg — список путей драйвера
	EvDraw    Kind = 8 // Canvas: нарисовать; arg — контекст рисования драйвера
	EvMouse   Kind = 9 // Canvas: событие мыши; arg — событие драйвера; ответ 1 — обработано
	// 10–15 — свободны
	KTimer Kind = 16 // источник GLib: таймер
	KWatch Kind = 17 // источник GLib: fd
	KDo    Kind = 18 // источник GLib: idle для очереди Do (id всегда 0)
)

const kindBits = 5

// UD кодирует id объекта и тип события в user_data — число, а не указатель (ADR-6).
func UD(id uint32, k Kind) uintptr { return uintptr(id)<<kindBits | uintptr(k) }

// Decode — обратное преобразование UD.
func Decode(ud uintptr) (uint32, Kind) {
	return uint32(ud >> kindBits), Kind(ud & (1<<kindBits - 1))
}
