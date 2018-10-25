// Package v1 - Description provided at doc.go
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

// Volume is a user's Request for a OpenEBS volume
type Volume struct {
	TypeMeta `json:",inline"`

	// Standard object's metadata
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// VolumeType holds the type of this volume
	// e.g. Jiva volume type or CStor volume type, etc
	VolumeType VolumeType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type,casttype=VolumeType"`

	// OrchProvider holds the container orchestrator that will
	// orchestrate OpenEBS volume for its provisioning & other
	// requirements
	OrchProvider OrchProvider `json:"orchestrator,omitempty" protobuf:"bytes,1,opt,name=orchestrator,casttype=OrchProvider"`

	// Namespace will hold the namespace where this Volume will exist
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`

	// Capacity will hold the capacity of this Volume
	Capacity string `json:"capacity,omitempty" protobuf:"bytes,1,opt,name=capacity"`

	// StoragePool is the name of the StoragePool where this volume
	// data will be stored. StoragePool will have the necessary storage
	// related properties.
	// +optional
	StoragePool string `json:"storagepool,omitempty" protobuf:"bytes,1,opt,name=storagepool"`

	// HostPath is directory where this volume data will be stored.
	// +optional
	HostPath string `json:"hostpath,omitempty" protobuf:"bytes,1,opt,name=hostpath"`

	// Monitor flags as well provides values for monitoring the volume
	// e.g. a value of:
	//  - `false` or empty value indicates monitoring is not required
	//  - `image: openebs/m-exporter:ci` indicates monitoring is enabled
	// and should use the provided image
	//  - `true` indicates monitoring is required & should use the defaults
	Monitor string `json:"monitor,omitempty" protobuf:"bytes,1,opt,name=monitor"`

	// VolumeClone is the specifications for vlone volume request.
	VolumeClone `json:"volumeClone,omitempty" protobuf:"bytes,4,opt,name=volumeClone"`

	// Specs contains the desired specifications the volume should have.
	// +optional
	Specs []VolumeSpec `json:"specs,omitempty" protobuf:"bytes,2,rep,name=specs"`

	// Status represents the current information/status of a volume
	Status VolumeStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// VolumeClone is the specifications for clone volume request.
type VolumeClone struct {
	// Defaults to false, true will enable the volume to be created as a clone
	Clone bool `json:"clone,omitempty"`
	// SourceVolume is snapshotted volume, required for extracting the clone
	// specific information, like storageclass, source-controller IP.
	SourceVolume string `json:"sourceVolume,omitempty"`
	// CloneIP is the source controller IP which will be used to make a sync and rebuild
	// request from the new clone replica.
	CloneIP string `json:"cloneIP,omitempty"`
	// SnapshotName is name of snapshot which is getting promoted as persistent
	// volume. Snapshot will be sync and rebuild to new replica volume.
	SnapshotName string `json:"snapshotName,omitempty"`
}

// VolumeList is a list of OpenEBS Volume items.
type VolumeList struct {
	TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// List of openebs volumes.
	Items []Volume `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// VolumeSpec provides various characteristics of a volume
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

	// Image represents the container image of this volume
	Image string `json:"image,omitempty" protobuf:"bytes,1,opt,name=image"`

	// Resources represents the actual resources of the volume
	//Capacity ResourceList
	// Source represents the location and type of a volume to mount.
	//VolumeSource
	// AccessModes contains all ways the volume can be mounted
	// +optional
	//AccessModes []VolumeAccessMode `json:"accessModes,omitempty" protobuf:"bytes,1,rep,name=accessModes,casttype=VolumeAccessMode"`
	// Name of StorageClass to which this persistent volume belongs. Empty value
	// means that this volume does not belong to any StorageClass.
	// +optional
	//StorageClassName string `json:"storageClassName,omitempty"`
}

// VolumeType defines the OpenEBS volume types that are
// supported by Maya
type VolumeType string

const (
	// JivaVolumeType represents a jiva volume
	JivaVolumeType VolumeType = "jiva"

	// CStorVolumeType represents a cstor volume
	CStorVolumeType VolumeType = "cstor"
)

// VolumeContext defines context of a volume
type VolumeContext string

const (
	// ReplicaVolumeContext represents a volume w.r.t
	// replica context
	ReplicaVolumeContext VolumeContext = "replica"

	// ControllerVolumeContext represents a volume w.r.t
	// controller context
	ControllerVolumeContext VolumeContext = "controller"
)

// OrchProvider defines the container orchestrators that
// will orchestrate the OpenEBS volumes
type OrchProvider string

const (
	// K8sOrchProvider represents Kubernetes orchestrator
	K8sOrchProvider OrchProvider = "kubernetes"
)

// K8sKind defines the various K8s Kinds that are understood
// by Maya
type K8sKind string

const (
	// DeploymentKK is a K8s Deployment Kind.
	DeploymentKK K8sKind = "deployment"
)

// VolumeSource represents the source type of the Openebs volume.
// NOTE:
//   Exactly one of its members must be set. Currently OpenEBS is the only
// member.
type VolumeSource struct {
	// OpenEBS represents an OpenEBS disk
	// +optional
	OpenEBS OpenEBS
}

// VolumeAccessMode defines different modes of volume access
type VolumeAccessMode string

const (
	// ReadWriteOnce - can be mounted read/write mode to exactly 1 host
	ReadWriteOnce VolumeAccessMode = "ReadWriteOnce"
	// ReadOnlyMany - can be mounted in read-only mode to many hosts
	ReadOnlyMany VolumeAccessMode = "ReadOnlyMany"
	// ReadWriteMany - can be mounted in read/write mode to many hosts
	ReadWriteMany VolumeAccessMode = "ReadWriteMany"
)

// VolumeStatus provides status of a volume
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

// VolumePhase defines phase of a volume
type VolumePhase string

const (
	// VolumePending - used for Volumes that are not available
	VolumePending VolumePhase = "Pending"
	// VolumeAvailable - used for Volumes that are not yet bound
	VolumeAvailable VolumePhase = "Available"
	// VolumeBound is used for Volumes that are bound
	VolumeBound VolumePhase = "Bound"
	// VolumeReleased - used for Volumes where the bound PersistentVol:syntime onumeClaim was deleted
	// released volumes must be recycled before becoming available again
	// this phase is used by the volume claim binder to signal to another process to reclaim the resource
	VolumeReleased VolumePhase = "Released"
	// VolumeFailed - used for Volumes that failed to be correctly recycled or deleted after being released from a claim
	VolumeFailed VolumePhase = "Failed"
)

// OpenEBS - Represents a Persistent Disk resource in OpenEBS.
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

// SnapshotAPISpec hsolds the config for creating asnapshot of volume
type SnapshotAPISpec struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		VolumeName string `yaml:"volumeName"`
	} `yaml:"spec"`
}

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

// VolumeSnapshotList - list of volume snapshots
type VolumeSnapshotList struct {
	TypeMeta `json:",inline"`
	Metadata ListMeta         `json:"metadata"`
	Items    []VolumeSnapshot `json:"items"`
}

// VolumeSnapshotSpec - The desired state of the volume snapshot
type VolumeSnapshotSpec struct {
	// PersistentVolumeClaimName is the name of the PVC being snapshotted
	// +optional
	VolumeName string `json:"volumeName" protobuf:"bytes,1,opt,name=persistentVolumeClaimName"`

	// SnapshotDataName binds the VolumeSnapshot object with the VolumeSnapshotData
	// +optional
	SnapshotDataName string `json:"snapshotDataName" protobuf:"bytes,2,opt,name=snapshotDataName"`
}

// VolumeSnapshotStatus defines the status of a Volume Snapshot
type VolumeSnapshotStatus struct {
	// The time the snapshot was successfully created
	// +optional
	CreationTimestamp Time `json:"creationTimestamp" protobuf:"bytes,1,opt,name=creationTimestamp"`

	// Represents the latest available observations about the volume snapshot
	Conditions []VolumeSnapshotCondition `json:"conditions" protobuf:"bytes,2,rep,name=conditions"`
}

// VolumeSnapshotConditionType - data type of volume snapshot condition
type VolumeSnapshotConditionType string

// These are valid conditions of a volume snapshot.
const (
	// VolumeSnapshotReady is added when the snapshot has been successfully created and is ready to be used.
	VolumeSnapshotConditionReady VolumeSnapshotConditionType = "Ready"
)

// VolumeSnapshotCondition describes the state of a volume snapshot at a certain point.
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
