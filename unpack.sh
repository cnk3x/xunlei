#!/usr/bin/env sh

set -e

# arch=$([ "$(uname -m)" = "x86_64" ] && echo x86_64 || echo armv8)
# SPK_TARGET=./rootfs/var/packages/pan-xunlei-com/target
# spk=$(find ./spk -type f -name "*${arch}.spk" | head -n1) \
#   && ([ -f "${spk}" ] && echo "spk check exist!" || exit 1) \
#   && mkdir -p ${SPK_TARGET} \
#   && (tar -Oxf ${spk} package.tgz | tar -xJC ${SPK_TARGET} --wildcards 'bin/bin/*' 'ui/index.cgi')

spk=$(find spk -type f -name "*x86_64.spk" | head -n1)

if [ -f "${spk}" ]; then
  echo "spk found!"
else
  echo "spk not found!"
  exit 1
fi

SPK_TARGET=testdata/target
mkdir -p ${SPK_TARGET}/package
tar -x -C ${SPK_TARGET} -f ${spk}
tar -x -J -C ${SPK_TARGET}/package -f ${SPK_TARGET}/package.tgz
