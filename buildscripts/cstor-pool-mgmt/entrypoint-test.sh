#!/bin/sh

set -ex

cd $GOPATH/src/github.com/openebs/maya
./buildscripts/cstor-pool-mgmt/test-cov-cstor.sh
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

if [ $SRC_REPO != $DST_REPO ];
then
        echo "Copying coverage.txt to $SRC_REPO"
        cp coverage.txt $SRC_REPO/
        cd $SRC_REPO
fi

