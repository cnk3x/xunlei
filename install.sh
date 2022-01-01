#!/usr/bin/env sh

latest=$(curl -fsSL https://gh.k3x.cn/api/repos/cnk3x/xunlei-from-syno/releases | grep browser_download_url | head -n 1 | grep -Eo https.+.tar.gz | sed 's/github.com/gh.k3x.cn/g')
curl -fsSL ${latest} | tar zx
./xunlei install $@
