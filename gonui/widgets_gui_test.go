//go:build gui

package nui

// Интеграционные тесты виджетов MVP (фаза 2): каждая строка api.md, §4–§15 —
// Set/Get с кириллицей, правило «программное изменение не вызывает On*» (ADR-8),
// уничтожение. Помощники ui, waitUntil, shot, expectPanic, alloc — в nui_gui_test.go.

import (
	"runtime"
	"testing"
	"time"

	"nui/internal/core"
	"nui/internal/drv"
)

// Строки, которые проходят через каждый текстовый Set/Get.
var textCases = []struct{ in, want string }{
	{"Привет", "Привет"},
	{"эмодзи 🙂", "эмодзи 🙂"},
	{"", ""},
	{"до\x00после", "до"},
	{"a\xffb", "a�b"},
}

func TestEntry(t *testing.T) {
	ui(func() {
		e := NewEntry()
		w := NewWindow("Entry", e)
		defer w.Close()
		for _, c := range textCases {
			e.SetText(c.in)
			if got := e.Text(); got != c.want {
				t.Errorf("SetText(%q) → %q, ожидалось %q", c.in, got, c.want)
			}
		}
		changes, enters := 0, 0
		e.OnChange(func() { changes++ })
		e.SetText("программно")
		if changes != 0 {
			t.Error("SetText вызвал OnChange")
		}
		drv.SimType(e.h, "ввод пользователя")
		if changes != 1 || e.Text() != "ввод пользователя" {
			t.Errorf("SimType: OnChange %d раз, текст %q", changes, e.Text())
		}
		e.OnEnter(func() { enters++ })
		drv.SimEnter(e.h)
		if enters != 1 {
			t.Errorf("Enter: OnEnter %d раз", enters)
		}
		e.SetPlaceholder("подсказка")
		e.SetPassword(true)
		e.SetPassword(false)
		e.SetReadOnly(true)
		e.SetReadOnly(false)
		e.SetWidthChars(30)
	})
}

// Enter в поле без OnEnter нажимает кнопку по умолчанию; с OnEnter — нет (api.md, §8).
// GtkButton «нажимается» не сразу: activate показывает нажатие и шлёт clicked через ~250 мс.
func TestDefaultButton(t *testing.T) {
	var w *Window
	var e *Entry
	var ce *Combo
	clicks := 0
	ui(func() {
		e = NewEntry()
		ce = NewComboEntry("один", "два")
		ok := NewButton("OK", func() { clicks++ })
		w = NewWindow("кнопка по умолчанию", VBox(e, ce, ok))
		w.SetDefault(ok)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { return drv.Visible(w.h) })
	ui(func() { drv.SimEnter(e.h) })
	waitUntil(t, 2000, func() bool { return clicks == 1 }) // Enter в Entry без OnEnter
	ui(func() { drv.SimEnter(comboEntryH(ce)) })
	waitUntil(t, 2000, func() bool { return clicks == 2 }) // Enter в поле ComboEntry
	ui(func() {
		e.OnEnter(func() {})
		drv.SimEnter(e.h)
	})
	time.Sleep(500 * time.Millisecond)
	ui(func() {
		if clicks != 2 {
			t.Error("с OnEnter кнопка по умолчанию не должна срабатывать")
		}
		e.OnEnter(nil)
		drv.SimEnter(e.h)
	})
	waitUntil(t, 2000, func() bool { return clicks == 3 }) // OnEnter(nil) возвращает кнопку по умолчанию
	ui(func() {
		w.SetDefault(nil)
		w.Close()
	})
}

func TestCheck(t *testing.T) {
	ui(func() {
		c := NewCheck("Согласен")
		w := NewWindow("Check", c)
		defer w.Close()
		for _, tc := range textCases {
			c.SetText(tc.in)
			if got := c.Text(); got != tc.want {
				t.Errorf("SetText(%q) → %q", tc.in, got)
			}
		}
		n := 0
		c.OnChange(func() { n++ })
		c.SetChecked(true)
		if !c.Checked() || n != 0 {
			t.Errorf("SetChecked: состояние %v, OnChange %d раз", c.Checked(), n)
		}
		drv.SimToggle(c.h)
		if c.Checked() || n != 1 {
			t.Errorf("SimToggle: состояние %v, OnChange %d раз", c.Checked(), n)
		}
	})
}

func TestCombo(t *testing.T) {
	ui(func() {
		c := NewCombo("один", "два", "три")
		ce := NewComboEntry("9600", "115200")
		w := NewWindow("Combo", VBox(c, ce))
		defer w.Close()
		if c.Selected() != -1 || c.Text() != "" || c.Len() != 3 || c.Item(1) != "два" || c.Item(7) != "" {
			t.Errorf("начало: выбор %d, текст %q, длина %d", c.Selected(), c.Text(), c.Len())
		}
		n := 0
		c.OnChange(func() { n++ })
		c.SetSelected(2)
		c.Add("четыре")
		c.SetText("один")
		if c.Selected() != 0 || c.Text() != "один" || c.Len() != 4 || n != 0 {
			t.Errorf("программно: выбор %d, текст %q, длина %d, OnChange %d", c.Selected(), c.Text(), c.Len(), n)
		}
		c.SetText("нет такого")
		c.SetSelected(99)
		if c.Selected() != -1 {
			t.Error("несуществующий элемент должен давать -1")
		}
		drv.SimSelect(c.h, 1)
		if n != 1 || c.Text() != "два" {
			t.Errorf("SimSelect: OnChange %d, текст %q", n, c.Text())
		}
		c.Clear()
		if c.Selected() != -1 || c.Len() != 0 || n != 1 {
			t.Errorf("Clear: выбор %d, длина %d, OnChange %d", c.Selected(), c.Len(), n)
		}

		// ComboEntry: текст поля, свободный ввод.
		m := 0
		ce.OnChange(func() { m++ })
		ce.SetSelected(1)
		if ce.Text() != "115200" || m != 0 {
			t.Errorf("ComboEntry SetSelected: текст %q, OnChange %d", ce.Text(), m)
		}
		for _, tc := range textCases {
			ce.SetText(tc.in)
			if got := ce.Text(); got != tc.want {
				t.Errorf("ComboEntry SetText(%q) → %q", tc.in, got)
			}
		}
		if m != 0 {
			t.Error("ComboEntry SetText вызвал OnChange")
		}
		drv.SimComboType(ce.h, "74880")
		if m == 0 || ce.Text() != "74880" || ce.Selected() != -1 {
			t.Errorf("ввод в ComboEntry: OnChange %d, текст %q, выбор %d", m, ce.Text(), ce.Selected())
		}
	})
}

func TestSlider(t *testing.T) {
	ui(func() {
		expectPanic(t, "nui: NewSlider: min должен быть меньше max", func() { NewSlider(5, 5) })
		s := NewSlider(0, 100)
		w := NewWindow("Slider", s)
		defer w.Close()
		n := 0
		s.OnChange(func() { n++ })
		s.SetValue(42)
		if s.Value() != 42 || n != 0 {
			t.Errorf("SetValue: %d, OnChange %d", s.Value(), n)
		}
		s.SetValue(1000)
		if s.Value() != 100 {
			t.Errorf("SetValue вне диапазона: %d, ожидалось 100", s.Value())
		}
		s.SetValue(-5)
		if s.Value() != 0 {
			t.Errorf("SetValue вне диапазона: %d, ожидалось 0", s.Value())
		}
		drv.SimSlide(s.h, 7)
		if s.Value() != 7 || n != 1 {
			t.Errorf("SimSlide: %d, OnChange %d", s.Value(), n)
		}
	})
}

func TestProgress(t *testing.T) {
	ui(func() {
		p := NewProgress()
		w := NewWindow("Progress", p)
		defer w.Close()
		p.SetValue(0.25)
		if v := p.Value(); v < 0.249 || v > 0.251 {
			t.Errorf("Value = %v", v)
		}
		p.SetValue(3)
		if p.Value() != 1 {
			t.Errorf("SetValue(3) → %v", p.Value())
		}
		p.SetValue(-1)
		if p.Value() != 0 {
			t.Errorf("SetValue(-1) → %v", p.Value())
		}
		p.Pulse()
		p.Pulse()
		p.SetText("Загрузка… 🙂")
		p.SetText("")
		p.SetValue(0.5)
		if v := p.Value(); v < 0.499 || v > 0.501 {
			t.Errorf("SetValue после Pulse: %v", v)
		}
	})
}

func TestTextView(t *testing.T) {
	ui(func() {
		tv := NewTextView()
		w := NewWindow("TextView", tv)
		defer w.Close()
		for _, c := range textCases {
			tv.SetText(c.in)
			if got := tv.Text(); got != c.want {
				t.Errorf("SetText(%q) → %q", c.in, got)
			}
		}
		n := 0
		tv.OnChange(func() { n++ })
		tv.SetText("строка\n")
		tv.Append("ещё")
		tv.Clear()
		if n != 0 || tv.Text() != "" {
			t.Errorf("программные SetText/Append/Clear: OnChange %d, текст %q", n, tv.Text())
		}
		// SetMaxLines(3): завершающий \n не считается лишней строкой.
		tv.SetMaxLines(3)
		for _, s := range []string{"a\n", "b\n", "c\n", "d\n", "e\n"} {
			tv.Append(s)
		}
		if got := tv.Text(); got != "c\nd\ne\n" {
			t.Errorf("MaxLines 3 с \\n: %q", got)
		}
		tv.SetText("")
		for _, s := range []string{"a", "\nb", "\nc", "\nd"} {
			tv.Append(s)
		}
		if got := tv.Text(); got != "b\nc\nd" {
			t.Errorf("MaxLines 3 без завершающего \\n: %q", got)
		}
		tv.SetMaxLines(0)
		tv.SetAutoScroll(true)
		tv.Append("\nконец")
		tv.SetReadOnly(true)
		tv.SetReadOnly(false)
		tv.SetMonospace(true)
		tv.SetWrap(true)
		tv.Focus()
		// Пользователь правит пустой буфер: один changed; непустой — два (удаление и вставка).
		tv.Clear()
		drv.SimEdit(tv.view, "правка")
		if n != 1 {
			t.Errorf("правка пустого буфера: OnChange %d, ожидался 1", n)
		}
		drv.SimEdit(tv.view, "ещё правка")
		want := 3 // GTK: удаление и вставка
		if runtime.GOOS == "windows" {
			want = 2 // Windows: один EN_CHANGE на замену текста
		}
		if n != want {
			t.Errorf("правка непустого буфера: OnChange %d, ожидалось %d", n, want)
		}
	})
}

func TestFrameScrollSeparator(t *testing.T) {
	ui(func() {
		old, nw := NewLabel("старый"), NewLabel("новый")
		f := NewFrame("Рамка 🙂", old)
		s := NewScroll(NewLabel("в прокрутке"))
		w := NewWindow("Frame", VBox(f, NewSeparator(), HBox(s, NewVSeparator(), NewLabel("справа"))))
		defer w.Close()
		f.SetChild(nw)
		if old.Alive() || !nw.Alive() {
			t.Error("Frame.SetChild должен уничтожить старого ребёнка")
		}
		f.SetTitle("")
		f.SetTitle("снова")
		f.SetChild(nil)
		if nw.Alive() {
			t.Error("Frame.SetChild(nil) должен уничтожить ребёнка")
		}
	})
}

func TestLabelStyle(t *testing.T) {
	ui(func() {
		l := NewLabel("Строка состояния")
		w := NewWindow("Label", l)
		defer w.Close()
		l.SetBold(true)
		l.SetColor(RGB(200, 0, 0))
		l.SetColor(NoColor)
		l.SetBold(false)
		l.SetSelectable(true)
		l.SetWrap(true)
		for _, a := range []Align{AlignStart, AlignCenter, AlignEnd, AlignFill} {
			l.SetTextAlign(a)
		}
		for _, e := range []Ellipsis{EllipsisStart, EllipsisMiddle, EllipsisEnd, EllipsisNone} {
			l.SetEllipsis(e)
		}
		if RGB(0x12, 0x34, 0x56) != 0x123456 {
			t.Error("RGB")
		}
	})
}

// Без многоточия метка не даёт окну стать уже текста; с многоточием — даёт (api.md, §6).
func TestLabelEllipsisWidth(t *testing.T) {
	long := "/очень/длинный/путь/к/файлу/журнала/esp32_logs/esp32_ttyUSB0_2026-10-08_12-00-00.log"
	widths := map[Ellipsis]int{}
	for _, e := range []Ellipsis{EllipsisNone, EllipsisMiddle} {
		var w *Window
		ui(func() {
			l := NewLabel(long)
			l.SetEllipsis(e)
			w = NewWindow("многоточие", l)
			w.SetSize(200, -1)
			w.Show()
		})
		waitUntil(t, 2000, func() bool { _, _, ww, _ := drv.Allocation(w.h); return ww > 1 })
		if e == EllipsisMiddle {
			shot(t, "label-ellipsis")
		}
		ui(func() {
			_, _, ww, _ := drv.Allocation(w.h)
			widths[e] = ww
			w.Close()
		})
	}
	if widths[EllipsisMiddle] > 210 || widths[EllipsisNone] <= 300 {
		t.Fatalf("ширина окна: с многоточием %d (ожидалось ≈200), без — %d (ожидалось больше текста)",
			widths[EllipsisMiddle], widths[EllipsisNone])
	}
}

// Grid (layout.md, §4): форма «подпись: поле», растягивается колонка 1.
func TestGridLayout(t *testing.T) {
	var g *Grid
	var l1, l2 *Label
	var e1, e2 *Entry
	var w *Window
	ui(func() {
		l1, l2 = NewLabel("Порт:"), NewLabel("Скорость:")
		e1, e2 = NewEntry(), NewEntry()
		g = NewGrid().Add(0, 0, l1).Add(1, 0, e1).Add(0, 1, l2)
		g.ExpandCol(1)  // до Add — флаг получит и будущий ребёнок колонки 1
		g.Add(1, 1, e2) // после ExpandCol
		g.SetMinSize(400, -1)
		w = NewWindow("Grid", g)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, _, gw, _ := alloc(g); return gw == 400 })
	shot(t, "grid")
	ui(func() {
		l1x, _, l1w, _ := alloc(l1)
		l2x, _, l2w, _ := alloc(l2)
		e1x, _, e1w, _ := alloc(e1)
		e2x, _, e2w, _ := alloc(e2)
		if l1x != 12 || l2x != 12 || l1w != l2w {
			t.Errorf("колонка 0: Порт[%d+%d] Скорость[%d+%d] — ширина колонки одна на всех", l1x, l1w, l2x, l2w)
		}
		if e1x != e2x || e1w != e2w || e1x != l1x+l1w+6 || e1x+e1w != 412 {
			t.Errorf("колонка 1 должна забрать всё свободное место: e1[%d+%d] e2[%d+%d]", e1x, e1w, e2x, e2w)
		}
		w.Close()
	})
}

func TestGridRules(t *testing.T) {
	ui(func() {
		g := NewGrid().AddSpan(0, 0, 2, 1, NewLabel("на две колонки"))
		w := NewWindow("Grid", g.Spacing(4, 2).Padding(3))
		defer w.Close()
		expectPanic(t, "nui: Grid: отрицательная колонка или строка", func() { g.Add(-1, 0, NewLabel("x")) })
		expectPanic(t, "nui: вставка nil вместо виджета", func() { g.Add(0, 1, nil) })
		if core.Debug {
			expectPanic(t, "nui: ячейки Grid пересекаются", func() { g.Add(1, 0, NewLabel("поверх")) })
		}
		g.ExpandRow(0).ExpandCol(5)
	})
}

// SetSize: растягиваемое окно — начальный размер; нерастягиваемое — фиксированный.
func TestWindowSize(t *testing.T) {
	var w *Window
	ui(func() {
		w = NewWindow("SetSize", NewLabel("окно"))
		w.SetSize(320, 200)
		w.SetPadding(20)
		w.SetIconName("utilities-terminal")
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, _, ww, wh := drv.Allocation(w.h); return ww == 320 && wh == 200 })
	ui(func() { w.SetResizable(false); w.SetSize(260, 150) })
	waitUntil(t, 2000, func() bool { _, _, ww, wh := drv.Allocation(w.h); return ww == 260 && wh == 150 })
	ui(func() { w.Close() })
}

// SetMaximized и SetFullscreen до и после показа. Разворачивает окно оконный менеджер:
// в make test (Xvfb без него) проверяется только, что вызовы не ломают окно; результат
// виден на снимках (docs/reports/window-states.md).
func TestWindowStates(t *testing.T) {
	var w *Window
	ui(func() {
		w = NewWindow("состояния окна", NewLabel("развёрнуто"))
		w.SetSize(300, 200)
		w.SetMaximized(true)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, _, ww, _ := drv.Allocation(w.h); return ww >= 300 })
	ui(func() {
		w.SetMaximized(false)
		w.SetFullscreen(true)
		w.SetFullscreen(false)
		w.SetFullscreen(false) // повторный — ничего не делает
		if !w.Alive() || !drv.Visible(w.h) {
			t.Error("окно пропало после смены состояний")
		}
		w.Close()
	})
}

func TestPhase2WidgetsDestroyed(t *testing.T) {
	var w *Window
	var e *Entry
	var c *Check
	var cb *Combo
	var s *Slider
	var p *Progress
	var tv *TextView
	var g *Grid
	var f *Frame
	ui(func() {
		e, c, cb, s, p, tv = NewEntry(), NewCheck("c"), NewCombo("a"), NewSlider(0, 10), NewProgress(), NewTextView()
		g = NewGrid().Add(0, 0, e)
		f = NewFrame("f", NewLabel("в рамке"))
		w = NewWindow("будет закрыто", VBox(g, c, cb, s, p, tv, f))
		w.Close()
	})
	time.Sleep(10 * time.Millisecond)
	ui(func() {
		if e.Alive() || c.Alive() || cb.Alive() || s.Alive() || p.Alive() || tv.Alive() || g.Alive() || f.Alive() {
			t.Error("виджеты живы после закрытия окна")
		}
		e.SetText("x")
		e.OnEnter(nil)
		c.SetChecked(true)
		cb.Add("b")
		cb.SetSelected(0)
		cb.SetText("a")
		s.SetValue(3)
		p.SetValue(0.5)
		p.Pulse()
		tv.Append("x")
		tv.Focus()
		g.Add(1, 1, NewLabel("x")).ExpandCol(0)
		f.SetChild(NewLabel("x"))
		w.SetSize(1, 1)
		w.SetResizable(false)
		w.SetDefault(nil)
		if e.Text() != "" || c.Checked() || cb.Selected() != -1 || cb.Text() != "" || s.Value() != 0 || p.Value() != 0 || tv.Text() != "" {
			t.Error("геттеры уничтоженных виджетов должны возвращать нулевые значения")
		}
	})
}

// comboEntryH — поле ввода ComboEntry, для имитации Enter в нём.
func comboEntryH(c *Combo) drv.H { return drv.ComboEntryForTest(c.h) }
