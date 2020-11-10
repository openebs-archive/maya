#!/bin/bash
# Copyright 2020 The OpenEBS Authors.
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

# Determine the arch/os we're building for
ARCH=$(uname -m)

if [ "${ARCH}" = "x86_64" ]; then
  ARCH_SUFFIX=""
elif [ "${ARCH}" = "aarch64" ]; then
  ARCH_SUFFIX="-arm64"
fi

curl --fail https://raw.githubusercontent.com/openebs/charts/gh-pages/scripts/release/buildscripts/push > ./buildscripts/push
chmod +x ./buildscripts/push

DIMAGE="${IMAGE_ORG}/m-apiserver${ARCH_SUFFIX}" ./buildscripts/push
DIMAGE="${IMAGE_ORG}/cstor-pool-mgmt${ARCH_SUFFIX}" ./buildscripts/push
DIMAGE="${IMAGE_ORG}/cstor-volume-mgmt${ARCH_SUFFIX}" ./buildscripts/push
DIMAGE="${IMAGE_ORG}/admission-server${ARCH_SUFFIX}" ./buildscripts/push
DIMAGE="${IMAGE_ORG}/m-upgrade${ARCH_SUFFIX}" ./buildscripts/push
