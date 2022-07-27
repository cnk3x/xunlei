FROM ubuntu:18.04 as spk

RUN sed -i 's/archive.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list
RUN sed -i 's/security.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list
RUN apt-get update && apt-get -y --no-install-recommends install ca-certificates xz-utils tzdata
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

COPY spk /spk
WORKDIR /var/packages/pan-xunlei-com/target
RUN if [ "$(uname -m)" = "aarch64" ]; then arch=armv8; else arch=$(uname -m); fi; \
    spkFn=$(find /spk -type f -name \*-${arch}.spk | head -n1); \
    if [ ! -f "${spkFn}" ]; then exit 1; fi; \
    tar --wildcards -Oxf ${spkFn} package.tgz | tar --wildcards -xJ 'bin/bin/*' 'ui/index.cgi'; \
    mv bin/bin/* bin; rm -rf bin/bin

RUN mkdir /rootfs && \
    cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs && \
    cp --parents -r /var/packages/pan-xunlei-com/target /rootfs && \
    cp --parents /etc/localtime /rootfs && \
    cp --parents /etc/timezone /rootfs

COPY bin/xlp-linux-x86_64 /rootfs/xlp

FROM ubuntu:18.04
LABEL maintainer="七月<wen@k3x.cn>"
COPY --from=spk /rootfs /
ENV XL_WEB_PORT=2345 XL_DEBUG=0
VOLUME [ "/downloads", "/data" ]
ENTRYPOINT [ "/xlp" ]

CMD ["run"]
