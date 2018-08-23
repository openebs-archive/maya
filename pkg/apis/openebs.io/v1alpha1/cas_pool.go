/*
Copyright 2017 The OpenEBS Authors.

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

// CasPool is a type which will be utilised by CAS engine to perform
// storagepool related operation
type CasPoolKey string
type CasPoolVals int

const (
	// HostNameCPK is the kubernetes host name label
	HostNameCPK CasPoolKey = "kubernetes.io/hostname"
	// StoragePoolClaimCPK is the storage pool claim label
	StoragePoolClaimCPK CasPoolKey = "openebs.io/storage-pool-claim"
	// DiskTypeCPK is the node-disk-manager disk type e.g. 'sparse' or 'disk'
	DiskTypeCPK CasPoolKey = "ndm.io/disk-type"
	// PoolTypeMirroredCPK is a key for mirrored for pool
	PoolTypeMirroredCPK CasPoolKey = "mirrored"
	// PoolTypeMirroredCPK is a key for striped for pool
	PoolTypeStripedCPK CasPoolKey = "striped"
	// TypeSparseCPK is a key for sparse disk pool
	TypeSparseCPK CasPoolKey = "sparse"
	// TypeDiskCPK is a key for physical disk pool
	TypeDiskCPK CasPoolKey = "disk"
	// StripedDiskCountCPK is the count for striped type pool
	StripedDiskCountCPK CasPoolVals = 1
	// MirroredDiskCountCPK is the count for mirrored type pool
	MirroredDiskCountCPK CasPoolVals = 2
)

type CasPool struct {
	// StoragePoolClaim is the name of the storagepoolclaim object
	StoragePoolClaim string

	// CasCreateTemplate is the cas template that will be used for storagepool create
	// operation
	CasCreateTemplate string

	// CasDeleteTemplate is the cas template that will be used for storagepool delete
	// operation
	CasDeleteTemplate string

	// Namespace can be passed via storagepoolclaim as labels to decide on the
	// execution of namespaced resources with respect to storagepool
	Namespace string

	// DiskList is the list of disks over which a storagepool will be provisioned
	DiskList []string

	// PoolType is the type of pool to be provisioned e.g. striped or mirrored
	PoolType string

	// MaxPool is the maximum number of pool that should be provisioned
	MaxPools int

	// MinPool is the minimum number of pool that should be provisioned
	MinPools int

	// Type is the CasPool type e.g. sparse or openebs-cstor
	Type string

	// reSync will decide whether the event is a reconciliation event
	ReSync bool

	// SparePoolCount is the number of pools that will be tried for creation as a part of reconciliation.
	SparePoolCount int
}
