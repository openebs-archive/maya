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

// CstorSnapshotArtifactsFor070 returns the cstor snapshot related artifacts
// corresponding to version 0.7.0
func CstorSnapshotArtifactsFor070() (list ArtifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(cstorSnapshotYamlsFor070)...)
	return
}

// cstorSnapshotYamlsFor070 returns all the yamls related to cstor snapshot in a
// string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func cstorSnapshotYamlsFor070() string {
	return `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  labels:
    openebs.io/version: 0.7.0
  name: cstor-snapshot-create-default-0.7.0
spec:
  defaultConfig:
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  output: cstor-snapshot-create-output-default-0.7.0
  run:
    tasks:
    - cstor-snapshot-create-listtargetservice-default-0.7.0
    - cstor-snapshot-create-createsnapshot-default-0.7.0
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  labels:
    openebs.io/version: 0.7.0
  name: cstor-snapshot-create-listtargetservice-default-0.7.0
  namespace: openebs
spec:
  meta: |
    runNamespace: {{.Config.RunNamespace.value}}
    apiVersion: v1
    id: readlistsvc
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/target-service=cstor-target-svc,openebs.io/persistent-volume={{ .Snapshot.volumeName }}
  post: |-
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistsvc.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistsvc.items | notFoundErr "target service not found" | saveIf "readlistsvc.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.clusterIP}" | trim | saveAs "readlistsvc.clusterIP" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-snapshot-create-createsnapshot-default-0.7.0
  namespace: openebs
spec:
  meta: |
    id: createcstorsnap
    kind: Command
  post: |
    {{- create cstor snapshot | withoption "ip" .TaskResult.readlistsvc.clusterIP | withoption "volname" .Snapshot.volumeName | withoption "snapname" .Snapshot.owner | run | saveas "createcstorsnap" .TaskResult -}}
    {{- $err := .TaskResult.createcstorsnap.error | default "" | toString -}}
    {{- $err | empty | not | verifyErr $err | saveIf "createcstorsnap.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  labels:
    openebs.io/version: 0.7.0
  name: cstor-snapshot-create-output-default-0.7.0
  namespace: openebs
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
      casType: cstor
      volumeName: .Snapshot.volumeName
---`
}
