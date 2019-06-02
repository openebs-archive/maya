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

const openEBSCRDYamls = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: castemplates.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: castemplates
    # singular name to be used as an alias on the CLI and for display
    singular: castemplate
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CASTemplate
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - cast
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: runtasks.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: runtasks
    # singular name to be used as an alias on the CLI and for display
    singular: runtask
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: RunTask
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - rtask
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  # storagepoolclaim will be deprecated 
  name: storagepoolclaims.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: storagepoolclaims
    # singular name to be used as an alias on the CLI and for display
    singular: storagepoolclaim
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: StoragePoolClaim
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - spc
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorpoolclusters.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: cstorpoolclusters
    # singular name to be used as an alias on the CLI and for display
    singular: cstorpoolcluster
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CStorPoolCluster
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - cspc
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: storagepools.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: storagepools
    # singular name to be used as an alias on the CLI and for display
    singular: storagepool
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: StoragePool
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - sp
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorpools.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: cstorpools
    # singular name to be used as an alias on the CLI and for display
    singular: cstorpool
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CStorPool
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - csp
  additionalPrinterColumns:
  - JSONPath: .status.capacity.used
    name: Allocated
    description: The amount of storage space within the pool that has been physically allocated
    type: string
  - JSONPath: .status.capacity.free
    name: Free
    description: The amount of free space available in the pool
    type: string
  - JSONPath: .status.capacity.total
    name: Capacity
    description: Total size of the storage pool
    type: string
  - JSONPath: .status.phase
    name: Status
    description: Identifies the current health of the pool
    type: string
  - JSONPath: .spec.poolSpec.poolType
    name: Type
    description: The type of the storage pool
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorvolumes.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CStorVolume
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: cstorvolumes
    # singular name to be used as an alias on the CLI and for display
    singular: cstorvolume
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - cstorvolume
  additionalPrinterColumns:
  - JSONPath: .status.phase
    name: Status
    description: Identifies the current health of the target
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorvolumereplicas.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CStorVolumeReplica
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: cstorvolumereplicas
    # singular name to be used as an alias on the CLI and for display
    singular: cstorvolumereplica
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - cvr
  additionalPrinterColumns:
  - JSONPath: .status.capacity.used
    name: Used
    description: The amount of space that is "logically" consumed by this dataset
    type: string
  - JSONPath: .status.capacity.totalAllocated
    name: Allocated
    description: The amount of disk space consumed by a dataset and all its descendents
    type: string
  - JSONPath: .status.phase
    name: Status
    description: Identifies the current health of the replicas
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: disks.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: disks
    # singular name to be used as an alias on the CLI and for display
    singular: disk
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: Disk
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - disk
  additionalPrinterColumns:
  - JSONPath: .spec.capacity.storage
    name: Size
    description: Identifies the disk size(in Bytes)
    type: string
  - JSONPath: .status.state
    name: Status
    description: Identifies the current health of the disk
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: blockdevices.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: blockdevices
    # singular name to be used as an alias on the CLI and for display
    singular: blockdevice
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: BlockDevice
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - bd
  additionalPrinterColumns:
  - JSONPath: .spec.capacity.storage
    name: Size
    description: Identifies the block device size(in Bytes)
    type: string
  - JSONPath: .status.claimState
    name: ClaimState
    description: Identifies the current claim status of the block device
    type: string
  - JSONPath: .status.state
    name: Status
    description: Identifies the current health of the block device
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: blockdeviceclaims.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: blockdeviceclaims
    # singular name to be used as an alias on the CLI and for display
    singular: blockdeviceclaim
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: BlockDeviceClaim
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - bdc
  additionalPrinterColumns:
  - JSONPath: .spec.blockDeviceName
    name: BlockDeviceName
    description: Identifies the block device associated with the claim
    type: string
  - JSONPath: .status.phase
    name: Phase
    description: Identifies the phase of block device claim
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cstorbackups.openebs.io
spec:
  group: openebs.io
  version: v1alpha1
  scope: Namespaced
  names:
    plural: cstorbackups
    singular: cstorbackup
    kind: CStorBackup
    shortNames:
    - cbkp
    - cbkps
    - cbackups
    - cbackup
  additionalPrinterColumns:
    - JSONPath: .spec.volumeName
      name: volume
      description: volume on which backup performed
      type: string
    - JSONPath: .spec.backupName
      name: backup/schedule
      description: Backup/schedule name
      type: string
    - JSONPath: .status
      name: Status
      description: Backup status
      type: string

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cstorcompletedbackups.openebs.io
spec:
  group: openebs.io
  version: v1alpha1
  scope: Namespaced
  names:
    plural: cstorcompletedbackups
    singular: cstorcompletedbackup
    kind: CStorCompletedBackup
    shortNames:
    - cbkpc
    - cbackupcompleted
  additionalPrinterColumns:
    - JSONPath: .spec.volumeName
      name: volume
      description: volume on which backup performed
      type: string
    - JSONPath: .spec.backupName
      name: backup/schedule
      description: Backup/schedule name
      type: string
    - JSONPath: .spec.prevSnapName
      name: lastSnap
      description: Last successful backup snapshot
      type: string
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cstorrestores.openebs.io
spec:
  group: openebs.io
  version: v1alpha1
  scope: Namespaced
  names:
    plural: cstorrestores
    singular: cstorrestore
    kind: CStorRestore
    shortNames:
    - crst
    - crsts
    - crestores
    - crestore
  additionalPrinterColumns:
    - JSONPath: .spec.restoreName
      name: backup
      description: backup name which is  restored
      type: string
    - JSONPath: .spec.volumeName
      name: volume
      description: volume on which restore performed
      type: string
    - JSONPath: .status
      name: Status
      description: Restore status
      type: string
---
`

// OpenEBSCRDArtifacts returns the CRDs required for latest version
func OpenEBSCRDArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(openEBSCRDs{})...)
	return
}

type openEBSCRDs struct{}

// FetchYamls returns all the CRD yamls related to 0.7.0
// in a string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (o openEBSCRDs) FetchYamls() string {
	return openEBSCRDYamls
}
