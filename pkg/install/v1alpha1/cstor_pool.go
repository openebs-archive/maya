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

const cstorPoolYamls = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-pool-create-default
spec:
  defaultConfig:
  # CstorPoolImage is the container image that executes zpool replication and
  # communicates with cstor iscsi target
  - name: CstorPoolImage
    value: {{env "OPENEBS_IO_CSTOR_POOL_IMAGE" | default "openebs/cstor-pool:latest"}}
  # CstorPoolMgmtImage runs cstor pool and cstor volume replica related CRUD
  # operations
  - name: CstorPoolMgmtImage
    value: {{env "OPENEBS_IO_CSTOR_POOL_MGMT_IMAGE" | default "openebs/cstor-pool-mgmt:latest"}}
  # HostPathType is a hostPath volume i.e. mounts a file or directory from the
  # host node’s filesystem into a Pod. 'DirectoryOrCreate' value  ensures
  # nothing exists at the given path i.e. an empty directory will be created.
  - name: HostPathType
    value: DirectoryOrCreate
  # SparseDir is a hostPath directory where to look for sparse files
  - name: SparseDir
    value: {{env "OPENEBS_IO_CSTOR_POOL_SPARSE_DIR" | default "/var/openebs/sparse"}}
  # RunNamespace is the namespace where namespaced resources related to pool
  # will be placed
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  # ServiceAccountName is the account name assigned to pool management pod
  # with permissions to view, create, edit, delete required custom resources
  - name: ServiceAccountName
    value: {{env "OPENEBS_SERVICE_ACCOUNT"}}
  # PoolResourceRequests allow you to specify resource requests that need to be available
  # before scheduling the containers. If not specified, the default is to use the limits
  # from PoolResourceLimits or the default requests set in the cluster. 
  - name: PoolResourceRequests
    value: "none"
  # PoolResourceLimits allow you to set the limits on memory and cpu for pool pods
  # The resource and limit value should be in the same format as expected by
  # Kubernetes. Example:
  #- name: PoolResourceLimits
  #  value: |-
  #      memory: 1Gi
  - name: PoolResourceLimits
    value: "none"
  # AuxResourceLimits allow you to set limits on side cars. Limits have to be specified
  # in the format expected by Kubernetes
  - name: AuxResourceLimits
    value: "none"
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    # Following are the list of run tasks executed in this order to
    # create a cstor storage pool
    - cstor-pool-create-getspcinfo-default
    - cstor-pool-create-listnode-default
    - cstor-pool-create-putcstorpoolcr-default
    - cstor-pool-create-putcstorpooldeployment-default
    - cstor-pool-create-putstoragepoolcr-default
    - cstor-pool-create-patchstoragepoolclaim-default
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cstor-pool-delete-default
spec:
  defaultConfig:
    # RunNamespace is the namespace to use to delete pool resources
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
    # Following are run tasks executed in this order to delete a storage pool
    - cstor-pool-delete-listcstorpoolcr-default
    - cstor-pool-delete-deletecstorpoolcr-default
    - cstor-pool-delete-listcstorpooldeployment-default
    - cstor-pool-delete-deletecstorpooldeployment-default
    - cstor-pool-delete-liststoragepoolcr-default
    - cstor-pool-delete-deletestoragepoolcr-default
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-create-getspcinfo-default
spec:
  meta: |
    id: getspcinfo
    apiVersion: openebs.io/v1alpha1
    kind: StoragePoolClaim
    objectName: {{.Storagepool.owner}}
    action: get
  post: |
    # For backward compatibility, getspcinfo.disk is saved as a task result
    {{- jsonpath .JsonResult "{range .spec.disks.diskList[*]}{$},{end}" | trim | saveAs "getspcinfo.disk" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.spec.poolSpec.poolType}" | trim | saveAs "getspcinfo.poolType" .TaskResult | noop -}}
    {{- jsonpath .JsonResult "{.spec.type}" | trim | saveAs "getspcinfo.type" .TaskResult | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-create-listnode-default
spec:
  meta: |
    id: listnode
    apiVersion: openebs.io/v1alpha1
    kind: Disk
    action: get
    repeatWith:
      metas:
      {{- $diskList := .TaskResult.getspcinfo.disk }}
      # To support backward compatibility
      # If .TaskResult.getspcinfo.disk is empty, get disk list from CAS engine top level property
      {{if $diskList}}
      {{- $diskList := .TaskResult.getspcinfo.disk | replace "," " "| trim | split " "}}
      {{ range $k,$v := $diskList }}
      - objectName: {{$v}}
      {{ end }}
      {{else}}
      {{- $diskList := .Storagepool.diskList}}
      {{ range $k,$v := $diskList }}
      - objectName: {{$v}}
      {{ end }}
      {{ end }}
  post: |
    {{- $diskDevLink:= jsonpath .JsonResult "{@.spec.devlinks[0].links[0]}"}}
    {{if $diskDevLink }}
    {{- $nodeDiskdevlinkList := jsonpath .JsonResult "pkey=node,{@.metadata.labels.kubernetes\\.io/hostname}={@.spec.devlinks[0].links[0]};" | trim | default "" | splitList ";" -}}
    {{- $nodeDiskdevlinkList | keyMap "nodeDiskdevlinkMap" .ListItems | noop -}}
    {{else}}
    {{- $nodeDiskdevlinkList := jsonpath .JsonResult "pkey=node,{@.metadata.labels.kubernetes\\.io/hostname}={@.spec.path};" | trim | default "" | splitList ";" -}}
    {{- $nodeDiskdevlinkList | keyMap "nodeDiskdevlinkMap" .ListItems | noop -}}
    {{end}}
    {{- $nodeDiskList := jsonpath .JsonResult "pkey=node,{@.metadata.labels.kubernetes\\.io/hostname}={@.metadata.name};" | trim | default "" | splitList ";" -}}
    {{- $nodeDiskList | keyMap "nodeDiskMap" .ListItems | noop -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-create-putcstorpoolcr-default
spec:
  meta: |
    apiVersion: openebs.io/v1alpha1
    kind: CStorPool
    action: put
    id: putcstorpoolcr
    repeatWith:
      resources:
      {{- range $k, $v := .ListItems.nodeDiskdevlinkMap.node}}
      - {{ $k }}
      {{- end }}
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | addTo "putcstorpoolcr.objectName" .TaskResult | noop -}}
    {{- $nodeUidList := jsonpath .JsonResult "pkey=nodeUid,{.metadata.labels.kubernetes\\.io/hostname}={.metadata.uid} {.metadata.name};" | trim | default "" | splitList ";" -}}
    {{- $nodeUidList | keyMap "nodeUidMap" .ListItems | noop -}}
  task: |
    apiVersion: openebs.io/v1alpha1
    kind: CStorPool
    metadata:
      name: {{.Storagepool.owner}}-{{randAlphaNum 4 |lower }}
      labels:
        openebs.io/storage-pool-claim: {{.Storagepool.owner}}
        kubernetes.io/hostname: {{ .ListItems.currentRepeatResource }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
    spec:
      disks:
        diskList: {{ pluck .ListItems.currentRepeatResource .ListItems.nodeDiskdevlinkMap.node }}
      poolSpec:
        poolType: {{.TaskResult.getspcinfo.poolType}}
        cacheFile: /tmp/{{.Storagepool.owner}}.cache
        overProvisioning: false
    status:
      phase: {{ .Storagepool.phase }}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-create-putcstorpooldeployment-default
spec:
  meta: |
    runNamespace: {{.Config.RunNamespace.value}}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: put
    id: putcstorpooldeployment
    repeatWith:
      resources:
      {{- range $k, $v := .ListItems.nodeUidMap.nodeUid }}
      - {{ $k }}
      {{- end }}
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | addTo "putcstorpooldeployment.objectName" .TaskResult | noop -}}
  task: |
    {{- $setResourceRequests := .Config.PoolResourceRequests.value | default "none" -}}
    {{- $resourceRequestsVal := fromYaml .Config.PoolResourceRequests.value -}}
    {{- $setResourceLimits := .Config.PoolResourceLimits.value | default "none" -}}
    {{- $resourceLimitsVal := fromYaml .Config.PoolResourceLimits.value -}}
    {{- $setAuxResourceLimits := .Config.AuxResourceLimits.value | default "none" -}}
    {{- $auxResourceLimitsVal := fromYaml .Config.AuxResourceLimits.value -}}
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: {{ pluck .ListItems.currentRepeatResource .ListItems.nodeUidMap.nodeUid |first | splitList " " | last}}
      labels:
        openebs.io/storage-pool-claim: {{.Storagepool.owner}}
        openebs.io/cstor-pool: {{ pluck .ListItems.currentRepeatResource .ListItems.nodeUidMap.nodeUid |first | splitList " " | last}}
        app: cstor-pool
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: cstor-pool
      template:
        metadata:
          labels:
            app: cstor-pool
        spec:
          serviceAccountName: {{ .Config.ServiceAccountName.value }}
          nodeSelector:
            kubernetes.io/hostname: {{ .ListItems.currentRepeatResource}}
          containers:
          - name: cstor-pool
            image: {{ .Config.CstorPoolImage.value }}
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
            - containerPort: 12000
              protocol: TCP
            - containerPort: 3233
              protocol: TCP
            - containerPort: 3232
              protocol: TCP
            securityContext:
              privileged: true
            volumeMounts:
            - name: device
              mountPath: /dev
            - name: tmp
              mountPath: /tmp
            - name: sparse
              mountPath: {{ .Config.SparseDir.value }}
            - name: udev
              mountPath: /run/udev
              # To avoid clash between terminating and restarting pod
              # in case older zrepl gets deleted faster, we keep initial delay
            lifecycle:
              postStart:
                 exec:
                    command: ["/bin/sh", "-c", "sleep 2"]
          - name: cstor-pool-mgmt
            image: {{ .Config.CstorPoolMgmtImage.value }}
            {{- if ne $setAuxResourceLimits "none" }}
            resources:
              limits:
              {{- range $rKey, $rLimit := $auxResourceLimitsVal }}
                {{ $rKey }}: {{ $rLimit }}
              {{- end }}
            {{- end }}
            ports:
            - containerPort: 9500
              protocol: TCP
            securityContext:
              privileged: true
            volumeMounts:
            - name: device
              mountPath: /dev
            - name: tmp
              mountPath: /tmp
            - name: sparse
              mountPath: {{ .Config.SparseDir.value }}
            - name: udev
              mountPath: /run/udev
            env:
              # OPENEBS_IO_CSTOR_ID env has UID of cStorPool CR.
            - name: OPENEBS_IO_CSTOR_ID
              value: {{ pluck .ListItems.currentRepeatResource .ListItems.nodeUidMap.nodeUid |first | splitList " " | first}}
          volumes:
          - name: device
            hostPath:
              # directory location on host
              path: /dev
              # this field is optional
              type: Directory
          - name: tmp
            hostPath:
              # From host, dir called /var/openebs/shared-<uid> is created to avoid clash if two replicas run on same node.
              path: /var/openebs/shared-{{.Storagepool.owner}}
              type: {{ .Config.HostPathType.value }}
          - name: sparse
            hostPath:
              path: {{ .Config.SparseDir.value }}
              type: {{ .Config.HostPathType.value }}
          - name: udev
            hostPath:
              path: /run/udev
              type: Directory
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-create-putstoragepoolcr-default
spec:
  meta: |
    apiVersion: openebs.io/v1alpha1
    kind: StoragePool
    action: put
    id: putstoragepool
    repeatWith:
      resources:
      {{- range $k, $v := .ListItems.nodeDiskdevlinkMap.node}}
      - {{ $k }}
      {{- end }}
  post: |
    {{- jsonpath .JsonResult "{.metadata.name}" | trim | addTo "putstoragepool.objectName" .TaskResult | noop -}}
  task: |
    apiVersion: openebs.io/v1alpha1
    kind: StoragePool
    metadata:
      name: {{ pluck .ListItems.currentRepeatResource .ListItems.nodeUidMap.nodeUid |first | splitList " " | last }}
      labels:
        openebs.io/storage-pool-claim: {{.Storagepool.owner}}
        openebs.io/cstor-pool: {{ pluck .ListItems.currentRepeatResource .ListItems.nodeUidMap.nodeUid |first | splitList " " | last}}
        openebs.io/cas-type: cstor
        kubernetes.io/hostname: {{ .ListItems.currentRepeatResource }}
        openebs.io/version: {{ .CAST.version }}
        openebs.io/cas-template-name: {{ .CAST.castName }}
    spec:
      disks:
        diskList: {{ pluck .ListItems.currentRepeatResource .ListItems.nodeDiskMap.node }}
      poolSpec:
        poolType: {{.TaskResult.getspcinfo.poolType}}
        cacheFile: /tmp/{{.Storagepool.owner}}.cache
        overProvisioning: false
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-create-patchstoragepoolclaim-default
spec:
  meta: |
    id: patchstoragepoolclaim
    apiVersion: openebs.io/v1alpha1
    kind: StoragePoolClaim
    objectName: {{.Storagepool.owner}}
    action: patch
  task: |
    type: merge
    pspec: |-
      status:
        phase: Online
---
# This run task lists all cstor pool CRs that need to be deleted
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-delete-listcstorpoolcr-default
spec:
  meta: |
    id: listcstorpoolcr
    apiVersion: openebs.io/v1alpha1
    kind: CStorPool
    action: list
    options: |-
      labelSelector: openebs.io/storage-pool-claim={{.Storagepool.owner}}
  post: |
    {{- $csps := jsonpath .JsonResult "{range .items[*]}pkey=csps,{@.metadata.name}=;{end}" | trim | default "" | splitList ";" -}}
    {{- $csps | notFoundErr "cstor pool cr not found" | saveIf "listcstorpoolcr.notFoundErr" .TaskResult | noop -}}
    {{- $csps | keyMap "csplist" .ListItems | noop -}}
---
# This run task delete all the required cstor pool CR
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-delete-deletecstorpoolcr-default
spec:
  meta: |
    apiVersion: openebs.io/v1alpha1
    kind: CStorPool
    action: delete
    id: deletecstorpoolcr
    objectName: {{ keys .ListItems.csplist.csps | join "," }}
---
# This run task lists all the required cstor pool deployments that need to be deleted
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-delete-listcstorpooldeployment-default
spec:
  meta: |
    id: listcstorpooldeployment
    apiVersion: extensions/v1beta1
    runNamespace: {{.Config.RunNamespace.value}}
    kind: Deployment
    action: list
    options: |-
      labelSelector: openebs.io/storage-pool-claim={{.Storagepool.owner}}
  post: |
    {{- $csds := jsonpath .JsonResult "{range .items[*]}pkey=csds,{@.metadata.name}=;{end}" | trim | default "" | splitList ";" -}}
    {{- $csds | notFoundErr "cstor pool deployment not found" | saveIf "listcstorpooldeployment.notFoundErr" .TaskResult | noop -}}
    {{- $csds | keyMap "csdlist" .ListItems | noop -}}
---
# This run task deletes all the required cstor pool deployments
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-delete-deletecstorpooldeployment-default
spec:
  meta: |
    id: deletecstorpooldeployment
    runNamespace: {{.Config.RunNamespace.value}}
    apiVersion: extensions/v1beta1
    kind: Deployment
    action: delete
    objectName: {{ keys .ListItems.csdlist.csds | join "," }}
---
# This run task lists all storage pool CRs that need to be deleted
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-delete-liststoragepoolcr-default
spec:
  meta: |
    id: liststoragepoolcr
    apiVersion: openebs.io/v1alpha1
    kind: StoragePool
    action: list
    options: |-
      labelSelector: openebs.io/storage-pool-claim={{.Storagepool.owner}}
  post: |
    {{- $sps := jsonpath .JsonResult "{range .items[*]}pkey=sps,{@.metadata.name}=;{end}" | trim | default "" | splitList ";" -}}
    {{- $sps | notFoundErr "storge pool cr not found" | saveIf "listcstorpoolcr.notFoundErr" .TaskResult | noop -}}
    {{- $sps | keyMap "splist" .ListItems | noop -}}
---
# This run task deletes the required storagepool object
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata:
  name: cstor-pool-delete-deletestoragepoolcr-default
spec:
  meta: |
    id: deletestoragepoolcr
    apiVersion: openebs.io/v1alpha1
    kind: StoragePool
    action: delete
    objectName: {{ keys .ListItems.splist.sps | join "," }}
---
`

// CstorPoolArtifacts returns the cstor pool related artifacts corresponding to
// latest version
func CstorPoolArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(cstorPools{})...)
	return
}

type cstorPools struct{}

// FetchYamls returns all the yamls related to cstor pool in a string
// format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (c cstorPools) FetchYamls() string {
	return cstorPoolYamls
}
