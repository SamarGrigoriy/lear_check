# Чистый Debian 12 с runtime GTK — для прогонов спайков. libgtk-3-dev здесь нет намеренно (ADR-3).
# Go в образ не ставится: скрипты монтируют Go с хоста (он собран статически).
#   docker build -t nui-bookworm -f tools/bookworm.Dockerfile tools
FROM debian:bookworm
ENV DEBIAN_FRONTEND=noninteractive
# Сборка и запуск: gcc, runtime GTK, Xvfb. Фотоотчёт: openbox (рамка с заголовком),
# netpbm (PPM → PNG), шрифты с кириллицей и эмодзи, русская локаль (кнопки «Да/Нет»).
# x11-apps и x11-utils (xwd, xwininfo, xprop) — для отладки.
RUN apt-get update -qq \
 && apt-get install -y -qq --no-install-recommends \
      gcc libc6-dev libgtk-3-0 xvfb xauth binutils \
      openbox x11-apps x11-utils netpbm fonts-dejavu-core fonts-noto-color-emoji adwaita-icon-theme locales \
 && sed -i 's/^# *ru_RU.UTF-8/ru_RU.UTF-8/' /etc/locale.gen && locale-gen \
 && rm -rf /var/lib/apt/lists/*
# Снимки экрана — своим rootshot, а не xwd -root: xwd теряет окна GTK с не-умолчательным
# визуалом (модальный диалог выходит чёрным). Подробности — в rootshot.c.
COPY rootshot.c /tmp/rootshot.c
RUN gcc -std=gnu11 -O2 /tmp/rootshot.c -o /usr/local/bin/rootshot -l:libX11.so.6 && rm /tmp/rootshot.c
# make — чтобы гонять цели Makefile в контейнере (make docker-ci). Отдельный слой, чтобы не перекачивать всё.
RUN apt-get update -qq \
 && apt-get install -y -qq --no-install-recommends make \
 && rm -rf /var/lib/apt/lists/*
# Загрузчик SVG для gdk-pixbuf: им Adwaita рисует точку переключателя и галочку флажка
# (bullet-symbolic.svg, check-symbolic.svg). На настоящем рабочем столе он всегда есть.
RUN apt-get update -qq \
 && apt-get install -y -qq --no-install-recommends librsvg2-common \
 && rm -rf /var/lib/apt/lists/*
