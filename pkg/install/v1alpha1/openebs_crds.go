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
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: castemplates.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  versions:
  - name: v1alpha1
    storage: true
    served: true
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
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
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: runtasks.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  versions:
  - name: v1alpha1
    storage: true
    served: true
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
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
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  # storagepoolclaim will be deprecated
  name: storagepoolclaims.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  versions:
  - name: v1alpha1
    storage: true
    served: true
    deprecated: true
    deprecationWarning: "openebs.io/v1alpha1 StoragePoolClaim is deprecated; see https://github.com/openebs/upgrade/blob/HEAD/README.md for instructions to migrate to cstor.openebs.io/v1 CStorPoolCluster"
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
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
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: storagepools.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  versions:
  - name: v1alpha1
    storage: true
    served: true
    deprecated: true
    deprecationWarning: "openebs.io/v1alpha1 StoragePool is deprecated; see https://github.com/openebs/upgrade/blob/HEAD/README.md for instructions to migrate to openebs.io/v1alpha1 JivaVolumePolicy"
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
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
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorpools.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  versions:
  - name: v1alpha1
    storage: true
    served: true
    deprecated: true
    deprecationWarning: "openebs.io/v1alpha1 CStorPool is deprecated; see https://github.com/openebs/upgrade/blob/HEAD/README.md for instructions to migrate to cstor.openebs.io/v1 CStorPoolInstance"
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
    additionalPrinterColumns:
    - jsonPath: .status.capacity.used
      name: Allocated
      description: The amount of storage space within the pool that has been physically allocated
      type: string
    - jsonPath: .status.capacity.free
      name: Free
      description: The amount of free space available in the pool
      type: string
    - jsonPath: .status.capacity.total
      name: Capacity
      description: Total size of the storage pool
      type: string
    - jsonPath: .status.phase
      name: Status
      description: Identifies the current health of the pool
      type: string
    - jsonPath: .status.readOnly
      description: Identifies the pool read only mode
      name: ReadOnly
      type: boolean
    - jsonPath: .spec.poolSpec.poolType
      name: Type
      description: The type of the storage pool
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
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
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorvolumes.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  versions:
  - name: v1alpha1
    storage: true
    served: true
    deprecated: true
    deprecationWarning: "openebs.io/v1alpha1 CStorVolume is deprecated; see https://github.com/openebs/upgrade/blob/HEAD/README.md for instructions to migrate to cstor.openebs.io/v1 CStorVolume"
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
    additionalPrinterColumns:
    - jsonPath: .status.phase
      name: Status
      description: Identifies the current health of the target
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    - jsonPath: .status.capacity
      description: Current volume capacity
      name: Capacity
      type: string
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
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorvolumereplicas.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  versions:
  - name: v1alpha1
    storage: true
    served: true
    deprecated: true
    deprecationWarning: "openebs.io/v1alpha1 CStorVolumeReplica is deprecated; see https://github.com/openebs/upgrade/blob/HEAD/README.md for instructions to migrate to cstor.openebs.io/v1 CStorVolumeReplica"
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
    additionalPrinterColumns:
    - jsonPath: .status.capacity.used
      name: Used
      description: The amount of space that is "logically" consumed by this dataset
      type: string
    - jsonPath: .status.capacity.totalAllocated
      name: Allocated
      description: The amount of disk space consumed by a dataset and all its descendents
      type: string
    - jsonPath: .status.phase
      name: Status
      description: Identifies the current health of the replicas
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
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
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: cstorbackups.openebs.io
  # either Namespaced or Cluster
spec:
  group: openebs.io
  versions:
  - name: v1alpha1
    storage: true
    served: true
    deprecated: true
    deprecationWarning: "openebs.io/v1alpha1 CStorBackups is deprecated; see https://github.com/openebs/upgrade/blob/HEAD/README.md for instructions to migrate to cstor.openebs.io/v1 CStorBackups"
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
    additionalPrinterColumns:
    - jsonPath: .spec.volumeName
      name: volume
      description: volume on which backup performed
      type: string
    - jsonPath: .spec.backupName
      name: backup/schedule
      description: Backup/schedule name
      type: string
    - jsonPath: .status
      name: Status
      description: Backup status
      type: string
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
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: cstorcompletedbackups.openebs.io
spec:
  group: openebs.io
  versions:
  - name: v1alpha1
    storage: true
    served: true
    deprecated: true
    deprecationWarning: "openebs.io/v1alpha1 CStorCompletedBackup is deprecated; see https://github.com/openebs/upgrade/blob/HEAD/README.md for instructions to migrate to cstor.openebs.io/v1 CStorCompletedBackup"
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
    additionalPrinterColumns:
    - jsonPath: .spec.volumeName
      name: volume
      description: volume on which backup performed
      type: string
    - jsonPath: .spec.backupName
      name: backup/schedule
      description: Backup/schedule name
      type: string
    - jsonPath: .spec.prevSnapName
      name: lastSnap
      description: Last successful backup snapshot
      type: string
  scope: Namespaced
  names:
    plural: cstorcompletedbackups
    singular: cstorcompletedbackup
    kind: CStorCompletedBackup
    shortNames:
    - cbkpc
    - cbackupcompleted
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: cstorrestores.openebs.io
spec:
  group: openebs.io
  versions:
  - name: v1alpha1
    storage: true
    served: true
    deprecated: true
    deprecationWarning: "openebs.io/v1alpha1 CStorRestore is deprecated; see https://github.com/openebs/upgrade/blob/HEAD/README.md for instructions to migrate to cstor.openebs.io/v1 CStorRestore"
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
    additionalPrinterColumns:
    - jsonPath: .spec.restoreName
      name: backup
      description: backup name which is  restored
      type: string
    - jsonPath: .spec.volumeName
      name: volume
      description: volume on which restore performed
      type: string
    - jsonPath: .status
      name: Status
      description: Restore status
      type: string
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
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: upgradetasks.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  versions:
  - name: v1alpha1
    storage: true
    served: true
    schema:
      openAPIV3Schema:
        x-kubernetes-preserve-unknown-fields: true
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: upgradetasks
    # singular name to be used as an alias on the CLI and for display
    singular: upgradetask
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: UpgradeTask
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - utask
---
---
`

// OpenEBSCRDArtifacts returns the CRDs required for latest version
func OpenEBSCRDArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamlsIf(openEBSCRDs{}, IsInstallCRDEnabled)...)
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
