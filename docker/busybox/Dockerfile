FROM --platform=$BUILDPLATFORM ubuntu:focal AS temp
ARG TARGETARCH

RUN [ "${TARGETARCH}" = "arm64" -o "${TARGETARCH}" = "amd64" ] && echo ok || exit 1

ENV LANG=C.UTF-8 LANG=zh_CN.UTF-8 LANGUAGE=zh_CN.UTF-8 LC_ALL=C

RUN sed -i 's/archive.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/security.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/ports.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && DEBIAN_FRONTEND=noninteractive apt-get update && apt-get -y --no-install-recommends install \
  tzdata ca-certificates xz-utils tree upx-ucl \
  && rm -rf /var/lib/apt/lists/* \
  && echo "Asia/Shanghai" >/etc/timezone \
  && rm -f /etc/localtime && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN mkdir -p /rootfs/etc/ssl/certs \
  && cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs/ \
  && cp --parents /etc/timezone /rootfs/ \
  && cp --parents /etc/localtime /rootfs/ \
  && cp --parents /etc/apt/sources.list /rootfs/

# busybox
COPY bin/xlp-${TARGETARCH} /rootfs/bin/xlp
# RUN upx /rootfs/bin/xlp >/dev/null

# copy libs
COPY spk/nasxunlei-*.spk /spk/
RUN tar -v -x -O -f /spk/nasxunlei-$(["${TARGETARCH}" = "arm64"] && echo "armv8" || echo "x86_64").spk package.tgz \
  | tar -v -x -J -C /spk --wildcards 'bin/bin/xunlei-pan-cli*' 'ui/index.cgi' \
  && mkdir -p /rootfs/lib \
  && ldd /spk/bin/bin/* /spk/ui/* | grep '=>' | awk '{ cmd="cp -L "$3" /rootfs/lib/"$1; print cmd; system(cmd) }'

# 中转，避免覆盖lib
FROM busybox:1.36.1 AS libs

COPY --from=temp /rootfs/ /rootfs/
RUN cp -Lf /lib/* /rootfs/lib/

FROM busybox:1.36.1

LABEL org.opencontainers.image.authors=cnk3x
LABEL org.opencontainers.image.source=https://github.com/cnk3x/xunlei

COPY --from=libs /rootfs/ /

ENV LANG=C.UTF-8 LANG=zh_CN.UTF-8 LANGUAGE=zh_CN.UTF-8 LC_ALL=C \
  XL_DEBUG=false \
  XL_DASHBOARD_PORT=2345 \
  XL_DASHBOARD_USERNAME= \
  XL_DIR_DOWNLOAD=/xunlei/downloads \
  XL_DIR_DATA=/xunlei/data \
  XL_DASHBOARD_PASSWORD= \
  XL_CHROOT=/xunlei \
  XL_PREVENT_UPDATE=true \
  XL_UID= \
  XL_GID=

CMD [ "/bin/xlp" ]
