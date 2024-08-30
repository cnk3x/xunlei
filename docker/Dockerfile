FROM --platform=$BUILDPLATFORM ubuntu:focal AS temp
ARG TARGETARCH

RUN [ "${TARGETARCH}" = "arm64" -o "${TARGETARCH}" = "amd64" ] && echo ok || exit 1

ENV LANG=C.UTF-8 LANG=zh_CN.UTF-8 LANGUAGE=zh_CN.UTF-8 LC_ALL=C

RUN sed -i 's/archive.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/security.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/ports.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && DEBIAN_FRONTEND=noninteractive apt-get update && apt-get -y --no-install-recommends install tzdata ca-certificates xz-utils \
  && rm -rf /var/lib/apt/lists/* \
  && echo "Asia/Shanghai" >/etc/timezone \
  && rm -f /etc/localtime && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN mkdir -p /rootfs/etc/ssl/certs \
  && cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs/ \
  && cp --parents /etc/timezone /rootfs/ \
  && cp --parents /etc/localtime /rootfs/ \
  && cp --parents /etc/apt/sources.list /rootfs/

WORKDIR /spk

# COPY spk/*.spk ./
# RUN SYS_ARCH=$([ "${TARGETARCH}" = "amd64" ] && echo "x86_64" || echo "armv8") \
#   && VER=$(ls | grep "${SYS_ARCH}" | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1) \
#   && NAME=$(ls | grep ${SYS_ARCH} | grep ${VER} | head -n1) \
#   && mkdir -p /rootfs/var/packages/pan-xunlei-com/target \
#   && [ -f "${NAME}" ] && tar -Oxf ${NAME} package.tgz | tar -JxC /rootfs/var/packages/pan-xunlei-com/target --wildcards 'bin/bin/*' 'ui/index.cgi' \
#   || exit 1

COPY bin/xlp-${TARGETARCH} /rootfs/usr/bin/xlp

FROM ubuntu:focal

LABEL org.opencontainers.image.authors=cnk3x
LABEL org.opencontainers.image.source=https://github.com/cnk3x/xunlei

COPY --from=temp /rootfs/ /

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
