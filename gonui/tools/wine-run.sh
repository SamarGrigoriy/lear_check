#!/bin/sh
# Запуск exe nui под Wine в Xvfb с openbox (образ nui-win, tools/win.Dockerfile).
#   tools/wine-run.sh exe…                 — selftest: каждый exe сам выходит, код должен быть 0
#   tools/wine-run.sh -shots DIR exe…      — снимок окна каждого exe, затем exe закрывается
#   -dpi N — масштаб Windows в точках на дюйм: 96 = 100 %, 144 = 150 %, 192 = 200 %
#   WINE_ARGS — аргументы exe (например, -test.v для тестового бинарника)
# Окно Wine ищется по классу (имени exe): заголовки не в ASCII xwininfo в локали C не печатает.
set -u
here=$(cd "$(dirname "$0")" && pwd)
if [ "${1:-}" != --inner ]; then
	# Без exec: xvfb-run не должен стать PID 1 контейнера — тогда он не дожидается Xvfb.
	xvfb-run -a -s "-screen 0 1024x768x24" "$0" --inner "$@"
	exit $?
fi
shift
shots="" dpi=96
while [ $# -gt 0 ]; do
	case "$1" in
	-shots) shots=$2; shift 2 ;;
	-dpi) dpi=$2; shift 2 ;;
	*) break ;;
	esac
done
export WINEPREFIX=${WINEPREFIX:-$HOME/.wine-nui} WINEDEBUG=-all WINEDLLOVERRIDES="mscoree,mshtml="
# Локаль UTF-8: в локали C Wine не может создать файл с именем не в ASCII («журнал.log»).
export LANG=ru_RU.UTF-8
openbox >/dev/null 2>&1 &
sleep 1
wineboot -i >/dev/null 2>&1
# Wine 8.0 не знает bcryptprimitives.dll (ProcessPrng), без неё рантайм Go ≥ 1.22 не стартует;
# на Windows 10+ она системная. Заглушка — только в префиксе Wine (spike/s4/RESULT.md).
stub=$WINEPREFIX/drive_c/windows/system32/bcryptprimitives.dll
[ -f "$stub" ] || x86_64-w64-mingw32-gcc -O2 -shared -o "$stub" "$here/wine/bcryptprimitives.c" \
	"$here/wine/bcryptprimitives.def" -ladvapi32
wine reg add 'HKCU\Control Panel\Desktop' /v LogPixels /t REG_DWORD /d "$dpi" /f >/dev/null 2>&1
fail=0
for f in "$@"; do
	n=$(basename "$f" .exe)
	wine "$f" ${WINE_ARGS:-} & pid=$!
	t=0 seen=нет
	while [ $t -lt 120 ]; do
		if xwininfo -root -tree 2>/dev/null | grep -q "(\"$n.exe\""; then seen=да; break; fi
		sleep 0.25; t=$((t+1))
	done
	if [ -n "$shots" ]; then
		sleep 1.5
		rootshot | pnmtopng > "$shots/$n.png"
		kill "$pid" 2>/dev/null
		wait "$pid" 2>/dev/null
		echo "wine: $n — окно: $seen, снимок $shots/$n.png"
		[ $seen = да ] || fail=1
	else
		wait "$pid"; code=$?
		echo "wine: $n — выход с кодом $code"
		[ $code = 0 ] || fail=1
	fi
done
wineserver -k 2>/dev/null
exit $fail
