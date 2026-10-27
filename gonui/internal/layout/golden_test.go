package layout_test

// Golden-тест (layout.md, §8): движок раскладывает каждый случай из testdata/cases так же,
// как GTK, — по снимкам из testdata/golden, без GTK, на любой ОС. Снимки пишет
// TestLayoutGTK в корневом пакете (make golden).

import (
	"os"
	"path/filepath"
	"testing"

	"nui/internal/layout/golden"
)

func TestGolden(t *testing.T) {
	cases, err := golden.Cases("testdata/cases")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("нет случаев в testdata/cases")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			g, err := golden.ReadGolden(filepath.Join("testdata/golden", c.Name+".txt"))
			if os.IsNotExist(err) {
				t.Fatal("нет golden-файла — снимите его: make golden")
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range golden.Check(c, g) {
				t.Error(e)
			}
		})
	}
}

// Каждому снимку — свой случай: снимки удалённых случаев не должны копиться.
func TestGoldenHasCase(t *testing.T) {
	files, _ := filepath.Glob("testdata/golden/*.txt")
	for _, f := range files {
		if _, err := os.Stat(filepath.Join("testdata/cases", filepath.Base(f))); err != nil {
			t.Errorf("%s: нет случая для снимка", f)
		}
	}
}
