FROM ubuntu:focal as prebuild
ARG TARGETARCH

RUN [ "${TARGETARCH}" = "arm64" -o "${TARGETARCH}" = "amd64" ] || exit 1

RUN sed -i 's@//.*archive.ubuntu.com@//mirrors.ustc.edu.cn@g' /etc/apt/sources.list \
  && sed -i 's/security.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list \
  && apt-get update \
  && apt-get -y --no-install-recommends install ca-certificates tzdata xz-utils curl zstd \
  && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /rootfs \
  && cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs/ \
  && cp /usr/share/zoneinfo/Asia/Shanghai /rootfs/etc/localtime \
  && echo "Asia/Shanghai" >/rootfs/etc/timezone

WORKDIR /spk
COPY spk/*.spk ./

RUN SPK_TARGET=/rootfs/var/packages/pan-xunlei-com/target \
  && mkdir -p ${SPK_TARGET} \
  && SYS_ARCH=$([ "${TARGETARCH}" = "amd64" ] && echo "x86_64" || echo "armv8") \
  && VER=$(ls | grep "${SYS_ARCH}" | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1 | sed 's/v//g') \
  && NAME=$(ls | grep ${SYS_ARCH} | grep ${VER} | head -n1) \
  && [ -f "${NAME}" ] && tar -x -O -f ${NAME} package.tgz | tar -x -J -C ${SPK_TARGET} --wildcards 'bin/bin/*' 'ui/index.cgi' \
  || exit 1

FROM ubuntu:focal
ARG TARGETARCH

COPY --from=prebuild /rootfs/ /
COPY bin/xlp-${TARGETARCH} /bin/xlp

ENV XL_DASHBOARD_PORT=2345 \
  XL_DASHBOARD_HOST= \
  XL_DASHBOARD_USERNAME= \
  XL_DASHBOARD_PASSWORD= \
  XL_DIR_DOWNLOAD=/xunlei/downloads \
  XL_DIR_DATA=/xunlei/data \
  XL_LOG=file \
  XL_LOG_MAXSIZE=5M \
  XL_LOG_COMPRESS=true

EXPOSE 2345
VOLUME [ "/xunlei/data", "/xunlei/downloads" ]
CMD [ "/bin/xlp" ]
