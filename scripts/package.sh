#!/bin/bash
set -x
BASE=$(dirname $0)
CODE_DIR=$(readlink -e "$BASE/../")

BUILD=$CODE_DIR/build

ARCH="$(uname -m)"
VERSION=$(git describe --long --always)

PACKAGE_NAME="${BUILD}/raintank-probe-${VERSION}_${ARCH}.deb"
mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/etc/init
mkdir -p ${BUILD}/etc/raintank

cp ${BASE}/etc/probe.ini ${BUILD}/etc/raintank/
cp $CODE_DIR/publicChecks.json ${BUILD}/etc/raintank/
cp ${BASE}/etc/init/raintank-probe.conf ${BUILD}/etc/init
mv ${BUILD}/raintank-probe ${BUILD}/usr/bin/

fpm -s dir -t deb \
  -v ${VERSION} -n raintank-probe -a ${ARCH} --description "Raintank Probe" \
  --deb-upstart ${BASE}/etc/init/raintank-probe.conf \
  --replaces node-raintank-collector --provides node-raintank-collector \
  -C ${BUILD} -p ${PACKAGE_NAME} .

