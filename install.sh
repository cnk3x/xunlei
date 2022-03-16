#!/usr/bin/env sh

set -eu

latest=$(curl -fsSL https://api.github.com/repos/cnk3x/xunlei/releases/latest | grep browser_download_url | grep $(uname -m) | head -n 1 | grep -Eo https.+.tar.gz | sed 's|github.com|mirror.ghproxy.com/https://github.com|g')
echo "download: $latest"
curl -fsSL ${latest} | tar zx
./xunlei $@

if [ "$(pwd)" != "/var/packages/pan-xunlei-com" ]; then
    rm -f ./xunlei
fi
