#!/bin/sh
# Критерии S1 в чистом Debian 12: только runtime-пакет GTK, без libgtk-3-dev.
# Образ nui-bookworm (tools/bookworm.Dockerfile) собирается один раз; Go монтируется с хоста
# (он собран статически). В Debian-пакете Go каталоги src, lib и др. — симлинки
# в /usr/share/go-*, поэтому монтируется и он, по тем же путям.
# Результаты: out/ — бинарники, журналы и report.txt; shots/ — снимки для фотоотчёта.
set -eu
cd "$(dirname "$0")"
docker build -q -t nui-bookworm -f ../../tools/bookworm.Dockerfile ../../tools >/dev/null
mkdir -p out
GOROOT_HOST=$(go env GOROOT)
GOSHARE=$(dirname "$(readlink -f "$GOROOT_HOST/src")")
docker run --rm \
	-v "$PWD":/src -w /src \
	-v "$GOROOT_HOST":"$GOROOT_HOST":ro -v "$GOSHARE":"$GOSHARE":ro \
	-e HOST_UID="$(id -u)" -e HOST_GID="$(id -g)" -e GOROOT_HOST="$GOROOT_HOST" \
	nui-bookworm sh -c '
set -eu
export PATH=$GOROOT_HOST/bin:$PATH GOTOOLCHAIN=local GOFLAGS=-buildvcs=false

fail=0
{
echo "== Система"
. /etc/os-release; echo "$PRETTY_NAME, $(ldd --version | head -1)"
echo "GTK: $(dpkg-query -W -f="\${Version}" libgtk-3-0), gcc: $(gcc -dumpversion), $(go version)"

echo "== 1. Сборка без libgtk-3-dev"
if dpkg -s libgtk-3-dev >/dev/null 2>&1; then echo "ОШИБКА: libgtk-3-dev установлен"; exit 1; fi
echo "libgtk-3-dev не установлен"
go build -trimpath -ldflags="-s -w -buildid=" -o out/s1-plain .
CGO_CFLAGS="-Os -ffunction-sections -fdata-sections" CGO_LDFLAGS="-Wl,--gc-sections -Wl,--as-needed" \
	go build -trimpath -ldflags="-s -w -buildid=" -o out/s1 .
go build -tags shots -o out/s1-shots . # для фотоотчёта; в замер размера не входит
echo "сборка: ok"

echo "== 2. Линковка по soname"
readelf -d out/s1 | grep NEEDED | sed "s/.*\[\(.*\)\]/\1/"

echo "== 3–4. Сценарий под Xvfb, G_DEBUG=fatal-criticals"
rc=0
NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals timeout 30 xvfb-run -a ./out/s1 -auto >out/auto.log 2>&1 || rc=$?
cat out/auto.log
[ "$rc" -eq 0 ] || { echo "ОШИБКА: сценарий завершился с кодом $rc"; exit 1; }

echo "== Фотоотчёт: тот же сценарий с оконным менеджером, LANG=ru_RU.UTF-8"
mkdir -p shots
rc=0
LANG=ru_RU.UTF-8 NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals timeout 30 \
	xvfb-run -a -s "-screen 0 640x360x24" sh -c "openbox >/dev/null 2>&1 & sleep 1; ./out/s1-shots -auto -shots shots" \
	>out/shots.log 2>&1 || rc=$?
[ "$rc" -eq 0 ] || { cat out/shots.log; echo "ОШИБКА: прогон со снимками завершился с кодом $rc"; exit 1; }
ls shots

echo "== 6–7. Размер (-s -w)"
echo "без флагов C-части:  $(stat -c %s out/s1-plain) байт"
echo "с CGO_CFLAGS/LDFLAGS: $(stat -c %s out/s1) байт"

echo "== 8. Версия glibc"
echo "максимальная GLIBC_: $(objdump -T out/s1 | grep -o "GLIBC_[0-9.]*" | sort -uV | tail -1)"
} >out/report.txt 2>&1 || fail=1
cat out/report.txt
chown -R "$HOST_UID:$HOST_GID" out shots 2>/dev/null || true
exit $fail
'
