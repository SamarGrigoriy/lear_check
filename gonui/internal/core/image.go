package core

// Fit — размер картинки w×h, вписанной в maxW×maxH с сохранением пропорций (Image.SetFit).
// 0 в maxW или maxH — по этой стороне без ограничения. Картинка только уменьшается и не
// бывает меньше 1×1; пустая (w или h ≤ 0) — 0×0.
func Fit(w, h, maxW, maxH int) (int, int) {
	if w <= 0 || h <= 0 {
		return 0, 0
	}
	num, den := 1, 1 // масштаб num/den: меньшее из 1, maxW/w, maxH/h
	if maxW > 0 && maxW*den < w*num {
		num, den = maxW, w
	}
	if maxH > 0 && maxH*den < h*num {
		num, den = maxH, h
	}
	return max(1, (w*num+den/2)/den), max(1, (h*num+den/2)/den)
}

// Orientation — поворот JPEG из EXIF (тег Orientation 0x0112): 1…8, как в стандарте;
// 1 — без поворота, тега нет или это не JPEG. На Windows GDI+ сам его не применяет.
func Orientation(b []byte) int {
	if len(b) < 4 || b[0] != 0xFF || b[1] != 0xD8 {
		return 1
	}
	for i := 2; i+4 <= len(b) && b[i] == 0xFF; {
		m := b[i+1]
		if m == 0xFF { // заполняющий байт перед маркером
			i++
			continue
		}
		n := int(b[i+2])<<8 | int(b[i+3])         // длина сегмента вместе с этими двумя байтами
		if m == 0xDA || n < 2 || i+2+n > len(b) { // SOS: дальше сжатые данные
			break
		}
		if m == 0xE1 {
			if o := exifOrientation(b[i+4 : i+2+n]); o != 0 {
				return o
			}
		}
		i += 2 + n
	}
	return 1
}

// exifOrientation — тег Orientation в IFD0 сегмента APP1; 0 — нет.
func exifOrientation(s []byte) int {
	if len(s) < 14 || string(s[:6]) != "Exif\x00\x00" {
		return 0
	}
	t := s[6:] // TIFF: порядок байт, 42, смещение IFD0
	be := false
	switch string(t[:2]) {
	case "II":
	case "MM":
		be = true
	default:
		return 0
	}
	u16 := func(p []byte) int {
		if be {
			return int(p[0])<<8 | int(p[1])
		}
		return int(p[1])<<8 | int(p[0])
	}
	off := u16(t[6:])<<16 | u16(t[4:]) // смещение — 32 бита
	if be {
		off = u16(t[4:])<<16 | u16(t[6:])
	}
	if off+2 > len(t) {
		return 0
	}
	for k, cnt := 0, u16(t[off:]); k < cnt; k++ {
		e := off + 2 + 12*k // запись: тег, тип, число, значение
		if e+12 > len(t) {
			return 0
		}
		if u16(t[e:]) == 0x0112 {
			if o := u16(t[e+8:]); u16(t[e+2:]) == 3 && o >= 1 && o <= 8 { // SHORT
				return o
			}
			return 0
		}
	}
	return 0
}
