FROM golang as builder

ARG TARGETARCH
ENV DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true
RUN if [ "${TARGETARCH}" != "arm64" -a "${TARGETARCH}" != "amd64" ]; then echo "arch ${TARGETARCH} is not supported"; exit 1; fi

RUN sed -i 's/deb.debian.org/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list && \
    sed -i 's/security.debian.org/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list
RUN apt-get update && apt-get -y --no-install-recommends install ca-certificates xz-utils tzdata
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

COPY spk /spk
WORKDIR /var/packages/pan-xunlei-com/target
RUN if [ "$(uname -m)" = "aarch64" ]; then arch=armv8; else arch=$(uname -m); fi; \
    spkFn=$(find /spk -type f -name \*-${arch}.spk | head -n1); \
    if [ ! -f "${spkFn}" ]; then exit 1; fi; \
    tar --wildcards -Oxf ${spkFn} package.tgz | tar --wildcards -xJ 'bin/bin/*' 'ui/*'; \
    mv bin/bin/* bin; rm -rf bin/bin

WORKDIR /goxlp
COPY xlp .
RUN GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 \
    go build -v -ldflags '-s -w -extldflags "-static"' -tags netgo -o /rootfs/xunlei/xlp ./

RUN cp --parents -r /var/packages/pan-xunlei-com/target /rootfs

FROM ubuntu:focal
LABEL maintainer="七月<wen@k3x.cn>"

ENV LANG=C.UTF-8 DEBIAN_FRONTEND=noninteractive LANG=zh_CN.UTF-8 LANGUAGE=zh_CN.UTF-8 LC_ALL=C

RUN sed -i 's/deb.debian.org/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \ 
    && apt-get update && apt-get -y --no-install-recommends install tzdata locales xfonts-wqy wget ca-certificates && rm -rf /var/lib/apt/lists/*

# 设置中文环境
RUN localedef -i zh_CN -c -f UTF-8 -A /usr/share/locale/locale.alias zh_CN.UTF-8 && locale-gen zh_CN.UTF-8 && \
    echo "Asia/Shanghai" > /etc/timezone && \
    dpkg-reconfigure -f noninteractive tzdata && \
    rm -f /etc/localtime && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    find /var/lib/apt/lists -type f -delete && \
    find /var/cache -type f -delete

COPY --from=builder /rootfs /
ENV XL_WEB_PORT=2345 XL_DEBUG=0
VOLUME [ "/xunlei/downloads", "/xunlei/data" ]
ENTRYPOINT [ "/xunlei/xlp" ]
CMD ["syno"]
