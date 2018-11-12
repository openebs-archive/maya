/*
Copyright 2018 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	"strconv"
)

// IsCstorSparsePoolEnabled reads from env variable to check wether cstor sparse pool
// should be created by default or not.
func IsCstorSparsePoolEnabled() (enabled bool) {
	enabled, _ = strconv.ParseBool(menv.Get(DefaultCstorSparsePool))
	return
}

// CstorSparsePoolSpcArtifactsFor070 returns the default storagepoolclaim
// and storageclass yaml corresponding to version 0.7.0 if cstor
// sparse pool creation is enabled as a part of openebs installation
func CstorSparsePoolSpcArtifactsFor070() (list ArtifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamlConditional(cstorSparsePoolSpcArtifactsFor070, IsCstorSparsePoolEnabled)...)
	return
}

// cstorPoolSpcForArtifacts070 returns all the yamls related to configuring
// a stoaragepoolclaim and storageclass in string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func cstorSparsePoolSpcArtifactsFor070() string {
	return `
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-cstor-sparse
  annotations:
    openebs.io/cas-type: cstor
    cas.openebs.io/config: |
      - name: StoragePoolClaim
        value: "cstor-sparse-pool"
      #- name: TargetResourceLimits
      #  value: |-
      #      memory: 1Gi
      #      cpu: 200m
      #- name: AuxResourceLimits
      #  value: |-
      #      memory: 0.5Gi
      #      cpu: 50m
provisioner: openebs.io/provisioner-iscsi
---
apiVersion: openebs.io/v1alpha1
kind: StoragePoolClaim
metadata:
  name: cstor-sparse-pool
  annotations:
    cas.openebs.io/config: |
      #For default sparse pool set the limit at 2Gi to safegaurd 
      # cstor pool from consuming more memory and causing the node 
      # to get into memory pressure condition. By default K8s will set the 
      # Requests to the same value as Limits. For example, when Limit is
      # set to 2Gi, the pool could get stuck in pending schedule state,
      # if node doesn't have Requested (2Gi) memory. 
      # Hence setting the Requests to a minimum (0.5Gi). 
      - name: PoolResourceRequests
        value: |-
            memory: 0.5Gi
            cpu: 100m
      - name: PoolResourceLimits
        value: |-
            memory: 2Gi
            cpu: 500m
      #- name: AuxResourceLimits
      #  value: |-
      #      memory: 1Gi
      #      cpu: 100m
spec:
  name: cstor-sparse-pool
  type: sparse
  maxPools: 3
  poolSpec:
    poolType: striped
---
`
}
