FROM golang:1.18.2 as builder

ARG TARGETARCH
ENV DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true
RUN if [ "${TARGETARCH}" != "arm64" -a "${TARGETARCH}" != "amd64" ]; then echo "arch ${TARGETARCH} is not supported"; exit 1; fi

RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && sed -i 's/security.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list
RUN apt-get update && apt-get -y --no-install-recommends install ca-certificates xz-utils tzdata zstd && rm -rf /var/lib/apt/lists/*
RUN echo "Asia/Shanghai" > /etc/timezone && cp -Lvr /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /spk
COPY spk .
RUN cp $(find /spk -type f -name \*-$(if [ "$(uname -m)" = "aarch64" -o "$(uname -m)" = "arm64" ]; then echo armv8; else uname -m; fi).spk | head -n1) xunlei.spk; \
    tar --wildcards -Oxf xunlei.spk package.tgz | tar --wildcards --strip-components=2 -xJ bin/bin/xunlei-pan-cli.\*.$(go env GOARCH);

WORKDIR /rootfs
RUN cp -Lv --parents /etc/timezone /etc/localtime /etc/ssl/certs/ca-certificates.crt .
RUN goarch=$(go env GOARCH); \
    for lib in $(ldd /spk/xunlei-pan-cli.*.${goarch} | grep = | awk '{print $3}'); do cp -Lrv --parents $lib .; done; \
    cp -Lrv --parents /lib64 .; \
    cp -Lrv --parents /usr/bin/ldd .; \
    sed -i 's^#! /bin/bash^#! /bin/sh^g' /rootfs/usr/bin/ldd; \
    mkdir -p /rootfs/xunlei && cp /spk/xunlei.spk /rootfs/xunlei/xunlei.spk;

WORKDIR /go/xlp
COPY xlp .
RUN GO111MODULE=on GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 go build -v -ldflags '-s -w -extldflags "-static"' -o /rootfs/xunlei/xlp ./

FROM busybox as vip
LABEL maintainer=七月<wen@k3x.cn>
COPY --from=builder /rootfs /
ENV XL_WEB_PORT=2345 XL_DEBUG=0
WORKDIR /xunlei
ENTRYPOINT [ "/xunlei/xlp" ]
CMD [ "run" ]

FROM vip as syno
CMD [ "syno" ]
