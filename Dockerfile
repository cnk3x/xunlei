FROM golang:1.18.2 as builder

ARG TARGETARCH
ENV DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true
RUN if [ "${TARGETARCH}" != "arm64" -a "${TARGETARCH}" != "amd64" ]; then echo "arch ${TARGETARCH} is not supported"; exit 1; fi

RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && sed -i 's/security.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list
RUN apt-get update && apt-get -y --no-install-recommends install ca-certificates xz-utils

COPY spk /spk
RUN targetDIR=/var/packages/pan-xunlei-com/target; mkdir -p ${targetDIR}; \
    if [ "$(uname -m)" = "aarch64" ]; then arch=armv8; else arch=$(uname -m); fi; \
    tar --wildcards -Oxf $(find /spk -type f -name \*-${arch}.spk | head -n1) package.tgz | tar --wildcards -xJC ${targetDIR} 'bin/bin/*' 'ui/index.cgi'; \
    mv ${targetDIR}/bin/bin/* ${targetDIR}/bin; rm -rf ${targetDIR}/bin/bin

WORKDIR /go/xlp
COPY xlp .
RUN GO111MODULE=on GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 go build -v -ldflags '-s -w -extldflags "-static"' ./

RUN mkdir -p /rootfs/xunlei && \
    cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs && \
    cp -r --parents /var/packages/pan-xunlei-com/target /rootfs && \
    cp /go/xlp/xlp /rootfs/xunlei/xlp

FROM ubuntu:18.04 as rootfs
RUN sed -i 's/archive.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && sed -i 's/security.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list
COPY --from=builder /rootfs /

# FROM busybox
# COPY --from=rootfs /var/packages /var/packages
# COPY --from=rootfs /xunlei /xunlei
# COPY --from=rootfs /lib /lib
# COPY --from=rootfs /lib64 /lib64
# COPY --from=rootfs /usr/lib /usr/lib

ENV XL_WEB_PORT=2345 XL_DEBUG=1
EXPOSE 2345
VOLUME [ "/data", "/downloads" ]
ENTRYPOINT [ "/xunlei/xlp" ]