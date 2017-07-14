#!/usr/bin/env bash
set -e

if [ -z "${CTLNAME}" ]; 
then
    CTLNAME="m-apiserver"
fi

# Create a temp dir and clean it up on exit
TEMPDIR=`mktemp -d -t m-apiserver-test.XXX`
trap "rm -rf $TEMPDIR" EXIT HUP INT QUIT TERM

# Build the Maya binary for the tests
echo "--> Building ${CTLNAME} ..."
go build -o $TEMPDIR/m-apiserver || exit 1


# Run the tests
echo "--> Running tests"
GOBIN="`which go`"

if [ -z "${TESTPKG}" ]; 
then
    TESTPKG=$($GOBIN list ./... | grep -v /vendor/)
else
    TESTPKG=$($GOBIN list ./... | grep -v /vendor/ | grep $TESTPKG)
fi

# Sample usage by running makefile
#
# $ TESTPKG=orchprovider/k8s  make
# --> Running go fmt
# --> Building m-apiserver ...
# --> Running tests
# ok  	github.com/openebs/mayaserver/lib/orchprovider/k8s	0.092s	coverage: 79.2% of statements
#
# $ TESTPKG=orchprovider/nomad  make
# --> Running go fmt
# --> Building m-apiserver ...
# --> Running tests
# ok  	github.com/openebs/mayaserver/lib/orchprovider/nomad	0.092s	coverage: 0.3% of statements
#
# NOTE:
#   This will reduce the test time by focusing on specific packages.
# Alternatively, below is an exhaustive unit test run.
# 
# $ make
# ...
# ...
sudo -E PATH=$TEMPDIR:$PATH  -E GOPATH=$GOPATH \
    $GOBIN test ${GOTEST_FLAGS:--cover -timeout=900s} $TESTPKG

