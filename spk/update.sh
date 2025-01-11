#! /usr/bin/env bash

# https://down.sandai.net/nas/nasxunlei-DSM7-armv8.spk
# https://down.sandai.net/nas/nasxunlei-DSM6-armv8.spk
# https://down.sandai.net/nas/nasxunlei-DSM6-x86_64.spk
# https://down.sandai.net/nas/nasxunlei-DSM7-x86_64.spk

set -e

function version_gt() { test "$(echo "$@" | tr " " "\n" | sort -V | head -n 1)" != "$1"; }

update() {
    local DOWNLOAD_URL=$1
    local SPK=nasxunlei.spk

    curl -SsL -o ${SPK} ${DOWNLOAD_URL}
    local PAN_CLI=$(tar -Oxf ${SPK} package.tgz | tar -Jt --wildcards 'bin/bin/xunlei-pan-cli.*')
    local VER=$(echo ${PAN_CLI} | grep -Eo "[0-9]+.[0-9]+.[0-9]+")
    local ARCH=$(echo ${PAN_CLI} | grep -q "amd64" && echo x86_64 || (echo ${PAN_CLI} | grep -q "arm64" && echo armv8))

    local ver_current=$(cat nasxunlei-${ARCH}.txt 2>/dev/null)

    if [ -z ${ver_current} ] || version_gt ${VER} ${ver_current}; then
        echo DOWNLOAD_URL: ${DOWNLOAD_URL} VER: ${VER}, ARCH: ${ARCH}
        mv ${SPK} nasxunlei-${ARCH}.spk
        echo ${VER} >nasxunlei-${ARCH}.txt
    else
        echo "skip nasxunlei-${VER}-${ARCH}.spk"
        rm ${SPK}
    fi
}

update https://down.sandai.net/nas/nasxunlei-DSM7-x86_64.spk
update https://down.sandai.net/nas/nasxunlei-DSM7-armv8.spk
