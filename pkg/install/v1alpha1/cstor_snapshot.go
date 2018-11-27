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

const cstorSnapshotYamls = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-snapshot-create-default
spec:
  defaultConfig:
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  output: cstor-snapshot-create-output-default
  run:
    tasks:
    - cstor-snapshot-create-listtargetservice-default
    - cstor-snapshot-create-createsnapshot-default
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-snapshot-create-listtargetservice-default
spec:
  meta: |
    {{- $runNamespace := .Config.RunNamespace.value -}}
    {{- $pvcServiceAccount := .Config.PVCServiceAccountName.value | default "" -}}
    {{- if ne $pvcServiceAccount "" }}
    runNamespace: {{ .Snapshot.runNamespace | saveAs "readlistsvc.derivedNS" .TaskResult }}
    {{ else }}
    runNamespace: {{ $runNamespace | saveAs "readlistsvc.derivedNS" .TaskResult }}
    {{- end }}
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
  name: cstor-snapshot-create-createsnapshot-default
spec:
  meta: |
    id: createcstorsnap
    kind: Command
  post: |
    {{- $runCommand := create cstor snapshot | withoption "ip" .TaskResult.readlistsvc.clusterIP -}}
    {{- $runCommand := $runCommand | withoption "volname" .Snapshot.volumeName -}}
    {{- $runCommand | withoption "snapname" .Snapshot.owner | run | saveas "createcstorsnap" .TaskResult -}}
    {{- $err := .TaskResult.createcstorsnap.error | default "" | toString -}}
    {{- $err | empty | not | verifyErr $err | saveIf "createcstorsnap.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-snapshot-create-output-default
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
      labels:
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
    spec:
      casType: cstor
      volumeName: {{ .Snapshot.volumeName }}
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-snapshot-delete-default
spec:
  defaultConfig:
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  output: cstor-snapshot-delete-output-default
  run:
    tasks:
    - cstor-snapshot-delete-listtargetservice-default
    - cstor-snapshot-delete-deletesnapshot-default
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-snapshot-delete-listtargetservice-default
spec:
  meta: |
    {{- $runNamespace := .Config.RunNamespace.value -}}
    {{- $pvcServiceAccount := .Config.PVCServiceAccountName.value | default "" -}}
    {{- if ne $pvcServiceAccount "" }}
    runNamespace: {{ .Snapshot.runNamespace | saveAs "readlistsvc.derivedNS" .TaskResult }}
    {{ else }}
    runNamespace: {{ $runNamespace | saveAs "readlistsvc.derivedNS" .TaskResult }}
    {{- end }}
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
  name: cstor-snapshot-delete-deletesnapshot-default
spec:
  meta: |
    id: deletecstorsnap
    kind: Command
  post: |
    {{- $runCommand := delete cstor snapshot | withoption "ip" .TaskResult.readlistsvc.clusterIP -}}
    {{- $runCommand := $runCommand | withoption "volname" .Snapshot.volumeName -}}
    {{- $runCommand | withoption "snapname" .Snapshot.owner | run | saveas "deletecstorsnap" .TaskResult -}}
    {{- $err := .TaskResult.deletecstorsnap.error | default "" | toString -}}
    {{- $err | empty | not | verifyErr $err | saveIf "deletecstorsnap.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-snapshot-delete-output-default
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
      volumeName: {{ .Snapshot.volumeName }}
---`

// CstorSnapshotArtifacts returns the cstor snapshot related artifacts
// corresponding to latest version
func CstorSnapshotArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(cstorSnapshots{})...)
	return
}

type cstorSnapshots struct{}

// FetchYamls returns all the yamls related to cstor snapshots in a string
// format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (s cstorSnapshots) FetchYamls() string {
	return cstorSnapshotYamls
}
