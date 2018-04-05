#!/bin/bash

set -ex

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

cp ./buildscripts/cstor-pool-mgmt/Dockerfile.test .
cp ./buildscripts/cstor-pool-mgmt/entrypoint-test.sh .
cp ./buildscripts/cstor-pool-mgmt/test-cov-cstor.sh .

#make golint-travis
#rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

sudo docker build -f Dockerfile.test -t openebs/cstor-test:ci .
kubectl apply -f ./buildscripts/cstor-pool-mgmt/pod-cstor-pool-test.yaml
