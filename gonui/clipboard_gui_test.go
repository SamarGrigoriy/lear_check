//go:build gui

package nui

import (
	"os"
	"os/exec"
	"slices"
	"testing"
)

// Другая программа (дочерний процесс) видит то, что положили в буфер мы.
func TestClipboardOtherProcess(t *testing.T) {
	const s = "из другого процесса: Ёж 🙂\nвторая строка"
	ui(func() { SetClipboard(s) })
	runHelperEnv(t, "clipboard", 0, []string{"NUI_CLIP=" + s}, nil)
}

// Буфер обмена: текст туда и обратно (кириллица, эмодзи, переводы строк — на Windows
// они в буфере \r\n), пустой текст; «Скопировать ответ» в чате — LastMessage.
func TestClipboard(t *testing.T) {
	ui(func() {
		for _, s := range []string{"Привет, мир 🙂", "строка 1\nстрока 2\n", ""} {
			SetClipboard(s)
			if got := Clipboard(); got != s {
				t.Errorf("Clipboard() = %q, ожидалось %q", got, s)
			}
		}
		c := NewChat()
		w := NewWindow("буфер", c)
		if c.LastMessage() != "" {
			t.Errorf("LastMessage пустого чата: %q", c.LastMessage())
		}
		c.Add("ИИ", "ответ\nв две строки")
		SetClipboard(c.LastMessage())
		if got := Clipboard(); got != "ответ\nв две строки" {
			t.Errorf("скопированный ответ: %q", got)
		}
		st := c.Stream("ИИ")
		st.WriteString("часть 1, ")
		st.WriteString("часть 2")
		st.Close()
		c.flush()
		if c.LastMessage() != "часть 1, часть 2" {
			t.Errorf("LastMessage после Stream: %q", c.LastMessage())
		}
		c.Clear()
		if c.LastMessage() != "" {
			t.Errorf("LastMessage после Clear: %q", c.LastMessage())
		}
		w.Close()
	})
}

// Args: аргументы командной строки — как os.Args (в gc), и русский путь с пробелами
// доходит без потерь (в TinyGo на Windows os.Args его портит — Args берёт GetCommandLineW).
func TestArgs(t *testing.T) {
	if a := Args(); !slices.Equal(a, os.Args) {
		t.Errorf("Args() = %q, os.Args = %q", a, os.Args)
	}
	const arg = `C:\Мои документы\фото с телефона.jpg`
	cmd := exec.Command(os.Args[0], "-test.run=^$", arg)
	cmd.Env = append(os.Environ(), "NUI_HELPER=args", "NUI_ARG="+arg)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("дочерний процесс: %v\n%s", err, out)
	}
}
