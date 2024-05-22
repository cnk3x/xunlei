FROM --platform=$BUILDPLATFORM debian:12-slim as spk
ARG TARGETARCH

RUN [ "${TARGETARCH}" = "arm64" -o "${TARGETARCH}" = "amd64" ] && echo ok || exit 1

RUN apt-get update \
  && apt-get -y --no-install-recommends install ca-certificates tzdata xz-utils curl \
  && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /rootfs/etc/ssl/certs \
  && cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs/ \
  && cp /usr/share/zoneinfo/Asia/Shanghai /rootfs/etc/localtime \
  && echo "Asia/Shanghai" >/rootfs/etc/timezone

WORKDIR /spk
COPY spk/*.spk ./

ENV SPK_TARGET=/rootfs/var/packages/pan-xunlei-com/target

RUN SYS_ARCH=$([ "${TARGETARCH}" = "amd64" ] && echo "x86_64" || echo "armv8") \
  && VER=$(ls | grep "${SYS_ARCH}" | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1 | sed 's/v//g') \
  && NAME=$(ls | grep ${SYS_ARCH} | grep ${VER} | head -n1) \
  && mkdir -p ${SPK_TARGET} \
  && [ -f "${NAME}" ] && tar -Oxf ${NAME} package.tgz | tar -JxC ${SPK_TARGET} --wildcards 'bin/bin/*' 'ui/index.cgi' \
  || exit 1

# 提取所需要的so库文件
RUN mkdir -p /rootfs/lib/ \
  && (find ${SPK_TARGET} -type f -exec ldd {} \; 2>/dev/null | grep "=>" | awk '{print $3}' | sort -u | xargs -I {} cp {} /rootfs/lib/)

COPY bin/xlp-${TARGETARCH} /rootfs/bin/xlp

# 通过一个临时的镜像，过滤掉busybox已经存在的so库文件
FROM busybox:latest as tmp
ARG TARGETARCH

COPY --from=spk /rootfs/ /rootfs/

RUN find /rootfs/lib -maxdepth 1 -type f -exec sh -c '[ -f "/lib/$(basename {})" ] && rm -f {}' \;

FROM busybox:latest
COPY --from=tmp /rootfs/ /

ENV XL_DASHBOARD_PORT=2345 \
  XL_DASHBOARD_USERNAME= \
  XL_DASHBOARD_PASSWORD= \
  XL_DEBUG=1

EXPOSE 2345
VOLUME [ "/xunlei/data", "/xunlei/downloads" ]
CMD [ "/bin/xlp", "--chroot", "/xunlei" ]
