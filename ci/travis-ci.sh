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

#./ci/helm_install_openebs.sh
# global env vars to be used in test scripts
export CI_BRANCH="v0.9.x"
export CI_TAG="ci"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"

./ci/build-maya.sh
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

curl https://raw.githubusercontent.com/openebs/openebs/master/k8s/ci/test-script.sh > test-script.sh
# append local tests to this script
cat ./ci/mayactl.sh >> ./test-script.sh
chmod +x test-script.sh && ./test-script.sh
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
