#!/usr/bin/env sh

#######################################################
## 提取群晖迅雷文件                                     ## 
## ./unpack.sh ../4月25日更新-v2.7.1-DSM7.x-armv8.spk  ## 
#######################################################

set -e
cd $(dirname $0)

spk=$(find ./spk -type f -name \*-$(uname -m).spk | head -1)
spk=${1:-$spk}

echo "群晖套包: ${spk}"

if [ -z "${spk}" ]; then
    echo "未找到套包"
    exit 1
fi

rm -rf target
mkdir -p target
echo "解压目录: ./target"

tar -Oxf ${spk} package.tgz | tar --wildcards -xJC target 'bin/bin/xunlei-pan-cli*' 'bin/bin/version' 'ui/index.cgi'
mv target/bin/bin/* target/bin
rm -rf target/bin/bin/

version=$(cat target/bin/version)
arch=$(ls target/bin | grep -Eo '(amd64|arm64)' | head -n 1)
echo "迅雷版本: v${version}-${arch}"
