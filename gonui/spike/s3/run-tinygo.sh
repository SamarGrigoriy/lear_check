#!/bin/sh
# S3: штатный TinyGo + свой target-файл (linux-glibc-amd64.json) вместо PR #5838.
# Собирает спайк S1 (../s1) в чистом Debian 12 (образ nui-bookworm) и проверяет критерии S2.
#   TINYGOROOT — распакованный выпуск TinyGo (по умолчанию ~/.cache/nui/tinygo-0.42.0/x/tinygo)
#   STRESS     — длительность проверки сборщика мусора, с (по умолчанию 60; в roadmap — 600)
# Результаты: out/ — бинарники и report.txt; shots/ — снимки фотоотчёта.
set -eu
cd "$(dirname "$0")"
TINYGOROOT=${TINYGOROOT:-$HOME/.cache/nui/tinygo-0.42.0/x/tinygo}
STRESS=${STRESS:-60}
docker build -q -t nui-bookworm -f ../../tools/bookworm.Dockerfile ../../tools >/dev/null
mkdir -p out shots
GOROOT_HOST=$(go env GOROOT)
GOSHARE=$(dirname "$(readlink -f "$GOROOT_HOST/src")")
docker run --rm \
	-v "$PWD/..":/spike -w /spike/s1 \
	-v "$TINYGOROOT":/opt/tinygo:ro \
	-v "$GOROOT_HOST":"$GOROOT_HOST":ro -v "$GOSHARE":"$GOSHARE":ro \
	-e HOST_UID="$(id -u)" -e HOST_GID="$(id -g)" -e GOROOT_HOST="$GOROOT_HOST" -e STRESS="$STRESS" \
	nui-bookworm sh -c '
set -eu
export PATH=/opt/tinygo/bin:$GOROOT_HOST/bin:$PATH GOTOOLCHAIN=local GOFLAGS=-buildvcs=false
T=../s3/linux-glibc-amd64.json
O=../s3/out
REL="-opt=z -no-debug -panic=trap"
X="xvfb-run -a"

fail=0
{
echo "== Система"
. /etc/os-release; echo "$PRETTY_NAME, $(ldd --version | head -1)"
echo "$(tinygo version)"

echo "== 1. Сборка и линковка (штатный TinyGo, свой target, без libgtk-3-dev)"
if dpkg -s libgtk-3-dev >/dev/null 2>&1; then echo "ОШИБКА: libgtk-3-dev установлен"; exit 1; fi
tinygo build -target=$T -o $O/s1-threads .
tinygo build -target=$T -scheduler=none -o $O/s1-none .
tinygo build -target=$T -tags shots -o $O/s1-shots .
echo "сборка: ok"
readelf -d $O/s1-none | grep NEEDED | sed "s/.*\[\(.*\)\]/NEEDED \1/"
readelf -l $O/s1-none | grep -o "interpreter: [^]]*"

echo "== 2–5, 7. Сценарий, threads (Do из горутины), G_DEBUG=fatal-criticals"
rc=0; NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals timeout 30 $X $O/s1-threads -auto >$O/threads.log 2>&1 || rc=$?
cat $O/threads.log; [ "$rc" -eq 0 ] || { echo "ОШИБКА: код $rc"; exit 1; }

echo "== 2–5. Сценарий, scheduler=none (Do из таймера), G_DEBUG=fatal-criticals"
rc=0; NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals timeout 30 $X $O/s1-none -auto >$O/none.log 2>&1 || rc=$?
cat $O/none.log; [ "$rc" -eq 0 ] || { echo "ОШИБКА: код $rc"; exit 1; }

echo "== 6. Сборщик мусора: $STRESS с, threads"
rc=0; NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals timeout $((STRESS + 60)) $X $O/s1-threads -stress $STRESS >$O/stress.log 2>&1 || rc=$?
cat $O/stress.log; [ "$rc" -eq 0 ] || { echo "ОШИБКА: код $rc"; exit 1; }

echo "== Фотоотчёт: threads, openbox, LANG=ru_RU.UTF-8"
rc=0
LANG=ru_RU.UTF-8 NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals timeout 30 \
	xvfb-run -a -s "-screen 0 640x360x24" sh -c "openbox >/dev/null 2>&1 & sleep 1; $O/s1-shots -auto -shots ../s3/shots" \
	>$O/shots.log 2>&1 || rc=$?
[ "$rc" -eq 0 ] || { cat $O/shots.log; echo "ОШИБКА: код $rc"; exit 1; }
ls ../s3/shots

echo "== 8. Размер (релиз: $REL)"
tinygo build -target=$T -scheduler=none $REL -o $O/s1-none-rel .
tinygo build -target=$T $REL -o $O/s1-threads-rel .
for f in s1-none-rel s1-threads-rel; do
	cp $O/$f $O/$f-stripped && strip --strip-all $O/$f-stripped
	echo "$f: $(stat -c %s $O/$f) байт, после strip: $(stat -c %s $O/$f-stripped) байт"
done
rc=0; NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals timeout 30 $X $O/s1-none-rel-stripped -auto >$O/none-rel.log 2>&1 || rc=$?
echo "релизная сборка none после strip проходит сценарий: $([ "$rc" -eq 0 ] && echo да || echo "НЕТ (код $rc)")"
[ "$rc" -eq 0 ] || exit 1

echo "== Версия glibc"
echo "максимальная GLIBC_: $(objdump -T $O/s1-none-rel | grep -o "GLIBC_[0-9.]*" | sort -uV | tail -1)"
} >$O/report.txt 2>&1 || fail=1
cat $O/report.txt
chown -R "$HOST_UID:$HOST_GID" ../s3/out ../s3/shots 2>/dev/null || true
exit $fail
'
