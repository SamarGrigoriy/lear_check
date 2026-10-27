# nui — сборка и проверки (docs/build.md, §4).
#
# Цели, которым нужны GTK под Xvfb и TinyGo, проще всего гонять в чистом Debian 12:
#   make docker-ci           — всё, что не требует libgtk-3-dev, в контейнере nui-bookworm
#   make ci                  — check на хосте (нужен libgtk-3-dev) + docker-ci
#   tools/docker-run.sh make test   — любая цель в контейнере

GO           ?= go
TINYGO       ?= tinygo
TINYGO_SCHED ?= none
ARCH         ?= amd64
EXAMPLES     := hello form chat viewer
GTK_DEBUG_ENV := NO_AT_BRIDGE=1 G_DEBUG=fatal-criticals
XVFB         ?= xvfb-run -a

# Флаги размера C-части — только через окружение: в #cgo gc их не пропускает (driver-gtk3.md, §9).
export CGO_CFLAGS  := -Os -ffunction-sections -fdata-sections
export CGO_LDFLAGS := -Wl,--gc-sections -Wl,--as-needed

GOFLAGS_REL := -trimpath -ldflags="-s -w -buildid="
# Выпущенный TinyGo + target-файл nui: системная glibc вместо musl (ADR-17).
TINYFLAGS   := -target=targets/linux-glibc-$(ARCH).json -opt=z -no-debug -panic=trap -gc=conservative -scheduler=$(TINYGO_SCHED)

.PHONY: all dev release-gc small test race selftest deps size abi check-protos vet check ci ci-container docker-ci shots golden win win-selftest win-test win-shots clean FORCE

all: dev

dev: $(EXAMPLES:%=bin/%)                     ## gc, для отладки
bin/%: FORCE
	$(GO) build -o $@ ./examples/$*

release-gc: $(EXAMPLES:%=bin/gc/%)           ## gc с релизными флагами
bin/gc/%: FORCE
	$(GO) build $(GOFLAGS_REL) -o $@ ./examples/$*

small: $(EXAMPLES:%=bin/small/%)             ## TinyGo, релиз
# strip обязателен: -no-debug убирает DWARF, но оставляет таблицу символов (size.md, §4).
bin/small/%: FORCE
	$(TINYGO) build $(TINYFLAGS) -o $@ ./examples/$*
	strip --strip-all $@

test:                                        ## юнит-тесты ядра и раскладки + интеграционные на GTK (обычная и отладочная сборка)
	$(GO) test -count=1 -cover ./internal/core/... ./internal/layout
	$(GTK_DEBUG_ENV) $(XVFB) $(GO) test -count=1 -tags gui .
	$(GTK_DEBUG_ENV) $(XVFB) $(GO) test -count=1 -tags 'gui nuidebug' .

race:                                        ## то же под детектором гонок (только gc)
	$(GO) test -count=1 -race ./internal/core/...
	$(GTK_DEBUG_ENV) $(XVFB) $(GO) test -count=1 -race -tags gui .

selftest:                                    ## примеры обоими тулчейнами: окно, цикл, выход с кодом 0
	@mkdir -p bin/selftest
	@for e in $(EXAMPLES); do \
		$(GO) build -tags selftest -o bin/selftest/$$e-gc ./examples/$$e && \
		$(TINYGO) build $(TINYFLAGS) -tags selftest -o bin/selftest/$$e-tinygo ./examples/$$e && \
		for t in gc tinygo; do \
			$(GTK_DEBUG_ENV) timeout 20 $(XVFB) ./bin/selftest/$$e-$$t || { echo "selftest: $$e ($$t) — ошибка"; exit 1; }; \
			echo "selftest: $$e ($$t) — ok"; \
		done; \
	done

deps:                                        ## белый список импортов, без горутин в библиотеке
	./tools/depcheck.sh

size: release-gc small                       ## размеры против size-budget.txt
	./tools/sizecheck.sh

abi: release-gc small                        ## glibc не новее Debian 12
	./tools/abicheck.sh

check-protos:                                ## прототипы против настоящих заголовков (нужен libgtk-3-dev)
	./tools/check-protos.sh

vet:
	$(GO) vet ./...
	$(GO) vet -tags gui .
	$(GO) vet -tags 'gui nuidebug' .

check: vet deps check-protos                 ## статические проверки на хосте

ci-container: deps test race selftest size abi   ## всё, что не требует libgtk-3-dev (внутри контейнера)

docker-ci:                                   ## ci-container в чистом Debian 12
	./tools/docker-run.sh make ci-container

ci: check docker-ci

# Golden-снимки раскладки (layout.md, §8): GTK раскладывает случаи из
# internal/layout/testdata/cases, снимки ложатся в testdata/golden. Только в контейнере —
# снимки не должны зависеть от шрифтов и версии GTK машины: tools/docker-run.sh make golden
golden:
	LANG=ru_RU.UTF-8 $(GTK_DEBUG_ENV) $(XVFB) -s "-screen 0 1024x768x24" sh -c '\
		openbox >/dev/null 2>&1 & sleep 1; \
		NUI_GOLDEN=write $(GO) test -count=1 -tags gui -run TestLayoutGTK .'
	$(GO) test -count=1 ./internal/layout

# Фотоотчёт: окна примеров (обоими тулчейнами) и окна тестов под Xvfb с openbox.
# Запускать в контейнере: tools/docker-run.sh make shots SHOTS=docs/reports/phase-N
SHOTS       ?= docs/reports/phase-5
SHOTS_TESTS ?= TestLayoutGTK
shots: release-gc small
	@mkdir -p $(SHOTS)
	LANG=ru_RU.UTF-8 $(GTK_DEBUG_ENV) $(XVFB) -s "-screen 0 800x600x24" sh -c '\
		openbox >/dev/null 2>&1 & sleep 1; \
		for e in $(EXAMPLES); do for t in gc small; do \
			./bin/$$t/$$e & pid=$$!; sleep 1.5; rootshot | pnmtopng > $(SHOTS)/$$e-$$t.png; kill $$pid; sleep 0.5; \
		done; done; \
		NUI_SHOTS=$(SHOTS) $(GO) test -count=1 -tags gui -run "$(SHOTS_TESTS)" .'
	@# Часть ручного чек-листа (testing.md, §7) — автоматически: тёмная тема и масштаб 2×.
	LANG=ru_RU.UTF-8 $(GTK_DEBUG_ENV) GTK_THEME=Adwaita:dark $(XVFB) -s "-screen 0 800x600x24" sh -c '\
		openbox >/dev/null 2>&1 & sleep 1; ./bin/small/form & pid=$$!; sleep 1.5; \
		rootshot | pnmtopng > $(SHOTS)/form-dark.png; kill $$pid'
	LANG=ru_RU.UTF-8 $(GTK_DEBUG_ENV) GDK_SCALE=2 $(XVFB) -s "-screen 0 1100x1100x24" sh -c '\
		openbox >/dev/null 2>&1 & sleep 1; ./bin/small/form & pid=$$!; sleep 1.5; \
		rootshot | pnmtopng > $(SHOTS)/form-scale2.png; kill $$pid'

# Windows (docs/driver-win32.md): кросс-сборка mingw-w64 и проверка под Wine — в образе nui-win:
#   NUI_IMAGE=nui-win tools/docker-run.sh make win win-selftest
# Флаги размера C-части для mingw свои: -fdata-sections раздувает exe (driver-win32.md, §1).
WIN_PREFIX ?= x86_64-w64-mingw32-
WIN_GO     := CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=$(WIN_PREFIX)gcc CGO_CFLAGS="-Os -ffunction-sections" CGO_LDFLAGS=
WIN_TINY   := GOOS=windows GOARCH=amd64 CGO_CFLAGS= CGO_LDFLAGS= $(TINYGO) build -opt=z -no-debug -panic=trap
WIN_SHOTS  ?= docs/reports/phase-5

win: $(EXAMPLES:%=bin/win/%.exe) $(EXAMPLES:%=bin/win/%-small.exe)   ## Windows: gc и TinyGo
# Манифест (app.rc → .syso для gc, rsrc.o для TinyGo: он не подхватывает .syso).
examples/%/rsrc_windows_amd64.syso: examples/%/app.rc examples/%/app.manifest
	cd examples/$* && $(WIN_PREFIX)windres app.rc -O coff -o rsrc_windows_amd64.syso && cp rsrc_windows_amd64.syso rsrc.o
bin/win/%.exe: examples/%/rsrc_windows_amd64.syso FORCE
	$(WIN_GO) $(GO) build -trimpath -ldflags="-s -w -H windowsgui -buildid=" -o $@ ./examples/$*
# TinyGo: подсистема GUI — после линковки (ld.lld не понимает -Wl,--subsystem из #cgo).
bin/win/%-small.exe: examples/%/rsrc_windows_amd64.syso FORCE
	@mkdir -p bin/win
	$(WIN_TINY) -o $@.tmp ./examples/$*
	$(WIN_PREFIX)objcopy --subsystem windows --strip-all $@.tmp $@ && rm -f $@.tmp

win-selftest: $(EXAMPLES:%=examples/%/rsrc_windows_amd64.syso)   ## примеры под Wine: окно, цикл, выход с кодом 0
	@mkdir -p bin/win/selftest
	@for e in $(EXAMPLES); do \
		$(WIN_GO) $(GO) build -tags selftest -ldflags="-H windowsgui" -o bin/win/selftest/$$e.exe ./examples/$$e && \
		$(WIN_TINY) -tags selftest -o bin/win/selftest/$$e-small.tmp ./examples/$$e && \
		$(WIN_PREFIX)objcopy --subsystem windows bin/win/selftest/$$e-small.tmp bin/win/selftest/$$e-small.exe || exit 1; \
	done
	./tools/wine-run.sh bin/win/selftest/*.exe

# Интеграционные тесты nui (-tags gui) на драйвере Win32 — те же, что на GTK, без
# привязанных к Linux (*_linux_test.go); обычная и отладочная сборка.
win-test:                                    ## интеграционные тесты под Wine
	@mkdir -p bin/win
	$(WIN_GO) CGO_CFLAGS= $(GO) test -c -tags gui -o bin/win/nui.test.exe .
	$(WIN_GO) CGO_CFLAGS= $(GO) test -c -tags 'gui nuidebug' -o bin/win/nui-debug.test.exe .
	WINE_ARGS="-test.timeout=300s" ./tools/wine-run.sh bin/win/nui.test.exe bin/win/nui-debug.test.exe

win-shots: win                               ## снимки примеров под Wine: 100 % и 150 %
	@mkdir -p $(WIN_SHOTS)
	./tools/wine-run.sh -shots $(WIN_SHOTS) bin/win/*.exe
	@mkdir -p bin/win/dpi144 && for f in bin/win/*.exe; do cp $$f bin/win/dpi144/$$(basename $$f .exe)-150.exe; done
	./tools/wine-run.sh -dpi 144 -shots $(WIN_SHOTS) bin/win/dpi144/*.exe

clean:
	rm -rf bin

FORCE:
