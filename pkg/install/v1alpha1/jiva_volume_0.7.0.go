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

// JivaVolumeArtifactsFor070 returns the jiva volume related artifacts
// corresponding to version 0.7.0
func JivaVolumeArtifactsFor070() (list ArtifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(jivaVolumeYamlsFor070)...)
	return
}

// jivaVolumeYamlsFor070 returns all the yamls related to jiva volume in a
// string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func jivaVolumeYamlsFor070() string {
	return `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-read-default-0.7.0
spec:
  taskNamespace: {{.installer.namespace}}
  run:
    tasks:
    - jiva-volume-read-listtargetservice-default-0.7.0
    - jiva-volume-read-listtargetpod-default-0.7.0
    - jiva-volume-read-listreplicapod-default-0.7.0
  output: jiva-volume-read-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-create-default-0.7.0
spec:
  defaultConfig:
  - name: ControllerImage
    value: openebs/jiva:0.6.0
  - name: ReplicaImage
    value: openebs/jiva:0.6.0
  - name: VolumeMonitorImage
    value: openebs/m-exporter:ci
  - name: ReplicaCount
    value: "3"
  - name: StoragePool
    value: default
  - name: VolumeMonitor
    enabled: "true"
  - name: EvictionTolerations
    value: |-
      t1:
        effect: NoExecute
        key: node.alpha.kubernetes.io/notReady
        operator: Exists
      t2:
        effect: NoExecute
        key: node.alpha.kubernetes.io/unreachable
        operator: Exists
      t3:
        effect: NoExecute
        key: node.kubernetes.io/not-ready
        operator: Exists
      t4:
        effect: NoExecute
        key: node.kubernetes.io/unreachable
        operator: Exists
      t5:
        effect: NoExecute
        key: node.kubernetes.io/out-of-disk
        operator: Exists
      t6:
        effect: NoExecute
        key: node.kubernetes.io/memory-pressure
        operator: Exists
      t7:
        effect: NoExecute
        key: node.kubernetes.io/disk-pressure
        operator: Exists
      t8:
        effect: NoExecute
        key: node.kubernetes.io/network-unavailable
        operator: Exists
      t9:
        effect: NoExecute
        key: node.kubernetes.io/unschedulable
        operator: Exists
      t10:
        effect: NoExecute
        key: node.cloudprovider.kubernetes.io/uninitialized
        operator: Exists
  - name: NodeAffinityRequiredSchedIgnoredExec
    value: |-
      t1:
        key: beta.kubernetes.io/os
        operator: In
        values:
        - linux
  - name: NodeAffinityPreferredSchedIgnoredExec
    value: |-
      t1:
        key: some-node-label-key
        operator: In
        values:
        - some-node-label-value
  taskNamespace: {{.installer.namespace}}
  run:
    tasks:
    - jiva-volume-create-getstorageclass-default-0.7.0
    - jiva-volume-create-puttargetservice-default-0.7.0
    - jiva-volume-create-getstoragepoolcr-default-0.7.0
    - jiva-volume-create-puttargetdeployment-default-0.7.0
    - jiva-volume-create-putreplicadeployment-default-0.7.0
  output: jiva-volume-create-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-delete-default-0.7.0
spec:
  taskNamespace: {{.installer.namespace}}
  run:
    tasks:
    - jiva-volume-delete-listtargetservice-default-0.7.0
    - jiva-volume-delete-listtargetdeployment-default-0.7.0
    - jiva-volume-delete-listreplicadeployment-default-0.7.0
    - jiva-volume-delete-deletetargetservice-default-0.7.0
    - jiva-volume-delete-deletetargetdeployment-default-0.7.0
    - jiva-volume-delete-deletereplicadeployment-default-0.7.0
  output: jiva-volume-delete-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-list-default-0.7.0
spec:
  taskNamespace: {{.installer.namespace}}
  run:
    tasks:
    - jiva-volume-list-listtargetservice-default-0.7.0
    - jiva-volume-list-listtargetpod-default-0.7.0
    - jiva-volume-list-listreplicapod-default-0.7.0
  output: jiva-volume-list-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listtargetservice-default-0.7.0
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
  name: jiva-volume-list-listtargetpod-default-0.7.0
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
  name: jiva-volume-list-listreplicapod-default-0.7.0
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
  name: jiva-volume-list-output-default-0.7.0
spec:
  meta: |
    id : listoutput
    action: output
    kind: CASVolumeList
    apiVersion: v1alpha1
  task: |
    kind: CASVolumeList
    items:
    {{- range $pkey, $map := .ListItems.volumeList }}
    {{- $capacity := pluck "capacity" $map | first | default "" | splitList ", " | first }}
    {{- $clusterIP := pluck "clusterIP" $map | first }}
    {{- $controllerIP := pluck "controllerIP" $map | first }}
    {{- $controllerStatus := pluck "controllerStatus" $map | first }}
    {{- $replicaIP := pluck "replicaIP" $map | first }}
    {{- $replicaStatus := pluck "replicaStatus" $map | first }}
    {{- $name := $pkey | splitList "/" | last }}
    {{- $ns := $pkey | splitList "/" | first }}
      - kind: CASVolume
        apiVersion: v1alpha1
        metadata:
          name: {{ $name }}
          namespace: {{ $ns }}
          annotations:
            vsm.openebs.io/controller-ips: {{ $controllerIP }}
            vsm.openebs.io/cluster-ips: {{ $clusterIP }}
            vsm.openebs.io/iqn: iqn.2016-09.com.openebs.jiva:{{ $name }}
            vsm.openebs.io/replica-count: {{ $replicaIP | default "" | splitList ", " | len }}
            vsm.openebs.io/volume-size: {{ $capacity }}
            vsm.openebs.io/replica-ips: {{ $replicaIP }}
            vsm.openebs.io/replica-status: {{ $replicaStatus | replace "true" "running" | replace "false" "notready" }}
            vsm.openebs.io/controller-status: {{ $controllerStatus | replace "true" "running" | replace "false" "notready" | replace " " "," }}
            vsm.openebs.io/targetportals: {{ $clusterIP }}:3260
        spec:
          capacity: {{ $capacity }}
          casType: jiva
    {{- end -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-listtargetservice-default-0.7.0
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
  name: jiva-volume-read-listtargetpod-default-0.7.0
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
    {{- jsonpath .JsonResult "{.items[*].status.podIP}" | trim | saveAs "readlistctrl.podIP" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.containerStatuses[*].ready}" | trim | saveAs "readlistctrl.status" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-listreplicapod-default-0.7.0
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
  name: jiva-volume-read-output-default-0.7.0
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
        vsm.openebs.io/iqn: iqn.2016-09.com.openebs.jiva:{{ .Volume.owner }}
        vsm.openebs.io/replica-count: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | len }}
        vsm.openebs.io/volume-size: {{ $capacity }}
        vsm.openebs.io/replica-ips: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | join "," }}
        vsm.openebs.io/replica-status: {{ .TaskResult.readlistrep.status | default "" | splitList " " | join "," | replace "true" "running" | replace "false" "notready" }}
        vsm.openebs.io/controller-status: {{ .TaskResult.readlistctrl.status | default "" | splitList " " | join "," | replace "true" "running" | replace "false" "notready" }}
        vsm.openebs.io/targetportals: {{ .TaskResult.readlistsvc.clusterIP }}:3260
    spec:
      capacity: {{ $capacity }}
      targetPortal: {{ .TaskResult.readlistsvc.clusterIP }}:3260
      iqn: iqn.2016-09.com.openebs.jiva:{{ .Volume.owner }}
      replicas: {{ .TaskResult.readlistrep.podIP | default "" | splitList " " | len }}
      casType: jiva
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-puttargetservice-default-0.7.0
spec:
  meta: |
    id: createputsvc
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Service
    action: put
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "createputsvc.objectName" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.spec.clusterIP}" | trim | saveAs "createputsvc.clusterIP" .TaskResult | noop -}}
  task: |
    apiVersion: v1
    Kind: Service
    metadata:
      labels:
        openebs/controller-service: jiva-controller-service
        openebs/volume-provisioner: jiva
        vsm: {{ .Volume.owner }}
        pvc: {{ .Volume.pvc }}
        openebs.io/storage-engine-type: jiva
        openebs.io/controller-service: jiva-controller-svc
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
      name: {{ .Volume.owner }}-ctrl-svc
    spec:
      ports:
      - name: iscsi
        port: 3260
        protocol: TCP
        targetPort: 3260
      - name: api
        port: 9501
        protocol: TCP
        targetPort: 9501
      selector:
        openebs/controller: jiva-controller
        vsm: {{ .Volume.owner }}
        openebs.io/controller: jiva-controller
        openebs.io/persistent-volume: {{ .Volume.owner }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-getstoragepoolcr-default-0.7.0
spec:
  meta: |
    id: creategetpath
    apiVersion: openebs.io/v1alpha1
    kind: StoragePool
    objectName: {{ .Config.StoragePool.value }}
    action: get
  post: |
    {{- jsonpath .JsonResult "{.spec.path}" | trim | saveAs "creategetpath.storagePoolPath" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-getstorageclass-default-0.7.0
spec:
  meta: |
    id: creategetsc
    apiVersion: storage.k8s.io/v1
    kind: StorageClass
    objectName: {{ .Volume.storageclass }}
    action: get
  post: |
    {{- $resourceVer := jsonpath .JsonResult "{.metadata.resourceVersion}" -}}
    {{- trim $resourceVer | saveAs "creategetsc.storageClassVersion" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-listreplicapod-default-0.7.0
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
  name: jiva-volume-create-patchreplicadeployment-default-0.7.0
spec:
  meta: |
    id: createpatchrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    objectName: {{ .Volume.owner }}-rep
    action: patch
  task: |
      {{- $isNodeAffinityRSIE := .Config.NodeAffinityRequiredSchedIgnoredExec.value | default "false" -}}
      {{- $nodeAffinityRSIEVal := fromYaml .Config.NodeAffinityRequiredSchedIgnoredExec.value -}}
      {{- $nodeNames := .TaskResult.createlistrep.nodeNames -}}
      type: strategic
      pspec: |-
        spec:
          template:
            spec:
              affinity:
                nodeAffinity:
                  {{- if ne $isNodeAffinityRSIE "false" }}
                  requiredDuringSchedulingIgnoredDuringExecution:
                    nodeSelectorTerms:
                    - matchExpressions:
                      {{- range $k, $v := $nodeAffinityRSIEVal }}
                      - 
                      {{- range $kk, $vv := $v }}
                        {{ $kk }}: {{ $vv }}
                      {{- end }}
                      {{- end }}
                      - key: kubernetes.io/hostname
                        operator: In
                        values:
                        {{- if ne $nodeNames "" }}
                        {{- $nodeNamesMap := $nodeNames | split " " }}
                        {{- range $k, $v := $nodeNamesMap }}
                        - {{ $v }}
                        {{- end }}
                        {{- end }}
                  {{- else }}
                  requiredDuringSchedulingIgnoredDuringExecution:
                    nodeSelectorTerms:
                    - matchExpressions:
                      - key: kubernetes.io/hostname
                        operator: In
                        values:
                        {{- if ne $nodeNames "" }}
                        {{- $nodeNamesMap := $nodeNames | split " " }}
                        {{- range $k, $v := $nodeNamesMap }}
                        - {{ $v }}
                        {{- end }}
                        {{- end }}
                  {{- end }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-puttargetdeployment-default-0.7.0
spec:
  meta: |
    id: createputctrl
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: put
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "createputctrl.objectName" .TaskResult | noop -}}
  task: |
    {{- $isMonitor := .Config.VolumeMonitor.enabled | default "true" | lower -}}
    apiVersion: extensions/v1beta1
    Kind: Deployment
    metadata:
      labels:
        openebs/volume-provisioner: jiva
        openebs/controller: jiva-controller
        vsm: {{ .Volume.owner }}
        pvc: {{ .Volume.pvc }}
        {{- if eq $isMonitor "true" }}
        monitoring: "volume_exporter_prometheus"
        {{- end}}
        openebs.io/storage-engine-type: jiva
        openebs.io/controller: jiva-controller
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
      annotations:
        {{- if eq $isMonitor "true" }}
        openebs.io/volume-monitor: "true"
        {{- end}}
        openebs.io/volume-type: jiva
      name: {{ .Volume.owner }}-ctrl
    spec:
      replicas: 1
      selector:
        matchLabels:
          openebs/volume-provisioner: jiva
          openebs/controller: jiva-controller
          vsm: {{ .Volume.owner }}
          pvc: {{ .Volume.pvc }}
          {{- if eq $isMonitor "true" }}
          monitoring: volume_exporter_prometheus
          {{- end}}
          openebs.io/controller: jiva-controller
          openebs.io/persistent-volume: {{ .Volume.owner }}
      template:
        metadata:
          labels:
            openebs/volume-provisioner: jiva
            openebs/controller: jiva-controller
            vsm: {{ .Volume.owner }}
            pvc: {{ .Volume.pvc }}
            {{- if eq $isMonitor "true" }}
            monitoring: volume_exporter_prometheus
            {{- end}}
            openebs.io/controller: jiva-controller
            openebs.io/persistent-volume: {{ .Volume.owner }}
            openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
        spec:
          containers:
          - args:
            - controller
            - --frontend
            - gotgt
            - --clusterIP
            - {{ .TaskResult.createputsvc.clusterIP }}
            - {{ .Volume.owner }}
            command:
            - launch
            image: {{ .Config.ControllerImage.value }}
            name: {{ .Volume.owner }}-ctrl-con
            env:
            - name: "REPLICATION_FACTOR"
              value: {{ .Config.ReplicaCount.value }}
            ports:
            - containerPort: 3260
              protocol: TCP
            - containerPort: 9501
              protocol: TCP
          {{- if eq $isMonitor "true" }}
          - args:
            - -c=http://127.0.0.1:9501
            command:
            - maya-exporter
            image: {{ .Config.VolumeMonitorImage.value }}
            name: maya-volume-exporter
            ports:
            - containerPort: 9500
              protocol: TCP
          {{- end}}
          tolerations:
          - effect: NoExecute
            key: node.alpha.kubernetes.io/notReady
            operator: Exists
            tolerationSeconds: 0
          - effect: NoExecute
            key: node.alpha.kubernetes.io/unreachable
            operator: Exists
            tolerationSeconds: 0
          - effect: NoExecute
            key: node.kubernetes.io/not-ready
            operator: Exists
            tolerationSeconds: 0
          - effect: NoExecute
            key: node.kubernetes.io/unreachable
            operator: Exists
            tolerationSeconds: 0
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-putreplicadeployment-default-0.7.0
spec:
  meta: |
    id: createputrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: put
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "createputrep.objectName" .TaskResult | noop -}}
  task: |
    {{- $isEvictionTolerations := .Config.EvictionTolerations.value | default "false" -}}
    {{- $evictionTolerationsVal := fromYaml .Config.EvictionTolerations.value -}}
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      labels:
        openebs/replica: jiva-replica
        openebs/volume-provisioner: jiva
        vsm: {{ .Volume.owner }}
        pvc: {{ .Volume.pvc }}
        openebs.io/storage-engine-type: jiva
        openebs.io/replica: jiva-replica
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
      annotations:
        openebs.io/capacity: {{ .Volume.capacity }}
        openebs.io/storage-pool: {{ .Config.StoragePool.value }}
      name: {{ .Volume.owner }}-rep
    spec:
      replicas: {{ .Config.ReplicaCount.value }}
      selector:
        matchLabels:
          openebs/replica: jiva-replica
          openebs/volume-provisioner: jiva
          vsm: {{ .Volume.owner }}
          pvc: {{ .Volume.pvc }}
          openebs.io/replica: jiva-replica
          openebs.io/persistent-volume: {{ .Volume.owner }}
      template:
        metadata:
          labels:
            openebs/replica: jiva-replica
            openebs/volume-provisioner: jiva
            vsm: {{ .Volume.owner }}
            pvc: {{ .Volume.pvc }}
            openebs.io/replica: jiva-replica
            openebs.io/persistent-volume: {{ .Volume.owner }}
            openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
          annotations:
            openebs.io/capacity: {{ .Volume.capacity }}
            openebs.io/storage-pool: {{ .Config.StoragePool.value }}
        spec:
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
              - labelSelector:
                  matchLabels:
                    openebs/replica: jiva-replica
                    openebs.io/replica: jiva-replica
                    vsm: {{ .Volume.owner }}
                    openebs.io/persistent-volume: {{ .Volume.owner }}
                topologyKey: kubernetes.io/hostname
          containers:
          - args:
            - replica
            - --frontendIP
            - {{ .TaskResult.createputsvc.clusterIP }}
            - --size
            - {{ .Volume.capacity }}
            - /openebs
            command:
            - launch
            image: {{ .Config.ReplicaImage.value }}
            name: {{ .Volume.owner }}-rep-con
            ports:
            - containerPort: 9502
              protocol: TCP
            - containerPort: 9503
              protocol: TCP
            - containerPort: 9504
              protocol: TCP
            volumeMounts:
            - name: openebs
              mountPath: /openebs
          tolerations:
          {{- if ne $isEvictionTolerations "false" }}
          {{- range $k, $v := $evictionTolerationsVal }}
          - 
          {{- range $kk, $vv := $v }}
            {{ $kk }}: {{ $vv }}
          {{- end }}
          {{- end }}
          {{- end }}
          volumes:
          - name: openebs
            hostPath:
              path: {{ .TaskResult.creategetpath.storagePoolPath }}/{{ .Volume.owner }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-output-default-0.7.0
spec:
  meta: |
    id: createoutput
    action: output
    kind: CASVolume
    apiVersion: v1alpha1
  task: |
    kind: CASVolume
    apiVersion: v1alpha1
    metadata:
      name: {{ .Volume.owner }}
      annotations:
        openebs.io/storageclass-version: {{ .TaskResult.creategetsc.storageClassVersion }}
    spec:
      capacity: {{ .Volume.capacity }}
      targetPortal: {{ .TaskResult.createputsvc.clusterIP }}:3260
      iqn: iqn.2016-09.com.openebs.jiva:{{ .Volume.owner }}
      replicas: {{ .Config.ReplicaCount.value }}
      casType: jiva
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listtargetservice-default-0.7.0
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
  name: jiva-volume-delete-listtargetdeployment-default-0.7.0
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
  name: jiva-volume-delete-listreplicadeployment-default-0.7.0
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
  name: jiva-volume-delete-deletetargetservice-default-0.7.0
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
  name: jiva-volume-delete-deletetargetdeployment-default-0.7.0
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
  name: jiva-volume-delete-deletereplicadeployment-default-0.7.0
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
  name: jiva-volume-delete-output-default-0.7.0
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
}
