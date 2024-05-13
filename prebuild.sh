#!/usr/bin/env sh

set -e

# arch=$([ "$(uname -m)" = "x86_64" ] && echo x86_64 || echo armv8)
# SPK_TARGET=./rootfs/var/packages/pan-xunlei-com/target
# spk=$(find ./spk -type f -name "*${arch}.spk" | head -n1) \
#   && ([ -f "${spk}" ] && echo "spk check exist!" || exit 1) \
#   && mkdir -p ${SPK_TARGET} \
#   && (tar -Oxf ${spk} package.tgz | tar -xJC ${SPK_TARGET} --wildcards 'bin/bin/*' 'ui/index.cgi')

cd $(dirname $0)
ROOT=$(pwd)

ARCH=$(go env GOARCH)
SYS_ARCH=$([ "${ARCH}" = "amd64" ] && echo "x86_64" || echo "armv8")
VER=$(ls spk | grep "${SYS_ARCH}" | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1 | sed 's/v//g')
NAME=$(ls spk | grep ${SYS_ARCH} | grep ${VER} | head -n1)

if [ -f "spk/${NAME}" ]; then
  SPK=spk/${NAME}
else
  echo "spk not found!"
  exit 1
fi

FILE_DIR=${ROOT}/embeds/nasxunlei-${ARCH}

echo extract ${SPK}
rm -rf ${FILE_DIR} && mkdir -p ${FILE_DIR}
tar -x -O -f ${SPK} package.tgz | tar -x -J -C ${FILE_DIR} --wildcards 'bin/bin/*' 'ui/index.cgi'

echo repack
DEST=${FILE_DIR}.rpk
rm -f ${DEST}

cd ${FILE_DIR}
tar --zstd -c -f ${DEST} bin/* ui/*
cd ${ROOT}

rm -rf ${FILE_DIR}

echo test
tar --zstd -t -f ${DEST}

echo packed file store in ${DEST}
