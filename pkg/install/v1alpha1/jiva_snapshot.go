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

// TODO
// Rename this file by removing the version suffix information
package v1alpha1

const jivaSnapshotYamls = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-snapshot-create-default
spec:
  defaultConfig:
  - name: JivaRunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-snapshot-isvalidversion-default
    - jiva-snapshot-create-listsourcetargetservice-default
    - jiva-snapshot-create-invokehttp-default
  output: jiva-snapshot-create-output-default
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-snapshot-isvalidversion-default
spec:
  meta: |
    id: is090jivavolume
    runNamespace: {{ .Config.JivaRunNamespace }}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/controller-service=jiva-controller-svc,openebs.io/version=0.9.0
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "is090jivavolume.name" .TaskResult | noop -}}
    {{- .TaskResult.is090jivavolume.name | empty | not | versionMismatchErr "is not a jiva volume of 0.9.0 version" | saveIf "is090jivavolume.versionMismatchErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-snapshot-create-listsourcetargetservice-default
spec:
  meta: |
    id: readSourceSvc
    runNamespace: {{ .Config.JivaRunNamespace.value }}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/controller-service=jiva-controller-svc,openebs.io/persistent-volume={{ .Snapshot.volumeName }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readSourceSvc.name" .TaskResult | noop -}}
    {{- .TaskResult.readSourceSvc.name | notFoundErr "source volume target service not found" | saveIf "readSourceSvc.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.clusterIP}" | trim | saveAs "readSourceSvc.clusterIP" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-snapshot-create-invokehttp-default
spec:
  meta: |
    id: createJivaSnap
    runNamespace: {{ .Config.JivaRunNamespace }}
    kind: Command
  post: |
    {{- $store :=  storeAt .TaskResult -}}
    {{- $runner := storeRunner $store -}}
    {{- $volsUrl := print "http://" .TaskResult.readSourceSvc.clusterIP ":9501/v1/volumes" -}}
    {{- $volID := print "{.data[?(@.name=='" .Snapshot.volumeName "')].id} as id" -}}
    {{- select $volID | get http | withoption "url" $volsUrl | runas "getVol" $runner -}}
    {{- $snapUrl := print $volsUrl "/" .TaskResult.getVol.result.id "?action=snapshot" -}}
    {{- $body := dict "name" .Snapshot.owner | toJsonObj -}}
    {{- post http | withoption "url" $snapUrl | withoption "body" $body | runas "createSnap" $runner -}}
    {{- $err := .TaskResult.createSnap.error | default "" | toString -}}
    {{- $err | empty | not | verifyErr $err | saveIf "createJivaSnap.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-snapshot-create-output-default
spec:
  meta: |
    action: output
    id: cstorsnapshotoutput
    kind: CASSnapshot
    apiVersion: v1alpha1
  task: |-
    kind: CASSnapshot
    apiVersion: v1alpha1
    metadata:
      name: {{ .Snapshot.owner }}
    spec:
      casType: jiva
      volumeName: {{ .Snapshot.volumeName }}
---
`

// JivaSnapshotArtifacts returns the jiva snapshot related artifacts
// corresponding to latest version
func JivaSnapshotArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(jivaSnapshots{})...)
	return
}

type jivaSnapshots struct{}

// FetchYamls returns all the yamls related to jiva snapshot in a string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (j jivaSnapshots) FetchYamls() string {
	return jivaSnapshotYamls
}
