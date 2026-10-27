package main

import "sync"

// Мини-ядро по docs/runtime.md: реестр, user_data = id<<5 | kind, диспетчер, очередь Do.
// Чистый Go, без cgo.

type Kind uint8

const (
	EvClick   Kind = 1
	EvClose   Kind = 4 // с ответом: 1 — не закрывать
	EvDestroy Kind = 5
	KTimer    Kind = 16
	KWatch    Kind = 17
	KDo       Kind = 18
)

type Receiver interface {
	Event(k Kind, arg uintptr) uintptr
}

var (
	objs = []Receiver{nil} // id 0 — «нет объекта»
	free []uint32
)

func register(r Receiver) uint32 {
	if n := len(free); n > 0 {
		id := free[n-1]
		free = free[:n-1]
		objs[id] = r
		return id
	}
	objs = append(objs, r)
	return uint32(len(objs) - 1)
}

func release(id uint32) {
	objs[id] = nil
	free = append(free, id)
}

func UD(id uint32, k Kind) uintptr { return uintptr(id)<<5 | uintptr(k) }

func dispatch(ud, arg uintptr) uintptr {
	id, k := uint32(ud>>5), Kind(ud&31)
	if k == KDo {
		drain()
		return 0 // idle-источник одноразовый
	}
	if int(id) >= len(objs) || objs[id] == nil {
		return 0 // гонка с уничтожением
	}
	return objs[id].Event(k, arg)
}

var (
	mu      sync.Mutex
	queue   []func()
	pending bool // idle-источник уже запланирован
)

// post можно вызывать из любого потока. true — нужно разбудить цикл.
func post(fn func()) bool {
	mu.Lock()
	defer mu.Unlock()
	queue = append(queue, fn)
	if pending {
		return false
	}
	pending = true
	return true
}

// drain выполняет пачку, накопленную к моменту вызова. Только в UI-потоке.
func drain() {
	mu.Lock()
	q := queue
	queue, pending = nil, false
	mu.Unlock()
	for _, fn := range q {
		fn()
	}
}
