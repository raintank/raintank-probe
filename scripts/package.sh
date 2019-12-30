#!/bin/bash
set -x
BASE=$(dirname $0)
CODE_DIR=$(readlink -e "$BASE/../")

sudo apt-get install rpm

BUILD_ROOT=$CODE_DIR/build

ARCH="$(uname -m)"
VERSION=$(git describe --long --always)
CONTACT="Grafana Labs <hello@grafana.com>"
VENDOR="grafana.com"
LICENSE="Apache2.0"

## ubuntu 14.04
BUILD=${BUILD_ROOT}/upstart

PACKAGE_NAME="${BUILD}/raintank-probe-${VERSION}_${ARCH}.deb"

mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/etc/init
mkdir -p ${BUILD}/etc/raintank

cp ${BASE}/config/probe.ini ${BUILD}/etc/raintank/
cp ${BUILD_ROOT}/raintank-probe ${BUILD}/usr/bin/

fpm -s dir -t deb \
  -v ${VERSION} -n raintank-probe -a ${ARCH} --description "Raintank probe monitoring agent" \
  --deb-upstart ${BASE}/config/upstart/raintank-probe \
  -m "$CONTACT" --vendor "$VENDOR" --license "$LICENSE" \
  -C ${BUILD} -p ${PACKAGE_NAME} .


## ubuntu 16.04, 18.04, Debian 8
BUILD=${BUILD_ROOT}/systemd
PACKAGE_NAME="${BUILD}/raintank-probe-${VERSION}_${ARCH}.deb"
mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/lib/systemd/system/
mkdir -p ${BUILD}/etc/raintank

cp ${BASE}/config/probe.ini ${BUILD}/etc/raintank/
cp ${BUILD_ROOT}/raintank-probe ${BUILD}/usr/bin/

fpm -s dir -t deb \
  -v ${VERSION} -n raintank-probe -a ${ARCH} --description "Raintank probe monitoring agent" \
  --deb-systemd ${BASE}/config/systemd/raintank-probe.service \
  -m "$CONTACT" --vendor "$VENDOR" --license "$LICENSE" \
  -C ${BUILD} -p ${PACKAGE_NAME} .


## CentOS 7
BUILD=${BUILD_ROOT}/systemd-centos7

mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/lib/systemd/system/
mkdir -p ${BUILD}/etc/raintank

cp ${BASE}/config/probe.ini ${BUILD}/etc/raintank/
cp ${BUILD_ROOT}/raintank-probe ${BUILD}/usr/bin/
cp ${BASE}/config/systemd/raintank-probe.service $BUILD/lib/systemd/system

PACKAGE_NAME="${BUILD}/raintank-probe-${VERSION}.el7.${ARCH}.rpm"

fpm -s dir -t rpm \
  -v ${VERSION} -n raintank-probe -a ${ARCH} --description "Raintank probe monitoring agent" \
  --config-files /etc/raintank/ \
  -m "$CONTACT" --vendor "$VENDOR" --license "$LICENSE" \
  -C ${BUILD} -p ${PACKAGE_NAME} .


## CentOS 6
BUILD=${BUILD_ROOT}/upstart-0.6.5

PACKAGE_NAME="${BUILD}/raintank-probe-${VERSION}.el6.${ARCH}.rpm"

mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/etc/init
mkdir -p ${BUILD}/etc/raintank

cp ${BASE}/config/probe.ini ${BUILD}/etc/raintank/
cp ${BUILD_ROOT}/raintank-probe ${BUILD}/usr/bin/
cp ${BASE}/config/upstart-0.6.5/raintank-probe.conf $BUILD/etc/init

fpm -s dir -t rpm \
  -v ${VERSION} -n raintank-probe -a ${ARCH} --description "Raintank probe monitoring agent" \
  --config-files /etc/raintank/ \
  -m "$CONTACT" --vendor "$VENDOR" --license "$LICENSE" \
   -C ${BUILD} -p ${PACKAGE_NAME} .
