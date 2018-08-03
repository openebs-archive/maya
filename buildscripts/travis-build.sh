#!/bin/bash

# Copyright 2017 The OpenEBS Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

MASTER="master"

# On travis the docker image tag will be of <upstreambranch>-ci format.
# `local` tag is at contributor's local machine for verifying with make command.
if [ "$TRAVIS_BRANCH" != "" ]; 
then
 if [ "$TRAVIS_BRANCH" == "$MASTER" ];
 then
  IMAGE_TAG="ci"
 else
  IMAGE_TAG="$TRAVIS_BRANCH-ci"
 fi
else
 IMAGE_TAG="local"
fi
export IMAGE_TAG

SRC_REPO="$TRAVIS_BUILD_DIR"
DST_REPO="$GOPATH/src/github.com/openebs/maya"

if [ "$SRC_REPO" != "$DST_REPO" ];
then
	echo "Copying from $SRC_REPO to $DST_REPO"
	# Get the git commit
	echo "But first, get the git revision from $SRC_REPO"
	GIT_COMMIT="$(git rev-parse HEAD)"
	echo $GIT_COMMIT >> $SRC_REPO/GITCOMMIT

	mkdir -p $DST_REPO
	cp -R $SRC_REPO/* $DST_REPO/
	cd $DST_REPO
fi

#make golint-travis
#rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

./buildscripts/test-cov.sh
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

make all
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

if [ $SRC_REPO != $DST_REPO ];
then
	echo "Copying coverage.txt to $SRC_REPO"
	cp coverage.txt $SRC_REPO/
	cd $SRC_REPO
fi
