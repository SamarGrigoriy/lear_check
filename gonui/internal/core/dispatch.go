package core

// Dispatch — единственная точка входа событий из ОС: её вызывают трамплины драйвера.
func Dispatch(ud, arg uintptr) uintptr {
	id, k := Decode(ud)
	if k == KDo {
		Drain()
		return 0 // idle-источник одноразовый
	}
	r := Lookup(id)
	if r == nil {
		return 0 // гонка с уничтожением — тихо игнорируем
	}
	return r.Event(k, arg)
}
