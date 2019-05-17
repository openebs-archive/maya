#!/usr/bin/env bash
set -e

# Create a temp dir and clean it up on exit
TEMPDIR=`mktemp -d -t maya-test.XXX`
trap "rm -rf $TEMPDIR" EXIT HUP INT QUIT TERM

# Build the Maya binary for the tests
echo "--> Building maya"
go build -o $TEMPDIR/maya || exit 1

# Run the tests
echo "--> Running tests"
GOBIN="`which go`"
PATH=$TEMPDIR:$PATH \
    $GOBIN test ${GOTEST_FLAGS:--cover -timeout=900s} $($GOBIN list ./... | grep -v 'vendor\|pkg/apis\|pkg/client/generated\|tests')

