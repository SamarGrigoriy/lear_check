#!/bin/sh
# Сверка собственных прототипов и констант драйвера с настоящими заголовками GTK
# (docs/driver-gtk3.md, §10). Нужны libgtk-3-dev и pkg-config — только здесь, не для сборки;
# для Win32 — mingw-w64 (driver-win32.md, §1).
# Ошибка в прототипе — неопределённое поведение без единого предупреждения компилятора.
set -eu
cd "$(dirname "$0")/.."
if ! pkg-config --exists gtk+-3.0; then
	echo "check-protos: нужен libgtk-3-dev (apt install libgtk-3-dev pkg-config)" >&2
	exit 2
fi
CFLAGS="-std=c11 -D_GNU_SOURCE -DNUI_CHECK -DNUI_GUI_TEST -Wall -Werror $(pkg-config --cflags gtk+-3.0)"

# 1. Каждый NUI_FN и NUI_CONST превращается в _Static_assert против настоящего объявления.
#    Исходник из stdin: «#include "x.h"» ищется сначала в текущем каталоге — поэтому cd.
(cd internal/drv && printf '#include <gtk/gtk.h>\n#include <glib-unix.h>\n#include <stdint.h>\n#include <string.h>\n#include <unistd.h>\n#include "gtk3_linux.h"\n' |
	cc $CFLAGS -x c -fsyntax-only -)

# 2. Обёртки (включая тестовые) компилируются против настоящих заголовков.
#    В режиме проверки NUI_FN ничего не объявляет, поэтому всё нужное — через -include.
cc $CFLAGS -I internal/drv -include gtk/gtk.h -include glib-unix.h -include stdint.h -include string.h \
	-fsyntax-only internal/drv/wrap_linux.c

# 3. Win32 (driver-win32.md, §1): win32_windows.h и обёртки против заголовков mingw-w64.
WINCC=${WINCC:-x86_64-w64-mingw32-gcc}
if command -v "$WINCC" >/dev/null; then
	WFLAGS="-std=c11 -DNUI_CHECK -DWINVER=0x0A00 -D_WIN32_WINNT=0x0A00 -Wall -Werror"
	WINC="-include windows.h -include commctrl.h -include shlobj.h -include commdlg.h -include gdiplus.h"
	"$WINCC" $WFLAGS $WINC -fsyntax-only -x c internal/drv/win32_windows.h
	"$WINCC" $WFLAGS $WINC -DNUI_GUI_TEST -I internal/drv -fsyntax-only internal/drv/wrap_windows.c
	"$WINCC" -std=c11 -Wall -Werror -DNUI_GUI_TEST -I internal/drv -fsyntax-only internal/drv/wrap_windows.c
	echo "check-protos: Win32 — ok"
else
	echo "check-protos: Win32 пропущен — нет $WINCC (apt install gcc-mingw-w64-x86-64)"
fi

echo "check-protos: ok"
