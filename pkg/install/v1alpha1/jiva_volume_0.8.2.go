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

const jivaVolumeYamls082 = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-read-default-0.8.2
  labels:
    openebs.io/version: "0.8.2"
spec:
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-isvalidversion-default-0.8.2
    - jiva-volume-read-listtargetservice-default-0.8.2
    - jiva-volume-read-listtargetpod-default-0.8.2
    - jiva-volume-read-listreplicapod-default-0.8.2
  output: jiva-volume-read-output-default-0.8.2
  fallback: jiva-volume-read-default-0.8.1
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-delete-default-0.8.2
  openebs.io/version: "0.8.2"
spec:
  defaultConfig:
  - name: ScrubImage
    value: "quay.io/openebs/openebs-tools:3.8"
  # RetainReplicaData specifies whether jiva replica data folder
  # should be cleared or retained.
  - name: RetainReplicaData
    enabled: "false"
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-isvalidversion-default-0.8.2
    - jiva-volume-delete-listtargetservice-default-0.8.2
    - jiva-volume-delete-listtargetdeployment-default-0.8.2
    - jiva-volume-delete-listreplicadeployment-default-0.8.2
    - jiva-volume-delete-deletetargetservice-default-0.8.2
    - jiva-volume-delete-deletetargetdeployment-default-0.8.2
    - jiva-volume-delete-listreplicapod-default-0.8.2
    - jiva-volume-delete-deletereplicadeployment-default-0.8.2
    - jiva-volume-delete-putreplicascrub-default-0.8.2
  output: jiva-volume-delete-output-default-0.8.2
  fallback: jiva-volume-delete-default-0.8.1
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-list-default-0.8.2
spec:
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-list-listtargetservice-default-0.8.2
    - jiva-volume-list-listtargetpod-default-0.8.2
    - jiva-volume-list-listreplicapod-default-0.8.2
    - jiva-volume-list-listpv-default-0.8.2
  output: jiva-volume-list-output-default-0.8.2
  fallback: jiva-volume-list-output-default-0.8.1
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-isvalidversion-default-0.8.2
spec:
  meta: |
    id: is081jivavolume
    runNamespace: {{.Volume.runNamespace}}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: vsm={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "is081jivavolume.name" .TaskResult | noop -}}
    {{- .TaskResult.is081jivavolume.name | empty | not | versionMismatchErr "is not a jiva volume of 0.8.1 version" | saveIf "is081jivavolume.versionMismatchErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listtargetservice-default-0.8.2
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
      labelSelector: openebs.io/controller-service=jiva-controller-svc
  post: |
    {{- $servicePairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.namespace}/{@.metadata.labels.openebs\\.io/persistent-volume},clusterIP={@.spec.clusterIP};{end}" | trim | default "" | splitList ";" -}}
    {{- $servicePairs | keyMap "volumeList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listtargetpod-default-0.8.2
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
      labelSelector: openebs.io/controller=jiva-controller
  post: |
    {{- $controllerPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.namespace}/{@.metadata.labels.openebs\\.io/persistent-volume},controllerIP={@.status.podIP},controllerStatus={@.status.containerStatuses[*].ready};{end}" | trim | default "" | splitList ";" -}}
    {{- $controllerPairs | keyMap "volumeList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listreplicapod-default-0.8.2
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
      labelSelector: openebs.io/replica=jiva-replica
  post: |
    {{- $replicaPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.namespace}/{@.metadata.labels.openebs\\.io/persistent-volume},replicaIP={@.status.podIP},replicaStatus={@.status.containerStatuses[*].ready},capacity={@.metadata.annotations.openebs\\.io/capacity};{end}" | trim | default "" | splitList ";" -}}
    {{- $replicaPairs | keyMap "volumeList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listpv-default-0.8.2
spec:
  meta: |
    id: listlistpv
    apiVersion: v1
    kind: PersistentVolume
    action: list
    options: |-
      labelSelector: openebs.io/cas-type=jiva
  post: |
     {{- $pvPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.name},accessModes={@.spec.accessModes[0]},storageClass={@.spec.storageClassName};{end}" | trim | default "" | splitList ";" -}}
     {{- $pvPairs | keyMap "pvList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-output-default-0.8.2
spec:
  meta: |
    id : listoutput
    action: output
    kind: CASVolumeList
    apiVersion: v1alpha1
  task: |
    kind: CASVolumeList
    items:
    {{- $pvList := .ListItems.pvList }}
    {{- range $pkey, $map := .ListItems.volumeList }}
    {{- $capacity := pluck "capacity" $map | first | default "" | splitList ", " | first }}
    {{- $clusterIP := pluck "clusterIP" $map | first }}
    {{- $controllerIP := pluck "controllerIP" $map | first }}
    {{- $controllerStatus := pluck "controllerStatus" $map | first }}
    {{- $replicaIP := pluck "replicaIP" $map | first }}
    {{- $replicaStatus := pluck "replicaStatus" $map | first }}
    {{- $name := $pkey | splitList "/" | last }}
    {{- $ns := $pkey | splitList "/" | first }}
    {{- $pvInfo := pluck $name $pvList | first }}
      - kind: CASVolume
        apiVersion: v1alpha1
        metadata:
          name: {{ $name }}
          namespace: {{ $ns }}
          annotations:
            openebs.io/storage-class: {{ $pvInfo.storageClass | default "" }}
            vsm.openebs.io/controller-ips: {{ $controllerIP }}
            vsm.openebs.io/cluster-ips: {{ $clusterIP }}
            vsm.openebs.io/iqn: iqn.2016-09.com.openebs.jiva:{{ $name }}
            vsm.openebs.io/replica-count: {{ $replicaIP | default "" | splitList ", " | len }}
            vsm.openebs.io/volume-size: {{ $capacity }}
            vsm.openebs.io/replica-ips: {{ $replicaIP }}
            vsm.openebs.io/replica-status: {{ $replicaStatus | replace "true" "running" | replace "false" "notready" }}
            vsm.openebs.io/controller-status: {{ $controllerStatus | replace "true" "running" | replace "false" "notready" | replace " " "," }}
            vsm.openebs.io/targetportals: {{ $clusterIP }}:3260
            openebs.io/controller-ips: {{ $controllerIP }}
            openebs.io/cluster-ips: {{ $clusterIP }}
            openebs.io/iqn: iqn.2016-09.com.openebs.jiva:{{ $name }}
            openebs.io/replica-count: {{ $replicaIP | default "" | splitList ", " | len }}
            openebs.io/volume-size: {{ $capacity }}
            openebs.io/replica-ips: {{ $replicaIP }}
            openebs.io/replica-status: {{ $replicaStatus | replace "true" "running" | replace "false" "notready" }}
            openebs.io/controller-status: {{ $controllerStatus | replace "true" "running" | replace "false" "notready" | replace " " "," }}
            openebs.io/targetportals: {{ $clusterIP }}:3260
        spec:
          accessMode: {{ $pvInfo.accessModes | default "" }}
          capacity: {{ $capacity }}
          iqn: iqn.2016-09.com.openebs.jiva:{{ $name }}
          targetPortal: {{ $clusterIP }}:3260
          replicas: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | len }}
          casType: jiva
          targetIP: {{ $clusterIP }}
          targetPort: 3260
    {{- end -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-listtargetservice-default-0.8.2
spec:
  meta: |
    id: readlistsvc
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/controller-service=jiva-controller-svc,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistsvc.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistsvc.items | notFoundErr "controller service not found" | saveIf "readlistsvc.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.clusterIP}" | trim | saveAs "readlistsvc.clusterIP" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-listtargetpod-default-0.8.2
spec:
  meta: |
    id: readlistctrl
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs.io/controller=jiva-controller,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistctrl.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistctrl.items | notFoundErr "controller pod not found" | saveIf "readlistctrl.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.nodeName}" | trim | saveAs "readlistctrl.targetNodeName" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.podIP}" | trim | saveAs "readlistctrl.podIP" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.containerStatuses[*].ready}" | trim | saveAs "readlistctrl.status" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].metadata.annotations.openebs\\.io/fs-type}" | trim | default "ext4" | saveAs "readlistctrl.fsType" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].metadata.annotations.openebs\\.io/lun}" | trim | default "0" | int | saveAs "readlistctrl.lun" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-listreplicapod-default-0.8.2
spec:
  meta: |
    id: readlistrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs.io/replica=jiva-replica,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistrep.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistrep.items | notFoundErr "replica pod(s) not found" | saveIf "readlistrep.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.podIP}" | trim | saveAs "readlistrep.podIP" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.containerStatuses[*].ready}" | trim | saveAs "readlistrep.status" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].metadata.annotations.openebs\\.io/capacity}" | trim | saveAs "readlistrep.capacity" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-output-default-0.8.2
spec:
  meta: |
    id : readoutput
    action: output
    kind: CASVolume
    apiVersion: v1alpha1
  task: |
    {{- $capacity := .TaskResult.readlistrep.capacity | default "" | splitList " " | first -}}
    kind: CASVolume
    apiVersion: v1alpha1
    metadata:
      name: {{ .Volume.owner }}
      annotations:
        vsm.openebs.io/controller-ips: {{ .TaskResult.readlistctrl.podIP | default "" | splitList " " | first }}
        vsm.openebs.io/cluster-ips: {{ .TaskResult.readlistsvc.clusterIP }}
        vsm.openebs.io/controller-node-name: {{ .TaskResult.readlistctrl.targetNodeName | default "" }}
        vsm.openebs.io/iqn: iqn.2016-09.com.openebs.jiva:{{ .Volume.owner }}
        vsm.openebs.io/replica-count: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | len }}
        vsm.openebs.io/volume-size: {{ $capacity }}
        vsm.openebs.io/replica-ips: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | join "," }}
        vsm.openebs.io/replica-status: {{ .TaskResult.readlistrep.status | default "" | splitList " " | join "," | replace "true" "running" | replace "false" "notready" }}
        vsm.openebs.io/controller-status: {{ .TaskResult.readlistctrl.status | default "" | splitList " " | join "," | replace "true" "running" | replace "false" "notready" }}
        vsm.openebs.io/targetportals: {{ .TaskResult.readlistsvc.clusterIP }}:3260
        openebs.io/controller-ips: {{ .TaskResult.readlistctrl.podIP | default "" | splitList " " | first }}
        openebs.io/cluster-ips: {{ .TaskResult.readlistsvc.clusterIP }}
        openebs.io/controller-node-name: {{ .TaskResult.readlistctrl.targetNodeName | default "" }}
        openebs.io/iqn: iqn.2016-09.com.openebs.jiva:{{ .Volume.owner }}
        openebs.io/replica-count: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | len }}
        openebs.io/volume-size: {{ $capacity }}
        openebs.io/replica-ips: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | join "," }}
        openebs.io/replica-status: {{ .TaskResult.readlistrep.status | default "" | splitList " " | join "," | replace "true" "running" | replace "false" "notready" }}
        openebs.io/controller-status: {{ .TaskResult.readlistctrl.status | default "" | splitList " " | join "," | replace "true" "running" | replace "false" "notready" }}
        openebs.io/targetportals: {{ .TaskResult.readlistsvc.clusterIP }}:3260
    spec:
      capacity: {{ $capacity }}
      targetPortal: {{ .TaskResult.readlistsvc.clusterIP }}:3260
      iqn: iqn.2016-09.com.openebs.jiva:{{ .Volume.owner }}
      replicas: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | len }}
      targetIP: {{ .TaskResult.readlistsvc.clusterIP }}
      targetPort: 3260
      lun: {{ .TaskResult.readlistctrl.lun }}
      fsType: {{ .TaskResult.readlistctrl.fsType }}
      casType: jiva
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-listreplicapod-default-0.8.2
spec:
  meta: |
    id: createlistrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs.io/replica=jiva-replica,openebs.io/persistent-volume={{ .Volume.owner }}
    retry: "12,10s"
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "createlistrep.items" .TaskResult | noop -}}
    {{- .TaskResult.createlistrep.items | empty | verifyErr "replica pod(s) not found" | saveIf "createlistrep.verifyErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.nodeName}" | trim | saveAs "createlistrep.nodeNames" .TaskResult | noop -}}
    {{- $expectedRepCount := .Config.ReplicaCount.value | int -}}
    {{- .TaskResult.createlistrep.nodeNames | default "" | splitList " " | isLen $expectedRepCount | not | verifyErr "number of replica pods does not match expected count" | saveIf "createlistrep.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listtargetservice-default-0.8.2
spec:
  meta: |
    id: deletelistsvc
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/controller-service=jiva-controller-svc,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistsvc.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistsvc.names | notFoundErr "controller service not found" | saveIf "deletelistsvc.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistsvc.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. of controller services is not 1" | saveIf "deletelistsvc.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listtargetdeployment-default-0.8.2
spec:
  meta: |
    id: deletelistctrl
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: list
    options: |-
      labelSelector: openebs.io/controller=jiva-controller,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistctrl.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistctrl.names | notFoundErr "controller deployment not found" | saveIf "deletelistctrl.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistctrl.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. of controller deployments is not 1" | saveIf "deletelistctrl.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listreplicadeployment-default-0.8.2
spec:
  meta: |
    id: deletelistrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: list
    options: |-
      labelSelector: openebs.io/replica=jiva-replica,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistrep.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistrep.names | notFoundErr "replica deployment not found" | saveIf "deletelistrep.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistrep.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. of replica deployments is not 1" | saveIf "deletelistrep.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-deletetargetservice-default-0.8.2
spec:
  meta: |
    id: deletedeletesvc
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Service
    action: delete
    objectName: {{ .TaskResult.deletelistsvc.names }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-deletetargetdeployment-default-0.8.2
spec:
  meta: |
    id: deletedeletectrl
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: delete
    objectName: {{ .TaskResult.deletelistctrl.names }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listreplicapod-default-0.8.2
spec:
  meta: |
    id: deletelistreppods
    runNamespace: {{ .Volume.runNamespace }}
    disable: {{ .Config.RetainReplicaData.enabled }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs.io/replica=jiva-replica,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- $nodesList := jsonpath .JsonResult "{range .items[*]}pkey=nodes,{@.spec.nodeName}={@.spec.volumes[?(@.name=='openebs')].hostPath.path};{end}" | trim | default "" | splitListTrim ";" -}}
    {{- $nodesList | keyMap "nodeJRPathList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-deletereplicadeployment-default-0.8.2
spec:
  meta: |
    id: deletedeleterep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: delete
    objectName: {{ .TaskResult.deletelistrep.names }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-putreplicascrub-default-0.8.2
spec:
  meta: |
    apiVersion: batch/v1
    runNamespace: {{ .Volume.runNamespace }}
    disable: {{ .Config.RetainReplicaData.enabled }}
    kind: Job
    action: put
    id: jivavolumedelreplicascrub
    {{- $nodeNames := keys .ListItems.nodeJRPathList.nodes }}
    repeatWith:
      resources:
      {{- range $k, $v := $nodeNames }}
      - {{ $v | quote }}
      {{- end }}
  task: |
    kind: Job
    apiVersion: batch/v1
    metadata:
      name: sjr-{{ .Volume.owner }}-{{randAlphaNum 4 |lower }}
      labels:
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/cas-type: jiva
    spec:
      backoffLimit: 4
      template:
        spec:
          restartPolicy: Never
          nodeSelector:
            kubernetes.io/hostname: {{ .ListItems.currentRepeatResource }}
          volumes:
          - name: replica-path
            hostPath:
              path: {{ pluck .ListItems.currentRepeatResource .ListItems.nodeJRPathList.nodes | first }}
              type: ""
          containers:
          - name: sjr
            image: {{ .Config.ScrubImage.value }}
            command: 
            - sh
            - -c
            - 'rm -rf /mnt/replica/*; sync; date > /mnt/replica/scrubbed.txt; sync;'
            volumeMounts:
            - mountPath: /mnt/replica
              name: replica-path
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | addTo "jivavolumedelreplicascrub.objectName" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-output-default-0.8.2
spec:
  meta: |
    id: deleteoutput
    action: output
    kind: CASVolume
    apiVersion: v1alpha1
  task: |
    kind: CASVolume
    apiVersion: v1alpha1
    metadata:
      name: {{ .Volume.owner }}
---
`

// JivaVolumeArtifactsFor082 returns the jiva volume related artifacts corresponding
// to latest version
func JivaVolumeArtifactsFor082() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(jivaVolumes082{})...)
	return
}

type jivaVolumes082 struct{}

// FetchYamls returns all the yamls related to jiva volume in a string
// format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (j jivaVolumes082) FetchYamls() string {
	return jivaVolumeYamls082
}
