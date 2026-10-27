//go:build gui

package nui

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// testPNG — PNG w×h: слева непрозрачный синий, справа полупрозрачный красный.
func testPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.NRGBA{40, 90, 200, 255}
			if x >= w/2 {
				c = color.NRGBA{220, 40, 40, 128}
			}
			img.SetNRGBA(x, y, c)
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// testJPEGRotated — JPEG w×h с EXIF Orientation = 6 («повернуть на 90° по часовой»):
// просмотрщик показывает его h×w.
func testJPEGRotated(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 200
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, nil); err != nil {
		t.Fatal(err)
	}
	// APP1 Exif: TIFF little-endian, одна запись IFD0 — 0x0112 SHORT 1 = 6.
	tiff := []byte{'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 0x01, 3, 0, 1, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0}
	app1 := append([]byte("Exif\x00\x00"), tiff...)
	n := len(app1) + 2
	seg := append([]byte{0xFF, 0xE1, byte(n >> 8), byte(n)}, app1...)
	j := b.Bytes()
	return append(append(append([]byte{}, j[:2]...), seg...), j[2:]...) // сразу после SOI
}

// Картинка: размер, SetFit, поворот по EXIF, файл, ошибки, Clear.
func TestImage(t *testing.T) {
	dir := tempDir(t)
	file := filepath.Join(dir, "картинка.png")
	if err := os.WriteFile(file, testPNG(t, 30, 10), 0o644); err != nil {
		t.Fatal(err)
	}
	var w *Window
	var img *Image
	ui(func() {
		img = NewImage()
		img.SetAlign(AlignStart, AlignStart) // место — ровно по картинке
		if err := img.SetData(testPNG(t, 120, 80)); err != nil {
			t.Fatalf("SetData(PNG): %v", err)
		}
		w = NewWindow("картинка", VBox(img, NewLabel("Подпись под картинкой")))
		w.SetSize(300, 200)
		w.Show()
	})
	size := func(wantW, wantH int) {
		t.Helper()
		waitUntil(t, 2000, func() bool { _, _, ww, hh := alloc(img); return ww == wantW && hh == wantH })
	}
	size(120, 80)
	shot(t, "image")
	ui(func() {
		if iw, ih := img.ImageSize(); iw != 120 || ih != 80 {
			t.Errorf("ImageSize = %d×%d, ожидалось 120×80", iw, ih)
		}
		img.SetFit(60, 0)
	})
	size(60, 40)
	ui(func() {
		if err := img.SetData([]byte("не картинка")); !errors.Is(err, ErrImage) {
			t.Errorf("SetData(мусор) = %v, ожидалось ErrImage", err)
		}
		if err := img.SetData(nil); !errors.Is(err, ErrImage) {
			t.Errorf("SetData(nil) = %v, ожидалось ErrImage", err)
		}
		if iw, ih := img.ImageSize(); iw != 120 || ih != 80 {
			t.Errorf("после ошибки ImageSize = %d×%d: прежняя картинка должна остаться", iw, ih)
		}
		img.SetFit(0, 0)
		if err := img.SetData(testJPEGRotated(t, 40, 20)); err != nil {
			t.Fatalf("SetData(JPEG): %v", err)
		}
		if iw, ih := img.ImageSize(); iw != 20 || ih != 40 {
			t.Errorf("JPEG с поворотом в EXIF: ImageSize = %d×%d, ожидалось 20×40", iw, ih)
		}
	})
	size(20, 40)
	ui(func() {
		if err := img.SetFile(file); err != nil {
			t.Errorf("SetFile: %v", err)
		}
		if err := img.SetFile(filepath.Join(dir, "нет такого.png")); err == nil || errors.Is(err, ErrImage) {
			t.Errorf("SetFile(нет файла) = %v, ожидалась ошибка чтения", err)
		}
	})
	size(30, 10)
	ui(func() {
		img.Clear()
		if iw, ih := img.ImageSize(); iw != 0 || ih != 0 {
			t.Errorf("после Clear ImageSize = %d×%d", iw, ih)
		}
		img.SetFit(10, 10) // на пустой картинке — ничего не ломает
		w.Close()
		if err := img.SetData(testPNG(t, 4, 4)); err != nil { // уничтоженная — молча ничего
			t.Errorf("SetData у уничтоженной картинки: %v", err)
		}
	})
}
