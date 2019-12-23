#!/bin/bash -e

#!/bin/bash
set -x
# Find the directory we exist within
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR}

GITVERSION=`git describe --long --always`
SOURCEDIR=${DIR}/..
BUILDDIR=$SOURCEDIR/build

# Disable CGO for builds.
export CGO_ENABLED=0

# Make dir
mkdir -p $BUILDDIR

# Clean build bin dir
rm -rf $BUILDDIR/*

# Build binary
cd ../cmd/raintank-probe
go build -mod=vendor -ldflags "-X main.GitHash=$GITVERSION" -o $BUILDDIR/raintank-probe
