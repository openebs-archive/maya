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

set -e

#./ci/helm_install_openebs.sh
# global env vars to be used in test scripts
export CI_BRANCH="master"
export CI_TAG="ci"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"

./ci/build-maya.sh
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

# curl https://raw.githubusercontent.com/openebs/openebs/master/k8s/ci/test-script.sh > test-script.sh

## Compile udev c code and build binary in /var/openebs/sparse
echo "Creating /var/openebs/sparse/udev_checks directory"
sudo mkdir -p /var/openebs/sparse/udev_checks
echo "Compiling and building the binary"
sudo gcc ci/udev_check.c -ludev -o /var/openebs/sparse/udev_checks/udev_check

# append mayactl tests to this script
cat ./ci/mayactl.sh >> ./ci/test-script.sh

# append local pv tests to this script
#cat ./ci/local_pv.sh >> ./test-script.sh

./ci/test-script.sh
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
