package nui

import (
	"errors"

	"nui/internal/core"
	"nui/internal/drv"
)

// Image — картинка (api.md, §19): PNG, JPEG, GIF (первый кадр) и BMP на всех ОС.
// Декодирует ОС: GdkPixbuf на Linux, GDI+ на Windows — свой декодер в бинарник не попадает.
// Размер — в логических пикселях: на экране с масштабом 200 % картинка 100×100 занимает
// 200×200 точек и остаётся чёткой, если в файле хватает точек. Фото с поворотом в EXIF
// показывается повёрнутым, как в просмотрщиках.
type Image struct {
	widget
	imgW, imgH int // размер картинки после поворота по EXIF; 0 — пусто
	fitW, fitH int // SetFit
}

// ErrImage — данные не распознаны как картинка (или формат не поддерживается).
var ErrImage = errors.New("nui: не удалось распознать картинку")

// NewImage создаёт пустую картинку (0×0).
func NewImage() *Image {
	checkCreate()
	i := &Image{}
	i.id = core.Register((*imageEv)(i))
	i.h = drv.ImageNew(i.id)
	return i
}

// SetData показывает картинку из байтов файла (PNG, JPEG…). Ошибка — ErrImage; тогда
// остаётся прежняя картинка. Данные копируются: срез потом можно менять.
func (i *Image) SetData(data []byte) error {
	if i.dead() {
		return nil
	}
	if len(data) == 0 {
		return ErrImage
	}
	w, h, ok := drv.ImageLoad(i.h, data)
	if !ok {
		return ErrImage
	}
	i.imgW, i.imgH = w, h
	i.show()
	return nil
}

// SetFile — то же из файла. Ошибка — файл не прочитан или ErrImage.
func (i *Image) SetFile(path string) error {
	if i.dead() {
		return nil
	}
	data, ok := drv.ReadFile(path)
	if !ok {
		return errors.New("nui: не удалось прочитать файл " + path)
	}
	return i.SetData(data)
}

// SetFit вписывает картинку в width×height с сохранением пропорций: большая уменьшается,
// маленькая остаётся как есть. 0 — по этой стороне без ограничения; (0, 0) — исходный
// размер (по умолчанию). Действует и на картинки, заданные позже.
func (i *Image) SetFit(width, height int) {
	if i.dead() {
		return
	}
	i.fitW, i.fitH = max(width, 0), max(height, 0)
	if i.imgW > 0 {
		i.show()
	}
}

// Clear убирает картинку: размер становится 0×0.
func (i *Image) Clear() {
	if i.dead() {
		return
	}
	drv.ImageLoad(i.h, nil)
	i.imgW, i.imgH = 0, 0
}

// ImageSize — размер картинки в пикселях файла (после поворота по EXIF), без SetFit;
// пусто — 0, 0.
func (i *Image) ImageSize() (width, height int) {
	if i.dead() {
		return 0, 0
	}
	return i.imgW, i.imgH
}

func (i *Image) show() {
	w, h := core.Fit(i.imgW, i.imgH, i.fitW, i.fitH)
	drv.ImageShow(i.h, w, h)
}

type imageEv Image

func (i *imageEv) Event(k core.Kind, _ uintptr) uintptr {
	if k == core.EvDestroy {
		i.destroyed()
	}
	return 0
}
