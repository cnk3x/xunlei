FROM --platform=${TARGETARCH} ubuntu:focal
ARG TARGETARCH
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
  && apt-get install --no-install-recommends -y ca-certificates tzdata \
  && rm -rf /var/lib/apt/lists/* \
  && mkdir -p /rootfs/etc/ssl/certs /rootfs/lib \
  && find /usr/lib \( -name libdl.so.2 -o -name libgcc_s.so.1 -o -name libstdc++.so.6 \) -exec cp -Lr {} /rootfs/lib/ \; \
  && cp -Lr /usr/share/zoneinfo/Asia/Chongqing /rootfs/etc/localtime \
  && echo "Asia/Chongqing" >/rootfs/etc/timezone \
  && cp -Lr --parents /etc/ssl/certs/ca-certificates.crt /rootfs/

COPY artifacts/xlp-${TARGETARCH} /rootfs/xlp
RUN chmod +x /rootfs/xlp

FROM --platform=${TARGETARCH} busybox:1.37
ARG TARGETARCH

LABEL org.opencontainers.image.authors=cnk3x \
  org.opencontainers.image.source=https://github.com/cnk3x/xunlei \
  org.opencontainers.image.description="迅雷远程下载服务(非官方)" \
  org.opencontainers.image.licenses=MIT

COPY --from=0 /rootfs /

ENV \
  XL_DASHBOARD_PORT=2345 \
  XL_DASHBOARD_IP= \
  XL_DASHBOARD_USERNAME= \
  XL_DIR_DOWNLOAD=/xunlei/downloads \
  XL_PREVENT_UPDATE= \
  XL_UID= \
  XL_GID= \
  XL_DEBUG= \
  XL_SPK_URL=

VOLUME [ "/xunlei/data", "/xunlei/var/packages/pan-xunlei-com" ]
EXPOSE 2345

CMD [ "/xlp" ]
