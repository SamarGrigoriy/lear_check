package core

import "unicode/utf8"

// Немногие нужные вещи вместо fmt, strconv, strings и path (ADR-11, size.md).

// Itoa — десятичная запись числа для сообщений.
func Itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	u := uint64(n)
	if neg {
		u = uint64(-n)
	}
	var b [20]byte
	i := len(b)
	for ; u > 0; u /= 10 {
		i--
		b[i] = byte('0' + u%10)
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// Atoi — целое число из десятичной записи: знак и цифры, пробелы по краям допустимы.
// false — не число или не помещается в int32 (Spin, api.md, §11).
func Atoi(s string) (int, bool) {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	neg := len(s) > 0 && (s[0] == '-' || s[0] == '+')
	if neg {
		neg = s[0] == '-'
		s = s[1:]
	}
	if s == "" || len(s) > 10 {
		return 0, false
	}
	v := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		v = v*10 + int(s[i]-'0')
	}
	if neg {
		v = -v
	}
	if v < -1<<31 || v > 1<<31-1 {
		return 0, false
	}
	return v, true
}

// FixUTF8 заменяет каждую битую последовательность UTF-8 на U+FFFD.
// GTK невалидный UTF-8 не принимает: пишет Gtk-CRITICAL и ничего не делает (ADR-13).
func FixUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	b := make([]byte, 0, len(s)+8)
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && n == 1 {
			b = append(b, "�"...)
			i++
			continue
		}
		b = append(b, s[i:i+n]...)
		i += n
	}
	return string(b)
}

// Dir — каталог пути: всё до последнего '/'. "/a/b" → "/a", "/a" → "/", "a" → ".".
func Dir(path string) string {
	i := len(path) - 1
	for i >= 0 && path[i] != '/' {
		i--
	}
	switch {
	case i < 0:
		return "."
	case i == 0:
		return "/"
	}
	return path[:i]
}
