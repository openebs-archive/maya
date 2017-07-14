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

// PersistentVolumeClaim is a user's REQUEST for and CLAIM to a persistent volume
type PersistentVolumeClaim struct {
	TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired characteristics of a volume requested by a pod author.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#persistentvolumeclaims
	// +optional
	Spec PersistentVolumeClaimSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status represents the current information/status of a persistent volume claim.
	// Read-only.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#persistentvolumeclaims
	// +optional
	Status PersistentVolumeClaimStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// PersistentVolumeClaimList is a list of PersistentVolumeClaim items.
type PersistentVolumeClaimList struct {
	TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#types-kinds
	// +optional
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// A list of persistent volume claims.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#persistentvolumeclaims
	Items []PersistentVolumeClaim `json:"items" protobuf:"bytes,2,rep,name=items"`
}

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

// PersistentVolume represents a named volume in OpenEBS that may be accessed
// by any container, VM, etc. This represents a CREATED resource.
type PersistentVolume struct {
	TypeMeta `json:",inline"`
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	//Spec defines a persistent volume owned by OpenEBS cluster
	// +optional
	Spec PersistentVolumeSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status represents the current information about persistent volume.
	// +optional
	Status PersistentVolumeStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// PersistentVolumeList is a list of PersistentVolume items.
type PersistentVolumeList struct {
	TypeMeta `json:",inline"`

	// Standard list metadata.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#types-kinds
	// +optional
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of persistent volumes.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes
	Items []PersistentVolume `json:"items" protobuf:"bytes,2,rep,name=items"`
}

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
type PersistentVolumeSpec struct {
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

type PersistentVolumeStatus struct {
	// Phase indicates if a volume is available, bound to a claim, or released by a claim
	// +optional
	Phase PersistentVolumePhase
	// A human-readable message indicating details about why the volume is in this state.
	// +optional
	Message string
	// Reason is a brief CamelCase string that describes any failure and is meant for machine parsing and tidy display in the CLI
	// +optional
	Reason string
}

type PersistentVolumePhase string

const (
	// used for PersistentVolumes that are not available
	VolumePending PersistentVolumePhase = "Pending"
	// used for PersistentVolumes that are not yet bound
	// Available volumes are held by the binder and matched to PersistentVolumeClaims
	VolumeAvailable PersistentVolumePhase = "Available"
	// used for PersistentVolumes that are bound
	VolumeBound PersistentVolumePhase = "Bound"
	// used for PersistentVolumes where the bound PersistentVolumeClaim was deleted
	// released volumes must be recycled before becoming available again
	// this phase is used by the persistent volume claim binder to signal to another process to reclaim the resource
	VolumeReleased PersistentVolumePhase = "Released"
	// used for PersistentVolumes that failed to be correctly recycled or deleted after being released from a claim
	VolumeFailed PersistentVolumePhase = "Failed"
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
