#!/bin/sh
# Бинарник требует glibc не старее той, на которой собран (версии символов GLIBC_*).
# Релиз должен запускаться на Debian 12, поэтому максимум — GLIBC_2.36 (docs/driver-gtk3.md, §9).
set -eu
cd "$(dirname "$0")/.."
limit=${GLIBC_LIMIT:-2.36}
fail=0
found=0
for f in bin/gc/* bin/small/*; do
	[ -f "$f" ] || continue
	found=1
	v=$(objdump -T "$f" | grep -o 'GLIBC_[0-9.]*' | sort -uV | tail -1)
	v=${v#GLIBC_}
	top=$(printf '%s\n%s\n' "$v" "$limit" | sort -V | tail -1)
	if [ "$top" = "$limit" ]; then
		echo "abi: $f — GLIBC_$v"
	else
		echo "abi: $f — GLIBC_$v новее $limit: собирайте релиз на Debian 12" >&2
		fail=1
	fi
done
[ "$found" -eq 1 ] || { echo "abi: нет бинарников — сначала make small release-gc" >&2; exit 2; }
exit $fail
