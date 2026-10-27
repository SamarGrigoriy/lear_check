package nui

// Color — цвет 0xRRGGBB: текст и фон виджетов (ADR-19). Цвета не следуют за тёмной темой
// ОС: задавая фон, задавайте и цвет текста.
type Color uint32

// RGB собирает цвет из компонент.
func RGB(r, g, b uint8) Color { return Color(r)<<16 | Color(g)<<8 | Color(b) }

// NoColor — цвет темы (по умолчанию).
const NoColor Color = 1 << 24
