// Description provided at doc.go
//
// NOTE:
//    There are references to Kubernetes (K8s) types & links. This reflects the
// similarity of OpenEBS design principles with K8s. These may not be a
// one-to-one mapping though.
//
//    We have not imported the K8s namespaces as-is, as OpenEBS will change
// these to suit its requirements.
//
// NOTE:
//    A volume in OpenEBS has the same design as a pod in K8s. Alternatively,
// a volume in OpenEBS can be considered as a StoragePod.
package v1

// PersistentVolumeClaimSpec describes the common attributes of storage devices
// and allows a Source for provider-specific attributes
type PersistentVolumeClaimSpec struct {
	// AccessModes contains the desired access modes the volume should have.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#access-modes-1
	// +optional
	AccessModes []PersistentVolumeAccessMode `json:"accessModes,omitempty" protobuf:"bytes,1,rep,name=accessModes,casttype=PersistentVolumeAccessMode"`
	// A label query over volumes to consider for binding.
	// +optional
	Selector *LabelSelector `json:"selector,omitempty" protobuf:"bytes,4,opt,name=selector"`
	// Resources represents the minimum resources the volume should have.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#resources
	// +optional
	Resources ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,2,opt,name=resources"`
	// VolumeName is the binding reference to the PersistentVolume backing this claim.
	// +optional
	VolumeName string `json:"volumeName,omitempty" protobuf:"bytes,3,opt,name=volumeName"`
	// Name of the StorageClass required by the claim.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#class-1
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// PersistentVolumeClaimStatus is the current status of a persistent volume claim.
type PersistentVolumeClaimStatus struct {
	// Phase represents the current phase of PersistentVolumeClaim.
	// +optional
	Phase PersistentVolumeClaimPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=PersistentVolumeClaimPhase"`
	// AccessModes contains the actual access modes the volume backing the PVC has.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#access-modes-1
	// +optional
	AccessModes []PersistentVolumeAccessMode `json:"accessModes,omitempty" protobuf:"bytes,2,rep,name=accessModes,casttype=PersistentVolumeAccessMode"`
	// Represents the actual resources of the underlying volume.
	// +optional
	Capacity ResourceList `json:"capacity,omitempty" protobuf:"bytes,3,rep,name=capacity,casttype=ResourceList,castkey=ResourceName"`
}

type PersistentVolumeClaimPhase string

const (
	// used for PersistentVolumeClaims that are not yet bound
	ClaimPending PersistentVolumeClaimPhase = "Pending"
	// used for PersistentVolumeClaims that are bound
	ClaimBound PersistentVolumeClaimPhase = "Bound"
	// used for PersistentVolumeClaims that lost their underlying
	// PersistentVolume. The claim was bound to a PersistentVolume and this
	// volume does not exist any longer and all data on it was lost.
	ClaimLost PersistentVolumeClaimPhase = "Lost"
)

// PersistentVolumeSource represents the source type of the persistent volume.
//
// NOTE:
//   Exactly one of its members must be set. Currently OpenEBS is the only
// member.
type PersistentVolumeSource struct {
	// OpenEBS represents an OpenEBS disk
	// +optional
	OpenEBS OpenEBS
}

// PersistentVolumeSpec provides various characteristics of a volume
// that can be mounted, used, etc.
//
// NOTE:
//    Only one of its members may be specified. Currently OpenEBS is the only
// member. There may be other members in future.
type VolumeSpec struct {
	// The context of this volume specification.
	// Examples: "controller", "replica". Implicitly inferred to be "replica"
	// if unspecified.
	// +optional
	Context VolumeContext `json:"context,omitempty" protobuf:"bytes,1,opt,name=context,casttype=VolumeContext"`

	// Number of desired replicas. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// Resources represents the actual resources of the volume
	Capacity ResourceList
	// Source represents the location and type of a volume to mount.
	PersistentVolumeSource
	// AccessModes contains all ways the volume can be mounted
	// +optional
	AccessModes []PersistentVolumeAccessMode
	// ClaimRef is part of a bi-directional binding between PersistentVolume and PersistentVolumeClaim.
	// ClaimRef is expected to be non-nil when bound.
	// claim.VolumeName is the authoritative bind between PV and PVC.
	// When set to non-nil value, PVC.Spec.Selector of the referenced PVC is
	// ignored, i.e. labels of this PV do not need to match PVC selector.
	// +optional
	ClaimRef *ObjectReference
	// Optional: what happens to a persistent volume when released from its claim.
	// +optional
	PersistentVolumeReclaimPolicy PersistentVolumeReclaimPolicy
	// Name of StorageClass to which this persistent volume belongs. Empty value
	// means that this volume does not belong to any StorageClass.
	// +optional
	StorageClassName string
}

type VolumeContext string

const (
	// ReplicaVolumeContext represents a volume w.r.t
	// replica context
	ReplicaVolumeContext VolumeContext = "Replica"

	// ControllerVolumeContext represents a volume w.r.t
	// controller context
	ControllerVolumeContext VolumeContext = "Controller"
)

// PersistentVolumeReclaimPolicy describes a policy for end-of-life maintenance of persistent volumes
type PersistentVolumeReclaimPolicy string

const (
	// PersistentVolumeReclaimRecycle means the volume will be recycled back into the pool of unbound persistent volumes on release from its claim.
	// The volume plugin must support Recycling.
	PersistentVolumeReclaimRecycle PersistentVolumeReclaimPolicy = "Recycle"
	// PersistentVolumeReclaimDelete means the volume will be deleted from Kubernetes on release from its claim.
	// The volume plugin must support Deletion.
	PersistentVolumeReclaimDelete PersistentVolumeReclaimPolicy = "Delete"
	// PersistentVolumeReclaimRetain means the volume will be left in its current phase (Released) for manual reclamation by the administrator.
	// The default policy is Retain.
	PersistentVolumeReclaimRetain PersistentVolumeReclaimPolicy = "Retain"
)

type PersistentVolumeAccessMode string

const (
	// can be mounted read/write mode to exactly 1 host
	ReadWriteOnce PersistentVolumeAccessMode = "ReadWriteOnce"
	// can be mounted in read-only mode to many hosts
	ReadOnlyMany PersistentVolumeAccessMode = "ReadOnlyMany"
	// can be mounted in read/write mode to many hosts
	ReadWriteMany PersistentVolumeAccessMode = "ReadWriteMany"
)

type VolumeStatus struct {
	// Phase indicates if a volume is available, bound to a claim, or released by a claim
	// +optional
	Phase VolumePhase
	// A human-readable message indicating details about why the volume is in this state.
	// +optional
	Message string
	// Reason is a brief CamelCase string that describes any failure and is meant for machine parsing and tidy display in the CLI
	// +optional
	Reason string
}

type VolumePhase string

const (
	// used for PersistentVolumes that are not available
	VolumePending VolumePhase = "Pending"
	// used for PersistentVolumes that are not yet bound
	// Available volumes are held by the binder and matched to PersistentVolumeClaims
	VolumeAvailable VolumePhase = "Available"
	// used for PersistentVolumes that are bound
	VolumeBound VolumePhase = "Bound"
	// used for PersistentVolumes where the bound PersistentVol:syntime onumeClaim was deleted
	// released volumes must be recycled before becoming available again
	// this phase is used by the persistent volume claim binder to signal to another process to reclaim the resource
	VolumeReleased VolumePhase = "Released"
	// used for PersistentVolumes that failed to be correctly recycled or deleted after being released from a claim
	VolumeFailed VolumePhase = "Failed"
)

// Represents a Persistent Disk resource in OpenEBS.
//
// An OpenEBS disk must exist before mounting to a container. An OpenEBS disk
// can only be mounted as read/write once. OpenEBS volumes support
// ownership management and SELinux relabeling.
type OpenEBS struct {
	// Unique ID of the persistent disk resource in OpenEBS.
	// More info: http://kubernetes.io/docs/user-guide/volumes#awselasticblockstore
	VolumeID string `json:"volumeID" protobuf:"bytes,1,opt,name=volumeID"`
	// Filesystem type of the volume that you want to mount.
	// Tip: Ensure that the filesystem type is supported by the host operating system.
	// Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
	// More info: http://kubernetes.io/docs/user-guide/volumes#awselasticblockstore
	// TODO: how do we prevent errors in the filesystem from compromising the machine
	// +optional
	FSType string `json:"fsType,omitempty" protobuf:"bytes,2,opt,name=fsType"`
	// The partition in the volume that you want to mount.
	// If omitted, the default is to mount by volume name.
	// Examples: For volume /dev/sda1, you specify the partition as "1".
	// Similarly, the volume partition for /dev/sda is "0" (or you can leave the property empty).
	// +optional
	Partition int32 `json:"partition,omitempty" protobuf:"varint,3,opt,name=partition"`
	// Specify "true" to force and set the ReadOnly property in VolumeMounts to "true".
	// If omitted, the default is "false".
	// More info: http://kubernetes.io/docs/user-guide/volumes#awselasticblockstore
	// +optional
	ReadOnly bool `json:"readOnly,omitempty" protobuf:"varint,4,opt,name=readOnly"`
}

// ObjectFieldSelector selects an APIVersioned field of an object.
type ObjectFieldSelector struct {
	// Version of the schema the FieldPath is written in terms of, defaults to "v1".
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,1,opt,name=apiVersion"`
	// Path of the field to select in the specified API version.
	FieldPath string `json:"fieldPath" protobuf:"bytes,2,opt,name=fieldPath"`
}

// ResourceRequirements describes the compute resource requirements.
type ResourceRequirements struct {
	// Limits describes the maximum amount of compute resources allowed.
	// +optional
	Limits ResourceList
	// Requests describes the minimum amount of compute resources required.
	// If Request is omitted for a container, it defaults to Limits if that is explicitly specified,
	// otherwise to an implementation-defined value
	// +optional
	Requests ResourceList
}

// ResourceName is the name identifying various resources in a ResourceList.
type ResourceName string

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[ResourceName]Quantity

// ObjectReference contains enough information to let you inspect or modify the referred object.
type ObjectReference struct {
	// +optional
	Kind string
	// +optional
	Namespace string
	// +optional
	Name string
	// +optional
	UID string
	// +optional
	APIVersion string
	// +optional
	ResourceVersion string

	// Optional. If referring to a piece of an object instead of an entire object, this string
	// should contain information to identify the sub-object. For example, if the object
	// reference is to a container within a pod, this would take on a value like:
	// "spec.containers{name}" (where "name" refers to the name of the container that triggered
	// the event) or if no container name is specified "spec.containers[2]" (container with
	// index 2 in this pod). This syntax is chosen only to have some well-defined way of
	// referencing a part of an object.
	// TODO: this design is not final and this field is subject to change in the future.
	// +optional
	FieldPath string
}

// VolumeAPISpec holds the config for creating a Volume
type VolumeAPISpec struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name   string `yaml:"name"`
		Labels struct {
			Storage string `yaml:"volumeprovisioner.mapi.openebs.io/storage-size"`
		}
	} `yaml:"metadata"`
}

//Volume is a user's Request for a Openebs volume
type Volume struct {
	TypeMeta `json:",inline"`

	// Standard object's metadata
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specs contains the desired specifications the volume should have.
	// +optional
	Specs []VolumeSpec `json:"specs,omitempty" protobuf:"bytes,2,rep,name=specs"`

	// Status represents the current information/status of a volume
	Status VolumeStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// VolumeList is a list of PersistentVolume items.
type VolumeList struct {
	TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// List of persistent volumes.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes
	Items []Volume `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// -------------Snapshot Structs ----------

// VolumeSnapshot is volume snapshot object accessible to the user. Upon successful creation of the actual
// snapshot by the volume provider it is bound to the corresponding VolumeSnapshotData through
// the VolumeSnapshotSpec
type VolumeSnapshot struct {
	TypeMeta `json:",inline"`
	Metadata ObjectMeta `json:"metadata"`

	// Spec represents the desired state of the snapshot
	// +optional
	Spec VolumeSnapshotSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// SnapshotName represents the name of the snapshot
	SnapshotName string `json:"snapshotName" protobuf:"bytes,1,opt,name=snapshotName"`

	// Status represents the latest observer state of the snapshot
	// +optional
	Status VolumeSnapshotStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

type VolumeSnapshotList struct {
	TypeMeta `json:",inline"`
	Metadata ListMeta         `json:"metadata"`
	Items    []VolumeSnapshot `json:"items"`
}

// The desired state of the volume snapshot
type VolumeSnapshotSpec struct {
	// PersistentVolumeClaimName is the name of the PVC being snapshotted
	// +optional
	VolumeName string `json:"volumeName" protobuf:"bytes,1,opt,name=persistentVolumeClaimName"`

	// SnapshotDataName binds the VolumeSnapshot object with the VolumeSnapshotData
	// +optional
	SnapshotDataName string `json:"snapshotDataName" protobuf:"bytes,2,opt,name=snapshotDataName"`
}

type VolumeSnapshotStatus struct {
	// The time the snapshot was successfully created
	// +optional
	CreationTimestamp Time `json:"creationTimestamp" protobuf:"bytes,1,opt,name=creationTimestamp"`

	// Representes the lates available observations about the volume snapshot
	Conditions []VolumeSnapshotCondition `json:"conditions" protobuf:"bytes,2,rep,name=conditions"`
}

type VolumeSnapshotConditionType string

// These are valid conditions of a volume snapshot.
const (
	// VolumeSnapshotReady is added when the snapshot has been successfully created and is ready to be used.
	VolumeSnapshotConditionReady VolumeSnapshotConditionType = "Ready"
)

// VolumeSnapshot Condition describes the state of a volume snapshot  at a certain point.
type VolumeSnapshotCondition struct {
	// Type of replication controller condition.
	Type VolumeSnapshotConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=VolumeSnapshotConditionType"`
	// Status of the condition, one of True, False, Unknown.
	//Status core_v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime Time `json:"lastTransitionTime" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message" protobuf:"bytes,5,opt,name=message"`
}
