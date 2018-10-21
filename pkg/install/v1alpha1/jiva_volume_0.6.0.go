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

const jivaVolumeYamls060 = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-read-default-0.6.0
spec:
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-read-getpv-default-0.6.0
    - jiva-volume-read-listtargetservice-default-0.6.0
    - jiva-volume-read-listtargetpod-default-0.6.0
    - jiva-volume-read-listreplicapod-default-0.6.0
  output: jiva-volume-read-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-delete-default-0.6.0
spec:
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-delete-listtargetservice-default-0.6.0
    - jiva-volume-delete-listtargetdeployment-default-0.6.0
    - jiva-volume-delete-listreplicadeployment-default-0.6.0
    - jiva-volume-delete-deletetargetservice-default-0.7.0
    - jiva-volume-delete-deletetargetdeployment-default-0.7.0
    - jiva-volume-delete-deletereplicadeployment-default-0.7.0
  output: jiva-volume-delete-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-list-default-0.6.0
spec:
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-list-listtargetservice-default-0.6.0
    - jiva-volume-list-listtargetpod-default-0.6.0
    - jiva-volume-list-listreplicapod-default-0.6.0
  output: jiva-volume-list-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listtargetservice-default-0.6.0
spec:
  meta: |
    {{- $nss := .Volume.runNamespace | default "" | splitList ", " -}}
    id: listlistsvc
    repeatWith:
      metas:
      {{- range $k, $ns := $nss }}
      - runNamespace: {{ $ns }}
      {{- end }}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/controller-service=jiva-controller-service
  post: |
    {{- $servicePairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.namespace}/{@.metadata.labels.vsm},clusterIP={@.spec.clusterIP};{end}" | trim | default "" | splitList ";" -}}
    {{- $servicePairs | keyMap "volumeList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listtargetpod-default-0.6.0
spec:
  meta: |
    {{- $nss := .Volume.runNamespace | default "" | splitList ", " -}}
    id: listlistctrl
    repeatWith:
      metas:
      {{- range $k, $ns := $nss }}
      - runNamespace: {{ $ns }}
      {{- end }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs/controller=jiva-controller
  post: |
    {{- $controllerPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.namespace}/{@.metadata.labels.vsm},controllerIP={@.status.podIP},controllerStatus={@.status.containerStatuses[*].ready};{end}" | trim | default "" | splitList ";" -}}
    {{- $controllerPairs | keyMap "volumeList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listreplicapod-default-0.6.0
spec:
  meta: |
    {{- $nss := .Volume.runNamespace | default "" | splitList ", " -}}
    id: listlistrep
    repeatWith:
      metas:
      {{- range $k, $ns := $nss }}
      - runNamespace: {{ $ns }}
      {{- end }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs/replica=jiva-replica
  post: |
    {{- $replicaPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.namespace}/{@.metadata.labels.vsm},replicaIP={@.status.podIP},replicaStatus={@.status.containerStatuses[*].ready},capacity=Unknown;{end}" | trim | default "" | splitList ";" -}}
    {{- $replicaPairs | keyMap "volumeList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-getpv-default-0.6.0
spec:
  meta: |
    id: readgetpv
    apiVersion: v1
    runNamespace: default
    kind: PersistentVolume
    objectName: {{ .Volume.owner }}
    action: get
  post: |
    {{- $capacity := jsonpath .JsonResult "{.spec.capacity.storage}" -}}
    {{- trim $capacity | saveAs "readgetpv.capacity" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-listtargetservice-default-0.6.0
spec:
  meta: |
    id: readlistsvc
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs/controller-service=jiva-controller-service,vsm={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistsvc.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistsvc.items | notFoundErr "controller service not found" | saveIf "readlistsvc.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.clusterIP}" | trim | saveAs "readlistsvc.clusterIP" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-listtargetpod-default-0.6.0
spec:
  meta: |
    id: readlistctrl
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs/controller=jiva-controller,vsm={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistctrl.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistctrl.items | notFoundErr "controller pod not found" | saveIf "readlistctrl.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.podIP}" | trim | saveAs "readlistctrl.podIP" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.containerStatuses[*].ready}" | trim | saveAs "readlistctrl.status" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-listreplicapod-default-0.6.0
spec:
  meta: |
    id: readlistrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs/replica=jiva-replica,vsm={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistrep.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistrep.items | notFoundErr "replica pod(s) not found" | saveIf "readlistrep.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.podIP}" | trim | saveAs "readlistrep.podIP" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.containerStatuses[*].ready}" | trim | saveAs "readlistrep.status" .TaskResult | noop -}}
    {{- .TaskResult.readgetpv.capacity | saveAs "readlistrep.capacity" .TaskResult -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listtargetservice-default-0.6.0
spec:
  meta: |
    id: deletelistsvc
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs/controller-service=jiva-controller-service,vsm={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistsvc.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistsvc.names | notFoundErr "controller service not found" | saveIf "deletelistsvc.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistsvc.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. of controller services is not 1" | saveIf "deletelistsvc.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listtargetdeployment-default-0.6.0
spec:
  meta: |
    id: deletelistctrl
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: list
    options: |-
      labelSelector: openebs/controller=jiva-controller,vsm={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistctrl.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistctrl.names | notFoundErr "controller deployment not found" | saveIf "deletelistctrl.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistctrl.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. of controller deployments is not 1" | saveIf "deletelistctrl.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listreplicadeployment-default-0.6.0
spec:
  meta: |
    id: deletelistrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: list
    options: |-
      labelSelector: openebs/replica=jiva-replica,vsm={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistrep.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistrep.names | notFoundErr "replica deployment not found" | saveIf "deletelistrep.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistrep.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. of replica deployments is not 1" | saveIf "deletelistrep.verifyErr" .TaskResult | noop -}}
---
`

// JivaVolumeArtifactsFor060 returns the jiva volume related artifacts
// corresponding to version 0.6.0
func JivaVolumeArtifactsFor060() (list ArtifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(jivaVolumeYamlsFor060)...)
	return
}

// jivaVolumeYamlsFor060 returns all the yamls related to jiva volume in a
// string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func jivaVolumeYamlsFor060() string {
	return jivaVolumeYamls060
}
