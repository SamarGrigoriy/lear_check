#!/bin/sh
# S4: hello на Win32 под gc и TinyGo — кросс-сборка mingw-w64 на Debian 12, запуск под Wine
# в Xvfb с openbox. Запуск из корня репозитория:
#   NUI_IMAGE=nui-win tools/docker-run.sh sh spike/s4/run.sh
# Результаты: out/ — exe и report.txt; shots/ — снимки фотоотчёта.
set -u
cd "$(dirname "$0")"
mkdir -p out shots
export GOTOOLCHAIN=local GOFLAGS=-buildvcs=false
export WINEPREFIX=/cache/wine WINEDEBUG=-all WINEDLLOVERRIDES="mscoree,mshtml="
fail=0
check() { if "$@"; then :; else echo "ОШИБКА: $*"; fail=1; fi; }

{
echo "== Система"
. /etc/os-release; echo "$PRETTY_NAME; $(x86_64-w64-mingw32-gcc --version | head -1)"
echo "$(go version); $(tinygo version 2>/dev/null || echo 'TinyGo не смонтирован')"
echo "$(wine --version)"

echo "== 1. Прототипы против <windows.h> mingw-w64 (NUI_CHECK)"
check x86_64-w64-mingw32-gcc -std=c11 -DNUI_CHECK -Wall -Werror -fsyntax-only \
	-include windows.h -include commctrl.h -x c win32.h
check x86_64-w64-mingw32-gcc -std=c11 -DNUI_CHECK -Wall -Werror -fsyntax-only \
	-include windows.h -include commctrl.h wrap.c
echo "прототипы: проверены"

echo "== 2. Ресурсы: манифест → .syso"
check x86_64-w64-mingw32-windres app.rc -O coff -o rsrc_windows_amd64.syso
cp rsrc_windows_amd64.syso rsrc.o # для TinyGo — через #cgo LDFLAGS (rsrc_tinygo.go)

echo "== 3. Сборка gc (cgo, mingw-w64)"
check env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_CFLAGS="-Os -ffunction-sections" \
	go build -trimpath -ldflags "-s -w -H windowsgui -buildid=" -o out/s4-gc.exe .

# Для сравнения в фотоотчёте — тот же exe без манифеста: .syso на время убирается.
mv rsrc_windows_amd64.syso rsrc.syso.off
check env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_CFLAGS="-Os -ffunction-sections" \
	go build -trimpath -ldflags "-s -w -H windowsgui -buildid=" -o out/s4-nomanifest.exe .
mv rsrc.syso.off rsrc_windows_amd64.syso

echo "== 4. Сборка TinyGo"
if command -v tinygo >/dev/null; then
	check env GOOS=windows GOARCH=amd64 tinygo build -opt=z -no-debug -panic=trap -o out/s4-tinygo-console.exe .
	# lld получает флаги #cgo как есть, а -Wl,--subsystem не понимает: подсистема GUI — после
	# линковки; заодно убираем таблицу символов COFF, которую -no-debug оставляет.
	check x86_64-w64-mingw32-objcopy --subsystem windows --strip-all out/s4-tinygo-console.exe out/s4-tinygo.exe
	rm -f out/s4-tinygo-console.exe
	# Другие планировщики и сборщики мусора TinyGo 0.42 под Windows не линкуются.
	for v in "none conservative" "threads conservative"; do
		set -- $v
		if env GOOS=windows GOARCH=amd64 tinygo build -opt=z -no-debug -panic=trap -scheduler=$1 -gc=$2 \
			-o out/v.exe . >out/v.log 2>&1; then echo "TinyGo -scheduler=$1 -gc=$2: собирается"
		else echo "TinyGo -scheduler=$1 -gc=$2: не собирается — $(grep -m1 -E 'error|undefined' out/v.log)"; fi
		rm -f out/v.exe out/v.log
	done
fi

echo "== 5. Что внутри exe"
for f in out/*.exe; do
	[ -f "$f" ] || continue
	sub=$(x86_64-w64-mingw32-objdump -p "$f" | sed -n 's/^Subsystem[[:space:]]*//p')
	dlls=$(x86_64-w64-mingw32-objdump -p "$f" | sed -n 's/^[[:space:]]*DLL Name: //p' | sort -u | tr '\n' ' ')
	rsrc=$(x86_64-w64-mingw32-objdump -h "$f" | grep -c '\.rsrc' || true)
	echo "$f: $(stat -c %s "$f") байт; подсистема $sub; .rsrc: $rsrc; DLL: $dlls"
done

echo "== 6. Запуск под Wine (Xvfb + openbox), selftest"
# В Wine 8.0 нет bcryptprimitives.dll, без неё рантайм Go ≥ 1.22 не стартует (на Windows 10+
# она есть) — в префикс кладётся заглушка ProcessPrng → RtlGenRandom (wine/bcryptprimitives.c).
check x86_64-w64-mingw32-gcc -O2 -shared -o out/bcryptprimitives.dll wine/bcryptprimitives.c \
	wine/bcryptprimitives.def -ladvapi32
xvfb-run -a -s "-screen 0 640x360x24" sh -c '
	openbox >/dev/null 2>&1 & sleep 1
	wineboot -i >/dev/null 2>&1
	cp out/bcryptprimitives.dll "$WINEPREFIX/drive_c/windows/system32/"
	for f in out/*.exe; do
		n=$(basename "$f" .exe)
		[ "$n" = bcryptprimitives ] && continue
		wine "$f" -selftest & pid=$!
		# Ищем по классу окна (имя exe): заголовок не в ASCII, и в локали C xwininfo его не печатает.
		t=0 seen=нет
		while [ $t -lt 60 ]; do
			if xwininfo -root -tree 2>/dev/null | grep -q "(\"$n.exe\""; then seen=да; break; fi
			sleep 0.25; t=$((t+1))
		done
		sleep 1.5 # таймер selftest успевает нажать кнопку
		rootshot | pnmtopng > shots/$n.png
		wait $pid; code=$?
		echo "$n: окно появилось — $seen, выход с кодом $code"
		[ $seen = да ] && [ $code = 0 ] || echo "ОШИБКА: $n"
	done
' 2>&1 | grep -v '^$'
echo "== Итог: $([ $fail = 0 ] && echo ok || echo ОШИБКИ)"
} 2>&1 | tee out/report.txt
