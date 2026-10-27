package nui

import "nui/internal/drv"

// Args — аргументы командной строки, как os.Args: Args()[0] — программа. На Windows
// они берутся из GetCommandLineW: в сборке TinyGo os.Args там приходят в кодировке ANSI,
// и путь «Открыть с помощью…» с русскими буквами портится. Можно вызывать до Run и
// из любой горутины.
func Args() []string { return drv.Args() }
