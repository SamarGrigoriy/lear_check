package core

import (
	"errors"
	"unicode/utf8"
)

// Маленький JSON для истории чата (Chat.LoadJSON, Chat.JSON): encoding/json не проходит
// белый список импортов (ADR-7) и раздул бы сборку TinyGo.

// JSON — разобранное значение.
type JSON struct {
	Kind byte   // 's' строка, 'n' число, 'b' true/false, 'z' null, 'a' массив, 'o' объект
	Str  string // строка; число — как записано; true/false — "true"/"false"
	Arr  []JSON // элементы массива
	Keys []string
	Vals []JSON // поля объекта в порядке записи
}

// Get — поле объекта; ok = false — поля нет или это не объект.
func (v JSON) Get(key string) (JSON, bool) {
	for i, k := range v.Keys {
		if k == key {
			return v.Vals[i], true
		}
	}
	return JSON{}, false
}

// ParseJSON разбирает документ JSON (RFC 8259). Ошибка — с номером байта.
func ParseJSON(b []byte) (JSON, error) {
	p := jsonParser{b: b}
	p.space()
	v, err := p.value(0)
	if err != nil {
		return JSON{}, err
	}
	p.space()
	if p.i < len(b) {
		return JSON{}, p.fail("лишнее после значения")
	}
	return v, nil
}

type jsonParser struct {
	b []byte
	i int
}

const jsonMaxDepth = 512 // глубже — ошибка, а не переполнение стека

func (p *jsonParser) fail(msg string) error {
	return errors.New("байт " + Itoa(p.i) + ": " + msg)
}

func (p *jsonParser) space() {
	for p.i < len(p.b) {
		switch p.b[p.i] {
		case ' ', '\t', '\n', '\r':
			p.i++
		default:
			return
		}
	}
}

// lit — слово true, false или null.
func (p *jsonParser) lit(word string, v JSON) (JSON, error) {
	if len(p.b)-p.i < len(word) || string(p.b[p.i:p.i+len(word)]) != word {
		return JSON{}, p.fail("неизвестное значение")
	}
	p.i += len(word)
	return v, nil
}

func (p *jsonParser) value(depth int) (JSON, error) {
	if depth > jsonMaxDepth {
		return JSON{}, p.fail("слишком глубокая вложенность")
	}
	if p.i >= len(p.b) {
		return JSON{}, p.fail("неожиданный конец")
	}
	switch c := p.b[p.i]; {
	case c == '"':
		s, err := p.str()
		return JSON{Kind: 's', Str: s}, err
	case c == '[':
		return p.array(depth)
	case c == '{':
		return p.object(depth)
	case c == 't':
		return p.lit("true", JSON{Kind: 'b', Str: "true"})
	case c == 'f':
		return p.lit("false", JSON{Kind: 'b', Str: "false"})
	case c == 'n':
		return p.lit("null", JSON{Kind: 'z'})
	case c == '-' || c >= '0' && c <= '9':
		return p.number()
	}
	return JSON{}, p.fail("неизвестное значение")
}

func (p *jsonParser) array(depth int) (JSON, error) {
	p.i++ // [
	v := JSON{Kind: 'a'}
	p.space()
	if p.i < len(p.b) && p.b[p.i] == ']' {
		p.i++
		return v, nil
	}
	for {
		p.space()
		e, err := p.value(depth + 1)
		if err != nil {
			return JSON{}, err
		}
		v.Arr = append(v.Arr, e)
		p.space()
		if p.i >= len(p.b) {
			return JSON{}, p.fail("массив не закрыт")
		}
		switch p.b[p.i] {
		case ',':
			p.i++
		case ']':
			p.i++
			return v, nil
		default:
			return JSON{}, p.fail("ожидалась «,» или «]»")
		}
	}
}

func (p *jsonParser) object(depth int) (JSON, error) {
	p.i++ // {
	v := JSON{Kind: 'o'}
	p.space()
	if p.i < len(p.b) && p.b[p.i] == '}' {
		p.i++
		return v, nil
	}
	for {
		p.space()
		if p.i >= len(p.b) || p.b[p.i] != '"' {
			return JSON{}, p.fail("ожидалось имя поля в кавычках")
		}
		k, err := p.str()
		if err != nil {
			return JSON{}, err
		}
		p.space()
		if p.i >= len(p.b) || p.b[p.i] != ':' {
			return JSON{}, p.fail("ожидалось «:»")
		}
		p.i++
		p.space()
		e, err := p.value(depth + 1)
		if err != nil {
			return JSON{}, err
		}
		v.Keys = append(v.Keys, k)
		v.Vals = append(v.Vals, e)
		p.space()
		if p.i >= len(p.b) {
			return JSON{}, p.fail("объект не закрыт")
		}
		switch p.b[p.i] {
		case ',':
			p.i++
		case '}':
			p.i++
			return v, nil
		default:
			return JSON{}, p.fail("ожидалась «,» или «}»")
		}
	}
}

// number — проверка формата; значение хранится как записано (чату числа не нужны).
func (p *jsonParser) number() (JSON, error) {
	start := p.i
	digits := func() int {
		n := 0
		for p.i < len(p.b) && p.b[p.i] >= '0' && p.b[p.i] <= '9' {
			p.i++
			n++
		}
		return n
	}
	if p.b[p.i] == '-' {
		p.i++
	}
	if p.i < len(p.b) && p.b[p.i] == '0' {
		p.i++
	} else if digits() == 0 {
		return JSON{}, p.fail("неверное число")
	}
	if p.i < len(p.b) && p.b[p.i] == '.' {
		p.i++
		if digits() == 0 {
			return JSON{}, p.fail("неверное число")
		}
	}
	if p.i < len(p.b) && (p.b[p.i] == 'e' || p.b[p.i] == 'E') {
		p.i++
		if p.i < len(p.b) && (p.b[p.i] == '+' || p.b[p.i] == '-') {
			p.i++
		}
		if digits() == 0 {
			return JSON{}, p.fail("неверное число")
		}
	}
	return JSON{Kind: 'n', Str: string(p.b[start:p.i])}, nil
}

// str — строка в кавычках с экранированием; битый UTF-8 заменяется на U+FFFD.
func (p *jsonParser) str() (string, error) {
	p.i++ // "
	var out []byte
	for {
		if p.i >= len(p.b) {
			return "", p.fail("строка не закрыта")
		}
		c := p.b[p.i]
		switch {
		case c == '"':
			p.i++
			return FixUTF8(string(out)), nil
		case c < 0x20:
			return "", p.fail("управляющий символ в строке")
		case c != '\\':
			out = append(out, c)
			p.i++
			continue
		}
		p.i++ // \
		if p.i >= len(p.b) {
			return "", p.fail("строка не закрыта")
		}
		e := p.b[p.i]
		p.i++
		switch e {
		case '"', '\\', '/':
			out = append(out, e)
		case 'b':
			out = append(out, '\b')
		case 'f':
			out = append(out, '\f')
		case 'n':
			out = append(out, '\n')
		case 'r':
			out = append(out, '\r')
		case 't':
			out = append(out, '\t')
		case 'u':
			r, ok := p.hex4()
			if !ok {
				return "", p.fail("неверная последовательность \\u")
			}
			if r >= 0xD800 && r < 0xDC00 { // старшая половина суррогатной пары
				save := p.i
				if p.i+1 < len(p.b) && p.b[p.i] == '\\' && p.b[p.i+1] == 'u' {
					p.i += 2
					if lo, ok := p.hex4(); ok && lo >= 0xDC00 && lo < 0xE000 {
						r = 0x10000 + (r-0xD800)<<10 + (lo - 0xDC00)
					} else {
						p.i = save
					}
				}
			}
			out = utf8.AppendRune(out, rune(r)) // одинокая половина пары → U+FFFD
		default:
			return "", p.fail("неверное экранирование")
		}
	}
}

func (p *jsonParser) hex4() (uint32, bool) {
	if len(p.b)-p.i < 4 {
		return 0, false
	}
	var r uint32
	for _, c := range p.b[p.i : p.i+4] {
		switch {
		case c >= '0' && c <= '9':
			r = r<<4 | uint32(c-'0')
		case c >= 'a' && c <= 'f':
			r = r<<4 | uint32(c-'a'+10)
		case c >= 'A' && c <= 'F':
			r = r<<4 | uint32(c-'A'+10)
		default:
			return 0, false
		}
	}
	p.i += 4
	return r, true
}

// QuoteJSON — строка JSON в кавычках: экранированы кавычка, обратная косая черта и
// управляющие символы; остальное (кириллица, эмодзи) — как есть, в UTF-8.
func QuoteJSON(s string) string {
	const hex = "0123456789abcdef"
	b := make([]byte, 0, len(s)+2)
	b = append(b, '"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"' || c == '\\':
			b = append(b, '\\', c)
		case c == '\n':
			b = append(b, '\\', 'n')
		case c == '\r':
			b = append(b, '\\', 'r')
		case c == '\t':
			b = append(b, '\\', 't')
		case c < 0x20:
			b = append(b, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xF])
		default:
			b = append(b, c)
		}
	}
	return string(append(b, '"'))
}
