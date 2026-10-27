package core

import "sync"

// Очередь Do: единственное, что можно трогать из любого потока.
var (
	mu      sync.Mutex
	queue   []func()
	pending bool // idle-источник уже запланирован
	closed  bool // Run завершился
)

// waker будит цикл событий (drv.Wake). Задаёт Run до setup; читается только в UI-потоке.
var waker func()

// SetWaker задаёт, как Drain будит цикл событий.
func SetWaker(f func()) { waker = f }

// Post ставит fn в очередь. true — очередь была пуста и цикл событий нужно разбудить
// (drv.Wake). Будим один раз на пачку, а не на каждый вызов.
func Post(fn func()) bool {
	mu.Lock()
	defer mu.Unlock()
	if closed {
		return false
	}
	queue = append(queue, fn)
	if pending {
		return false
	}
	pending = true
	return true
}

// Drain выполняет функции, накопленные к моменту вызова, по одной. Только в UI-потоке.
//
// Функция может открыть модальный диалог — тогда она не вернётся, пока диалог открыт,
// а внутри крутится вложенный цикл. Чтобы остальные функции очереди не ждали закрытия
// диалога, перед вызовом каждой Drain будит цикл заранее, если в очереди есть ещё:
// их выполнит вложенный Drain. Каждую функцию забирает ровно один Drain, поэтому порядок
// сохраняется. Функции, добавленные во время Drain, — в следующем: бесконечного цикла нет.
func Drain() {
	mu.Lock()
	n := len(queue)
	pending = false
	mu.Unlock()
	for ; n > 0; n-- {
		mu.Lock()
		if len(queue) == 0 { // остальное выполнил вложенный Drain
			mu.Unlock()
			return
		}
		fn := queue[0]
		queue = queue[1:]
		if len(queue) == 0 {
			queue = nil
		}
		wake := len(queue) > 0 && !pending
		if wake {
			pending = true
		}
		mu.Unlock()
		if wake && waker != nil {
			waker()
		}
		fn()
	}
}

// Close вызывается после завершения Run: дальнейшие Post игнорируются.
func Close() {
	mu.Lock()
	closed = true
	queue = nil
	mu.Unlock()
}
