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

// OldVolumeLabels is a set of labels set against the volume structure
// This is specifically for backward compatibility
type OldVolumeLabels struct {
	// CapacityOld contains the volume capacity value
	CapacityOld string `json:"volumeprovisioner.mapi.openebs.io/storage-size,omitempty" protobuf:"bytes,1,opt,name=volumeprovisioner.mapi.openebs.io/storage-size"`

	// ReplicaImageOld contains the jiva replica image
	ReplicaImageOld string `json:"volumeprovisioner.mapi.openebs.io/replica-image,omitempty" protobuf:"bytes,1,opt,name=volumeprovisioner.mapi.openebs.io/replica-image"`

	// ControllerImageOld contains the jiva controller image
	ControllerImageOld string `json:"volumeprovisioner.mapi.openebs.io/controller-image,omitempty" protobuf:"bytes,1,opt,name=volumeprovisioner.mapi.openebs.io/controller-image"`

	// ReplicasOld contains the replica count
	// + optional
	ReplicasOld *int32 `json:"volumeprovisioner.mapi.openebs.io/replica-count,omitempty" protobuf:"varint,1,opt,name=volumeprovisioner.mapi.openebs.io/replica-count"`

	// ControllersOld contains the controller count
	ControllersOld *int32 `json:"volumeprovisioner.mapi.openebs.io/controller-count,omitempty" protobuf:"varint,1,opt,name=volumeprovisioner.mapi.openebs.io/controller-count"`

	// ApplicationOld contains the application label associated with volume
	ApplicationOld string `json:"volumeprovisioner.mapi.openebs.io/application,omitempty" protobuf:"bytes,1,opt,name=volumeprovisioner.mapi.openebs.io/application"`

	// ReplicaTopologyKeyDomainOld contains the domain that needs to be specified for Replica Deployment PodAntiAffinity Topology Key
	ReplicaTopologyKeyDomainOld string `json:"volumeprovisioner.mapi.openebs.io/replica-topology-key-domain,omitempty" protobuf:"bytes,1,opt,name=volumeprovisioner.mapi.openebs.io/replica-topology-key-domain"`

	// ReplicaTopologyKeyTypeOld contains the type that needs to be specified for Replica Deployment PodAntiAffinity Topology Key
	ReplicaTopologyKeyTypeOld string `json:"volumeprovisioner.mapi.openebs.io/replica-topology-key-type,omitempty" protobuf:"bytes,1,opt,name=volumeprovisioner.mapi.openebs.io/replica-topology-key-type"`
}

// K8sVolumeLabels is a typed structure that consists of
// various K8s related info. These are typically used during the
// **registration** phase of volume provisioning (using K8s as its
// orchestration provider).
type K8sVolumeLabels struct {

	// K8sStorageClassEnabled flags if fetching policy from K8s storage
	// class is enabled. A value of true implies fetching of volume
	// policies from K8s storage class must be undertaken.
	//
	// NOTE:
	//  This is an optional setting
	K8sStorageClassEnabled bool `json:"k8s.io/storage-class-enabled,omitempty" protobuf:"varint,4,opt,name=k8s.io/storage-class-enabled"`

	// K8sStorageClass contains the name of the K8s storage class
	// which will be used during volume operations. A K8s storage
	// class will typically have various volume policies set in it.
	K8sStorageClass string `json:"k8s.io/storage-class,omitempty" protobuf:"bytes,1,opt,name=k8s.io/storage-class"`

	// K8sOutCluster contains the external K8s cluster information
	// where the volume operations will be executed
	K8sOutCluster string `json:"k8s.io/out-cluster,omitempty" protobuf:"bytes,1,opt,name=k8s.io/out-cluster"`

	// K8sNamespace contains the K8s namespace where volume operations
	// will be executed
	K8sNamespace string `json:"k8s.io/namespace,omitempty" protobuf:"bytes,1,opt,name=k8s.io/namespace"`

	// K8sPersistentVolumeClaim is the pvc used to create this volume
	K8sPersistentVolumeClaim string `json:"k8s.io/pvc,omitempty" protobuf:"bytes,1,opt,name=k8s.io/pvc"`
}

// VolumeLabels is a typed structure that consists of
// various openebs volume related info. These are typically used
// during the **registration** phase of volume provisioning
type VolumeLabels struct {
	// VolumeType contains the openebs volume type
	VolumeType VolumeType `json:"openebs.io/volume-type,omitempty" protobuf:"bytes,3,opt,name=openebs.io/volume-type,casttype=VolumeType"`
}

// VolumeKey is a typed string used to represent openebs
// volume related policy keys. These keys along with their
// values will be fetched from various K8s Kinds
type VolumeKey string

const (
	// CASTemplateVK is the key to fetch name of CASTemplate custom resource
	// to create a cas volume
	CASTemplateVK VolumeKey = "cas.openebs.io/template"

	// CASTemplateForReadVK is the key to fetch name of CASTemplate custom
	// resource to read a cas volume
	CASTemplateForReadVK VolumeKey = "cas.openebs.io/read-template"

	// CASTemplateForDeleteVK is the key to fetch name of CASTemplate custom
	// resource to delete a cas volume
	CASTemplateForDeleteVK VolumeKey = "cas.openebs.io/delete-template"

	// CASTemplateForListVK is the key to fetch name of CASTemplate custom
	// resource to list cas volumes
	CASTemplateForListVK VolumeKey = "cas.openebs.io/list-template"

	// CASConfigVK is the key to fetch configurations w.r.t a CAS volume
	CASConfigVK VolumeKey = "cas.openebs.io/config"

	// CapacityVK is the key to fetch volume capacity
	CapacityVK VolumeKey = "openebs.io/capacity"

	// JivaIQNVK is the key to fetch volume iqn
	JivaIQNVK VolumeKey = "openebs.io/jiva-iqn"

	// IsK8sServiceVK is the key to fetch a boolean indicating
	// if a K8s service is required during volume provisioning
	IsK8sServiceVK VolumeKey = "openebs.io/is-k8s-service"

	// K8sTargetKindVK is the key to fetch K8s Kind value.
	// It suggests the K8s Kind object a volume is supposed to
	// be transformed to.
	K8sTargetKindVK VolumeKey = "openebs.io/k8s-target-kind"

	// JivaReplicaImageVK is the key to fetch the jiva replica image
	JivaReplicaImageVK VolumeKey = "openebs.io/jiva-replica-image"

	// JivaControllerImageVK is the key to fetch the jiva controller image
	JivaControllerImageVK VolumeKey = "openebs.io/jiva-controller-image"

	// JivaReplicasVK is the key to fetch jiva replica count
	JivaReplicasVK VolumeKey = "openebs.io/jiva-replica-count"

	// JivaControllersVK is the key to fetch jiva controller count
	JivaControllersVK VolumeKey = "openebs.io/jiva-controller-count"

	// JivaReplicaIPsVK is the key to fetch jiva replica IP Addresses
	JivaReplicaIPsVK VolumeKey = "openebs.io/jiva-replica-ips"

	// JivaControllerIPsVK is the key to fetch jiva controller IP Addresses
	JivaControllerIPsVK VolumeKey = "openebs.io/jiva-controller-ips"

	// JivaReplicaStatusVK is the key to fetch jiva replica status(-es)
	JivaReplicaStatusVK VolumeKey = "openebs.io/jiva-replica-status"

	// JivaControllerStatusVK is the key to fetch jiva controller status(-es)
	JivaControllerStatusVK VolumeKey = "openebs.io/jiva-controller-status"

	// JivaControllerClusterIPVK is the key to fetch jiva controller cluster ip
	JivaControllerClusterIPVK VolumeKey = "openebs.io/jiva-controller-cluster-ip"

	// JivaTargetPortalVK is the key to fetch jiva target portal address
	JivaTargetPortalVK VolumeKey = "openebs.io/jiva-target-portal"

	// StoragePoolVK is the key to fetch the name of storage pool
	StoragePoolVK VolumeKey = "openebs.io/storage-pool"

	// MonitorVK is the key to fetch the monitoring details
	MonitorVK VolumeKey = "openebs.io/volume-monitor"

	// VolumeTypeVK is the key to fetch the volume type
	VolumeTypeVK VolumeKey = "openebs.io/volume-type"

	// ControllerContainerStatusVK is the key to fetch the status of
	// controller containers in a pod
	ControllerContainerStatusVK VolumeKey = "openebs.io/controller-container-status"

	// ReplicaContainerStatusVK is the key to fetch the status of
	// replica containers in a pod
	ReplicaContainerStatusVK VolumeKey = "openebs.io/replica-container-status"
	// VolumeStorageClassVK is the key to fetch the storage class of openebs
	// volume
	VolumeStorageClassVK VolumeKey = "openebs.io/storage-class"
)

// VolumeValue is a typed string used to represent openebs
// volume related policy values.
type VolumeValue string

const (
	// NilVV is the value that represents is Nil i.e. null value
	// for any storage policy key. This can be a representation of
	// blank, empty or in-progress status of a storage policy.
	NilVV VolumeValue = "nil"

	// ContainerWaitingVV represents container waiting state
	ContainerWaitingVV VolumeValue = "Waiting"

	// ContainerTerminatedVV represents container terminated state
	ContainerTerminatedVV VolumeValue = "Terminated"

	// ContainerRunningVV represents container running state
	ContainerRunningVV VolumeValue = "Running"

	// ContainerNotRunningVV represents container not-running state
	ContainerNotRunningVV VolumeValue = "NotRunning"

	// VolumeRunningVV represents container running state
	VolumeRunningVV VolumeValue = "Running"

	// VolumeNotRunningVV represents container not-running state
	VolumeNotRunningVV VolumeValue = "NotRunning"
)
