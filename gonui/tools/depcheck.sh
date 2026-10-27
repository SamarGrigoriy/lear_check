#!/bin/sh
# Правила кода библиотеки (docs/size.md, §3; ADR-7, ADR-11), только не-тестовый код nui и internal/*:
#  1. прямые импорты — из белого списка;
#  2. ни одной горутины.
# internal/layout/golden — код golden-тестов (его импортируют только тесты), правила к нему не относятся.
set -eu
cd "$(dirname "$0")/.."
allowed=" unsafe sync sync/atomic errors runtime unicode/utf8 C "
bad=""

# 1. Импорты. .Imports — без тестовых (.TestImports не смотрим): тестам можно всё.
#    Обе ОС драйвера: у каждой свои файлы *_linux.go и *_windows.go.
list=$(for os in linux windows; do
	GOOS=$os CGO_ENABLED=1 go list -f '{{.ImportPath}}{{range .Imports}} {{.}}{{end}}' . ./internal/...
done)
while read -r pkg imports; do
	[ "$pkg" = nui/internal/layout/golden ] && continue
	for imp in $imports; do
		case "$imp" in
		nui/internal/layout/golden)
			bad="$bad
deps: $pkg импортирует код тестов $imp"
			continue ;;
		nui/internal/*) continue ;;
		esac
		case "$allowed" in *" $imp "*) continue ;; esac
		bad="$bad
deps: $pkg импортирует $imp — нет в белом списке"
	done
done <<EOF
$list
EOF

# 2. Горутины: оператор go в коде (строки-комментарии пропускаем).
for f in $(find . -maxdepth 1 -name '*.go' ! -name '*_test.go') $(find internal -name '*.go' ! -name '*_test.go' ! -path 'internal/layout/golden/*'); do
	hits=$(grep -nE '(^|[^[:alnum:]_.])go +(func|[[:alnum:]_.]+ *\()' "$f" | grep -vE '^[0-9]+:[[:space:]]*//' || true)
	[ -z "$hits" ] || bad="$bad
deps: горутина в библиотеке (ADR-7): $f:$hits"
done

if [ -n "$bad" ]; then
	echo "$bad" | sed '/^$/d' >&2
	exit 1
fi
echo "deps: ok"
