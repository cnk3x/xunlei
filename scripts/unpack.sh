#!/usr/bin/env sh

cd $(dirname $0)
cd ..

SPK=$1
SPK_TARGET=testdata/spk

if [ -z "${SPK}" ]; then
  echo "usage: $0 <spk>"
  exit 1
fi

mkdir -p ${SPK_TARGET}/package

tar -v -x -f ${SPK} -C ${SPK_TARGET}
tar -v -x -J -f ${SPK_TARGET}/package.tgz -C ${SPK_TARGET}/package
