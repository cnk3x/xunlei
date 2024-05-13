#!/usr/bin/env sh

set -e

cd $(dirname $0)
ROOT=$(pwd)

echo "# download and unpack"
TARGET=${ROOT}/embeds/nasxunlei
rm -rf ${TARGET}
mkdir -p ${TARGET}
curl -SsL https://down.sandai.net/nas/nasxunlei-DSM7-x86_64.spk | tar -xvO package.tgz | tar -xvJC ${TARGET} --wildcards 'bin/bin/*' 'ui/index.cgi'

cd ${TARGET}
RPK=${TARGET}-latest.rpk
echo
echo "# pack to ${RPK}"

rm -f ${RPK}
tar --zstd -cvf ${RPK} bin/* ui/*
cd ${ROOT}

rm -rf ${TARGET}
echo

echo "# test"
tar --zstd -tf ${TARGET}.rpk
