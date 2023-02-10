FROM --platform=$BUILDPLATFORM cnk3x/xunlei:latest

ARG TARGETPLATFORM \
    TARGETARCH\
    CLI_UPDATE_VERSION=3.3.1 \
    CLI_UPDATE_VERSION_DATE=20230202 \
    CLI_UPDATE_URL_PREFIX=https://static.xbase.cloud/file/2rvk4e3gkdnl7u1kl0k/pancli \
    BIN_ROOT=/var/packages/pan-xunlei-com/target/bin 

ENV CLI_VERSION=${CLI_UPDATE_VERSION} \
    PLATFORM=${TARGETPLATFORM} \
    ARCH=${TARGETARCH} \
    CLI_DOWNLOAD_URL=${CLI_UPDATE_URL_PREFIX}/${TARGETARCH}/xunlei-pan-cli.${CLI_UPDATE_VERSION}-${CLI_UPDATE_VERSION_DATE}.${TARGETARCH}

RUN echo "${CLI_UPDATE_VERSION}" > /var/packages/pan-xunlei-com/target/bin/version && \
    rm -f /var/packages/pan-xunlei-com/target/bin/xunlei-pan-cli.* &&\
    wget -O ${BIN_ROOT}/xunlei-pan-cli.${CLI_UPDATE_VERSION}.${ARCH} ${CLI_DOWNLOAD_URL} && \
    chmod +x ${BIN_ROOT}/xunlei-pan-cli.${CLI_UPDATE_VERSION}.${ARCH}
