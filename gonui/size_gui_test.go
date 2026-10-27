//go:build gui

package nui

import (
	"strconv"
	"testing"
)

// Size: 0 до показа и у скрытого, после раскладки — место виджета; размер окна.
func TestSize(t *testing.T) {
	var w *Window
	var l *Label
	var b *Button
	ui(func() {
		l = NewLabel("Подпись")
		b = NewButton("Кнопка", nil)
		b.SetMinSize(120, 40)
		w = NewWindow("размер", VBox(l, b))
		w.SetSize(300, 200)
		if lw, lh := l.Size(); lw != 0 || lh != 0 {
			t.Errorf("до показа окна Size = %d×%d", lw, lh)
		}
		w.Show()
	})
	waitUntil(t, 2000, func() bool { bw, _ := b.Size(); return bw > 0 })
	ui(func() {
		if ww, wh := w.Size(); ww != 300 || wh != 200 {
			t.Errorf("окно %d×%d, ожидалось 300×200", ww, wh)
		}
		bw, bh := b.Size()
		if bw != 300-24 || bh < 40 { // VBox растягивает по ширине: окно минус отступы 12
			t.Errorf("кнопка %d×%d, ожидалась ширина 276 и высота от 40", bw, bh)
		}
		b.SetVisible(false)
	})
	waitUntil(t, 2000, func() bool { bw, bh := b.Size(); return bw == 0 && bh == 0 })
	ui(func() { w.Close() })
}

// Свои размеры листьев без SetMinSize и Expand — одинаковые на всех ОС: у холста своего
// размера нет, список — 46×71, как GtkTreeView в прокрутке.
func TestLeafSizes(t *testing.T) {
	var w *Window
	var c *Canvas
	var l *List
	ui(func() {
		c = NewCanvas(nil)
		l = NewList("Имя")
		l.SetAlign(AlignStart, AlignFill)
		w = NewWindow("размеры листьев", VBox(c, l, NewLabel("под ними")))
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, lh := l.Size(); return lh >= 71 }) // без исправления на Win32 — 0
	ui(func() {
		if _, ch := c.Size(); ch > 1 {
			t.Errorf("холст без SetMinSize высотой %d, ожидалось 0 (GTK даёт 1)", ch)
		}
		if lw, lh := l.Size(); lw != 46 || lh != 71 {
			t.Errorf("список %d×%d, ожидалось 46×71", lw, lh)
		}
		w.Close()
	})
}

// Кнопки добавляются в VBox внутри Scroll: окно и прокрутка не растут, ряд прокручивается.
func TestScrollKeepsWindowSize(t *testing.T) {
	var w *Window
	var list *Box
	var scroll *Scroll
	ui(func() {
		list = VBox()
		scroll = NewScroll(list)
		w = NewWindow("кнопки", VBox(NewLabel("Кнопки:"), Expand(scroll)))
		w.SetSize(300, 250)
		w.Show()
	})
	waitUntil(t, 2000, func() bool { _, sh := scroll.Size(); return sh > 0 })
	var sh0 int
	ui(func() {
		_, sh0 = scroll.Size()
		for i := 1; i <= 60; i++ {
			list.Add(NewButton("Кнопка "+strconv.Itoa(i), nil))
		}
	})
	waitUntil(t, 2000, func() bool { _, lh := list.Size(); return lh > 1000 })
	shot(t, "scroll-buttons")
	ui(func() {
		if ww, wh := w.Size(); ww != 300 || wh != 250 {
			t.Errorf("окно выросло до %d×%d, ожидалось 300×250", ww, wh)
		}
		if _, sh := scroll.Size(); sh != sh0 {
			t.Errorf("прокрутка %d в высоту, была %d", sh, sh0)
		}
		w.Close()
	})
}
