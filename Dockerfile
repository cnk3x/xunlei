FROM ubuntu:jammy
ARG TARGETARCH
ENV DEBIAN_FRONTEND=noninteractive

RUN apt update
RUN apt install --no-install-recommends -y ca-certificates tzdata curl xz-utils

ENV spk=/tmp/xl.spk spk_tmp=/tmp/xl-tmp rootfs=/rootfs
RUN mkdir -p ${spk_tmp} ${rootfs}/etc/ssl/certs
RUN curl -kLo ${spk} "https://down.sandai.net/nas/nasxunlei-DSM7-$([ "${TARGETARCH}" = "arm64" ] && echo x86_64 || echo armv8).spk"
RUN tar -xvOf ${spk} package.tgz | tar -xvJC ${spk_tmp} ui/index.cgi bin
RUN ldd ${spk_tmp}/ui/index.cgi ${spk_tmp}/bin/bin/* 2>/dev/null | grep -v 'not found' | awk '{print $3}' | sort | uniq | xargs -I '{}' cp -v '{}' ${rootfs}/lib
RUN cp -Lr /usr/share/zoneinfo/Asia/Chongqing ${rootfs}/etc/localtime
RUN echo "Asia/Chongqing" >${rootfs}/etc/timezone
RUN cp -Lr --parents /etc/ssl/certs/ca-certificates.crt ${rootfs}

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

CMD [ "/xlp", "-r", "/xunlei" ]
