#!/bin/sh
set -eux

EXCLUDE="bin|vendor|tools|^\..*"

ln -sfnv $(readlink -f .) ${HOME}/.go/src/github.com/kizkoh/rcc

mkdir -pv ./bin
for i in $(find . -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | egrep -v ${EXCLUDE})
do
    go build -o ./bin/$i ./$i
done
