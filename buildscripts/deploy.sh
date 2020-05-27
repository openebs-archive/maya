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
  APISERVER_IMG="${IMAGE_ORG}/m-apiserver"
  M_EXPORTER_IMG="${IMAGE_ORG}/m-exporter"
  CSTOR_POOL_MGMT_IMG="${IMAGE_ORG}/cstor-pool-mgmt"
  CSPI_MGMT_IMG="${IMAGE_ORG}/cspi-mgmt"
  CSTOR_VOLUME_MGMT_IMG="${IMAGE_ORG}/cstor-volume-mgmt"
  ADMISSION_SERVER_IMG="${IMAGE_ORG}/admission-server"
  CSPC_OPERATOR_IMG="${IMAGE_ORG}/cspc-operator"
  UPGRADE_IMG="${IMAGE_ORG}/m-upgrade"
  PROVISIONER_LOCALPV="${IMAGE_ORG}/provisioner-localpv"
  CVC_OPERATOR_IMG="${IMAGE_ORG}/cvc-operator"
elif [ "${ARCH}" = "aarch64" ]; then
  APISERVER_IMG="${IMAGE_ORG}/m-apiserver-arm64"
  M_EXPORTER_IMG="${IMAGE_ORG}/m-exporter-arm64"
  CSTOR_POOL_MGMT_IMG="${IMAGE_ORG}/cstor-pool-mgmt-arm64"
  CSPI_MGMT_IMG="${IMAGE_ORG}/cspi-mgmt-arm64"
  CSTOR_VOLUME_MGMT_IMG="${IMAGE_ORG}/cstor-volume-mgmt-arm64"
  ADMISSION_SERVER_IMG="${IMAGE_ORG}/admission-server-arm64"
  CSPC_OPERATOR_IMG="${IMAGE_ORG}/cspc-operator-arm64"
  UPGRADE_IMG="${IMAGE_ORG}/m-upgrade-arm64"
  PROVISIONER_LOCALPV="${IMAGE_ORG}/provisioner-localpv-arm64"
  CVC_OPERATOR_IMG="${IMAGE_ORG}/cvc-operator-arm64"
elif [ "${ARCH}" = "ppc64le" ]; then
  PROVISIONER_LOCALPV="${IMAGE_ORG}/provisioner-localpv-ppc64le"
fi

# tag and push all the images
if [ "${ARCH}" = "ppc64le" ]; then
  DIMAGE="${PROVISIONER_LOCALPV}" ./buildscripts/push
else
  DIMAGE="${APISERVER_IMG}" ./buildscripts/push
  DIMAGE="${M_EXPORTER_IMG}" ./buildscripts/push
  DIMAGE="${CSTOR_POOL_MGMT_IMG}" ./buildscripts/push
  DIMAGE="${CSPI_MGMT_IMG}" ./buildscripts/push
  DIMAGE="${CSTOR_VOLUME_MGMT_IMG}" ./buildscripts/push
  DIMAGE="${ADMISSION_SERVER_IMG}" ./buildscripts/push
  DIMAGE="${CSPC_OPERATOR_IMG}" ./buildscripts/push
  DIMAGE="${UPGRADE_IMG}" ./buildscripts/push
  DIMAGE="${PROVISIONER_LOCALPV}" ./buildscripts/push
  DIMAGE="${CVC_OPERATOR_IMG}" ./buildscripts/push
fi
