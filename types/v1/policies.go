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

package v1

// K8sVolumeLabelKey is a typed string used to represent
// volume related policy keys w.r.t K8s context.
type K8sVolumeLabelKey string

const (
	// StorageClassKVLK is the key to fetch the name of K8s storage
	// class.
	StorageClassKVLK K8sVolumeLabelKey = "k8s.io/storage-class"

	// NamespaceKVLK is the key to fetch the name of K8s namespace
	NamespaceKVLK K8sVolumeLabelKey = "k8s.io/namespace"

	// OutClusterKVLK is the key to fetch the out-cluster value.
	// In other words it's value suggests if the volume operation
	// will be executed outside the K8s cluster
	OutClusterKVLK K8sVolumeLabelKey = "k8s.io/out-cluster"
)

// K8sVolumeKey is a typed structure that consists of
// various volume related policy keys that are understood w.r.t K8s
// context.
type K8sVolumeKey struct {

	// StorageClass contains the name of the K8s storage class
	// which will be used during volume operations. A K8s storage
	// class will typically have various volume policies set in it.
	StorageClass string `yaml:"k8s.io/storage-class"`

	// Namespace contains the K8s namespace where the volume
	// operations will be executed
	Namespace string `yaml:"k8s.io/namespace"`

	// OutCluster contains the external K8s cluster information where the
	// volume operations will be executed
	OutCluster string `yaml:"k8s.io/out-cluster"`
}

// VolumeLabelKey is a typed string used to represent openebs
// volume related policy keys.
type VolumeLabelKey string

const (
	// CapacityOldVLK is the key to fetch volume capacity
	// TODO Deprecate in favour of CapacityVLK
	CapacityOldVLK VolumeLabelKey = "volumeprovisioner.mapi.openebs.io/storage-size"

	// CapacityVLK is the key to fetch volume capacity
	CapacityVLK VolumeLabelKey = "openebs.io/capacity"
)

// VolumeKey is a typed structure that consists of
// various volume related policy keys that are understood w.r.t OpenEBS
// context.
type VolumeKey struct {
	// CapacityOld contains the capacity of the volume.
	// TODO Deprecate in favour of Capacity
	CapacityOld string `yaml:"volumeprovisioner.mapi.openebs.io/storage-size"`

	// Capacity containts the capacity of the volume
	Capacity string `yaml:"openebs.io/capacity"`
}

// ReplicaPlacementLabelKey is a typed string used to represent
// replica placement related policy keys.
type ReplicaPlacementLabelKey string

const (
	// ReplicasOldRPLK is the key to fetch replica count
	// TODO Deprecate in favour of ReplicasRPLK
	ReplicasOldRPLK ReplicaPlacementLabelKey = "volumeprovisioner.mapi.openebs.io/replica-count"

	// ReplicasRPLK is the key to fetch replica count
	ReplicasRPLK ReplicaPlacementLabelKey = "openebs.io/replica-count"
)

// ReplicaPlacementKey is a typed structure that consists of
// various replica placement related policy keys.
type ReplicaPlacementKey struct {

	// ReplicasOld contains the replica count
	// TODO Deprecate in favour of Replicas
	ReplicasOld string `yaml:"volumeprovisioner.mapi.openebs.io/replica-count"`

	// Replicas contains the replica count
	Replicas string `yaml:"openebs.io/replica-count"`
}
