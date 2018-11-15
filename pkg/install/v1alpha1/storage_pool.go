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

const storagePoolYaml = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: storage-pool-list-default
spec: 
  taskNamespace: {{ env "OPENEBS_NAMESPACE" }}
  run:
    tasks:
    - storage-pool-list-default
  output: storage-pool-list-output-default
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: storage-pool-list-default
spec: 
  meta: |
    id: liststoragepool
    apiVersion: openebs.io/v1alpha1
    kind: StoragePool
    action: list
    options: |-
      labelSelector: openebs.io/cas-type=cstor
  post: |
    {{- .JsonResult | saveAs "liststoragepool.list" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: storage-pool-list-output-default
spec:
  meta: |
    id: liststoragepooloutput
    action: output
    kind: CASStoragePoolList
    apiVersion: v1alpha1
  task: | 
    {{ .TaskResult.liststoragepool.list | toString }}
`

// StoragePoolArtifacts returns the CRDs required for latest version
func StoragePoolArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(StoragePool{})...)
	return
}

type StoragePool struct{}

// FetchYamls returns volume stats yamls
func (v StoragePool) FetchYamls() string {
	return storagePoolYaml
}
