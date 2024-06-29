#!/usr/bin/env sh

set -e

cd $(dirname $0)
cd ..

WD=$(pwd)

SPK_ROOT=${WD}/spk

checkspk() {
  if [ -z "$1" ]; then
    echo "spk path required" >&2
  elif SPK=$(realpath $1) && [ -f "${SPK}" ]; then
    echo ${SPK} >&1
  else
    echo $1 not found >&2
  fi
}

checkver() {
  local SPK=$(checkspk $1)
  [ -z "${SPK}" ] && exit 1
  local PAN_CLI=$(tar -Oxf ${SPK} package.tgz | tar -Jt --wildcards 'bin/bin/xunlei-pan-cli.*')
  local VER=$(echo ${PAN_CLI} | grep -Eo "[0-9]+.[0-9]+.[0-9]+")
  local ARCH=$(echo ${PAN_CLI} | grep -q "amd64" && echo x86_64 || (echo ${PAN_CLI} | grep -q "arm64" && echo armv8))
  echo $(basename ${SPK}): VER: ${VER}, ARCH: ${ARCH}
}

unpack() {
  SPK=$(checkspk $1)
  [ -z "${SPK}" ] && exit 1

  FILE_DIR=$2
  if [ -z "${FILE_DIR}" ]; then
    FILE_DIR=${SPK%.*}-All
  fi

  echo extract ${SPK} to ${FILE_DIR}
  mkdir -p ${FILE_DIR}/package
  tar -zxf ${SPK} -C ${FILE_DIR}
  tar -Jxf ${FILE_DIR}/package.tgz -C ${FILE_DIR}/package
}

mkrpk() {
  SPK=$(checkspk $1)
  [ -z "${SPK}" ] && exit 1

  FILE_DIR=$2
  if [ -z "${FILE_DIR}" ]; then
    FILE_DIR=${SPK%.*}
  fi

  echo extract $(basename ${SPK}) to ${FILE_DIR}
  rm -rf ${FILE_DIR} && mkdir -p ${FILE_DIR}
  tar -Oxf ${SPK} package.tgz | tar -JxC ${FILE_DIR} --wildcards 'bin/bin/*' 'ui/index.cgi'

  TARGET=${FILE_DIR}.tar.zst
  echo repack to $(basename ${TARGET})
  rm -f ${TARGET}

  cd ${FILE_DIR}
  tar --zstd -cf ${TARGET} bin/* ui/*
  cd ${WD}

  rm -rf ${FILE_DIR}

  # echo test $(basename ${TARGET})
  # tar --zstd -tf ${TARGET}

  echo rpk file store in ${TARGET}
}

spkRename() {
  local SPK=$(checkspk $1)
  [ -z "${SPK}" ] && exit 1

  local PAN_CLI=$(tar -Oxf ${SPK} package.tgz | tar -Jt --wildcards 'bin/bin/xunlei-pan-cli.*')
  local VER=$(echo ${PAN_CLI} | grep -Eo "[0-9]+.[0-9]+.[0-9]+")
  local ARCH=$(echo ${PAN_CLI} | grep -q "amd64" && echo x86_64 || (echo ${PAN_CLI} | grep -q "arm64" && echo armv8))
  local DSM=$(echo ${PAN_CLI} | grep -q "DSM" && echo DSM || echo DSM7)
  local NEW_SPK=$(dirname $SPK)/nasxunlei-v${VER}-${ARCH}.spk

  if [ -z "${VER}" ]; then
    echo "spkRename $(basename $SPK), miss VER" >&2
  elif [ -z "${ARCH}" ]; then
    echo "spkRename $(basename $SPK), miss ARCH" >&2
  elif [ "${SPK}" = "${NEW_SPK}" ]; then
    echo "spkRename $(basename $SPK), skip" >&2
  else
    echo spkRename $(basename $SPK) to $(basename ${NEW_SPK})
    mv ${SPK} ${NEW_SPK}
  fi
}

uniname() {
  for SPK in $(ls ${SPK_ROOT} | grep .spk); do
    spkRename ${SPK_ROOT}/${SPK}
  done
}

download() {
  local DOWNLOADED_SPK=${SPK_ROOT}/nasxunlei.spk
  curl -SsL -o ${DOWNLOADED_SPK} https://down.sandai.net/nas/nasxunlei-DSM7-x86_64.spk
  spkRename ${DOWNLOADED_SPK}
}

findspk() {
  local ARCH=$([ -n "$1" ] && echo $1 || echo x86_64)
  local V_LATEST=$(ls ${WD}/spk | grep ${ARCH} | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1)
  local NAME=$(ls ${WD}/spk | grep ${ARCH} | grep ${V_LATEST} | head -n1)
  [ -n "${NAME}" ] && echo ${WD}/spk/${NAME}
}

prebuild() {
  local ARCH=
  local GOARCH=
  [ -z "$1" ] && (echo "need input arch") && exit 1

  case $1 in
  x86_64 | amd64)
    ARCH=x86_64
    GOARCH=amd64
    ;;
  armv8 | arm64)
    ARCH=armv8
    GOARCH=arm64
    ;;
  *)
    echo need input arch
    exit 1
    ;;
  esac

  NAME=$(findspk ${ARCH})
  if [ -f "${NAME}" ]; then
    mkrpk ${NAME} ${SPK_ROOT}/nasxunlei-${ARCH}
  else
    echo ${NAME} not found!
    exit 1
  fi
}

case $1 in
mkrpk)
  mkrpk $2 $3
  ;;
checkver)
  checkver $2
  ;;
unpack)
  unpack $2 $3
  ;;
uniname)
  uniname
  ;;
rename)
  spkRename $2
  ;;
latest)
  findspk x86_64
  findspk armv8
  ;;
findspk)
  findspk $2
  ;;
prebuild)
  if [ -z "$2" ]; then
    prebuild amd64
    prebuild arm64
  else
    prebuild $2
  fi
  ;;
download)
  download
  ;;
*)
  echo "$0 mkrpk | checkver | unpack | uniname | latest | findspk | prebuild | download"
  ;;
esac
