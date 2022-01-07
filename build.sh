#!/usr/bin/env sh

set -eu

version=$(cat target/version)

echo $version
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags "-static"' -o bin/xunlei ./

if command -v upx >/dev/null; then
    upx bin/xunlei
fi

tgz="xunlei-v${version}.$(uname | tr 'A-Z' 'a-z').$(uname -m).tar.gz"
echo $tgz
tar -C bin -cf bin/$tgz xunlei
tar -tf bin/$tgz
