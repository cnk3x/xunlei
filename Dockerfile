FROM golang:1.18.1 as builder

ARG TARGETARCH

RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && \
    sed -i 's/security.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y install upx-ucl xz-utils ca-certificates tzdata

WORKDIR /app

ENV XTARGET=/xunlei/var/packages/pan-xunlei-com/target
COPY spk/xunlei-${TARGETARCH}.spk ./spk/
RUN mkdir -p ${XTARGET}
RUN tar Oxvf ./spk/xunlei-${TARGETARCH}.spk package.tgz | \
    tar --wildcards -xvJC ${XTARGET} 'bin/bin/*' 'ui/index.cgi' && \
    mv ${XTARGET}/bin/bin/* ${XTARGET}/bin && \
    rm -rf ${XTARGET}/bin/bin

RUN ldd ${XTARGET}/bin/* ${XTARGET}/ui/* | grep -v dynamic | grep '=>' | cut -d ' ' -f3 | sed 's/://g' | sort | uniq | xargs -I {} cp --parents {} /xunlei
RUN if [ "${TARGETARCH}" = "amd64" ]; then cp --parents /lib64/ld-linux-x86-64.so.* /xunlei; fi
RUN if [ "${TARGETARCH}" = "arm64" ]; then cp --parents /lib/ld-linux-aarch64.so.* /xunlei; fi

COPY src .
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct 
RUN CGO_ENABLED=0 go build -a -v -ldflags '-s -w -extldflags "-static"' -o /xunlei/xlp 
RUN upx /xunlei/xlp

FROM busybox as bb1

COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /xunlei /

RUN mkdir -p /usr/syno/synoman/webman/modules && \
    echo "#!/bin/sh" > /usr/syno/synoman/webman/modules/authenticate.cgi && \
    echo "echo OK" >> /usr/syno/synoman/webman/modules/authenticate.cgi && \
    chmod +x /usr/syno/synoman/webman/modules/authenticate.cgi

FROM busybox

COPY --from=bb1 / /xunlei/

RUN mv /xunlei/xlp /xlp && \
    mknod /xunlei/dev/null c 1 3 && \
    mknod /xunlei/dev/zero c 1 5 && \
    mknod /xunlei/dev/tty  c 5 0 && \
    chmod 0666 /xunlei/dev/null && \
    chmod 0666 /xunlei/dev/tty && \
    chmod 0666 /xunlei/dev/zero && \
    chown root.tty /xunlei/dev/tty && \
    echo "127.0.0.1 localhost" > /xunlei/etc/hosts && \
    echo "nameserver 114.114.114.114" > /xunlei/etc/resolv.conf && \
    echo "search lan" >> /xunlei/etc/resolv.conf && \
    echo "Asia/Shanghai" > /xunlei/etc/timezone

# FROM scratch
# COPY --from=bb2 /xunlei /xunlei
# COPY --from=bb2 /xlp /xlp

VOLUME [ "/data", "/downloads" ]

CMD [ "/xlp", "-port=2345" ]
