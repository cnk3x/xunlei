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

# RUN curl -SsL https://github.com/upx/upx/releases/download/v4.2.3/upx-4.2.3-${TARGETARCH}_linux.tar.xz | tar -xJC /bin --strip-components 1 --wildcards '*/upx'

COPY spk/*.spk /spk/

ENV SPK_TARGET=/rootfs/var/packages/pan-xunlei-com/target

# 从spk中提取所需要的文件
RUN arch=$([ "${TARGETARCH}" = "arm64" ] && echo "armv8" || uname -m) \
  && spk=/spk/$(ls /spk | grep "${arch}.spk" | head -n1) \
  && ([ -f "${spk}" ] && echo "spk found ${spk}" || exit 1) \
  && mkdir -p ${SPK_TARGET} \
  && (tar -Oxf ${spk} package.tgz | tar --wildcards -xJC ${SPK_TARGET} 'bin/bin/*' 'ui/index.cgi')

# 提取所需要的so库文件
RUN mkdir -p /rootfs/lib/ \
  && (find ${SPK_TARGET} -type f -exec ldd {} \; 2>/dev/null | grep "=>" | awk '{print $3}' | sort -u | xargs -I {} cp {} /rootfs/lib/)
# RUN upx ${SPK_TARGET}/ui/index.cgi && ls ${SPK_TARGET}/bin/bin/xunlei* | xargs -I {} upx  ${SPK_TARGET}/bin/bin/{}

COPY bin/xlp-${TARGETARCH} /rootfs/bin/xlp
# RUN upx /rootfs/bin/xlp

# 通过一个临时的镜像，过滤掉busybox已经存在的so库文件
FROM busybox:latest as tmp
ARG TARGETARCH

RUN mkdir -p /rootfs/bin \
  && echo "LD_TRACE_LOADED_OBJECTS=1 exec \$@" >/rootfs/bin/ldd && chmod +x /rootfs/bin/ldd

COPY --from=spk /rootfs/ /rootfs/

RUN find /rootfs/lib -maxdepth 1 -type f -exec sh -c '[ -f "/lib/$(basename {})" ] && rm -f {}' \;

FROM busybox:latest
COPY --from=tmp /rootfs/ /

ENV XL_DASHBOARD_PORT=2345 \
  XL_DASHBOARD_HOST= \
  XL_DASHBOARD_USERNAME= \
  XL_DASHBOARD_PASSWORD= \
  XL_DIR_DOWNLOAD=/downloads \
  XL_DIR_DATA=/data \
  XL_LOG=file \
  XL_LOG_MAXSIZE=5M \
  XL_LOG_COMPRESS=true

EXPOSE 2345
VOLUME [ "/data", "/downloads" ]
CMD [ "/bin/xlp" ]
