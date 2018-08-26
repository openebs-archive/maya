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

// CstorVolumeArtifactsFor070 returns the cstor volume related artifacts
// corresponding to version 0.7.0
func CstorVolumeArtifactsFor070() (list ArtifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(cstorVolumeYamlsFor070)...)
	return
}

// cstorVolumeYamlsFor070 returns all the yamls related to cstor volume in a
// string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func cstorVolumeYamlsFor070() string {
	return `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-volume-create-default-0.7.0
spec:
  defaultConfig:
  - name: VolumeControllerImage
    value: openebs/cstor-volume-mgmt:ci
  - name: VolumeTargetImage
    value: openebs/cstor-istgt:ci
  - name: VolumeMonitorImage
    value: openebs/m-exporter:ci
  - name: ReplicaCount
    value: "3"
  taskNamespace: {{.installer.namespace}}
  run:
    tasks:
    - cstor-volume-create-listcstorpoolcr-default-0.7.0
    - cstor-volume-create-puttargetservice-default-0.7.0
    - cstor-volume-create-putcstorvolumecr-default-0.7.0
    - cstor-volume-create-puttargetdeployment-default-0.7.0
    - cstor-volume-create-putcstorvolumereplicacr-default-0.7.0
  output: cstor-volume-create-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-volume-delete-default-0.7.0
spec:
  taskNamespace: {{.installer.namespace}}
  run:
    tasks:
    - cstor-volume-delete-listcstorvolumecr-default-0.7.0
    - cstor-volume-delete-listtargetservice-default-0.7.0
    - cstor-volume-delete-listtargetdeployment-default-0.7.0
    - cstor-volume-delete-listcstorvolumereplicacr-default-0.7.0
    - cstor-volume-delete-deletetargetservice-default-0.7.0
    - cstor-volume-delete-deletetargetdeployment-default-0.7.0
    - cstor-volume-delete-deletecstorvolumereplicacr-default-0.7.0
    - cstor-volume-delete-deletecstorvolumecr-default-0.7.0
  output: cstor-volume-delete-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-volume-read-default-0.7.0
spec:
  taskNamespace: {{.installer.namespace}}
  run:
    tasks:
    - cstor-volume-read-listtargetservice-default-0.7.0
    - cstor-volume-read-listcstorvolumereplicacr-default-0.7.0
    - cstor-volume-read-listtargetpod-default-0.7.0
  output: cstor-volume-read-output-default-0.7.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-volume-list-default-0.7.0
spec:
  taskNamespace: {{.installer.namespace}}
  run:
    tasks:
    - cstor-volume-list-listtargetservice-default-0.7.0
    - cstor-volume-list-listtargetpod-default-0.7.0
    - cstor-volume-list-listcstorvolumereplicacr-default-0.7.0
  output: cstor-volume-list-output-default-0.7.0
---
# runTask to list cstor pools
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-listcstorpoolcr-default-0.7.0
spec:
  meta: |
    id: cvolcreatelistpool
    runNamespace: openebs
    apiVersion: openebs.io/v1alpha1
    kind: CStorPool
    action: list
    options: |-
      labelSelector: openebs.io/storage-pool-claim={{ .Config.StoragePoolClaim.value }}
  post: |
    {{/*
    Check if enough online pools are present to create replicas.
    If pools are not present error out.
    Save the cstorpool's uid:name into .ListItems.cvolPoolList otherwise
    */}}
    {{- $replicaCount := int64 .Config.ReplicaCount.value | saveAs "rc" .ListItems -}}
    {{- $poolsList := jsonpath .JsonResult "{range .items[?(@.status.phase=="Online")]}pkey=pools,{@.metadata.uid}={@.metadata.name};{end}" | trim | default "" | splitList ";" -}}
    {{- $poolsList | saveAs "pl" .ListItems -}}
    {{- len $poolsList | gt $replicaCount | verifyErr "not enough pools available to create replicas" | saveAs "cvolcreatelistpool.verifyErr" .TaskResult | noop -}}
    {{- $poolsList | keyMap "cvolPoolList" .ListItems | noop -}}
---
# runTask to create cStor target service
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-puttargetservice-default-0.7.0
spec:
  meta: |
    apiVersion: v1
    kind: Service
    action: put
    id: cvolcreateputsvc
    runNamespace: openebs
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "cvolcreateputsvc.objectName" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.spec.clusterIP}" | trim | saveAs "cvolcreateputsvc.clusterIP" .TaskResult | noop -}}
  task: |
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        openebs.io/target-service: cstor-target-svc
        openebs.io/storage-engine-type: cstor
        openebs.io/persistent-volume: {{ .Volume.owner }}
      name: {{ .Volume.owner }}
    spec:
      ports:
      - name: cstor-iscsi
        port: 3260
        protocol: TCP
        targetPort: 3260
      - name: mgmt
        port: 6060
        targetPort: 6060
        protocol: TCP
      selector:
        openebs.io/target: cstor-target
        openebs.io/persistent-volume: {{ .Volume.owner }}
        app: cstor-volume-manager
---
# runTask to create cStorVolume
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-putcstorvolumecr-default-0.7.0
spec:
  meta: |
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolume
    id: cvolcreateputvolume
    runNamespace: openebs
    action: put
  post: |
    {{- jsonpath .JsonResult "{.metadata.uid}" | trim | saveAs "cvolcreateputvolume.cstorid" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "cvolcreateputvolume.objectName" .TaskResult | noop -}}
  task: |
    {{- $replicaCount := .Config.ReplicaCount.value | int64 -}}
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolume
    metadata:
      name: {{ .Volume.owner }}
      labels:
        openebs.io/persistent-volume: {{ .Volume.owner }}
    spec:
      targetIP: {{ .TaskResult.cvolcreateputsvc.clusterIP }}
      capacity: {{ .Volume.capacity }}
      nodeBase: iqn.2016-09.com.openebs.cstor
      iqn: iqn.2016-09.com.openebs.cstor:{{ .Volume.owner }}
      targetPortal: {{ .TaskResult.cvolcreateputsvc.clusterIP }}:3260
      targetPort: 3260
      status: ""
      replicationFactor: {{ $replicaCount }}
      consistencyFactor: {{ div $replicaCount 2 | floor | add1 }}
---
# runTask to create cStor target deployment
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-puttargetdeployment-default-0.7.0
spec:
  meta: |
    runNamespace: openebs
    apiVersion: apps/v1beta1
    kind: Deployment
    action: put
    id: cvolcreateputctrl
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "cvolcreateputctrl.objectName" .TaskResult | noop -}}
  task: |
    {{- $isMonitor := .Config.VolumeMonitor.enabled | default "true" | lower -}}
    apiVersion: apps/v1beta1
    Kind: Deployment
    metadata:
      name: {{ .Volume.owner }}-target
      labels:
        app: cstor-volume-manager
        openebs.io/storage-engine-type: cstor
        openebs.io/target: cstor-target
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
      annotations:
        {{- if eq $isMonitor "true" }}
        openebs.io/volume-monitor: "true"
        {{- end}}
        openebs.io/volume-type: cstor
    spec:
      replicas: 1
      selector:
        matchLabels:
          {{- if eq $isMonitor "true" }}
          monitoring: volume_exporter_prometheus
          {{- end}}
          openebs.io/target: cstor-target
          openebs.io/persistent-volume: {{ .Volume.owner }}
          app: cstor-volume-manager
      template:
        metadata:
          labels:
            {{- if eq $isMonitor "true" }}
            monitoring: volume_exporter_prometheus
            {{- end}}
            openebs.io/target: cstor-target
            openebs.io/persistent-volume: {{ .Volume.owner }}
            app: cstor-volume-manager
        spec:
          serviceAccountName: openebs-maya-operator
          containers:
          - image: {{ .Config.VolumeTargetImage.value }}
            name: cstor-istgt
            imagePullPolicy: IfNotPresent
            ports:
            - containerPort: 3260
              protocol: TCP
            securityContext:
              privileged: true
            volumeMounts:
            - name: sockfile
              mountPath: /var/run
            - name: conf
              mountPath: /usr/local/etc/istgt
            - name: dummyfile
              mountPath: /tmp/cstor
          {{- if eq $isMonitor "true" }}
          - image: {{ .Config.VolumeMonitorImage.value }}
            name: maya-volume-exporter
            args:
            - "-e=cstor"
            command: ["maya-exporter"]
            ports:
            - containerPort: 9500
              protocol: TCP
            volumeMounts:
            - name: sockfile
              mountPath: /configs
          {{- end}}
          - name: cstor-volume-mgmt
            image: {{ .Config.VolumeControllerImage.value }}
            imagePullPolicy: IfNotPresent
            ports:
            - containerPort: 80
            env:
            - name: OPENEBS_IO_CSTOR_VOLUME_ID
              value: {{ .TaskResult.cvolcreateputvolume.cstorid }}
            securityContext:
              privileged: true
            volumeMounts:
            - name: sockfile
              mountPath: /var/run
            - name: conf
              mountPath: /usr/local/etc/istgt
            - name: dummyfile
              mountPath: /tmp/cstor
          volumes:
          - name: sockfile
            emptyDir: {}
          - name: conf
            emptyDir: {}
          - name: dummyfile
            emptyDir: {}
---
# runTask to create cStorVolumeReplica/(s)
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-putcstorvolumereplicacr-default-0.7.0
spec:
  meta: |
    apiVersion: openebs.io/v1alpha1
    runNameSpace: openebs
    kind: CStorVolumeReplica
    action: put
    id: cstorvolumecreatereplica
    {{/*
    Fetch all the cStorPool uids into a list.
    Calculate the replica count
    Add as many poolUid to resources as there is replica count
    */}}
    {{- $poolUids := keys .ListItems.cvolPoolList.pools }}
    {{- $replicaCount := int64 .Config.ReplicaCount.value }}
    repeatWith:
      resources:
      {{- range $k, $v := $poolUids }}
      {{- if lt $k $replicaCount }}
      - {{ $v }}
      {{- end }}
      {{- end }}
  task: |
    kind: CStorVolumeReplica
    apiVersion: openebs.io/v1alpha1
    metadata:
      {{/*
      We pluck the cStorPool name from the map[uid]name:
      { "uid1":"name1","uid2":"name2","uid2":"name2" }
      The .ListItems.currentRepeatResource gives us the uid of one
      of the pools from resources list
      */}}
      name: {{ .Volume.owner }}-{{ pluck .ListItems.currentRepeatResource .ListItems.cvolPoolList.pools | first }}
      labels:
        cstorpool.openebs.io/name: {{ pluck .ListItems.currentRepeatResource .ListItems.cvolPoolList.pools | first }}
        cstorpool.openebs.io/uid: {{ .ListItems.currentRepeatResource }}
        cstorvolume.openebs.io/name: {{ .Volume.owner }}
        openebs.io/persistent-volume: {{ .Volume.owner }}
      finalizers: ["cstorvolumereplica.openebs.io/finalizer"]
    spec:
      capacity: {{ .Volume.capacity }}
      targetIP: {{ .TaskResult.cvolcreateputsvc.clusterIP }}
    status:
      # phase would be update by appropriate target
      phase: ""
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | addTo "cstorvolumecreatereplica.objectName" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.metadata.spec.capacity}" | trim | saveAs "cstorvolumecreatereplica.capacity" .TaskResult | noop -}}
    {{- $replicaPair := jsonpath .JsonResult "pkey=replicas,{@.metadata.name}={@.spec.capacity};" | trim | default "" | splitList ";" -}}
    {{- $replicaPair | keyMap "replicaList" .ListItems | noop -}}
---
# runTask to render volume create output as CASVolume
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-output-default-0.7.0
spec:
  meta: |
    action: output
    id: cstorvolumeoutput
    kind: CASVolume
    apiVersion: v1alpha1
  task: |
    kind: CASVolume
    apiVersion: v1alpha1
    metadata:
      name: {{ .Volume.owner }}
    spec:
      capacity: {{ .Volume.capacity }}
      iqn: iqn.2016-09.com.openebs.cstor:{{ .Volume.owner }}
      targetPortal: {{ .TaskResult.cvolcreateputsvc.clusterIP }}:3260
      targetIP: {{ .TaskResult.cvolcreateputsvc.clusterIP }}
      targetPort: 3260
      replicas: {{ .ListItems.replicaList.replicas | len }}
      casType: cstor
---
# runTask to list all cstor target deployment services
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-list-listtargetservice-default-0.7.0
spec:
  meta: |
    {{- /*
    Create and save list of namespaces to $nss.
    Iterate over each namespace and perform list task
    */ -}}
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
      labelSelector: openebs.io/target-service=cstor-target-svc
  post: |
    {{/*
    We create a pair of "clusterIP"=xxxxx and save it for corresponding volume
    The per volume is servicePair is identified by unique "namespace/vol-name" key
    */}}
    {{- $servicePairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.labels.openebs\\.io/pv},clusterIP={@.spec.clusterIP};{end}" | trim | default "" | splitList ";" -}}
    {{- $servicePairs | keyMap "volumeList" .ListItems | noop -}}
---
# runTask to list all cstor target pods
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-list-listtargetpod-default-0.7.0
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
      labelSelector: openebs.io/target=cstor-target
  post: |
    {{/*
    We create a pair of "targetIP"=xxxxx and save it for corresponding volume
    The per volume is servicePair is identified by unique "namespace/vol-name" key
    */}}
    {{- $targetPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.labels.openebs\\.io/pv},targetIP={@.status.podIP},targetStatus={@.status.containerStatuses[*].ready};{end}" | trim | default "" | splitList ";" -}}
    {{- $targetPairs | keyMap "volumeList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-list-listcstorvolumereplicacr-default-0.7.0
spec:
  meta: |
    runNamespace: openebs
    id: listlistrep
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolumeReplica
    action: list
  post: |
    {{- $replicaPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.labels.openebs\\.io/pv},replicaName={@.metadata.name},capacity={@.spec.capacity};{end}" | trim | default "" | splitList ";" -}}
    {{- $replicaPairs | keyMap "volumeList" .ListItems | noop -}}
---
# runTask to render volume list output
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-list-output-default-0.7.0
spec:
  meta: |
    id : listoutput
    action: output
    kind: CASVolumeList
    apiVersion: v1alpha1
  task: |
    kind: CASVolumeList
    items:
    {{/*
    We have a unique key for each volume in .ListItems.volumeList
    We iterate over it to extract various volume properties. These
    properties were set in preceeding list tasks,
    */}}
    {{- range $pkey, $map := .ListItems.volumeList }}
    {{- $capacity := pluck "capacity" $map | first | default "" | splitList ", " | first }}
    {{- $clusterIP := pluck "clusterIP" $map | first }}
    {{- $targetStatus := pluck "targetStatus" $map | first }}
    {{- $replicaName := pluck "replicaName" $map | first }}
    {{- $name := $pkey }}
      - kind: CASVolume
        apiVersion: v1alpha1
        metadata:
          name: {{ $name }}
          annotations:
            openebs.io/cluster-ips: {{ $clusterIP }}
            openebs.io/volume-size: {{ $capacity }}
            openebs.io/controller-status: {{ $targetStatus | replace "true" "running" | replace "false" "notready" }}
        spec:
          capacity: {{ $capacity }}
          iqn: iqn.2016-09.com.openebs.cstor:{{ $name }}
          targetPortal: {{ $clusterIP }}:3260
          targetIP: {{ $clusterIP }}
          targetPort: 3260
          replicas: {{ $replicaName | default "" | splitList ", " | len }}
          casType: cstor
    {{- end -}}
---
# runTask to list cStor target deployment service
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-listtargetservice-default-0.7.0
spec:
  meta: |
    runNamespace: openebs
    apiVersion: v1
    id: readlistsvc
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/target-service=cstor-target-svc,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistsvc.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistsvc.items | notFoundErr "target service not found" | saveIf "readlistsvc.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.clusterIP}" | trim | saveAs "readlistsvc.clusterIP" .TaskResult | noop -}}
---
# runTask to list all replicas of a volume
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-listcstorvolumereplicacr-default-0.7.0
spec:
  meta: |
    id: readlistrep
    runNamespace: openebs
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolumeReplica
    action: list
    options: |-
      labelSelector: openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistrep.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistrep.items | notFoundErr "replicas not found" | saveIf "readlistrep.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.capacity}" | trim | saveAs "readlistrep.capacity" .TaskResult | noop -}}
---
# runTask to list cStor volume target pods
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-listtargetpod-default-0.7.0
spec:
  meta: |
    runNamespace: openebs
    apiVersion: v1
    kind: Pod
    action: list
    id: readlistctrl
    options: |-
      labelSelector: openebs.io/target=cstor-target,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistctrl.items" .TaskResult | noop -}}
    {{- .TaskResult.readlistctrl.items | notFoundErr "target pod not found" | saveIf "readlistctrl.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.podIP}" | trim | saveAs "readlistctrl.podIP" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.containerStatuses[*].ready}" | trim | saveAs "readlistctrl.status" .TaskResult | noop -}}
---
# runTask to render output of read volume task as CAS Volume
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-output-default-0.7.0
spec:
  meta: |
    id : readoutput
    action: output
    kind: CASVolume
    apiVersion: v1alpha1
  task: |
    {{/* We calculate capacity of the volume here. Pickup capacity from cvr */}}
    {{- $capacity := .TaskResult.readlistrep.capacity | default "" | splitList " " | first -}}
    kind: CASVolume
    apiVersion: v1alpha1
    metadata:
      name: {{ .Volume.owner }}
      {{/* Render other values into annotation */}}
      annotations:
        openebs.io/controller-ips: {{ .TaskResult.readlistctrl.podIP | default "" | splitList " " | first }}
        openebs.io/controller-status: {{ .TaskResult.readlistctrl.status | default "" | splitList " " | join "," | replace "true" "running" | replace "false" "notready" }}
    spec:
      capacity: {{ $capacity }}
      iqn: iqn.2016-09.com.openebs.cstor:{{ .Volume.owner }}
      targetPortal: {{ .TaskResult.readlistsvc.clusterIP }}:3260
      targetIP: {{ .TaskResult.readlistsvc.clusterIP }}
      targetPort: 3260
      replicas: {{ .TaskResult.readlistrep.capacity | default "" | splitList " " | len }}
      casType: cstor
---
# runTask to list the cstorvolume that has to be deleted
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-listcstorvolumecr-default-0.7.0
spec:
  meta: |
    runNamespace: openebs
    id: deletelistcsv
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolume
    action: list
    options: |-
      labelSelector: openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistcsv.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistcsv.names | notFoundErr "cstor volume not found" | saveIf "deletelistcsv.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistcsv.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. cstor volume is not 1" | saveIf "deletelistcsv.verifyErr" .TaskResult | noop -}}
---
# runTask to list target service of volume to delete
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-listtargetservice-default-0.7.0
spec:
  meta: |
    id: deletelistsvc
    runNamespace: openebs
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/target-service=cstor-target-svc,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{/*
    Save the name of the service. Error if service is missing or more
    than one service exists
    */}}
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistsvc.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistsvc.names | notFoundErr "target service not found" | saveIf "deletelistsvc.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistsvc.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. of target services is not 1" | saveIf "deletelistsvc.verifyErr" .TaskResult | noop -}}
---
# runTask to list target deployment of volume to delete
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-listtargetdeployment-default-0.7.0
spec:
  meta: |
    id: deletelistctrl
    runNamespace: openebs
    apiVersion: apps/v1beta1
    kind: Deployment
    action: list
    options: |-
      labelSelector: openebs.io/target=cstor-target,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "deletelistctrl.names" .TaskResult | noop -}}
    {{- .TaskResult.deletelistctrl.names | notFoundErr "target deployment not found" | saveIf "deletelistctrl.notFoundErr" .TaskResult | noop -}}
    {{- .TaskResult.deletelistctrl.names | default "" | splitList " " | isLen 1 | not | verifyErr "total no. of target deployments is not 1" | saveIf "deletelistctrl.verifyErr" .TaskResult | noop -}}
---
# runTask to list cstorvolumereplica of volume to delete
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-listcstorvolumereplicacr-default-0.7.0
spec:
  meta: |
    id: deletelistcvr
    runNamespace: openebs
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolumeReplica
    action: list
    options: |-
      labelSelector: openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{/*
    List the names of the cstorvolumereplicas. Error if
    cstorvolumereplica is missing, save to a map cvrlist otherwise
    */}}
    {{- $cvrs := jsonpath .JsonResult "{range .items[*]}pkey=cvrs,{@.metadata.name}=;{end}" | trim | default "" | splitList ";" -}}
    {{- $cvrs | notFoundErr "cstor volume replica not found" | saveIf "deletelistcvr.notFoundErr" .TaskResult | noop -}}
    {{- $cvrs | keyMap "cvrlist" .ListItems | noop -}}
---
# runTask to delete cStor volume target service
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-deletetargetservice-default-0.7.0
spec:
  meta: |
    id: deletedeletesvc
    runNamespace: openebs
    apiVersion: v1
    kind: Service
    action: delete
    objectName: {{ .TaskResult.deletelistsvc.names }}
---
# runTask to delete cStor volume target deployment
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-deletetargetdeployment-default-0.7.0
spec:
  meta: |
    id: deletedeletectrl
    runNamespace: openebs
    apiVersion: apps/v1beta1
    kind: Deployment
    action: delete
    objectName: {{ .TaskResult.deletelistctrl.names }}
---
# runTask to delete cstorvolumereplica
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-deletecstorvolumereplicacr-default-0.7.0
spec:
  meta: |
    runNamespace: openebs
    id: deletedeletecvr
    action: delete
    kind: CStorVolumeReplica
    objectName: {{ keys .ListItems.cvrlist.cvrs | join "," }}
    apiVersion: openebs.io/v1alpha1
---
# runTask to delete cstorvolume
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-deletecstorvolumecr-default-0.7.0
spec:
  meta: |
    runNamespace: openebs
    id: deletedeletecsv
    action: delete
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolume
    objectName: {{ pluck "names" .TaskResult.deletelistcsv | first }}
---
# runTask to render output of deleted volume.
# This task only returns the name of volume that is deleted
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-output-default-0.7.0
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
