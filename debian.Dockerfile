FROM --platform=${TARGETARCH} debian:stable-slim
ARG TARGETARCH

LABEL org.opencontainers.image.authors=cnk3x
LABEL org.opencontainers.image.source=https://github.com/cnk3x/xunlei

RUN apt update && apt install --no-install-recommends -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/* && \
  rm -f /etc/localtime /etc/timezone && \
  cp -Lr /usr/share/zoneinfo/Asia/Chongqing /etc/localtime && \
  echo "Asia/Chongqing" >/etc/timezone

COPY artifacts/xlp-${TARGETARCH} /xlp

ENV XL_DASHBOARD_PORT=2345 \
  XL_DASHBOARD_IP= \
  XL_DASHBOARD_USERNAME= \
  XL_DIR_DOWNLOAD=/xunlei/downloads \
  XL_PREVENT_UPDATE= \
  XL_SPK_URL= \
  XL_UID= \
  XL_GID= \
  XL_DEBUG=

CMD [ "/xlp" ]

