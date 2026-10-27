# Образ для драйвера Win32 (фаза 5): кросс-компилятор mingw-w64 и Wine, чтобы собирать
# и запускать Windows-программы на Linux под Xvfb. Снимки — тем же rootshot.
#   docker build -t nui-win -f tools/win.Dockerfile tools
FROM debian:bookworm
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -qq \
 && apt-get install -y -qq --no-install-recommends \
      gcc libc6-dev make binutils gcc-mingw-w64-x86-64 binutils-mingw-w64-x86-64 \
      wine wine64 xvfb xauth openbox x11-utils netpbm fonts-dejavu-core locales \
 && sed -i 's/^# *ru_RU.UTF-8/ru_RU.UTF-8/' /etc/locale.gen && locale-gen \
 && rm -rf /var/lib/apt/lists/*
COPY rootshot.c /tmp/rootshot.c
RUN gcc -std=gnu11 -O2 /tmp/rootshot.c -o /usr/local/bin/rootshot -l:libX11.so.6 && rm /tmp/rootshot.c
