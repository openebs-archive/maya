/*
Copyright 2018 The OpenEBS Authors.
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

package k8s

import (
	openebs_v1 "github.com/openebs/CITF/pkg/apis/openebs.io/v1alpha1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateStoragePoolClaim takes StoragePoolClaim as an argument and creates it.
func (k8s K8S) CreateStoragePoolClaim(storagePoolClaim *openebs_v1.StoragePoolClaim) (*openebs_v1.StoragePoolClaim, error) {
	spcClient := k8s.OpenebsClientSet.OpenebsV1alpha1().StoragePoolClaims()
	return spcClient.Create(storagePoolClaim)
}

// GetStoragePoolClaim returns the StoragePoolClaim object for given spcName.
func (k8s K8S) GetStoragePoolClaim(spcName string, opts meta_v1.GetOptions) (*openebs_v1.StoragePoolClaim, error) {
	spcClient := k8s.OpenebsClientSet.OpenebsV1alpha1().StoragePoolClaims()
	return spcClient.Get(spcName, opts)
}

// ListStoragePoolClaims returns an object of StoragePoolClaimList.
func (k8s K8S) ListStoragePoolClaims(opts meta_v1.ListOptions) (*openebs_v1.StoragePoolClaimList, error) {
	spcClient := k8s.OpenebsClientSet.OpenebsV1alpha1().StoragePoolClaims()
	return spcClient.List(opts)
}

// DeleteStoragePoolClaim deletes a StoragePoolClaim with the given name.
func (k8s K8S) DeleteStoragePoolClaim(spcName string, opts *meta_v1.DeleteOptions) error {
	spcClient := k8s.OpenebsClientSet.OpenebsV1alpha1().StoragePoolClaims()
	return spcClient.Delete(spcName, opts)
}

// GetCStorPool returns the CStorPool object for given cStorPoolName.
// :return: *openebs_v1.CStorPool: Pointer to CStorPool objects.
func (k8s K8S) GetCStorPool(cStorPoolName string, opts meta_v1.GetOptions) (*openebs_v1.CStorPool, error) {
	cStorPoolClient := k8s.OpenebsClientSet.OpenebsV1alpha1().CStorPools()
	return cStorPoolClient.Get(cStorPoolName, opts)
}

// ListCStorPool returns all CStorPool objects.
func (k8s K8S) ListCStorPool(opts meta_v1.ListOptions) (*openebs_v1.CStorPoolList, error) {
	cStorPoolClient := k8s.OpenebsClientSet.OpenebsV1alpha1().CStorPools()
	return cStorPoolClient.List(opts)
}

// DeleteCStorPool deletes a CStorPool with the given name.
func (k8s K8S) DeleteCStorPool(cStorPoolName string, opts *meta_v1.DeleteOptions) error {
	cStorePoolClient := k8s.OpenebsClientSet.OpenebsV1alpha1().CStorPools()
	return cStorePoolClient.Delete(cStorPoolName, opts)
}

// GetStoragePool returns the StoragePool object for the give storagePoolName
func (k8s K8S) GetStoragePool(storagePoolName string, opts meta_v1.GetOptions) (*openebs_v1.StoragePool, error) {
	storagePoolClient := k8s.OpenebsClientSet.OpenebsV1alpha1().StoragePools()
	return storagePoolClient.Get(storagePoolName, opts)
}

// ListStoragePool returns all the StoragePool objects
func (k8s K8S) ListStoragePool(opts meta_v1.ListOptions) (*openebs_v1.StoragePoolList, error) {
	storagePoolClient := k8s.OpenebsClientSet.OpenebsV1alpha1().StoragePools()
	return storagePoolClient.List(opts)
}

// DeleteStoragePool deletes a StoragePool object with the given storagePoolName
func (k8s K8S) DeleteStoragePool(storagePoolName string, opts *meta_v1.DeleteOptions) error {
	storagePoolClient := k8s.OpenebsClientSet.OpenebsV1alpha1().StoragePools()
	return storagePoolClient.Delete(storagePoolName, opts)
}

// GetDisk returns the Disk object for the given disk name
func (k8s K8S) GetDisk(diskName string, opts meta_v1.GetOptions) (*openebs_v1.Disk, error) {
	diskClient := k8s.OpenebsClientSet.OpenebsV1alpha1()
	return diskClient.Disks().Get(diskName, opts)
}

// ListDisks list all the Disk objects
func (k8s K8S) ListDisks(opts meta_v1.ListOptions) (*openebs_v1.DiskList, error) {
	diskClient := k8s.OpenebsClientSet.OpenebsV1alpha1()
	return diskClient.Disks().List(opts)
}

// DeleteDisk deletes a Disk object with the given diskname
func (k8s K8S) DeleteDisk(diskName string, opts *meta_v1.DeleteOptions) error {
	diskClient := k8s.OpenebsClientSet.OpenebsV1alpha1()
	return diskClient.Disks().Delete(diskName, opts)
}

// GetCStorVolumeReplica returns the CStorVolumeReplica object for given CStorVolumeReplicaName and namespace
func (k8s K8S) GetCStorVolumeReplica(cvrName, namespace string, opts meta_v1.GetOptions) (*openebs_v1.CStorVolumeReplica, error) {
	cvrClient := k8s.OpenebsClientSet.OpenebsV1alpha1().CStorVolumeReplicas(namespace)
	return cvrClient.Get(cvrName, opts)
}

// ListCStorVolumeReplica returns all the CStorVolumeReplicaList for given namespace
func (k8s K8S) ListCStorVolumeReplica(namespace string, opts meta_v1.ListOptions) (*openebs_v1.CStorVolumeReplicaList, error) {
	cvrClient := k8s.OpenebsClientSet.OpenebsV1alpha1().CStorVolumeReplicas(namespace)
	return cvrClient.List(opts)
}

// DeleteCStorVolumeReplica deletes a CStorVolumeReplica with the given CStorVolumeReplicaName and namespace
func (k8s K8S) DeleteCStorVolumeReplica(cvrName, namespace string, opts *meta_v1.DeleteOptions) error {
	cvrClient := k8s.OpenebsClientSet.OpenebsV1alpha1().CStorVolumeReplicas(namespace)
	return cvrClient.Delete(cvrName, opts)
}
