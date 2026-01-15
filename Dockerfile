FROM debian:stable-slim
ARG TARGETARCH
ENV DEBIAN_FRONTEND=noninteractive

RUN apt update
RUN apt install --no-install-recommends -y ca-certificates tzdata curl wget xz-utils

# ENV spk=/tmp/xl.spk spk_tmp=/tmp/xl-tmp rootfs=/rootfs
# RUN mkdir -p ${spk_tmp} ${rootfs}/etc/ssl/certs
# RUN wget -O ${spk} "https://down.sandai.net/nas/nasxunlei-DSM7-$([ "${TARGETARCH}" = "arm64" ] && echo x86_64 || echo armv8).spk"
# RUN tar -xvOf ${spk} package.tgz | tar -xvJC ${spk_tmp} ui/index.cgi bin
# RUN ldd ${spk_tmp}/ui/index.cgi ${spk_tmp}/bin/bin/xunlei-pan-cli* 2>/dev/null | grep '=>' | awk '{printf "cp %s /rootfs/lib/%s;\n", $3, $1}' | sh

ENV rootfs=/rootfs
RUN mkdir -p ${rootfs}/etc/ssl/certs ${rootfs}/lib
RUN find /usr/lib \( -name libdl.so.2 -o -name libgcc_s.so.1 -o -name libstdc++.so.6 \) -exec cp -Lr {} ${rootfs}/lib/ \;

RUN cp -Lr /usr/share/zoneinfo/Asia/Chongqing ${rootfs}/etc/localtime
RUN echo "Asia/Chongqing" >${rootfs}/etc/timezone
RUN cp -Lr --parents /etc/ssl/certs/ca-certificates.crt ${rootfs}/

COPY artifacts/xlp-${TARGETARCH} /rootfs/xlp
RUN chmod +x /rootfs/xlp

FROM busybox:latest
ARG TARGETARCH

COPY --from=0 /rootfs /rootfs/
RUN cp -Lr --parents /lib /rootfs/

FROM busybox:latest
ARG TARGETARCH

LABEL org.opencontainers.image.authors=cnk3x
LABEL org.opencontainers.image.source=https://github.com/cnk3x/xunlei

COPY --from=1 /rootfs /

ENV XL_DASHBOARD_PORT=2345 \
  XL_DASHBOARD_IP= \
  XL_DASHBOARD_USERNAME= \
  XL_DIR_DOWNLOAD=/xunlei/downloads \
  XL_PREVENT_UPDATE= \
  XL_SPK_URL= \
  XL_UID= \
  XL_GID= \
  XL_DEBUG=

VOLUME [ "/xunlei/data", "/xunlei/var/packages/pan-xunlei-com" ]
EXPOSE 2345

CMD [ "/xlp" ]
