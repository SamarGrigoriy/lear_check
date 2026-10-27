#!/bin/sh
# Сверка прототипов и констант s1.h с настоящими заголовками GTK (нужен libgtk-3-dev).
# Ошибка в прототипе — неопределённое поведение без единого предупреждения, поэтому проверка обязательна.
set -eu
cd "$(dirname "$0")"
CFLAGS="-std=c11 -D_GNU_SOURCE -DNUI_CHECK -Wall -Werror $(pkg-config --cflags gtk+-3.0)"

# 1. Каждый NUI_FN и NUI_CONST превращается в _Static_assert против настоящего объявления.
printf '#include <gtk/gtk.h>\n#include <glib-unix.h>\n#include <stdint.h>\n#include <string.h>\n#include <unistd.h>\n#include "s1.h"\n' |
	cc $CFLAGS -I . -x c -fsyntax-only -

# 2. Обёртки компилируются против настоящих заголовков.
cc $CFLAGS -I . -include gtk/gtk.h -include glib-unix.h -include stdint.h -fsyntax-only wrap.c

echo "check-protos: ok"
