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

// JivaPoolArtifactsFor070 returns the default jiva pool and storage
// class related artifacts corresponding to version 0.7.0
func JivaPoolArtifactsFor070() (list ArtifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(jivaPoolYamlsFor070)...)
	return
}

// jivaPoolYamlsFor070 returns all the yamls related to jiva pool and 
// storage class in a string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func jivaPoolYamlsFor070() string {
	return `
---
apiVersion: openebs.io/v1alpha1
kind: StoragePool
metadata:
  name: default
  type: hostdir
spec:
  path: "/var/openebs"
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-standard
  annotations:
    cas.openebs.io/config: |
      - name: ReplicaCount
        value: "3"
      - name: StoragePool
        value: default
      #- name: TargetResourceLimits
      #  value: |-
      #      memory: 1Gi
      #      cpu: 100m
      #- name: AuxResourceLimits
      #  value: |-
      #      memory: 0.5Gi
      #      cpu: 50m
      #- name: ReplicaResourceLimits
      #  value: |-
      #      memory: 2Gi
provisioner: openebs.io/provisioner-iscsi
---
`
}
