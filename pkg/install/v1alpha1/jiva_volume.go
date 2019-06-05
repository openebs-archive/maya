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

const jivaVolumeYamls = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-read-default
spec:
  defaultConfig:
  - name: OpenEBSNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-isvalidversion-default
    - jiva-volume-podsinopenebsns-default
    - jiva-volume-read-listtargetservice-default
    - jiva-volume-read-listtargetpod-default
    - jiva-volume-read-listreplicapod-default
    - jiva-volume-read-verifyreplicationfactor-default
    - jiva-volume-read-patchreplicadeployment-default
  output: jiva-volume-read-output-default
  fallback: jiva-volume-read-default-0.6.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-create-default
spec:
  defaultConfig:
  - name: OpenEBSNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  # The value will be filled by the installer. 
  # Administrator can select to deploy the jiva pods
  # in the openebs namespace instead of target namespace
  # for the following reasons:
  # - avoid granting access to hostpath in user namespace
  # - manage all the storage pods in a single namespace
  # By default, this is set to false to retain 
  # backward compatability. However in the future releases
  # if more and more deployments prefer to use this option, 
  # the default can be set to deploy in openebs. 
  - name: DeployInOpenEBSNamespace
    enabled: "false"
  - name: ControllerImage
    value: {{env "OPENEBS_IO_JIVA_CONTROLLER_IMAGE" | default "openebs/jiva:latest"}}
  - name: ReplicaImage
    value: {{env "OPENEBS_IO_JIVA_REPLICA_IMAGE" | default "openebs/jiva:latest"}}
  - name: VolumeMonitorImage
    value: {{env "OPENEBS_IO_VOLUME_MONITOR_IMAGE" | default "openebs/m-exporter:latest"}}
  - name: ReplicaCount
    value: {{env "OPENEBS_IO_JIVA_REPLICA_COUNT" | default "3" | quote }}
  - name: StoragePool
    value: "default"
  - name: VolumeMonitor
    enabled: "true"
  # TargetTolerations allows you to specify the tolerations for target
  # Example:
  # - name: TargetTolerations
  #   value: |-
  #      t1:
  #        key: "key1"
  #        operator: "Equal"
  #        value: "value1"
  #        effect: "NoSchedule"
  #      t2:
  #        key: "key1"
  #        operator: "Equal"
  #        value: "value1"
  #        effect: "NoExecute"
  - name: TargetTolerations
    value: "none"
  # ReplicaTolerations allows you to specify the tolerations for target
  # Example:
  # - name: ReplicaTolerations
  #   value: |-
  #      t1:
  #        key: "key1"
  #        operator: "Equal"
  #        value: "value1"
  #        effect: "NoSchedule"
  #      t2:
  #        key: "key1"
  #        operator: "Equal"
  #        value: "value1"
  #        effect: "NoExecute"
  - name: ReplicaTolerations
    value: "none"
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
  # TargetResourceRequests allow you to specify resource requests that need to be available
  # before scheduling the containers. If not specified, the default is to use the limits
  # from TargetResourceLimits or the default requests set in the cluster.
  - name: TargetResourceRequests
    value: "none"
  # TargetResourceLimits allow you to set the limits on memory and cpu for jiva
  # target pods. The resource and limit value should be in the same format as
  # expected by Kubernetes. Example:
  #- name: TargetResourceLimits
  #  value: |-
  #      memory: 1Gi
  #      cpu: 200m
  # By default, the resource limits are disabled.
  - name: TargetResourceLimits
    value: "none"
  # ReplicaResourceRequests allow you to specify resource requests that need to be available
  # before scheduling the containers. If not specified, the default is to use the limits
  # from ReplicaResourceLimits or the default requests set in the cluster.
  - name: ReplicaResourceRequests
    value: "none"
  # ReplicaResourceLimits allow you to set the limits on memory and cpu for jiva
  # replica pods. The resource and limit value should be in the same format as
  # expected by Kubernetes. Example:
  - name: ReplicaResourceLimits
    value: "none"
  # AuxResourceRequests allow you to set requests on side cars. Requests have to be specified
  # in the format expected by Kubernetes
  - name: AuxResourceRequests
    value: "none"
  # AuxResourceLimits allow you to set limits on side cars. Limits have to be specified
  # in the format expected by Kubernetes
  - name: AuxResourceLimits
    value: "none"
  # ReplicaAntiAffinityTopoKey is used to schedule replica pods
  # of a given volume/application, such that they are:
  # - not co-located on the same node. (kubernetes.io/hostname)
  # - not co-located on the same availability zone.(failure-domain.beta.kubernetes.io/zone)
  # The value for toplogy key can be anything supported by Kubernetes
  # clusters. It is possible that some cluster might support topology schemes
  # like the rack or floor.
  #
  # Examples:
  #   kubernetes.io/hostname (default)
  #   failure-domain.beta.kubernetes.io/zone
  #   failure-domain.beta.kubernetes.io/region
  - name: ReplicaAntiAffinityTopoKey
    value: "kubernetes.io/hostname"
  # TargetNodeSelector allows you to specify the nodes where
  # openebs targets have to be scheduled. To use this feature,
  # the nodes should already be labeled with the key=value. For example:
  # "kubectl label nodes <node-name> nodetype=storage"
  # Note: It is recommended that node selector for replica specify
  # nodes that have disks/ssds attached to them. Example:
  #- name: TargetNodeSelector
  #  value: |-
  #      nodetype: storage
  - name: TargetNodeSelector
    value: "none"
  # ReplicaNodeSelector allows you to specify the nodes where
  # openebs replicas have to be scheduled. To use this feature,
  # the nodes should already be labeled with the key=value. For example:
  # "kubectl label nodes <node-name> nodetype=storage"
  # Note: It is recommended that node selector for replica specify
  # nodes that have disks/ssds attached to them. Example:
  #- name: ReplicaNodeSelector
  #  value: |-
  #      nodetype: storage
  - name: ReplicaNodeSelector
    value: "none"
  # FSType specifies the format type that Kubernetes should use to
  # mount the Persistent Volume. Note that there are no validations
  # done to check the validity of the FsType
  - name: FSType
    value: "ext4"
  # Lun specifies the lun number with which Kubernetes should login
  # to iSCSI Volume (i.e OpenEBS Persistent Volume)
  - name: Lun
    value: "0"
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-create-getstorageclass-default
    - jiva-volume-create-getpvc-default
    - jiva-volume-create-puttargetservice-default
    - jiva-volume-create-getstoragepoolcr-default
    - jiva-volume-create-putreplicadeployment-default
    - jiva-volume-create-puttargetdeployment-default
  output: jiva-volume-create-output-default
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-delete-default
spec:
  defaultConfig:
  - name: OpenEBSNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  - name: ScrubImage
    value: "quay.io/openebs/openebs-tools:3.8"
  # RetainReplicaData specifies whether jiva replica data folder
  # should be cleared or retained.
  - name: RetainReplicaData
    enabled: "false"
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-isvalidversion-default
    - jiva-volume-podsinopenebsns-default
    - jiva-volume-delete-listtargetservice-default
    - jiva-volume-delete-listtargetdeployment-default
    - jiva-volume-delete-listreplicadeployment-default
    - jiva-volume-delete-deletetargetservice-default
    - jiva-volume-delete-deletetargetdeployment-default
    - jiva-volume-delete-listreplicapod-default
    - jiva-volume-delete-deletereplicadeployment-default
    - jiva-volume-delete-putreplicascrub-default
  output: jiva-volume-delete-output-default
  fallback: jiva-volume-delete-default-0.6.0
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: jiva-volume-list-default
spec:
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - jiva-volume-list-listtargetservice-default
    - jiva-volume-list-listtargetpod-default
    - jiva-volume-list-listreplicapod-default
    - jiva-volume-list-listpv-default
  output: jiva-volume-list-output-default
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-isvalidversion-default
spec:
  meta: |
    id: is070jivavolume
    runNamespace: {{.Volume.runNamespace}}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: vsm={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "is070jivavolume.name" .TaskResult | noop -}}
    {{- .TaskResult.is070jivavolume.name | empty | not | versionMismatchErr "is not a jiva volume of 0.7.0 version" | saveIf "is070jivavolume.versionMismatchErr" .TaskResult | noop -}}
---
# Use this generic task in jiva operations like 
# read, delete or snapshot to determine if the 
# jiva pods were created in openebs namespace or
# pvc namespace. This task will check if the service
# is deployed in openebs and saves the result. 
# Each of the further run tasks, will check on this
# saved result to determine if the read operations
# should be performed on openebs or pvc namespace. 
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-podsinopenebsns-default
spec:
  meta: |
    id: jivapodsinopenebsns
    runNamespace: {{ .Config.OpenEBSNamespace.value }}
    apiVersion: v1
    kind: Service
    action: list
    options: |-
      labelSelector: openebs.io/controller-service=jiva-controller-svc,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.namespace}" | trim | saveAs "jivapodsinopenebsns.ns" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-list-listtargetservice-default
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
  name: jiva-volume-list-listtargetpod-default
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
  name: jiva-volume-list-listreplicapod-default
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
  name: jiva-volume-list-listpv-default
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
  name: jiva-volume-list-output-default
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
  name: jiva-volume-read-listtargetservice-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: readlistsvc
    runNamespace: {{ $jivapodsns }}
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
  name: jiva-volume-read-listtargetpod-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: readlistctrl
    runNamespace: {{ $jivapodsns }}
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
  name: jiva-volume-read-listreplicapod-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: readlistrep
    runNamespace: {{ $jivapodsns }}
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
    {{- jsonpath .JsonResult "{.items[*].spec.nodeName}" | trim | saveAs "readlistrep.nodeNames" .TaskResult | noop -}}
    {{- .TaskResult.readlistrep.nodeNames | default "" | splitListLen " " | saveAs "readlistrep.noOfReplicas" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-verifyreplicationfactor-default
spec:
  meta: |
    {{ $isPatchValNotEmpty := ne .Volume.isPatchJivaReplicaNodeAffinity "" }}
    {{ $isPatchValEnabled := eq .Volume.isPatchJivaReplicaNodeAffinity "enabled" }}
    {{ $shouldPatch := and $isPatchValNotEmpty $isPatchValEnabled | toString }}
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: verifyreplicationfactor
    runNamespace: {{ $jivapodsns }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: list
    disable: {{ ne $shouldPatch "true" }}
    options: |-
      labelSelector: openebs.io/replica=jiva-replica,openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "verifyreplicationfactor.items" .TaskResult | noop -}}
    {{- $errMsg := printf "replica deployment not found" -}}
    {{- .TaskResult.verifyreplicationfactor.items | notFoundErr $errMsg | saveIf "verifyreplicationfactor.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.replicas}" | trim | saveAs "verifyreplicationfactor.noOfReplicas" .TaskResult | noop -}}
    {{- $expectedRepCount := .TaskResult.verifyreplicationfactor.noOfReplicas | int -}}
    {{- $msg := printf "expected %v no of replica pod(s), found only %v replica pod(s)" $expectedRepCount .TaskResult.readlistrep.noOfReplicas -}}
    {{- .TaskResult.readlistrep.nodeNames | default "" | splitList " " | isLen $expectedRepCount | not | verifyErr $msg | saveIf "verifyreplicationfactor.verifyErr" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-patchreplicadeployment-default
spec:
  meta: |
    {{ $isPatchValNotEmpty := ne .Volume.isPatchJivaReplicaNodeAffinity "" }}
    {{ $isPatchValEnabled := eq .Volume.isPatchJivaReplicaNodeAffinity "enabled" }}
    {{ $shouldPatch := and $isPatchValNotEmpty $isPatchValEnabled | toString }}
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: readpatchrep
    runNamespace: {{ $jivapodsns }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    objectName: {{ .Volume.owner }}-rep
    disable: {{ ne $shouldPatch "true" }}
    action: patch
  task: |
      {{- $nodeNames := .TaskResult.readlistrep.nodeNames -}}
      type: strategic
      pspec: |-
        spec:
          template:
            spec:
              affinity:
                nodeAffinity:
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
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-read-output-default
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
#Creating a Target Service is the first operation in 
#creating K8s objects for the given PVC. Determine
#the namespace and save it for further create options.
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-puttargetservice-default
spec:
  meta: |
    {{- $deployInOpenEBSNamespace := .Config.DeployInOpenEBSNamespace.enabled | default "false" | lower -}}
    id: createputsvc
    {{- if eq $deployInOpenEBSNamespace "false" }}
    runNamespace: {{ .Volume.runNamespace | trim | saveAs "createputsvc.jivapodsns" .TaskResult }}
    {{ else }}
    runNamespace: {{ .Config.OpenEBSNamespace.value | trim | saveAs "createputsvc.jivapodsns" .TaskResult }}
    {{ end }}
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
      annotations:
        openebs.io/storage-class-ref: |
          name: {{ .Volume.storageclass }}
          resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
      labels:
        openebs.io/storage-engine-type: jiva
        openebs.io/cas-type: jiva
        openebs.io/controller-service: jiva-controller-svc
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
        pvc: {{ .Volume.pvc }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
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
      - name: exporter
        port: 9500
        protocol: TCP
        targetPort: 9500
      selector:
        openebs.io/controller: jiva-controller
        openebs.io/persistent-volume: {{ .Volume.owner }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-getstoragepoolcr-default
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
  name: jiva-volume-create-getstorageclass-default
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
    {{- $stsTargetAffinity := jsonpath .JsonResult "{.metadata.labels.openebs\\.io/sts-target-affinity}" | trim | default "none" -}}
    {{- $stsTargetAffinity | saveAs "stsTargetAffinity" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-getpvc-default
spec:
  meta: |
    id: creategetpvc
    apiVersion: v1
    runNamespace: {{ .Volume.runNamespace }}
    kind: PersistentVolumeClaim
    objectName: {{ .Volume.pvc }}
    action: get
  post: |
    {{- $replicaAntiAffinity := jsonpath .JsonResult "{.metadata.labels.openebs\\.io/replica-anti-affinity}" | trim | default "none" -}}
    {{- $replicaAntiAffinity | saveAs "creategetpvc.replicaAntiAffinity" .TaskResult | noop -}}
    {{- $targetAffinity := jsonpath .JsonResult "{.metadata.labels.openebs\\.io/target-affinity}" | trim | default "none" -}}
    {{- $targetAffinity | saveAs "creategetpvc.targetAffinity" .TaskResult | noop -}}
    {{- $stsTargetAffinity := jsonpath .JsonResult "{.metadata.labels.openebs\\.io/sts-target-affinity}" | trim | default "none" -}}
    {{- if ne $stsTargetAffinity "none" -}}
    {{- $stsTargetAffinity | saveAs "stsTargetAffinity" .TaskResult | noop -}}
    {{- end -}}
    {{- if ne .TaskResult.stsTargetAffinity "none" -}}
    {{- printf "%s-%s" .TaskResult.stsTargetAffinity ((splitList "-" .Volume.pvc) | last) | default "none" | saveAs "sts.applicationName" .TaskResult -}}
    {{- end -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-listreplicapod-default
spec:
  meta: |
    id: createlistrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: v1
    kind: Pod
    action: list
    options: |-
      labelSelector: openebs.io/replica=jiva-replica,openebs.io/persistent-volume={{ .Volume.owner }}
    retry: "24,5s"
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
  name: jiva-volume-create-patchreplicadeployment-default
spec:
  meta: |
    id: createpatchrep
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    objectName: {{ .Volume.owner }}-rep
    action: patch
  task: |
      {{- $isNodeAffinityRSIE := .Config.NodeAffinityRequiredSchedIgnoredExec.value | default "none" -}}
      {{- $nodeAffinityRSIEVal := fromYaml .Config.NodeAffinityRequiredSchedIgnoredExec.value -}}
      {{- $nodeNames := .TaskResult.createlistrep.nodeNames -}}
      type: strategic
      pspec: |-
        spec:
          template:
            spec:
              affinity:
                nodeAffinity:
                  {{- if ne $isNodeAffinityRSIE "none" }}
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
  name: jiva-volume-create-puttargetdeployment-default
spec:
  meta: |
    id: createputctrl
    runNamespace: {{ .TaskResult.createputsvc.jivapodsns }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: put
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "createputctrl.objectName" .TaskResult | noop -}}
  task: |
    {{- $isMonitor := .Config.VolumeMonitor.enabled | default "true" | lower -}}
    {{- $setResourceRequests := .Config.TargetResourceRequests.value | default "none" -}}
    {{- $resourceRequestsVal := fromYaml .Config.TargetResourceRequests.value -}}
    {{- $setResourceLimits := .Config.TargetResourceLimits.value | default "none" -}}
    {{- $resourceLimitsVal := fromYaml .Config.TargetResourceLimits.value -}}
    {{- $setAuxResourceRequests := .Config.AuxResourceRequests.value | default "none" -}}
    {{- $auxResourceRequestsVal := fromYaml .Config.AuxResourceRequests.value -}}
    {{- $setAuxResourceLimits := .Config.AuxResourceLimits.value | default "none" -}}
    {{- $auxResourceLimitsVal := fromYaml .Config.AuxResourceLimits.value -}}
    {{- $hasNodeSelector := .Config.TargetNodeSelector.value | default "none" -}}
    {{- $nodeSelectorVal := fromYaml .Config.TargetNodeSelector.value -}}
    {{- $targetAffinityVal := .TaskResult.creategetpvc.targetAffinity -}}
    {{- $hasTargetToleration := .Config.TargetTolerations.value | default "none" -}}
    {{- $targetTolerationVal := fromYaml .Config.TargetTolerations.value -}}
    apiVersion: extensions/v1beta1
    Kind: Deployment
    metadata:
      labels:
        {{- if eq $isMonitor "true" }}
        monitoring: "volume_exporter_prometheus"
        {{- end}}
        openebs.io/storage-engine-type: jiva
        openebs.io/cas-type: jiva
        openebs.io/controller: jiva-controller
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
      annotations:
        openebs.io/storage-class-ref: |
          name: {{ .Volume.storageclass }}
          resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
        {{- if eq $isMonitor "true" }}
        openebs.io/volume-monitor: "true"
        {{- end}}
        openebs.io/volume-type: jiva
        openebs.io/fs-type: {{ .Config.FSType.value }}
        openebs.io/lun: {{ .Config.Lun.value }}
      name: {{ .Volume.owner }}-ctrl
    spec:
      replicas: 1
      strategy:
        type: Recreate
      selector:
        matchLabels:
          openebs.io/controller: jiva-controller
          openebs.io/persistent-volume: {{ .Volume.owner }}
      template:
        metadata:
          labels:
            {{- if eq $isMonitor "true" }}
            monitoring: volume_exporter_prometheus
            {{- end}}
            openebs.io/controller: jiva-controller
            openebs.io/persistent-volume: {{ .Volume.owner }}
            openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
            openebs.io/version: {{ .CAST.version }}
          annotations:
            openebs.io/storage-class-ref: |
                name: {{ .Volume.storageclass }}
                resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
            openebs.io/fs-type: {{ .Config.FSType.value }}
            openebs.io/lun: {{ .Config.Lun.value }}
            {{- if eq $isMonitor "true" }}
            prometheus.io/path: /metrics
            prometheus.io/port: "9500"
            prometheus.io/scrape: "true"
            {{- end}}
        spec:
          {{- if ne $hasNodeSelector "none" }}
          nodeSelector:
            {{- range $sK, $sV := $nodeSelectorVal }}
              {{ $sK }}: {{ $sV }}
            {{- end }}
          {{- end}}
          {{- if ne (.TaskResult.sts.applicationName | default "") "" }}
          affinity:
            podAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
              - labelSelector:
                  matchExpressions:
                  - key: statefulset.kubernetes.io/pod-name
                    operator: In
                    values:
                    - {{ .TaskResult.sts.applicationName }}
                topologyKey: kubernetes.io/hostname
          {{- else if ne $targetAffinityVal "none" }}
          affinity:
            podAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
              - labelSelector:
                  matchExpressions:
                  - key: openebs.io/target-affinity
                    operator: In
                    values:
                    - {{ $targetAffinityVal }}
                topologyKey: kubernetes.io/hostname
          {{- end }}
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
            resources:
              {{- if ne $setResourceLimits "none" }}
              limits:
              {{- range $rKey, $rLimit := $resourceLimitsVal }}
                {{ $rKey }}: {{ $rLimit }}
              {{- end }}
              {{- end }}
              {{- if ne $setResourceRequests "none" }}
              requests:
              {{- range $rKey, $rReq := $resourceRequestsVal }}
                {{ $rKey }}: {{ $rReq }}
              {{- end }}
              {{- end }}
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
            resources:
              {{- if ne $setAuxResourceRequests "none" }}
              requests:
              {{- range $rKey, $rLimit := $auxResourceRequestsVal }}
                {{ $rKey }}: {{ $rLimit }}
              {{- end }}
              {{- end }}
              {{- if ne $setAuxResourceLimits "none" }}
              limits:
              {{- range $rKey, $rLimit := $auxResourceLimitsVal }}
                {{ $rKey }}: {{ $rLimit }}
              {{- end }}
              {{- end }}
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
          {{- if ne $hasTargetToleration "none" }}
          {{- range $k, $v := $targetTolerationVal }}
          -
          {{- range $kk, $vv := $v }}
            {{ $kk }}: {{ $vv }}
          {{- end }}
          {{- end }}
          {{- end }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-create-putreplicadeployment-default
spec:
  meta: |
    id: createputrep
    runNamespace: {{ .TaskResult.createputsvc.jivapodsns }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: put
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "createputrep.objectName" .TaskResult | noop -}}
  task: |
    {{- $isEvictionTolerations := .Config.EvictionTolerations.value | default "none" -}}
    {{- $evictionTolerationsVal := fromYaml .Config.EvictionTolerations.value -}}
    {{- $isCloneEnable := .Volume.isCloneEnable | default "false" -}}
    {{- $setResourceRequests := .Config.ReplicaResourceRequests.value | default "none" -}}
    {{- $resourceRequestsVal := fromYaml .Config.ReplicaResourceRequests.value -}}
    {{- $setResourceLimits := .Config.ReplicaResourceLimits.value | default "none" -}}
    {{- $resourceLimitsVal := fromYaml .Config.ReplicaResourceLimits.value -}}
    {{- $replicaAntiAffinityVal := .TaskResult.creategetpvc.replicaAntiAffinity -}}
    {{- $hasNodeSelector := .Config.ReplicaNodeSelector.value | default "none" -}}
    {{- $nodeSelectorVal := fromYaml .Config.ReplicaNodeSelector.value -}}
    {{- $hasReplicaToleration := .Config.ReplicaTolerations.value | default "none" -}}
    {{- $replicaTolerationVal := fromYaml .Config.ReplicaTolerations.value -}}
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      labels:
        openebs.io/storage-engine-type: jiva
        openebs.io/cas-type: jiva
        openebs.io/replica: jiva-replica
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
      annotations:
        openebs.io/storage-class-ref: |
          name: {{ .Volume.storageclass }}
          resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
        openebs.io/capacity: {{ .Volume.capacity }}
        openebs.io/storage-pool: {{ .Config.StoragePool.value }}
      name: {{ .Volume.owner }}-rep
    spec:
      replicas: {{ .Config.ReplicaCount.value }}
      strategy:
        type: Recreate
      selector:
        matchLabels:
          openebs.io/replica: jiva-replica
          openebs.io/persistent-volume: {{ .Volume.owner }}
      template:
        metadata:
          labels:
            openebs.io/replica: jiva-replica
            openebs.io/persistent-volume: {{ .Volume.owner }}
            openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
            openebs.io/version: {{ .CAST.version }}
            {{- if ne $replicaAntiAffinityVal "none" }}
            openebs.io/replica-anti-affinity: {{ $replicaAntiAffinityVal }}
            {{- end }}
          annotations:
            openebs.io/storage-class-ref: |
              name: {{ .Volume.storageclass }}
              resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
            openebs.io/capacity: {{ .Volume.capacity }}
            openebs.io/storage-pool: {{ .Config.StoragePool.value }}
        spec:
          {{- if ne $hasNodeSelector "none" }}
          nodeSelector:
            {{- range $sK, $sV := $nodeSelectorVal }}
              {{ $sK }}: {{ $sV }}
            {{- end }}
          {{- end}}
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
              - labelSelector:
                  matchLabels:
                    openebs.io/replica: jiva-replica
                    {{/* If PVC object has a replica anti-affinity value. Use it.
                         This is usually the case for STS that creates PVCs from a
                         PVC Template. So, a STS can have multiple PVs with their
                         unique id. To schedule/spread out replicas belonging to
                         different PV, a unique label associated with the STS should
                         be passed to all the PVCs tied to the STS. */}}
                    {{- if ne $replicaAntiAffinityVal "none" }}
                    openebs.io/replica-anti-affinity: {{ $replicaAntiAffinityVal }}
                    {{- else }}
                    openebs.io/persistent-volume: {{ .Volume.owner }}
                    {{- end }}
                topologyKey: {{ .Config.ReplicaAntiAffinityTopoKey.value }}
          containers:
          - args:
            - replica
            - --frontendIP
            - {{ .TaskResult.createputsvc.clusterIP }}
            {{- if ne $isCloneEnable "false" }}
            - --cloneIP
            - {{ .Volume.sourceVolumeTargetIP }}
            - --type
            - "clone"
            - --snapName
            - {{ .Volume.snapshotName }}
            {{- end }}
            - --size
            - {{ .Volume.capacity }}
            - /openebs
            securityContext:
                privileged: true
            command:
            - launch
            image: {{ .Config.ReplicaImage.value }}
            name: {{ .Volume.owner }}-rep-con
            resources:
              {{- if ne $setResourceLimits "none" }}
              limits:
              {{- range $rKey, $rLimit := $resourceLimitsVal }}
                {{ $rKey }}: {{ $rLimit }}
              {{- end }}
              {{- end }}
              {{- if ne $setResourceRequests "none" }}
              requests:
              {{- range $rKey, $rReq := $resourceRequestsVal }}
                {{ $rKey }}: {{ $rReq }}
              {{- end }}
              {{- end }}
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
          {{- if ne $isEvictionTolerations "none" }}
          {{- range $k, $v := $evictionTolerationsVal }}
          -
          {{- range $kk, $vv := $v }}
            {{ $kk }}: {{ $vv }}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- if ne $hasReplicaToleration "none" }}
          {{- range $k, $v := $replicaTolerationVal }}
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
  name: jiva-volume-create-output-default
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
      targetIP: {{ .TaskResult.readlistsvc.clusterIP }}
      targetPort: 3260
      casType: jiva
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listtargetservice-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: deletelistsvc
    runNamespace: {{ $jivapodsns }}
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
  name: jiva-volume-delete-listtargetdeployment-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: deletelistctrl
    runNamespace: {{ $jivapodsns }}
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
  name: jiva-volume-delete-listreplicadeployment-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: deletelistrep
    runNamespace: {{ $jivapodsns }}
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
  name: jiva-volume-delete-deletetargetservice-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: deletedeletesvc
    runNamespace: {{ $jivapodsns }}
    apiVersion: v1
    kind: Service
    action: delete
    objectName: {{ .TaskResult.deletelistsvc.names }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-deletetargetdeployment-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: deletedeletectrl
    runNamespace: {{ $jivapodsns }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: delete
    objectName: {{ .TaskResult.deletelistctrl.names }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-listreplicapod-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: deletelistreppods
    runNamespace: {{ $jivapodsns }}
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
  name: jiva-volume-delete-deletereplicadeployment-default
spec:
  meta: |
    {{- $jivapodsns := .TaskResult.jivapodsinopenebsns.ns | default .Volume.runNamespace -}}
    id: deletedeleterep
    runNamespace: {{ $jivapodsns }}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: delete
    objectName: {{ .TaskResult.deletelistrep.names }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: jiva-volume-delete-putreplicascrub-default
spec:
  meta: |
    apiVersion: batch/v1
    runNamespace: {{ .Config.OpenEBSNamespace.value }}
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
      {{- if kubeVersionGte .CAST.kubeVersion "v1.12.0" }}
      ttlSecondsAfterFinished: 0
      {{- end }}
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
  name: jiva-volume-delete-output-default
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

// JivaVolumeArtifacts returns the jiva volume related artifacts corresponding
// to latest version
func JivaVolumeArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(jivaVolumes{})...)
	return
}

type jivaVolumes struct{}

// FetchYamls returns all the yamls related to jiva volume in a string
// format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (j jivaVolumes) FetchYamls() string {
	return jivaVolumeYamls
}
