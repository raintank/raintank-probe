#!/bin/bash

set -x
# Find the directory we exist within
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR}

VERSION=`git describe --long --always`

mkdir build
cp ../build/raintank-probe build/

docker build -t raintank/raintank-probe:$VERSION .
docker tag raintank/raintank-probe:$VERSION raintank/raintank-probe:latest
