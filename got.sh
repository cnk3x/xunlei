#!/usr/bin/env bash

findMod() {
  dir=$1
  pkg=$2
  modFile=${dir}/go.mod
  if [ -f "${modFile}" ]; then
    mod=$(cat ${modFile} | grep "module" | awk '{print $2}')
    echo ${mod}$pkg
  else
    findMod $(dirname $dir) /$(basename $dir)${pkg}
  fi
}

name=$1
pkg=$(findMod $(pwd))

if [ -z "$name" ]; then
  echo "usage: got <test name>"
  exit 1
fi

if [[ ! $name = Test* ]]; then
  name=Test$name
fi

echo "test ${pkg}.${name}"

go test -timeout 30s -run ^${name}\$ ${pkg} -v -count=1
