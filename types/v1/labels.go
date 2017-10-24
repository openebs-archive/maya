package v1

// NomadEnvironmentVariable is a typed label that defines environment variables
// that are understood by Nomad
type NomadEnvironmentVariable string

const (
	// NomadAddressEnvKey is the environment variable that determines the
	// Nomad server address where the Job request can be directed to.
	NomadAddressEnvKey NomadEnvironmentVariable = "NOMAD_ADDR"
	// NomadRegionEnvKey is the environment variable that determines the Nomad region
	// where the Job request can be directed to.
	NomadRegionEnvKey NomadEnvironmentVariable = "NOMAD_REGION"
)

// EnvironmentVariableLabel is a typed label that defines environment variable
// labels that are passed as request options during provisioning.
type EnvironmentVariableLabel string

const (
	// EnvVariableContextLbl is the label that can be optionally set as one of the
	// request option during VSM provisioning operations. Its value is used
	// to set the context (/ prefix) against the environment variables for that
	// particular request.
	EnvVariableContextLbl EnvironmentVariableLabel = "env.mapi.openebs.io/env-var-ctx"
)

// EnvironmentVariableDefaults is a typed label that defines the environment variable
// defaults
type EnvironmentVariableDefaults string

const (
	// Default value for environment variable context
	EnvVariableContextDef EnvironmentVariableDefaults = "DEFAULT"
)

// EnvironmentVariableKey is a typed label that define the environment variables
type EnvironmentVariableKey string

const (
	// PVPProfileNameEnvVarKey is the environment variable key for persistent
	// volume provisioner's profile name
	//
	// Usage:
	// <CTX>_PVP_PROFILE_NAME = <some value>
	PVPProfileNameEnvVarKey EnvironmentVariableKey = "_PVP_PROFILE_NAME"
	// PVPNameEnvVarKey is the environment variable key for persistent volume
	// provisioner's name
	//
	// Usage:
	// <CTX>_PVP_NAME = <some value>
	PVPNameEnvVarKey EnvironmentVariableKey = "_PVP_NAME"
	// PVPControllerImageEnvVarKey is the environment variable key for persistent
	// volume provisioner's controller image
	//
	// Usage:
	// <CTX>_CONTROLLER_IMAGE = <some value>
	PVPControllerImageEnvVarKey EnvironmentVariableKey = "_CONTROLLER_IMAGE"
	// PVPPersistentPathEnvVarKey is the environment variable key for persistent
	// volume provisioner's replica persistent path
	//
	// Usage:
	// <CTX>_PERSISTENT_PATH = <some value>
	PVPPersistentPathEnvVarKey EnvironmentVariableKey = "_PERSISTENT_PATH"
	// PVPStorageSizeEnvVarKey is the environment variable key for persistent
	// volume provisioner's replica size
	//
	// Usage:
	// <CTX>_STORAGE_SIZE = <some value>
	PVPStorageSizeEnvVarKey EnvironmentVariableKey = "_STORAGE_SIZE"
	// PVPReplicaCountEnvVarKey is the environment variable key for persistent
	// volume provisioner's replica count
	//
	// Usage:
	// <CTX>_REPLICA_COUNT = <some value>
	PVPReplicaCountEnvVarKey EnvironmentVariableKey = "_REPLICA_COUNT"
	// PVPReplicaImageEnvVarKey is the environment variable key for persistent
	// volume provisioner's replica image
	//
	// Usage:
	// <CTX>_REPLICA_IMAGE = <some value>
	PVPReplicaImageEnvVarKey EnvironmentVariableKey = "_REPLICA_IMAGE"
	// PVPControllerCountEnvVarKey is the environment variable key for persistent
	// volume provisioner's controller count
	//
	// Usage:
	// <CTX>_CONTROLLER_COUNT = <some value>
	PVPControllerCountEnvVarKey EnvironmentVariableKey = "_CONTROLLER_COUNT"
	// PVPReplicaTopologyKeyEnvVarKey is the environment variable key for persistent
	// volume provisioner's replica topology key
	//
	// Usage:
	// <CTX>_REPLICA_TOPOLOGY_KEY = <some value>
	PVPReplicaTopologyKeyEnvVarKey EnvironmentVariableKey = "_REPLICA_TOPOLOGY_KEY"

	// PVPControllerNodeTaintTolerationEnvVarKey is the environment variable key
	// for persistent volume provisioner's node taint toleration
	//
	// Usage:
	// <CTX>_CONTROLLER_NODE_TAINT_TOLERATION = <some value>
	PVPControllerNodeTaintTolerationEnvVarKey EnvironmentVariableKey = "_CONTROLLER_NODE_TAINT_TOLERATION"

	// PVPReplicaNodeTaintTolerationEnvVarKey is the environment variable key for
	// persistent volume provisioner's node taint toleration
	//
	// Usage:
	// <CTX>__REPLICA_NODE_TAINT_TOLERATION = <some value>
	PVPReplicaNodeTaintTolerationEnvVarKey EnvironmentVariableKey = "_REPLICA_NODE_TAINT_TOLERATION"

	// OrchestratorNameEnvVarKey is the environment variable key for
	// orchestration provider's name
	//
	// Usage:
	// <CTX>_ORCHESTRATOR_NAME = <some value>
	OrchestratorNameEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_NAME"
	// OrchestratorRegionEnvVarKey is the environment variable key for orchestration
	// provider's region
	//
	// Usage:
	// <CTX>_ORCHESTRATOR_REGION = <some value>
	OrchestratorRegionEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_REGION"
	// OrchestratorDCEnvVarKey is the environment variable key for orchestration
	// provider's datacenter
	//
	// Usage:
	// <CTX>_ORCHESTRATOR_DC = <some value>
	OrchestratorDCEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_DC"
	// OrchestratorAddressEnvVarKey is the environment variable key for orchestration
	// provider's address
	//
	// Usage:
	// <CTX>_<REGION>_<DC>_ORCHESTRATOR_ADDR = 10.20.1.1
	OrchestratorAddressEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_ADDR"
	// OrchestratorCNTypeEnvVarKey is the environment variable key for orchestration
	// provider's network type
	//
	// Usage:
	// <CTX>_ORCHESTRATOR_CN_TYPE = <some value>
	OrchestratorCNTypeEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_CN_TYPE"
	// OrchestratorCNInterfaceEnvVarKey is the environment variable key for orchestration
	// provider's network interface
	//
	// Usage:
	// <CTX>_ORCHESTRATOR_CN_INTERFACE = <some value>
	OrchestratorCNInterfaceEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_CN_INTERFACE"
	// OrchestratorCNAddrEnvVarKey is the environment variable key for orchestration
	// provider's network address
	//
	// Usage:
	// <CTX>_ORCHESTRATOR_CN_ADDRESS = <some value>
	OrchestratorCNAddrEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_CN_ADDRESS"
	// OrchestratorNSEnvVarKey is the environment variable key for orchestration
	// provider's namespace
	//
	// Usage:
	// <CTX>_ORCHESTRATOR_NS = <some value>
	OrchestratorNSEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_NS"
	// OrchestratorInClusterEnvVarKey is the environment variable key for orchestration
	// provider's in-cluster flag
	//
	// Usage:
	// <CTX>_ORCHESTRATOR_IN_CLUSTER = <some value>
	OrchestratorInClusterEnvVarKey EnvironmentVariableKey = "_ORCHESTRATOR_IN_CLUSTER"
)

// OrchProviderProfileLabel is a typed label to determine orchestration provider
// profile's values.
type OrchProviderProfileLabel string

const (
	// Label / Tag for an orchestrator profile name
	OrchProfileNameLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/profile-name"
	// Label / Tag for an orchestrator region
	OrchRegionLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/region"
	// Label / Tag for an orchestrator datacenter
	OrchDCLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/dc"
	// OrchAddrLbl is the Label / Tag for an orchestrator address
	OrchAddrLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/address"
	// Label / Tag for an orchestrator namespace
	OrchNSLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/ns"
	// OrchInClusterLbl is the label for setting the in cluster flag. This is used
	// during provisioning operations. It sets if the provisioning is meant to be
	// within cluster or outside the cluster.
	OrchInClusterLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/in-cluster"
	// OrchCNTypeLbl is the Label / Tag for an orchestrator's networking type
	OrchCNTypeLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/cn-type"
	// OrchCNNetworkAddrLbl is the Label / Tag for an orchestrator's network address
	// in CIDR notation
	OrchCNNetworkAddrLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/cn-addr"
	// OrchCNSubnetLbl is the Label / Tag for an orchestrator's network subnet
	OrchCNSubnetLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/cn-subnet"
	// OrchCNInterfaceLbl is the Label / Tag for an orchestrator's network interface
	OrchCNInterfaceLbl OrchProviderProfileLabel = "orchprovider.mapi.openebs.io/cn-interface"
)

// OrchProviderDefaults is a typed label to provide default values w.r.t
// orchestration provider properties.
type OrchProviderDefaults string

const (
	// Default value for orchestrator's network address
	// NOTE: Should be in valid CIDR notation
	OrchNetworkAddrDef OrchProviderDefaults = "172.28.128.1/24"
	// Default value for orchestrator's in-cluster flag
	OrchInClusterDef OrchProviderDefaults = "true"
	// Default value for orchestrator namespace
	OrchNSDef OrchProviderDefaults = "default"
	// OrchRegionDef is the default value of orchestrator region
	OrchRegionDef OrchProviderDefaults = "global"
	// OrchDCDef is the default value of orchestrator datacenter
	OrchDCDef OrchProviderDefaults = "dc1"
	// OrchAddressDef is the default value of orchestrator address
	OrchAddressDef OrchProviderDefaults = "127.0.0.1"
	// OrchCNTypeDef is the default value of orchestrator network type
	OrchCNTypeDef OrchProviderDefaults = "host"
	// OrchCNInterfaceDef is the default value of orchestrator network interface
	OrchCNInterfaceDef OrchProviderDefaults = "enp0s8"
)

// VolumeProvisionerProfileLabel is a typed label to determine volume provisioner
// profile values.
type VolumeProvisionerProfileLabel string

const (
	// Label / Tag for a persistent volume provisioner profile's name
	PVPProfileNameLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/profile-name"
	// Label / Tag for a persistent volume provisioner's replica support
	PVPReqReplicaLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/req-replica"
	// Label / Tag for a persistent volume provisioner's networking support
	PVPReqNetworkingLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/req-networking"

	// Deprecate
	// Label / Tag for a persistent volume provisioner's replica count
	PVPReplicaCountLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/replica-count"
	// Label / Tag for a persistent volume provisioner's persistent path count
	PVPPersistentPathCountLbl VolumeProvisionerProfileLabel = PVPReplicaCountLbl
	// Label / Tag for a persistent volume provisioner's storage size
	PVPStorageSizeLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/storage-size"
	// Label / Tag for a persistent volume provisioner's replica IPs
	PVPReplicaIPsLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/replica-ips"
	// Label / Tag for a persistent volume provisioner's replica image
	PVPReplicaImageLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/replica-image"
	// Label / Tag for a persistent volume provisioner's controller count
	PVPControllerCountLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/controller-count"
	// Label / Tag for a persistent volume provisioner's controller image
	PVPControllerImageLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/controller-image"
	// Label / Tag for a persistent volume provisioner's controller IPs
	PVPControllerIPsLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/controller-ips"
	// Label / Tag for a persistent volume provisioner's persistent path
	PVPPersistentPathLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/persistent-path"
	// Label / Tag for a persistent volume provisioner's controller node taint toleration
	PVPControllerNodeTaintTolerationLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/controller-node-taint-toleration"
	// Label / Tag for a persistent volume provisioner's replica node taint toleration
	PVPReplicaNodeTaintTolerationLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/replica-node-taint-toleration"

	// PVPReplicaTopologyKeyLbl is the label for a persistent volume provisioner's
	// VSM replica topology key
	PVPReplicaTopologyKeyLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/replica-topology-key"

	// PVPNodeAffinityExpressionsLbl is the label to determine the node affinity
	// of the replica(s).
	//
	// NOTE:
	//    1. These are comma separated key value pairs, where each
	// key & value is separated by an operator e.g. In, NotIn, Exists, DoesNotExist
	//
	//    2. The key & value should have been labeled against a node or group of
	// nodes belonging to the K8s cluster
	//
	//    3. The replica count should match the number of of pairs provided
	//
	// Usage:
	// For OpenEBS volume with 2 replicas:
	// volumeprovisioner.mapi.openebs.io/node-affinity-expressions=
	//    "<replica-identifier>=kubernetes.io/hostname:In:node1,
	//     <another-replica-identifier>=kubernetes.io/hostname:In:node2"
	//
	// Usage:
	// For OpenEBS volume with 3 replicas:
	// volumeprovisioner.mapi.openebs.io/node-affinity-expressions=
	//    "<replica-identifier>=kubernetes.io/hostname:In:node1,
	//     <another-replica-identifier>=kubernetes.io/hostname:In:node2,
	//     <yet-another-replica-identifier>=kubernetes.io/hostname:In:node3"
	//
	// Usage:
	// For OpenEBS volume with 3 replicas:
	// volumeprovisioner.mapi.openebs.io/node-affinity-expressions=
	//    "<replica-identifier>=volumeprovisioner.mapi.openebs.io/replica-zone-1-ssd-1:In:zone-1-ssd-1,
	//     <another-replica-identifier>=openebs.io/replica-zone-1-ssd-2:In:zone-1-ssd-2,
	//     <yet-another-replica-identifier>=openebs.io/replica-zone-2-ssd-1:In:zone-2-ssd-1"
	//
	// Usage:
	// For OpenEBS volume with 3 replicas:
	// volumeprovisioner.mapi.openebs.io/node-affinity-expressions=
	//    "<replica-identifier>=openebs.io/replica-zone-1-grp-1:In:zone-1-grp-1,
	//     <another-replica-identifier>=openebs.io/replica-zone-1-grp-2:In:zone-1-grp-2,
	//     <yet-another-replica-identifier>=openebs.io/replica-zone-2-grp-1:In:zone-2-grp-1"
	//PVPNodeAffinityExpressionsLbl VolumeProvisionerProfileLabel = "volumeprovisioner.mapi.openebs.io/node-affinity-expressions"

	// PVPNodeSelectorKeyLbl is the label to build the node affinity
	// of the replica based on the key & the replica identifier
	//
	// NOTE:
	//  PVPNodeAffinityExpressionsLbl is used here as key is a part of the expressions
	//PVPNodeSelectorKeyLbl VolumeProvisionerProfileLabel = PVPNodeAffinityExpressionsLbl

	// PVPNodeSelectorOpLbl is the label to build the node affinity
	// of the replica based on the operator & the replica identifier
	//
	// NOTE:
	//  PVPNodeAffinityExpressionsLbl is used here as operator is a part of the expressions
	//PVPNodeSelectorOpLbl VolumeProvisionerProfileLabel = PVPNodeAffinityExpressionsLbl

	// PVPNodeSelectorValueLbl is the label to build the node affinity
	// of the replica based on the operator & the replica identifier
	//
	// NOTE:
	//  PVPNodeAffinityExpressionsLbl is used here as value is a part of the expressions
	//PVPNodeSelectorValueLbl VolumeProvisionerProfileLabel = PVPNodeAffinityExpressionsLbl
)

// Deprecate
type MayaAPIServiceOutputLabel string

// Deprecate all these constants
const (
	ReplicaStatusAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/replica-status"

	ControllerStatusAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/controller-status"

	TargetPortalsAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/targetportals"

	ClusterIPsAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/cluster-ips"

	ReplicaIPsAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/replica-ips"

	ControllerIPsAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/controller-ips"

	IQNAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/iqn"

	VolumeSizeAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/volume-size"

	// Deprecate
	ReplicaCountAPILbl MayaAPIServiceOutputLabel = "vsm.openebs.io/replica-count"
)

// VolumeProvsionerDefaults is a typed label to provide default values w.r.t
// volume provisioner properties.
type VolumeProvisionerDefaults string

const (
	// Default value for persistent volume provisioner's controller count
	PVPControllerCountDef VolumeProvisionerDefaults = "1"
	// Default value for persistent volume provisioner's replica count
	PVPReplicaCountDef VolumeProvisionerDefaults = "2"
	// Default value for persistent volume provisioner's persistent path count
	// This should be equal to persistent volume provisioner's replica count
	PVPPersistentPathCountDef VolumeProvisionerDefaults = PVPReplicaCountDef
	// Default value for persistent volume provisioner's controller image
	PVPControllerImageDef VolumeProvisionerDefaults = "openebs/jiva:latest"
	// Default value for persistent volume provisioner's support for replica
	PVPReqReplicaDef VolumeProvisionerDefaults = "true"
	// Default value for persistent volume provisioner's replica image
	PVPReplicaImageDef VolumeProvisionerDefaults = "openebs/jiva:latest"
	// Default value for persistent volume provisioner's networking support
	PVPReqNetworkingDef VolumeProvisionerDefaults = "false"
	// PVPPersistentPathDef is the default value for persistent volume provisioner's
	// replica persistent path
	PVPPersistentPathDef VolumeProvisionerDefaults = "/var/openebs"
	// PVPStorageSizeDef is the default value for persistent volume provisioner's
	// replica size
	PVPStorageSizeDef VolumeProvisionerDefaults = "1G"

	// PVPNodeSelectorKeyDef is the default value for volume replica's node selector
	// key
	//PVPNodeSelectorKeyDef VolumeProvisionerDefaults = "kubernetes.io/hostname"

	// PVPNodeSelectorOpDef is the default value for volume replica's node selector
	// operator
	//PVPNodeSelectorOpDef VolumeProvisionerDefaults = "In"
)

// NameLabel type will be used to identify various maya api service components
// via this typed label
type NameLabel string

const (
	// Label / Tag for an orchestrator name
	OrchestratorNameLbl NameLabel = "orchprovider.mapi.openebs.io/name"
	// Label / Tag for a persistent volume provisioner name
	VolumeProvisionerNameLbl NameLabel = "volumeprovisioner.mapi.openebs.io/name"
)

// OrchestratorRegistry type will be used to register various maya api service
// orchestrators.
type OrchProviderRegistry string

const (
	// K8sOrchestrator states Kubernetes as orchestration provider plugin.
	// This is used for registering Kubernetes as an orchestration provider in maya
	// api server.
	K8sOrchestrator OrchProviderRegistry = "kubernetes"
	// NomadOrchestrator states Nomad as orchestration provider plugin.
	// This is used for registering Nomad as an orchestration provider in maya api
	// server.
	NomadOrchestrator OrchProviderRegistry = "nomad"
	// DefaultOrchestrator provides the default orchestration provider
	DefaultOrchestrator = K8sOrchestrator
)

// VolumeProvisionerRegistry type will be used to register various maya api
// service volume provisioners.
type VolumeProvisionerRegistry string

const (
	// JivaVolumeProvisioner states Jiva as persistent volume provisioner plugin.
	// This is used for registering Jiva as a volume provisioner in maya api server.
	JivaVolumeProvisioner VolumeProvisionerRegistry = "jiva"
	// DefaultVolumeProvisioner provides the default persistent volume provisioner
	// plugin.
	DefaultVolumeProvisioner VolumeProvisionerRegistry = JivaVolumeProvisioner
)

// OrchProviderProfileRegistry type will be used to register various maya api
// service orchestrator profiles
type OrchProviderProfileRegistry string

const (
	// This is the name of PVC as orchestration provider profile
	// This is used for labelling PVC as a orchestration provider profile
	PVCOrchestratorProfile OrchProviderProfileRegistry = "pvc"
)

// VolumeProvisionerProfileRegistry type will be used to register various maya api service
// persistent volume provisioner profiles
type VolumeProvisionerProfileRegistry string

const (
	// This is the name of PVC as persistent volume provisioner profile
	// This is used for labelling PVC as a persistent volume provisioner profile
	PVCProvisionerProfile VolumeProvisionerProfileRegistry = "pvc"
)

type GenericAnnotations string

const (
	// VolumeProvisionerSelectorKey is used to filter VSMs
	VolumeProvisionerSelectorKey GenericAnnotations = "openebs/volume-provisioner"

	// ControllerSelectorKey is used to filter controllers
	ControllerSelectorKey GenericAnnotations = "openebs/controller"

	// ControllerSelectorKeyEquals is used to filter controller when
	// selector logic is used
	ControllerSelectorKeyEquals GenericAnnotations = ControllerSelectorKey + "="

	// ReplicaCountSelectorKey is used to filter replicas
	//ReplicaCountSelectorKey GenericAnnotations = "openebs/replica-count"

	// ReplicaSelectorKey is used to filter replicas
	ReplicaSelectorKey GenericAnnotations = "openebs/replica"

	// ReplicaSelectorKeyEquals is used to filter replica when
	// selector logic is used
	ReplicaSelectorKeyEquals GenericAnnotations = ReplicaSelectorKey + "="

	// ServiceSelectorKey is used to filter services
	ServiceSelectorKey GenericAnnotations = "openebs/controller-service"

	// ServiceSelectorKeyEquals is used to filter services when selector logic is
	// used
	ServiceSelectorKeyEquals GenericAnnotations = ServiceSelectorKey + "="

	// SelectorEquals is used to filter
	SelectorEquals GenericAnnotations = "="

	// VSMSelectorKey is used to filter vsm
	VSMSelectorKey GenericAnnotations = "vsm"

	// VSMSelectorKeyEquals is used to filter vsm when selector logic is used
	VSMSelectorKeyEquals GenericAnnotations = VSMSelectorKey + "="

	// ControllerSuffix is used as a suffix for controller related names
	ControllerSuffix GenericAnnotations = "-ctrl"

	// ReplicaSuffix is used as a suffix for replica related names
	ReplicaSuffix GenericAnnotations = "-rep"

	// ServiceSuffix is used as a suffix for service related names
	ServiceSuffix GenericAnnotations = "-svc"

	// ContainerSuffix is used as a suffix for container related names
	ContainerSuffix GenericAnnotations = "-con"
)

// TODO
// Move these to jiva folder
//
// JivaAnnotations will be used to provide filtering options like
// named-labels, named-suffix, named-prefix, constants, etc.
//
// NOTE:
//    These value(s) are generally used / remembered by the consumers of
// maya api service
type JivaAnnotations string

// TODO
// Rename these const s.t. they start with Jiva as Key Word
const (
	// JivaVolumeProvisionerSelectorValue is used to filter jiva based objects
	JivaVolumeProvisionerSelectorValue JivaAnnotations = "jiva"

	// JivaControllerSelectorValue is used to filter jiva controller objects
	JivaControllerSelectorValue JivaAnnotations = "jiva-controller"

	// JivaReplicaSelectorValue is used to filter jiva replica objects
	JivaReplicaSelectorValue JivaAnnotations = "jiva-replica"

	// JivaServiceSelectorValue is used to filter jiva service objects
	JivaServiceSelectorValue JivaAnnotations = "jiva-controller-service"

	// PortNameISCSI is the name given to iscsi ports
	PortNameISCSI JivaAnnotations = "iscsi"

	// PortNameAPI is the name given to api ports
	PortNameAPI JivaAnnotations = "api"

	// JivaCtrlIPHolder is used as a placeholder for persistent volume controller's
	// IP address
	//
	// NOTE:
	//    This is replaced at runtime
	JivaClusterIPHolder JivaAnnotations = "__CLUSTER_IP__"

	// JivaStorageSizeHolder is used as a placeholder for persistent volume's
	// storage capacity
	//
	// NOTE:
	//    This is replaced at runtime
	JivaStorageSizeHolder JivaAnnotations = "__STOR_SIZE__"

	//
	JivaVolumeNameHolder JivaAnnotations = "__VOLUME_NAME__"
)

// JivaDefaults is a typed label to provide DEFAULT values to Jiva based
// persistent volume properties
type JivaDefaults string

const (
	// JivaControllerFrontendDef is used to provide default frontend for jiva
	// persistent volume controller
	JivaControllerFrontendDef JivaDefaults = "gotgt"

	// Jiva's iSCSI Qualified IQN value.
	JivaIqnFormatPrefix JivaDefaults = "iqn.2016-09.com.openebs.jiva"

	// JivaISCSIPortDef is used to provide default iscsi port value for jiva
	// based persistent volumes
	JivaISCSIPortDef JivaDefaults = "3260"

	// JivaPersistentMountPathDef is the default mount path used by jiva based
	// persistent volumes
	JivaPersistentMountPathDef JivaDefaults = "/openebs"

	// JivaPersistentMountNameDef is the default mount path name used by jiva based
	// persistent volumes
	JivaPersistentMountNameDef JivaDefaults = "openebs"

	// JivaAPIPortDef is used to provide management port for persistent volume
	// storage
	JivaAPIPortDef JivaDefaults = "9501"

	// JivaReplicaPortOneDef is used to provide port for jiva based persistent
	// volume replica
	JivaReplicaPortOneDef JivaDefaults = "9502"

	// JivaReplicaPortTwoDef is used to provide port for jiva based persistent
	// volume replica
	JivaReplicaPortTwoDef JivaDefaults = "9503"

	// JivaReplicaPortThreeDef is used to provide port for jiva based persistent
	// volume replica
	JivaReplicaPortThreeDef JivaDefaults = "9504"

	// JivaBackEndIPPrefixLbl is used to provide the label for VSM replica IP on
	// Nomad
	JivaBackEndIPPrefixLbl JivaDefaults = "JIVA_REP_IP_"
)

// These will be used to provide array based constants that are
// related to jiva volume provisioner
var (
	// JivaCtrlCmd is the command used to start jiva controller
	JivaCtrlCmd = []string{"launch"}

	// JivaCtrlArgs is the set of arguments provided to JivaCtrlCmd
	//JivaCtrlArgs = []string{"controller", "--frontend", string(JivaControllerFrontendDef), string(JivaVolumeNameDef)}
	JivaCtrlArgs = []string{"controller", "--frontend", string(JivaControllerFrontendDef), "--clusterIP", string(JivaClusterIPHolder), string(JivaVolumeNameHolder)}

	// JivaReplicaCmd is the command used to start jiva replica
	JivaReplicaCmd = []string{"launch"}

	// JivaReplicaArgs is the set of arguments provided to JivaReplicaCmd
	JivaReplicaArgs = []string{"replica", "--frontendIP", string(JivaClusterIPHolder), "--size", string(JivaStorageSizeHolder), string(JivaPersistentMountPathDef)}
)

// TODO
// Move these to k8s folder
//
// K8sAnnotations will be used to provide string based constants that are
// related to kubernetes as orchestration provider
type K8sAnnotations string

const (
	// K8sKindPod is used to state the k8s Pod
	K8sKindPod K8sAnnotations = "Pod"
	// K8sKindDeployment is used to state the k8s Deployment
	K8sKindDeployment K8sAnnotations = "Deployment"
	// K8sKindService is used to state the k8s Service
	K8sKindService K8sAnnotations = "Service"
	// K8sServiceVersion is used to state the k8s Service version
	K8sServiceVersion K8sAnnotations = "v1"
	// K8sPodVersion is used to state the k8s Pod version
	K8sPodVersion K8sAnnotations = "v1"
	// K8sDeploymentVersion is used to state the k8s Deployment version
	K8sDeploymentVersion K8sAnnotations = "extensions/v1beta1"
	// K8sHostnameTopologyKey is used to specify the hostname as topology key
	K8sHostnameTopologyKey K8sAnnotations = "kubernetes.io/hostname"
)
