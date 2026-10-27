//go:build gui

package nui

import (
	"testing"

	"nui/internal/drv"
)

// Оформление (ADR-19): цвет и фон, отступ цветной панели, крупный шрифт подписи,
// цветная кнопка по умолчанию; сброс в NoColor.
func TestStyles(t *testing.T) {
	var w *Window
	var small, big, inPanel *Label
	var panel *Box
	var e *Entry
	clicks := 0
	ui(func() {
		small, big = NewLabel("Заголовок"), NewLabel("Заголовок")
		big.SetFontScale(2)
		big.SetBold(true)
		big.SetColor(RGB(255, 255, 255))
		inPanel = NewLabel("в панели")
		panel = VBox(inPanel).Padding(10).Background(RGB(40, 70, 120))
		e = NewEntry()
		e.SetColor(RGB(0, 0, 160))
		e.SetBackground(RGB(255, 250, 220))
		ok := NewButton("OK", func() { clicks++ })
		ok.SetBackground(RGB(70, 130, 200))
		ok.SetColor(RGB(255, 255, 255))
		tv := NewTextView()
		tv.SetColor(RGB(30, 30, 30))
		tv.SetBackground(RGB(250, 250, 245))
		w = NewWindow("оформление", VBox(panel, small, big, e, Expand(tv), ok).Spacing(6))
		w.SetBackground(RGB(245, 245, 250))
		w.SetDefault(ok)
		w.SetSize(360, 360)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, _, ww, _ := alloc(panel); return ww > 1 })
	shot(t, "styles")
	ui(func() {
		// Отступ цветной панели — по-прежнему отступ: на GTK он стал CSS padding.
		_, py, _, _ := drv.Allocation(panel.h)
		x, y, _, _ := alloc(inPanel)
		if x != 12+10 || y != 12+10 {
			t.Errorf("подпись в цветной панели в (%d, %d), ожидалось (22, 22); панель y = %d", x, y, py)
		}
		_, _, _, hs := alloc(small)
		_, _, _, hb := alloc(big)
		if hb < hs*16/10 {
			t.Errorf("SetFontScale(2): высота %d при обычной %d", hb, hs)
		}
	})
	ui(func() { drv.SimEnter(e.h) }) // цветная кнопка по умолчанию нажимается Enter
	waitUntil(t, 2000, func() bool { return clicks == 1 })
	ui(func() {
		panel.Background(NoColor) // отступ возвращается в border-width — раскладка та же
		big.SetFontScale(1)
		e.SetColor(NoColor)
		e.SetBackground(NoColor)
		w.SetBackground(NoColor)
		panel.Padding(4)
	})
	waitUntil(t, 2000, func() bool { x, _, _, _ := alloc(inPanel); return x == 12+4 })
	ui(func() {
		_, _, _, hb := alloc(big)
		_, _, _, hs := alloc(small)
		if hb > hs+hs/4 { // жирный, но обычного размера
			t.Errorf("SetFontScale(1): высота %d при обычной %d", hb, hs)
		}
		w.Close()
	})
}
