#!/bin/sh
# Сверка размеров релизных бинарников с size-budget.txt (docs/size.md, §7).
# Бинарники: bin/small/<пример> (TinyGo, после strip) и bin/gc/<пример> — make small release-gc.
# Превышение — код 1. Размер меньше 90 % бюджета — напоминание снизить бюджет.
set -eu
cd "$(dirname "$0")/.."
fail=0
note=""
printf '%-18s %-7s %10s %10s %10s\n' "пример" "тулчейн" "бюджет" "факт" "запас"
while read -r ex tc budget _; do
	case "$ex" in "" | \#*) continue ;; esac
	case "$tc" in
	tinygo) bin=bin/small/${ex##*/} ;;
	gc) bin=bin/gc/${ex##*/} ;;
	*) echo "size: неизвестный тулчейн $tc" >&2; exit 2 ;;
	esac
	if [ ! -f "$bin" ]; then
		echo "size: нет $bin — сначала make small release-gc" >&2
		exit 2
	fi
	size=$(stat -c %s "$bin")
	printf '%-18s %-7s %10s %10s %10s\n' "$ex" "$tc" "$budget" "$size" "$((budget - size))"
	if [ "$size" -gt "$budget" ]; then
		fail=1
	elif [ $((size * 10)) -lt $((budget * 9)) ]; then
		note="$note
size: $ex ($tc) меньше бюджета больше чем на 10 % — снизьте бюджет, чтобы запас не «съели» незаметно"
	fi
done <size-budget.txt
[ -z "$note" ] || echo "$note" | sed '/^$/d'
if [ "$fail" -ne 0 ]; then
	echo "size: превышен бюджет" >&2
	exit 1
fi
echo "size: ok"
