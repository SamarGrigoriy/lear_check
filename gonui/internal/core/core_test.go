package core

import (
	"sync"
	"testing"
)

// reset возвращает ядро в начальное состояние: тесты не должны зависеть друг от друга.
func reset() {
	objs = []Receiver{nil}
	free = nil
	waker = nil
	mu.Lock()
	queue, pending, closed = nil, false, false
	mu.Unlock()
}

type rec struct {
	kinds []Kind
	args  []uintptr
	ret   uintptr
}

func (r *rec) Event(k Kind, arg uintptr) uintptr {
	r.kinds = append(r.kinds, k)
	r.args = append(r.args, arg)
	return r.ret
}

func TestRegisterReusesIDsAndNeverReturnsZero(t *testing.T) {
	reset()
	a, b := &rec{}, &rec{}
	ida, idb := Register(a), Register(b)
	if ida == 0 || idb == 0 || ida == idb {
		t.Fatalf("id: %d, %d", ida, idb)
	}
	Release(ida)
	if Lookup(ida) != nil {
		t.Fatal("после Release объект всё ещё в реестре")
	}
	c := &rec{}
	if idc := Register(c); idc != ida {
		t.Fatalf("освобождённый id не переиспользован: %d вместо %d", idc, ida)
	}
	if Lookup(idb) != b {
		t.Fatal("Lookup вернул чужой объект")
	}
}

func TestReleaseIsIdempotentAndIgnoresBadIDs(t *testing.T) {
	reset()
	id := Register(&rec{})
	Release(id)
	Release(id) // повторный — ничего не делает, id в стеке свободных один раз
	Release(0)
	Release(1000)
	if len(free) != 1 {
		t.Fatalf("free = %v", free)
	}
	a, b := Register(&rec{}), Register(&rec{})
	if a == b {
		t.Fatal("один id выдан дважды")
	}
}

func TestLookupOutOfRange(t *testing.T) {
	reset()
	if Lookup(0) != nil || Lookup(5) != nil || Lookup(^uint32(0)) != nil {
		t.Fatal("Lookup вне диапазона должен давать nil")
	}
}

func TestUDDecodeRoundTrip(t *testing.T) {
	ids := []uint32{0, 1, 2, 31, 1<<27 - 1}
	kinds := []Kind{EvClick, EvChange, EvEnter, EvClose, EvDestroy, KTimer, KWatch, KDo, 31}
	for _, id := range ids {
		for _, k := range kinds {
			gid, gk := Decode(UD(id, k))
			if gid != id || gk != k {
				t.Fatalf("UD(%d, %d) → (%d, %d)", id, k, gid, gk)
			}
		}
	}
}

func TestDispatch(t *testing.T) {
	reset()
	r := &rec{ret: 1}
	id := Register(r)
	if got := Dispatch(UD(id, EvClose), 42); got != 1 {
		t.Fatalf("ответ %d, ожидался 1", got)
	}
	if len(r.kinds) != 1 || r.kinds[0] != EvClose || r.args[0] != 42 {
		t.Fatalf("событие дошло не так: %v %v", r.kinds, r.args)
	}
	if got := Dispatch(UD(id+100, EvClick), 0); got != 0 {
		t.Fatal("неизвестный id должен давать 0")
	}
	Release(id)
	if got := Dispatch(UD(id, EvClick), 0); got != 0 || len(r.kinds) != 1 {
		t.Fatal("освобождённый объект получил событие")
	}
}

func TestDispatchKDoDrains(t *testing.T) {
	reset()
	ran := 0
	Post(func() { ran++ })
	if got := Dispatch(UD(0, KDo), 0); got != 0 || ran != 1 {
		t.Fatalf("Dispatch(KDo): ответ %d, выполнено %d", got, ran)
	}
}

func TestPostWakesOncePerBatch(t *testing.T) {
	reset()
	got := []bool{Post(func() {}), Post(func() {}), Post(func() {})}
	if !got[0] || got[1] || got[2] {
		t.Fatalf("Post ×3 = %v, ожидалось true, false, false", got)
	}
	// Drain пачки из трёх заранее будит цикл (в очереди оставались функции): idle уже
	// запланирован, и Post будить не должен. Когда этот idle отработает — снова должен.
	wakes := 0
	SetWaker(func() { wakes++ })
	Drain()
	if wakes != 1 || Post(func() {}) {
		t.Fatalf("после Drain пачки: разбужено %d раз, Post будит повторно", wakes)
	}
	Drain() // разбуженный idle
	Drain()
	if !Post(func() {}) {
		t.Fatal("когда idle отработал и очередь пуста, первый Post снова должен будить цикл")
	}
}

func TestPostInsideDrainRunsInNextBatch(t *testing.T) {
	reset()
	wakes := 0
	SetWaker(func() { wakes++ })
	var order []int
	var woke bool
	Post(func() {
		order = append(order, 1)
		woke = Post(func() { order = append(order, 3) })
	})
	Post(func() { order = append(order, 2) })
	Drain()
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Fatalf("первая пачка: %v", order)
	}
	// Цикл должен быть разбужен для следующей пачки — либо самим Post, либо заранее Drain.
	if !woke && wakes == 0 {
		t.Fatal("после Post внутри Drain цикл никто не разбудил")
	}
	Drain()
	if len(order) != 3 || order[2] != 3 {
		t.Fatalf("вторая пачка: %v", order)
	}
}

// Функция из очереди открыла «модальный диалог» (вложенный цикл): остальные функции
// той же пачки выполняет вложенный Drain, не дожидаясь возврата. Порядок сохраняется.
func TestDrainNestedLoop(t *testing.T) {
	reset()
	nested := 0
	SetWaker(func() { nested++ })
	var order []int
	Post(func() {
		order = append(order, 1)
		if nested != 1 {
			t.Fatal("перед функцией, которая может заблокироваться, цикл должен быть разбужен")
		}
		Drain() // вложенный цикл: разбуженный idle выполняет остаток очереди
		order = append(order, 99)
	})
	Post(func() { order = append(order, 2) })
	Post(func() { order = append(order, 3) })
	Drain()
	want := []int{1, 2, 3, 99}
	if len(order) != len(want) {
		t.Fatalf("порядок %v, ожидался %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("порядок %v, ожидался %v", order, want)
		}
	}
}

func TestPostAfterClose(t *testing.T) {
	reset()
	Post(func() { t.Fatal("функция из очереди до Close выполнилась") })
	Close()
	if Post(func() { t.Fatal("функция после Close выполнилась") }) {
		t.Fatal("Post после Close должен давать false")
	}
	Drain()
}

// 100 горутин × 1000 Post: без гонок (-race), порядок внутри каждой горутины сохранён.
func TestPostConcurrentOrder(t *testing.T) {
	reset()
	const G, N = 100, 1000
	last := make([]int, G)
	for i := range last {
		last[i] = -1
	}
	var wg sync.WaitGroup
	for g := 0; g < G; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < N; i++ {
				i := i
				Post(func() {
					if i != last[g]+1 {
						t.Errorf("горутина %d: %d после %d", g, i, last[g])
					}
					last[g] = i
				})
			}
		}(g)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	for drained := false; !drained; {
		select {
		case <-done:
			drained = true
		default:
		}
		Drain() // «UI-поток»: только эта горутина выполняет функции
	}
	Drain()
	for g, v := range last {
		if v != N-1 {
			t.Fatalf("горутина %d: выполнено до %d из %d", g, v, N-1)
		}
	}
}

func TestAtoi(t *testing.T) {
	for s, want := range map[string]int{"0": 0, "42": 42, " -17 ": -17, "+8080": 8080, "\t65535": 65535,
		"2147483647": 2147483647, "-2147483648": -2147483648} {
		if got, ok := Atoi(s); !ok || got != want {
			t.Errorf("Atoi(%q) = %d, %v; ожидалось %d", s, got, ok, want)
		}
	}
	for _, s := range []string{"", " ", "-", "+", "1a", "4 2", "--1", "2147483648", "-2147483649", "12345678901"} {
		if got, ok := Atoi(s); ok {
			t.Errorf("Atoi(%q) = %d — не число", s, got)
		}
	}
}

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 7: "7", 42: "42", -5: "-5", 1234567890: "1234567890", -9223372036854775808: "-9223372036854775808"}
	for n, want := range cases {
		if got := Itoa(n); got != want {
			t.Errorf("Itoa(%d) = %q, ожидалось %q", n, got, want)
		}
	}
}

func TestFixUTF8(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"Привет 🙂", "Привет 🙂"},
		{"a\xffb", "a�b"},
		{"\xc3", "�"},
		{"ok\xe2\x82", "ok��"},
		{"�", "�"}, // настоящий U+FFFD не трогаем
	}
	for _, c := range cases {
		if got := FixUTF8(c.in); got != c.want {
			t.Errorf("FixUTF8(%q) = %q, ожидалось %q", c.in, got, c.want)
		}
	}
}

func TestDir(t *testing.T) {
	cases := map[string]string{"/a/b": "/a", "/a": "/", "a": ".", "a/b/c.log": "a/b", "/": "/", "": "."}
	for in, want := range cases {
		if got := Dir(in); got != want {
			t.Errorf("Dir(%q) = %q, ожидалось %q", in, got, want)
		}
	}
}

func TestFit(t *testing.T) {
	cases := []struct{ w, h, maxW, maxH, wantW, wantH int }{
		{4000, 3000, 0, 0, 4000, 3000},   // без ограничений
		{4000, 3000, 400, 0, 400, 300},   // по ширине
		{4000, 3000, 0, 300, 400, 300},   // по высоте
		{4000, 3000, 400, 200, 267, 200}, // высота строже ширины; округление
		{4000, 3000, 200, 400, 200, 150}, // ширина строже высоты
		{100, 50, 400, 300, 100, 50},     // меньше рамки — не увеличивается
		{1000, 1, 10, 0, 10, 1},          // не меньше 1 пикселя
		{0, 10, 100, 100, 0, 0},          // пустая
		{10, -1, 100, 100, 0, 0},
	}
	for _, c := range cases {
		if w, h := Fit(c.w, c.h, c.maxW, c.maxH); w != c.wantW || h != c.wantH {
			t.Errorf("Fit(%d, %d, %d, %d) = %d×%d, ожидалось %d×%d", c.w, c.h, c.maxW, c.maxH, w, h, c.wantW, c.wantH)
		}
	}
}

// exifJPEG — начало JPEG: SOI, APP1 с TIFF-заголовком tiff, затем SOS.
func exifJPEG(app1 []byte) []byte {
	n := len(app1) + 2
	b := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 4, 1, 2} // SOI и APP0 (пропускается)
	b = append(b, 0xFF, 0xE1, byte(n>>8), byte(n))
	b = append(b, app1...)
	return append(b, 0xFF, 0xDA, 0, 2)
}

func TestOrientation(t *testing.T) {
	le := func(o byte) []byte { // IFD0: другой тег, потом Orientation
		return []byte("Exif\x00\x00II\x2a\x00\x08\x00\x00\x00\x02\x00" +
			"\x0f\x01\x02\x00\x01\x00\x00\x00\x00\x00\x00\x00" +
			"\x12\x01\x03\x00\x01\x00\x00\x00" + string([]byte{o, 0}) + "\x00\x00\x00\x00\x00\x00")
	}
	be := []byte("Exif\x00\x00MM\x00\x2a\x00\x00\x00\x08\x00\x01" +
		"\x01\x12\x00\x03\x00\x00\x00\x01\x00\x08\x00\x00\x00\x00\x00\x00")
	badType := le(6)
	badType[6+8+2+12+2] = 4 // LONG вместо SHORT
	cases := []struct {
		name string
		b    []byte
		want int
	}{
		{"II, 6", exifJPEG(le(6)), 6},
		{"MM, 8", exifJPEG(be), 8},
		{"значение вне 1…8", exifJPEG(le(9)), 1},
		{"тип не SHORT", exifJPEG(badType), 1},
		{"нет тега", exifJPEG(append(append([]byte{}, le(6)[:6+8]...), "\x01\x00\x0f\x01\x02\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"...)), 1},
		{"IFD обрезан", exifJPEG(le(6)[:6+8+2+12]), 1},
		{"обрезанная запись", exifJPEG(le(6)[:6+8+2+12+5]), 1},
		{"смещение за концом", exifJPEG([]byte("Exif\x00\x00II\x2a\x00\xff\x00\x00\x00\x00\x00")), 1},
		{"не TIFF", exifJPEG([]byte("Exif\x00\x00XX\x2a\x00\x08\x00\x00\x00\x00\x00")), 1},
		{"не Exif", exifJPEG([]byte("XMP\x00 какой-то текст")), 1},
		{"заполняющие FF", append([]byte{0xFF, 0xD8, 0xFF}, exifJPEG(le(3))[2:]...), 3},
		{"сегмент длиннее файла", []byte{0xFF, 0xD8, 0xFF, 0xE1, 0x10, 0}, 1},
		{"не JPEG", []byte("\x89PNG\r\n"), 1},
		{"пусто", nil, 1},
	}
	for _, c := range cases {
		if got := Orientation(c.b); got != c.want {
			t.Errorf("%s: Orientation = %d, ожидалось %d", c.name, got, c.want)
		}
	}
}

func TestDebugHooksAreHarmlessWithoutUIThread(t *testing.T) {
	MarkUIThread(nil)
	CheckThread() // без зарегистрированной проверки — ничего не делает
	if Debug {
		MarkUIThread(func() bool { return false })
		defer MarkUIThread(nil)
		defer func() {
			if recover() == nil {
				t.Fatal("в отладочной сборке вызов не из UI-потока должен паниковать")
			}
		}()
		CheckThread()
	}
}

func TestParseJSON(t *testing.T) {
	v, err := ParseJSON([]byte(` { "messages" : [ {"role":"user","content":"Привет\n\"мир\" \u0041\ud83d\ude42\/"},
		{"n": -1.5e+3, "t": true, "f": false, "z": null, "e": [], "o": {}} ] } `))
	if err != nil {
		t.Fatal(err)
	}
	m, ok := v.Get("messages")
	if !ok || m.Kind != 'a' || len(m.Arr) != 2 {
		t.Fatalf("messages: %+v", m)
	}
	if c, _ := m.Arr[0].Get("content"); c.Str != "Привет\n\"мир\" A🙂/" {
		t.Errorf("content = %q", c.Str)
	}
	o := m.Arr[1]
	for k, want := range map[string]JSON{"n": {Kind: 'n', Str: "-1.5e+3"}, "t": {Kind: 'b', Str: "true"},
		"f": {Kind: 'b', Str: "false"}, "z": {Kind: 'z'}} {
		if got, _ := o.Get(k); got.Kind != want.Kind || got.Str != want.Str {
			t.Errorf("%s = %+v, ожидалось %+v", k, got, want)
		}
	}
	if e, _ := o.Get("e"); e.Kind != 'a' || len(e.Arr) != 0 {
		t.Errorf("пустой массив: %+v", e)
	}
	if e, _ := o.Get("o"); e.Kind != 'o' || len(e.Keys) != 0 {
		t.Errorf("пустой объект: %+v", e)
	}
	if _, ok := o.Get("нет"); ok {
		t.Error("Get несуществующего поля")
	}
	// Экранирования, одинокие половины суррогатных пар, числа.
	for in, want := range map[string]string{
		`"\b\f\r\t\\"`:   "\b\f\r\t\\",
		`"\ud83d\ude42"`: "🙂",
		`"\ud83d"`:       "\uFFFD",
		`"\ud83dx"`:      "\uFFFDx",
		`"\ud83d\u0041"`: "\uFFFDA",
		`"\ud83d\uZZZZ"`: "",
		`"\uDE42"`:       "\uFFFD",
		`"\u00e9\u00C9"`: "éÉ",
		"\"\xff\"":       "\uFFFD",
	} {
		v, err := ParseJSON([]byte(in))
		if want == "" {
			if err == nil {
				t.Errorf("%s: ожидалась ошибка", in)
			}
			continue
		}
		if err != nil || v.Str != want {
			t.Errorf("%s = %q, %v; ожидалось %q", in, v.Str, err, want)
		}
	}
	for _, n := range []string{"0", "-0", "12", "0.5", "1e5", "1E-2", "-3.25e+10"} {
		if v, err := ParseJSON([]byte(n)); err != nil || v.Kind != 'n' || v.Str != n {
			t.Errorf("число %s: %+v, %v", n, v, err)
		}
	}
	deep := make([]byte, 0, 2*(jsonMaxDepth+2))
	for i := 0; i <= jsonMaxDepth+1; i++ {
		deep = append(deep, '[')
	}
	for i := 0; i <= jsonMaxDepth+1; i++ {
		deep = append(deep, ']')
	}
	bad := []string{"", " ", "[1,", "[1 2]", "[", `{"a"`, `{"a":1`, `{"a":1 "b":2}`, `{1:2}`, `{"a" 1}`,
		`{"a":}`, `[}`, "[1", `{"a`, "tru", "nul", "x", `"abc`, `"a\`, `"\x"`, `"\u12"`, "\"a\x01\"", "-", "01x", "1.",
		"1e", "1e+", "[1]x", `{"a":[`, string(deep), `["a",`}
	for _, in := range bad {
		if _, err := ParseJSON([]byte(in)); err == nil {
			t.Errorf("%q: ожидалась ошибка", in)
		}
	}
	if _, err := ParseJSON([]byte("[1,")); err == nil || err.Error() != "байт 3: неожиданный конец" {
		t.Errorf("текст ошибки: %v", err)
	}
}

func TestQuoteJSON(t *testing.T) {
	s := "Привет \"мир\"\\\n\r\t\x01 🙂"
	q := QuoteJSON(s)
	if q != `"Привет \"мир\"\\\n\r\t\u0001 🙂"` {
		t.Errorf("QuoteJSON = %s", q)
	}
	if v, err := ParseJSON([]byte(q)); err != nil || v.Str != s {
		t.Errorf("туда и обратно: %q, %v", v.Str, err)
	}
}
