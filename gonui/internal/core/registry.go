package core

// Receiver получает события своего объекта. Результат нужен только событиям
// «с ответом» (EvClose) и источникам GLib (1 — продолжать, 0 — удалить).
type Receiver interface {
	Event(k Kind, arg uintptr) uintptr
}

// Реестр трогают только из UI-потока, поэтому мьютекс не нужен.
// Пока объект в реестре, Go-объект достижим и сборщик его не тронет (ADR-6).
var (
	objs = []Receiver{nil} // objs[0] == nil навсегда: id 0 означает «нет объекта»
	free []uint32          // стек освободившихся id
)

// Register кладёт объект в реестр и возвращает его id (никогда не 0).
func Register(r Receiver) uint32 {
	CheckThread()
	if n := len(free); n > 0 {
		id := free[n-1]
		free = free[:n-1]
		objs[id] = r
		return id
	}
	objs = append(objs, r)
	return uint32(len(objs) - 1)
}

// Release освобождает id. Повторный Release и id 0 ничего не делают.
// После Release ОС больше не вызывает колбэков с этим id: виджет уничтожен,
// источник GLib удалён. Поэтому повторно выданный id чужих событий не получит.
func Release(id uint32) {
	CheckThread()
	if id == 0 || int(id) >= len(objs) || objs[id] == nil {
		return
	}
	objs[id] = nil
	free = append(free, id)
}

// Lookup возвращает объект или nil, если id вне диапазона или освобождён.
func Lookup(id uint32) Receiver {
	if int(id) >= len(objs) {
		return nil
	}
	return objs[id]
}
