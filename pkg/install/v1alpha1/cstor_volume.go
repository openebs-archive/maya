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

const cstorVolumeYamls = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-volume-create-default
spec:
  defaultConfig:
  - name: VolumeControllerImage
    value: {{env "OPENEBS_IO_CSTOR_VOLUME_MGMT_IMAGE" | default "openebs/cstor-volume-mgmt:latest"}}
  - name: VolumeTargetImage
    value: {{env "OPENEBS_IO_CSTOR_TARGET_IMAGE" | default "openebs/cstor-istgt:latest"}}
  - name: VolumeMonitorImage
    value: {{env "OPENEBS_IO_VOLUME_MONITOR_IMAGE" | default "openebs/m-exporter:latest"}}
  - name: ReplicaCount
    value: "3"
  # Target Dir is a hostPath directory for target pod
  - name: TargetDir
    value: {{env "OPENEBS_IO_CSTOR_TARGET_DIR" | default "/var/openebs"}}
  # TargetResourceRequests allow you to specify resource requests that need to be available
  # before scheduling the containers. If not specified, the default is to use the limits
  # from TargetResourceLimits or the default requests set in the cluster.
  - name: TargetResourceRequests
    value: "none"
  # TargetResourceLimits allow you to set the limits on memory and cpu for target pods
  # The resource and limit value should be in the same format as expected by
  # Kubernetes. Example:
  #- name: TargetResourceLimits
  #  value: |-
  #      memory: 1Gi
  #      cpu: 200m
  # By default, the resource limits are disabled.
  - name: TargetResourceLimits
    value: "none"
  # AuxResourceRequests allow you to set requests on side cars. Requests have to be specified
  # in the format expected by Kubernetes
  - name: AuxResourceRequests
    value: "none"
  # AuxResourceLimits allow you to set limits on side cars. Limits have to be specified
  # in the format expected by Kubernetes
  - name: AuxResourceLimits
    value: "none"
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  # ServiceAccountName is the account name assigned to volume management pod
  # with permissions to view, create, edit, delete required custom resources
  - name: ServiceAccountName
    value: {{env "OPENEBS_SERVICE_ACCOUNT"}}
  # FSType specifies the format type that Kubernetes should use to
  # mount the Persistent Volume. Note that there are no validations
  # done to check the validity of the FsType
  - name: FSType
    value: "ext4"
  # Lun specifies the lun number with which Kubernetes should login
  # to iSCSI Volume (i.e OpenEBS Persistent Volume)
  - name: Lun
    value: "0"
  # ResyncInterval specifies duration after which a controller should
  # resync the resource status
  - name: ResyncInterval
    value: "30"
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
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - cstor-volume-create-getstorageclass-default
    - cstor-volume-create-getpvc-default
    - cstor-volume-create-listclonecstorvolumereplicacr-default
    - cstor-volume-create-listcstorpoolcr-default
    - cstor-volume-create-puttargetservice-default
    - cstor-volume-create-putcstorvolumecr-default
    - cstor-volume-create-puttargetdeployment-default
    - cstor-volume-create-putcstorvolumereplicacr-default
  output: cstor-volume-create-output-default
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-volume-delete-default
spec:
  defaultConfig:
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - cstor-volume-delete-listcstorvolumecr-default
    - cstor-volume-delete-listtargetservice-default
    - cstor-volume-delete-listtargetdeployment-default
    - cstor-volume-delete-listcstorvolumereplicacr-default
    - cstor-volume-delete-deletetargetservice-default
    - cstor-volume-delete-deletetargetdeployment-default
    - cstor-volume-delete-deletecstorvolumereplicacr-default
    - cstor-volume-delete-deletecstorvolumecr-default
  output: cstor-volume-delete-output-default
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-volume-read-default
spec:
  defaultConfig:
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - cstor-volume-read-listtargetservice-default
    - cstor-volume-read-listcstorvolumecr-default
    - cstor-volume-read-listcstorvolumereplicacr-default
    - cstor-volume-read-listtargetpod-default
  output: cstor-volume-read-output-default
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-volume-list-default
spec:
  defaultConfig:
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    - cstor-volume-list-listtargetservice-default
    - cstor-volume-list-listtargetpod-default
    - cstor-volume-list-listcstorvolumereplicacr-default
    - cstor-volume-list-listpv-default
  output: cstor-volume-list-output-default
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-getpvc-default
spec:
  meta: |
    id: creategetpvc
    apiVersion: v1
    runNamespace: {{ .Volume.runNamespace }}
    kind: PersistentVolumeClaim
    objectName: {{ .Volume.pvc }}
    action: get
  post: |
    {{- $hostName := jsonpath .JsonResult "{.metadata.annotations.volume\\.kubernetes\\.io/selected-node}" | trim | default "" -}}
    {{- $hostName | saveAs "creategetpvc.hostName" .TaskResult | noop -}}
    {{- $replicaAntiAffinity := jsonpath .JsonResult "{.metadata.labels.openebs\\.io/replica-anti-affinity}" | trim | default "" -}}
    {{- $replicaAntiAffinity | saveAs "creategetpvc.replicaAntiAffinity" .TaskResult | noop -}}
    {{- $preferredReplicaAntiAffinity := jsonpath .JsonResult "{.metadata.labels.openebs\\.io/preferred-replica-anti-affinity}" | trim | default "" -}}
    {{- $preferredReplicaAntiAffinity | saveAs "creategetpvc.preferredReplicaAntiAffinity" .TaskResult | noop -}}
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
# This RunTask is meant to be run only during clone create requests.
# However, clone & volume creation follow the same CASTemplate specifications.
# As of today, RunTask can not be run based on conditions. Hence, it contains
# a logic which will list empty pools
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-listclonecstorvolumereplicacr-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .Config.RunNamespace.value }}
    id: cvolcreatelistclonecvr
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolumeReplica
    action: list
    options: |-
    {{- if ne $isClone "false" }}
      labelSelector: openebs.io/persistent-volume={{ .Volume.sourceVolume }}
    {{- else }}
      labelSelector: openebs.io/ignore=false
    {{- end }}
  post: |
    {{- $poolsList := jsonpath .JsonResult "{range .items[*]}pkey=pools,{@.metadata.labels.cstorpool\\.openebs\\.io/uid}={@.metadata.labels.cstorpool\\.openebs\\.io/name};{end}" | trim | default "" | splitListTrim ";" -}}
    {{- $poolsList | saveAs "pl" .ListItems -}}
    {{- $poolsList | keyMap "cvolPoolList" .ListItems | noop -}}
    {{- $poolsNodeList := jsonpath .JsonResult "{range .items[*]}pkey=pools,{@.metadata.labels.cstorpool\\.openebs\\.io/uid}={@.metadata.annotations.cstorpool\\.openebs\\.io/hostname};{end}" | trim | default "" | splitList ";" -}}
    {{- $poolsNodeList | keyMap "cvolPoolNodeList" .ListItems | noop -}}
---
# runTask to list cstor pools
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-listcstorpoolcr-default
spec:
  meta: |
    id: cvolcreatelistpool
    runNamespace: {{.Config.RunNamespace.value}}
    apiVersion: openebs.io/v1alpha1
    kind: CStorPool
    action: list
    options: |-
      labelSelector: openebs.io/storage-pool-claim={{ .Config.StoragePoolClaim.value }}
  post: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    {{/*
    If clone is not enabled then override changes of previous runtask
    */}}
    {{- if eq $isClone "false" }}
    {{- $replicaCount := int64 .Config.ReplicaCount.value | saveAs "rc" .ListItems -}}
    {{- $poolsList := jsonpath .JsonResult "{range .items[*]}pkey=pools,{@.metadata.uid}={@.metadata.name};{end}" | trim | default "" | splitListTrim ";" -}}
    {{- $poolsList | saveAs "pl" .ListItems -}}
    {{- len $poolsList | gt $replicaCount | verifyErr "not enough pools available to create replicas" | saveAs "cvolcreatelistpool.verifyErr" .TaskResult | noop -}}
    {{- $poolsList | keyMap "cvolPoolList" .ListItems | noop -}}
    {{- $poolsNodeList := jsonpath .JsonResult "{range .items[*]}pkey=pools,{@.metadata.uid}={@.metadata.labels.kubernetes\\.io/hostname};{end}" | trim | default "" | splitList ";" -}}
    {{- $poolsNodeList | keyMap "cvolPoolNodeList" .ListItems | noop -}}
    {{- end }}
---
#runTask to get storageclass info
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-getstorageclass-default
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
# runTask to create cStor target service
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-puttargetservice-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    {{- $runNamespace := .Config.RunNamespace.value -}}
    {{- $pvcServiceAccount := .Config.PVCServiceAccountName.value | default "" -}}
    {{- if ne $pvcServiceAccount "" }}
    runNamespace: {{ .Volume.runNamespace | saveAs "cvolcreateputsvc.derivedNS" .TaskResult }}
    {{ else }}
    runNamespace: {{ $runNamespace | saveAs "cvolcreateputsvc.derivedNS" .TaskResult }}
    {{- end }}
    apiVersion: v1
    kind: Service
    action: put
    id: cvolcreateputsvc
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "cvolcreateputsvc.objectName" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.spec.clusterIP}" | trim | saveAs "cvolcreateputsvc.clusterIP" .TaskResult | noop -}}
  task: |
    apiVersion: v1
    kind: Service
    metadata:
      annotations:
        openebs.io/storage-class-ref: |
          name: {{ .Volume.storageclass }}
          resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
      labels:
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
        openebs.io/target-service: cstor-target-svc
        openebs.io/storage-engine-type: cstor
        openebs.io/cas-type: cstor
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
      name: {{ .Volume.owner }}
    spec:
      ports:
      - name: cstor-iscsi
        port: 3260
        protocol: TCP
        targetPort: 3260
      - name: cstor-grpc
        port: 7777
        protocol: TCP
        targetPort: 7777
      - name: mgmt
        port: 6060
        targetPort: 6060
        protocol: TCP
      - name: exporter
        port: 9500
        targetPort: 9500
        protocol: TCP
      selector:
        app: cstor-volume-manager
        openebs.io/target: cstor-target
        openebs.io/persistent-volume: {{ .Volume.owner }}
---
# runTask to create cStorVolume
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-putcstorvolumecr-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.cvolcreateputsvc.derivedNS }}
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolume
    id: cvolcreateputvolume
    action: put
  post: |
    {{- jsonpath .JsonResult "{.metadata.uid}" | trim | saveAs "cvolcreateputvolume.cstorid" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "cvolcreateputvolume.objectName" .TaskResult | noop -}}
  task: |
    {{- $replicaCount := .Config.ReplicaCount.value | int64 -}}
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolume
    metadata:
      name: {{ .Volume.owner }}
      annotations:
        openebs.io/fs-type: {{ .Config.FSType.value }}
        openebs.io/lun: {{ .Config.Lun.value }}
        openebs.io/storage-class-ref: |
          name: {{ .Volume.storageclass }}
          resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
        {{- if ne $isClone "false" }}
        openebs.io/snapshot: {{ .Volume.snapshotName }}
        {{- end }}

      labels:
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
        {{- if ne $isClone "false" }}
        openebs.io/source-volume: {{ .Volume.sourceVolume }}
        {{- end }}

    spec:
      targetIP: {{ .TaskResult.cvolcreateputsvc.clusterIP }}
      capacity: {{ .Volume.capacity }}
      nodeBase: iqn.2016-09.com.openebs.cstor
      iqn: iqn.2016-09.com.openebs.cstor:{{ .Volume.owner }}
      targetPortal: {{ .TaskResult.cvolcreateputsvc.clusterIP }}:3260
      targetPort: 3260
      status: "Init"
      replicationFactor: {{ $replicaCount }}
      consistencyFactor: {{ div $replicaCount 2 | floor | add1 }}
---
# runTask to create cStor target deployment
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-puttargetdeployment-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.cvolcreateputsvc.derivedNS }}
    apiVersion: apps/v1beta1
    kind: Deployment
    action: put
    id: cvolcreateputctrl
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | saveAs "cvolcreateputctrl.objectName" .TaskResult | noop -}}
  task: |
    {{- $isMonitor := .Config.VolumeMonitorImage.enabled | default "true" | lower -}}
    {{- $setResourceRequests := .Config.TargetResourceRequests.value | default "none" -}}
    {{- $resourceRequestsVal := fromYaml .Config.TargetResourceRequests.value -}}
    {{- $setResourceLimits := .Config.TargetResourceLimits.value | default "none" -}}
    {{- $resourceLimitsVal := fromYaml .Config.TargetResourceLimits.value -}}
    {{- $setAuxResourceRequests := .Config.AuxResourceRequests.value | default "none" -}}
    {{- $auxResourceRequestsVal := fromYaml .Config.AuxResourceRequests.value -}}
    {{- $setAuxResourceLimits := .Config.AuxResourceLimits.value | default "none" -}}
    {{- $auxResourceLimitsVal := fromYaml .Config.AuxResourceLimits.value -}}
    {{- $targetAffinityVal := .TaskResult.creategetpvc.targetAffinity -}}
    {{- $hasNodeSelector := .Config.TargetNodeSelector.value | default "none" -}}
    {{- $nodeSelectorVal := fromYaml .Config.TargetNodeSelector.value -}}
    {{- $hasTargetToleration := .Config.TargetTolerations.value | default "none" -}}
    {{- $targetTolerationVal := fromYaml .Config.TargetTolerations.value -}}
    {{- $isQueueDepth := .Config.QueueDepth.value | default "" -}}
    {{- $isLuworkers := .Config.Luworkers.value | default "" -}}
    apiVersion: apps/v1beta1
    Kind: Deployment
    metadata:
      name: {{ .Volume.owner }}-target
      labels:
        app: cstor-volume-manager
        openebs.io/storage-engine-type: cstor
        openebs.io/cas-type: cstor
        openebs.io/target: cstor-target
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
        openebs.io/storage-pool-claim: {{ .Config.StoragePoolClaim.value }}
      annotations:
        {{- if eq $isMonitor "true" }}
        openebs.io/volume-monitor: "true"
        {{- end}}
        openebs.io/volume-type: cstor
        openebs.io/storage-class-ref: |
          name: {{ .Volume.storageclass }}
          resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
    spec:
      replicas: 1
      strategy:
        type: Recreate
      selector:
        matchLabels:
          app: cstor-volume-manager
          openebs.io/target: cstor-target
          openebs.io/persistent-volume: {{ .Volume.owner }}
      template:
        metadata:
          labels:
            {{- if eq $isMonitor "true" }}
            monitoring: volume_exporter_prometheus
            {{- end}}
            app: cstor-volume-manager
            openebs.io/target: cstor-target
            openebs.io/persistent-volume: {{ .Volume.owner }}
            openebs.io/storage-class: {{ .Volume.storageclass }}
            openebs.io/persistent-volume-claim: {{ .Volume.pvc }}
            openebs.io/version: {{ .CAST.version }}
          annotations:
            openebs.io/storage-class-ref: |
              name: {{ .Volume.storageclass }}
              resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
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
          serviceAccountName: {{ .Config.PVCServiceAccountName.value | default .Config.ServiceAccountName.value }}
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
                namespaces:
                - {{ .Volume.runNamespace }}
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
                namespaces: [{{.Volume.runNamespace}}]
          {{- end }}
          tolerations:
          - effect: NoExecute
            key: node.alpha.kubernetes.io/notReady
            operator: Exists
            tolerationSeconds: 30
          - effect: NoExecute
            key: node.alpha.kubernetes.io/unreachable
            operator: Exists
            tolerationSeconds: 30
          - effect: NoExecute
            key: node.kubernetes.io/not-ready
            operator: Exists
            tolerationSeconds: 30
          - effect: NoExecute
            key: node.kubernetes.io/unreachable
            operator: Exists
            tolerationSeconds: 30
          {{- if ne $hasTargetToleration "none" }}
          {{- range $k, $v := $targetTolerationVal }}
          -
          {{- range $kk, $vv := $v }}
            {{ $kk }}: {{ $vv }}
          {{- end }}
          {{- end }}
          {{- end }}
          containers:
          - image: {{ .Config.VolumeTargetImage.value }}
            name: cstor-istgt
            imagePullPolicy: IfNotPresent
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
            - containerPort: 3260
              protocol: TCP
            env:
            {{- if ne $isQueueDepth "" }}
            - name: QueueDepth
              value: {{ .Config.QueueDepth.value }}
            {{- end }}
            {{- if ne $isLuworkers "" }}
            - name: Luworkers
              value: {{ .Config.Luworkers.value }}
            {{- end }}
            securityContext:
              privileged: true
            volumeMounts:
            - name: sockfile
              mountPath: /var/run
            - name: conf
              mountPath: /usr/local/etc/istgt
            - name: tmp
              mountPath: /tmp
              mountPropagation: Bidirectional
          {{- if eq $isMonitor "true" }}
          - image: {{ .Config.VolumeMonitorImage.value }}
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
            args:
            - "-e=cstor"
            command: ["maya-exporter"]
            ports:
            - containerPort: 9500
              protocol: TCP
            volumeMounts:
            - name: sockfile
              mountPath: /var/run
            - name: conf
              mountPath: /usr/local/etc/istgt
          {{- end }}
          - name: cstor-volume-mgmt
            image: {{ .Config.VolumeControllerImage.value }}
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
            imagePullPolicy: IfNotPresent
            ports:
            - containerPort: 80
            env:
            - name: OPENEBS_IO_CSTOR_VOLUME_ID
              value: {{ .TaskResult.cvolcreateputvolume.cstorid }}
            - name: RESYNC_INTERVAL
              value: {{ .Config.ResyncInterval.value }}
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            securityContext:
              privileged: true
            volumeMounts:
            - name: sockfile
              mountPath: /var/run
            - name: conf
              mountPath: /usr/local/etc/istgt
            - name: tmp
              mountPath: /tmp
              mountPropagation: Bidirectional
          volumes:
          - name: sockfile
            emptyDir: {}
          - name: conf
            emptyDir: {}
          - name: tmp
            hostPath:
              path: {{ .Config.TargetDir.value }}/shared-{{ .Volume.owner }}-target
              type: DirectoryOrCreate
---
# runTask to create cStorVolumeReplica/(s)
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-create-putcstorvolumereplicacr-default
spec:
  meta: |
    apiVersion: openebs.io/v1alpha1
    runNamespace: {{.Config.RunNamespace.value}}
    kind: CStorVolumeReplica
    action: put
    id: cstorvolumecreatereplica
    {{/*
    Fetch all the cStorPool uids into a list.
    Calculate the replica count
    Add as many poolUid to resources as there is replica count
    */}}
    {{- $hostName := .TaskResult.creategetpvc.hostName -}}
    {{- $replicaAntiAffinity := .TaskResult.creategetpvc.replicaAntiAffinity }}
    {{- $preferredReplicaAntiAffinity := .TaskResult.creategetpvc.preferredReplicaAntiAffinity }}
    {{- $antiAffinityLabelSelector := printf "openebs.io/replica-anti-affinity=%s" $replicaAntiAffinity | IfNotNil $replicaAntiAffinity }}
    {{- $preferredAntiAffinityLabelSelector := printf "openebs.io/preferred-replica-anti-affinity=%s" $preferredReplicaAntiAffinity | IfNotNil $preferredReplicaAntiAffinity }}
    {{- $preferedScheduleOnHostAnnotationSelector := printf "volume.kubernetes.io/selected-node=%s" $hostName | IfNotNil $hostName }}
    {{- $selectionPolicies := cspGetPolicies $antiAffinityLabelSelector $preferredAntiAffinityLabelSelector $preferedScheduleOnHostAnnotationSelector }}
    {{- $pools :=  createCSPListFromUIDNodeMap (getMapofString .ListItems.cvolPoolNodeList "pools") }}
    {{- $poolUids := cspFilterPoolIDs $pools $selectionPolicies | randomize }}
    {{- $replicaCount := .Config.ReplicaCount.value | int64 -}}
    {{- if lt (len $poolUids) $replicaCount -}}
    {{- printf "Not enough pools to provision replica: expected replica count %d actual count %d" $replicaCount (len $poolUids) | fail -}}
    {{- end -}}
    repeatWith:
      resources:
      {{- range $k, $v := $poolUids }}
      {{- if lt $k $replicaCount }}
      - {{ $v | quote }}
      {{- end }}
      {{- end }}
  task: |
    {{- $replicaAntiAffinity := .TaskResult.creategetpvc.replicaAntiAffinity -}}
    {{- $preferredReplicaAntiAffinity := .TaskResult.creategetpvc.preferredReplicaAntiAffinity }}
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    {{- $zvolWorkers := .Config.ZvolWorkers.value | default "" -}}
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
      finalizers: ["cstorvolumereplica.openebs.io/finalizer"]
      labels:
        cstorpool.openebs.io/name: {{ pluck .ListItems.currentRepeatResource .ListItems.cvolPoolList.pools | first }}
        cstorpool.openebs.io/uid: {{ .ListItems.currentRepeatResource }}
        cstorvolume.openebs.io/name: {{ .Volume.owner }}
        openebs.io/persistent-volume: {{ .Volume.owner }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
        {{- if ne $isClone "false" }}
        openebs.io/cloned: true
        {{- end }}
        {{- if ne $replicaAntiAffinity  "" }}
        openebs.io/replica-anti-affinity: {{ .TaskResult.creategetpvc.replicaAntiAffinity }}
        {{- end }}
        {{- if ne $preferredReplicaAntiAffinity  "" }}
        openebs.io/preferred-replica-anti-affinity: {{ .TaskResult.creategetpvc.preferredReplicaAntiAffinity }}
        {{- end }}
      annotations:
        {{- if ne $isClone "false" }}
        openebs.io/snapshot: {{ .Volume.snapshotName }}
        openebs.io/source-volume: {{ .Volume.sourceVolume }}
        {{- end }}
        cstorpool.openebs.io/hostname: {{ pluck .ListItems.currentRepeatResource .ListItems.cvolPoolNodeList.pools | first }}
        isRestoreVol: {{ .Volume.isRestoreVol }}
        openebs.io/storage-class-ref: |
          name: {{ .Volume.storageclass }}
          resourceVersion: {{ .TaskResult.creategetsc.storageClassVersion }}
    spec:
      capacity: {{ .Volume.capacity }}
      targetIP: {{ .TaskResult.cvolcreateputsvc.clusterIP }}
      {{- if ne $zvolWorkers  "" }}
      zvolWorkers: {{ .Config.ZvolWorkers.value }}
      {{- end }}
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
  name: cstor-volume-create-output-default
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
      labels:
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
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
  name: cstor-volume-list-listtargetservice-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    id: listlistsvc
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
    {{- $servicePairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.labels.openebs\\.io/persistent-volume},clusterIP={@.spec.clusterIP};{end}" | trim | default "" | splitList ";" -}}
    {{- $servicePairs | keyMap "volumeList" .ListItems | noop -}}
---
#runTask to list all cstor pv
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-list-listpv-default
spec:
  meta: |
    id: listlistpv
    apiVersion: v1
    kind: PersistentVolume
    action: list
    options: |-
      labelSelector: openebs.io/cas-type=cstor
  post: |
      {{- $pvPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.name},accessModes={@.spec.accessModes[0]},storageClass={@.spec.storageClassName};{end}" | trim | default "" | splitList ";" -}}
      {{- $pvPairs | keyMap "volumeList" .ListItems | noop -}}
    ---
# runTask to list all cstor target pods
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-list-listtargetpod-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    id: listlistctrl
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
    {{- $targetPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.labels.openebs\\.io/persistent-volume},targetIP={@.status.podIP},namespace={@.metadata.namespace},targetStatus={@.status.containerStatuses[*].ready};{end}" | trim | default "" | splitList ";" -}}
    {{- $targetPairs | keyMap "volumeList" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-list-listcstorvolumereplicacr-default
spec:
  meta: |
    runNamespace: {{.Config.RunNamespace.value}}
    id: listlistrep
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolumeReplica
    action: list
  post: |
    {{- $replicaPairs := jsonpath .JsonResult "{range .items[*]}pkey={@.metadata.labels.openebs\\.io/persistent-volume},replicaName={@.metadata.name},capacity={@.spec.capacity};{end}" | trim | default "" | splitList ";" -}}
    {{- $replicaPairs | keyMap "volumeList" .ListItems | noop -}}
---
# runTask to render volume list output
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-list-output-default
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
    properties were set in preceding list tasks,
    */}}
    {{- range $pkey, $map := .ListItems.volumeList }}
    {{- $capacity := pluck "capacity" $map | first | default "" | splitList ", " | first }}
    {{- $clusterIP := pluck "clusterIP" $map | first }}
    {{- $targetStatus := pluck "targetStatus" $map | first }}
    {{- $replicaName := pluck "replicaName" $map | first }}
    {{- $namespace := pluck "namespace" $map | first }}
    {{- $accessMode :=  pluck "accessModes" $map | first }}
    {{- $storageClass := pluck "storageClass" $map | first }}
    {{- $name := $pkey }}
      - kind: CASVolume
        apiVersion: v1alpha1
        metadata:
          name: {{ $name }}
          namespace: {{ $namespace }}
          annotations:
            openebs.io/storage-class: {{ $storageClass | default "" }}
            openebs.io/cluster-ips: {{ $clusterIP }}
            openebs.io/volume-size: {{ $capacity }}
            openebs.io/controller-status: {{ $targetStatus | default "" | replace "true" "running" | replace "false" "notready" }}
        spec:
          capacity: {{ $capacity }}
          iqn: iqn.2016-09.com.openebs.cstor:{{ $name }}
          targetPortal: {{ $clusterIP }}:3260
          targetIP: {{ $clusterIP }}
          targetPort: 3260
          replicas: {{ $replicaName | default "" | splitList ", " | len }}
          casType: cstor
          accessMode: {{ $accessMode | default "" }}
    {{- end -}}
---
# runTask to list cStor target deployment service
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-listtargetservice-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    {{- $runNamespace := .Config.RunNamespace.value -}}
    {{- $pvcServiceAccount := .Config.PVCServiceAccountName.value | default "" -}}
    {{- if ne $pvcServiceAccount "" }}
    runNamespace: {{ .Volume.runNamespace | saveAs "readlistsvc.derivedNS" .TaskResult }}
    {{ else }}
    runNamespace: {{ $runNamespace | saveAs "readlistsvc.derivedNS" .TaskResult }}
    {{- end }}
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
    {{- jsonpath .JsonResult "{.items[*].metadata.annotations.openebs\\.io/pvc-namespace}" | default "" | trim | saveAs "readlistsvc.pvcNs" .TaskResult | noop -}}
---
# runTask to list cstor volume cr
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-listcstorvolumecr-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.readlistsvc.derivedNS }}
    id: readlistcv
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolume
    action: list
    options: |-
      labelSelector: openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistcv.names" .TaskResult | noop -}}
    {{- .TaskResult.readlistcv.names | notFoundErr "cStor Volume CR not found" | saveIf "readlistcv.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].metadata.annotations.openebs\\.io/fs-type}" | trim | default "ext4" | saveAs "readlistcv.fsType" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].metadata.annotations.openebs\\.io/lun}" | trim | default "0" | int | saveAs "readlistcv.lun" .TaskResult | noop -}}
---
# runTask to list all replica crs of a volume
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-listcstorvolumereplicacr-default
spec:
  meta: |
    id: readlistrep
    runNamespace: {{.Config.RunNamespace.value}}
    apiVersion: openebs.io/v1alpha1
    kind: CStorVolumeReplica
    action: list
    options: |-
      labelSelector: openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
    {{- jsonpath .JsonResult "{.items[*].metadata.name}" | trim | saveAs "readlistrep.items" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].metadata.annotations.cstorpool\\.openebs\\.io/hostname}" | trim | saveAs "readlistrep.hostname" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].metadata.labels.cstorpool\\.openebs\\.io/name}" | trim | saveAs "readlistrep.poolname" .TaskResult | noop -}}
    {{- .TaskResult.readlistrep.items | notFoundErr "replicas not found" | saveIf "readlistrep.notFoundErr" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].spec.capacity}" | trim | saveAs "readlistrep.capacity" .TaskResult | noop -}}
---
# runTask to list cStor volume target pods
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-listtargetpod-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.readlistsvc.derivedNS }}
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
    {{- jsonpath .JsonResult "{.items[*].spec.nodeName}" | trim | saveAs "readlistctrl.targetNodeName" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.items[*].status.containerStatuses[*].ready}" | trim | saveAs "readlistctrl.status" .TaskResult | noop -}}
---
# runTask to render output of read volume task as CAS Volume
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-read-output-default
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
        openebs.io/cvr-names: {{ .TaskResult.readlistrep.items | default "" | splitList " " | join "," }}
        openebs.io/node-names: {{ .TaskResult.readlistrep.hostname | default "" | splitList " " | join "," }}
        openebs.io/pool-names: {{ .TaskResult.readlistrep.poolname | default "" | splitList " " | join "," }}
        openebs.io/controller-node-name: {{ .TaskResult.readlistctrl.targetNodeName | default ""}}
    spec:
      capacity: {{ $capacity }}
      iqn: iqn.2016-09.com.openebs.cstor:{{ .Volume.owner }}
      targetPortal: {{ .TaskResult.readlistsvc.clusterIP }}:3260
      targetIP: {{ .TaskResult.readlistsvc.clusterIP }}
      targetPort: 3260
      lun: {{ .TaskResult.readlistcv.lun }}
      fsType: {{ .TaskResult.readlistcv.fsType }}
      replicas: {{ .TaskResult.readlistrep.capacity | default "" | splitList " " | len }}
      casType: cstor
---
# runTask to list the cstorvolume that has to be deleted
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-listcstorvolumecr-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    {{- $runNamespace := .Config.RunNamespace.value -}}
    {{- $pvcServiceAccount := .Config.PVCServiceAccountName.value | default "" -}}
    {{- if ne $pvcServiceAccount "" }}
    runNamespace: {{ .Volume.runNamespace | saveAs "deletelistcsv.derivedNS" .TaskResult }}
    {{ else }}
    runNamespace: {{ $runNamespace | saveAs "deletelistcsv.derivedNS" .TaskResult }}
    {{- end }}
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
  name: cstor-volume-delete-listtargetservice-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.deletelistcsv.derivedNS }}
    id: deletelistsvc
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
  name: cstor-volume-delete-listtargetdeployment-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.deletelistcsv.derivedNS }}
    id: deletelistctrl
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
  name: cstor-volume-delete-listcstorvolumereplicacr-default
spec:
  meta: |
    id: deletelistcvr
    runNamespace: {{.Config.RunNamespace.value}}
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
    {{- $cvrs := jsonpath .JsonResult "{range .items[*]}pkey=cvrs,{@.metadata.name}='';{end}" | trim | default "" | splitList ";" -}}
    {{- $cvrs | notFoundErr "cstor volume replica not found" | saveIf "deletelistcvr.notFoundErr" .TaskResult | noop -}}
    {{- $cvrs | keyMap "cvrlist" .ListItems | noop -}}
---
# runTask to delete cStor volume target service
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-deletetargetservice-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.deletelistcsv.derivedNS }}
    id: deletedeletesvc
    apiVersion: v1
    kind: Service
    action: delete
    objectName: {{ .TaskResult.deletelistsvc.names }}
---
# runTask to delete cStor volume target deployment
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-deletetargetdeployment-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.deletelistcsv.derivedNS }}
    id: deletedeletectrl
    apiVersion: apps/v1beta1
    kind: Deployment
    action: delete
    objectName: {{ .TaskResult.deletelistctrl.names }}
---
# runTask to delete cstorvolumereplica
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-volume-delete-deletecstorvolumereplicacr-default
spec:
  meta: |
    runNamespace: {{.Config.RunNamespace.value}}
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
  name: cstor-volume-delete-deletecstorvolumecr-default
spec:
  meta: |
    {{- $isClone := .Volume.isCloneEnable | default "false" -}}
    runNamespace: {{ .TaskResult.deletelistcsv.derivedNS }}
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
  name: cstor-volume-delete-output-default
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

// CstorVolumeArtifacts returns the cstor volume related artifacts
// corresponding to latest version
func CstorVolumeArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(cstorVolumes{})...)
	return
}

type cstorVolumes struct{}

// FetchYamls returns all the yamls related to cstor volume in a string
// format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (c cstorVolumes) FetchYamls() string {
	return cstorVolumeYamls
}
