FROM golang:1.18.2 as builder

ARG TARGETARCH
ENV DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true
RUN if [ "${TARGETARCH}" != "arm64" -a "${TARGETARCH}" != "amd64" ]; then echo "arch ${TARGETARCH} is not supported"; exit 1; fi

RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && sed -i 's/security.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list
RUN apt-get update && apt-get -y --no-install-recommends install ca-certificates xz-utils tzdata
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

COPY spk /spk
WORKDIR /var/packages/pan-xunlei-com/target
RUN if [ "$(uname -m)" = "aarch64" ]; then arch=armv8; else arch=$(uname -m); fi; \
    tar --wildcards -Oxf $(find /spk -type f -name \*-${arch}.spk | head -n1) package.tgz | tar --wildcards -xJ 'bin/bin/*' 'ui/index.cgi'; \
    mv bin/bin/* bin; rm -rf bin/bin

WORKDIR /go/xlp
COPY xlp .
RUN GO111MODULE=on GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 go build -v -ldflags '-s -w -extldflags "-static"' ./

RUN mkdir -p /rootfs/xunlei && \
    cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs && \
    cp -r --parents /var/packages/pan-xunlei-com/target /rootfs && \
    cp /go/xlp/xlp /rootfs/xunlei/xlp

FROM ubuntu:18.04 as vip

LABEL maintainer=七月<wen@k3x.cn>

COPY --from=builder /rootfs /
COPY --from=builder /etc/localtime /etc/timezone /etc/

ENV XL_WEB_PORT=2345 XL_DEBUG=0
EXPOSE 2345
VOLUME [ "/data", "/downloads" ]
ENTRYPOINT [ "/xunlei/xlp" ]
CMD ["run"]

FROM vip as syno
CMD ["syno"]
