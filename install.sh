#!/usr/bin/env sh

set -eu

latest=$(curl -fsSL https://gh.k3x.cn/api/repos/cnk3x/xunlei/releases/latest | grep browser_download_url | grep $(uname -m) | head -n 1 | grep -Eo https.+.tar.gz | sed 's/github.com/gh.k3x.cn/g')
curl -fsSL ${latest} | tar zx
./xunlei $@

if [ "$(pwd)" != "/var/packages/pan-xunlei-com" ]; then
    rm -f ./xunlei
fi
