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
  APISERVER_IMG="openebs/m-apiserver"
  M_EXPORTER_IMG="openebs/m-exporter"
  CSTOR_POOL_MGMT_IMG="openebs/cstor-pool-mgmt"
  CSPI_MGMT_IMG="openebs/cspi-mgmt"
  CSTOR_VOLUME_MGMT_IMG="openebs/cstor-volume-mgmt"
  ADMISSION_SERVER_IMG="openebs/admission-server"
  CSPC_OPERATOR_IMG="openebs/cspc-operator"
  UPGRADE_IMG="openebs/m-upgrade"
  PROVISIONER_LOCALPV="openebs/provisioner-localpv"
  CVC_OPERATOR_IMG="openebs/cvc-operator"
elif [ "${ARCH}" = "aarch64" ]; then
  APISERVER_IMG="openebs/m-apiserver-arm64"
  M_EXPORTER_IMG="openebs/m-exporter-arm64"
  CSTOR_POOL_MGMT_IMG="openebs/cstor-pool-mgmt-arm64"
  CSPI_MGMT_IMG="openebs/cspi-mgmt-arm64"
  CSTOR_VOLUME_MGMT_IMG="openebs/cstor-volume-mgmt-arm64"
  ADMISSION_SERVER_IMG="openebs/admission-server-arm64"
  CSPC_OPERATOR_IMG="openebs/cspc-operator-arm64"
  UPGRADE_IMG="openebs/m-upgrade-arm64"
  PROVISIONER_LOCALPV="openebs/provisioner-localpv-arm64"
  CVC_OPERATOR_IMG="openebs/cvc-operator-arm64"
fi

# tag and push all the images
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