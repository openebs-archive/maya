package v1

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/nethelper"
)

// GetPVPNodeSelectorKey gets the not nil value of volume provisioner volume
// replica's node selector key
//func GetPVPNodeSelectorKey(repIdentifier string, profileMap map[string]string) string {
//	return PVPNodeSelectorKey(repIdentifier, profileMap)
//}

// PVPNodeSelectorKey will fetch the value specified against volume provisioner
// volume replica's node selector key if available otherwise will return blank.
//func PVPNodeSelectorKey(repIdentifier string, profileMap map[string]string) string {
//	val := ""
//	if profileMap != nil {
//		val = strings.TrimSpace(profileMap[string(PVPNodeSelectorKeyLbl)])
//	}

//	if val == "" {
//		return val
//	}

// TODO
// Make use of repIdentifier to extract the specific key

//	return val
//}

// DefaultPVPNodeSelectorKey will fetch the default value for volume provisioner
// volume replica's node selector key
//func DefaultPVPNodeSelectorKey() string {
//	return string(PVPNodeSelectorKeyDef)
//}

// GetPVPNodeSelectorOp gets the not nil value of volume provisioner volume
// replica's node selector operator
//func GetPVPNodeSelectorOp(repIdentifier string, profileMap map[string]string) string {
//	return PVPNodeSelectorOp(repIdentifier, profileMap)
//}

// PVPNodeSelectorOp will fetch the value specified against volume provisioner
// volume replica's node selector operator if available otherwise will return blank.
//func PVPNodeSelectorOp(repIdentifier string, profileMap map[string]string) string {
//	val := ""
//	if profileMap != nil {
//		val = strings.TrimSpace(profileMap[string(PVPNodeSelectorOpLbl)])
//	}

//	if val == "" {
//		return val
//	}

// TODO
// Make use of replica name to extract the specific operator

//	return val
//}

// DefaultPVPNodeSelectorOp will fetch the default value for volume provisioner
// volume replica's node selector operator
//func DefaultPVPNodeSelectorOp() string {
//	return string(PVPNodeSelectorOpDef)
//}

// GetPVPNodeSelectorValue gets the not nil value of volume provisioner volume
// replica's node selector value
//func GetPVPNodeSelectorValue(repIdentifier string, profileMap map[string]string) string {
//	return PVPNodeSelectorValue(repIdentifier, profileMap)
//}

// PVPNodeSelectorValue will fetch the value specified against volume provisioner
// volume replica's node selector value if available otherwise will return blank.
//func PVPNodeSelectorValue(repIdentifier string, profileMap map[string]string) string {
//	val := ""
//	if profileMap != nil {
//		val = strings.TrimSpace(profileMap[string(PVPNodeSelectorValueLbl)])
//	}

//	if val == "" {
//		return val
//	}

// TODO
// Make use of repIdentifier to extract the specific operator

//	return val
//}

// GetPVPReplicaTopologyKey gets the not nil value of PVP's VSM Replica topology
// key
func GetPVPReplicaTopologyKey(profileMap map[string]string) string {
	val := PVPReplicaTopologyKey(profileMap)
	if val == "" {
		val = DefaultPVPReplicaTopologyKey()
	}

	return val
}

// PVPReplicaTopologyKey will fetch the value specified against PVP's VSM
// Replica topology key if available otherwise will return blank.
func PVPReplicaTopologyKey(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPReplicaTopologyKeyLbl)])
	}

	if val != "" {
		return val
	}

	return OSGetEnv(string(PVPReplicaTopologyKeyEnvVarKey), profileMap)
}

// DefaultPVPReplicaTopologyKey will fetch the default value for PVP's VSM
// Replica topology key
func DefaultPVPReplicaTopologyKey() string {
	// TODO
	// else get based on the replica count & current replica index
	// e.g.
	// if replica count = 2 then use K8sHostnameTopologyKey for 2 replicas
	// if replica count = 1 then use K8sHostnameTopologyKey for the replica
	// if replica count = 3 then use K8sHostnameTopologyKey for 2 replicas & use
	// failure-domain.beta.kubernetes.io/zone for 1 replica if zone is really available
	// if replica count > 3 then use K8sHostnameTopologyKey for n-1 replicas & use
	// failure-domain.beta.kubernetes.io/zone for 1 replica if zone is really available
	return string(K8sHostnameTopologyKey)
}

// GetPVPControllerCountInt gets the not nil value of PVP's VSM Controller count
// in int
func GetPVPControllerCountInt(profileMap map[string]string) (int, error) {
	return strconv.Atoi(GetPVPControllerCount(profileMap))
}

// GetPVPControllerCount gets the not nil value of PVP's VSM Controller count
func GetPVPControllerCount(profileMap map[string]string) string {
	val := PVPControllerCount(profileMap)
	if val == "" {
		val = DefaultPVPControllerCount()
	}

	return val
}

// PVPControllerCount will fetch the value specified against PVP's VSM
// Controller count if available otherwise will return blank.
func PVPControllerCount(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPControllerCountLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(PVPControllerCountEnvVarKey), profileMap)
}

// DefaultPVPControllerCount will fetch the default value for PVP's VSM
// Controller count
func DefaultPVPControllerCount() string {
	return string(PVPControllerCountDef)
}

// VSMName will fetch the value specified against persistent volume
// VSM name if available otherwise will return blank.
//
// NOTE:
//    This utility function does not validate & just returns if not capable of
// performing
func VSMName(pvcName string) string {
	// Name of PVC is the name of VSM
	return pvcName
}

// OrchProfileName will fetch the value specified against persistent volume's
// orchestrator profile name if available otherwise will return blank.
//
// NOTE:
//    This utility function does not validate & just returns if not capable of
// performing
func OrchProfileName(profileMap map[string]string) string {
	if profileMap == nil {
		return ""
	}

	// Extract orchestrator profile name
	return profileMap[string(OrchProfileNameLbl)]
}

// VolumeProvisionerProfileName will fetch the name of volume provisioner
// profile if available otherwise will return blank.
//
// NOTE:
//    This utility function makes the best attempt to get the value from
// provided profileMap or from the machine's environment variable
func VolumeProvisionerProfileName(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPProfileNameLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(PVPProfileNameEnvVarKey), profileMap)
}

// VolumeProvisionerName will fetch the name of volume provisioner
// if available otherwise will return blank.
//
// NOTE:
//    This utility function makes the best attempt to get the value from
// provided profileMap or from the machine's environment variable
func VolumeProvisionerName(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(VolumeProvisionerNameLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(PVPNameEnvVarKey), profileMap)
}

// DefaultVolumeProvisionerName gets the default name of persistent volume
// provisioner plugin used to cater the provisioning requests to maya api
// service
//
// NOTE:
//    This returns the hard coded default set in this pkg
func DefaultVolumeProvisionerName() VolumeProvisionerRegistry {
	return DefaultVolumeProvisioner
}

// GetOrchestratorRegion gets the not nil name of orchestrator
func GetOrchestratorName(profileMap map[string]string) OrchProviderRegistry {
	val := OrchestratorName(profileMap)
	if val == "" {
		val = DefaultOrchestratorName()
	}

	return OrchProviderRegistry(val)
}

// OrchestratorName will fetch the value specified against persistent
// volume's orchestrator name if available otherwise will return blank.
//
// NOTE:
//    This utility function does not validate & just returns if not capable of
// performing
func OrchestratorName(profileMap map[string]string) string {

	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchestratorNameLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(OrchestratorNameEnvVarKey), profileMap)
}

// DefaultOrchestratorName gets the default name of orchestration provider
//
// NOTE:
//    This utility function does not validate & just returns if not capable of
// performing
func DefaultOrchestratorName() string {
	return string(DefaultOrchestrator)
}

// GetOrchestratorAddress fetches the not nil orchestrator address
func GetOrchestratorAddress(profileMap map[string]string) string {
	val := OrchestratorAddress(profileMap)
	if val == "" {
		val = DefaultOrchestratorAddress()
	}

	return val
}

// OrchestratorAddress fetches the value specified against persistent
// volume's orchestrator address if available otherwise will return blank.
//
// NOTE:
//  A region is composed of one or more [datacenter : address]
func OrchestratorAddress(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchAddrLbl)])
	}

	if val != "" {
		return val
	}

	reg := GetOrchestratorRegion(profileMap)
	dc := GetOrchestratorDC(profileMap)

	// else get from environment variable
	// Note env context && region && dc needs to be considered at the same time
	val = OSGetEnv("_"+strings.ToUpper(reg)+"_"+strings.ToUpper(dc)+string(OrchestratorAddressEnvVarKey), profileMap)
	if val != "" {
		return val
	}

	oName := GetOrchestratorName(profileMap)
	// Nomad Specific
	if oName == NomadOrchestrator {
		// Nomad understands this env variable
		// No need to prefix with any maya api specific context
		return strings.TrimSpace(os.Getenv(string(NomadAddressEnvKey)))
	}

	return val
}

func DefaultOrchestratorAddress() string {
	return string(OrchAddressDef)
}

// GetOrchestratorRegion gets the not nil region name of orchestrator
func GetOrchestratorRegion(profileMap map[string]string) string {
	val := OrchestratorRegion(profileMap)
	if val == "" {
		val = DefaultOrchestratorRegion()
	}

	return val
}

// OrchestratorRegion will fetch the value specified against the
// orchestrator region if available otherwise will return blank.
func OrchestratorRegion(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchRegionLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	val = OSGetEnv(string(OrchestratorRegionEnvVarKey), profileMap)
	if val != "" {
		return val
	}

	oName := GetOrchestratorName(profileMap)
	// Nomad Specific
	if oName == NomadOrchestrator {
		// Nomad understands this env variable
		// No need to prefix with any maya api specific context
		return strings.TrimSpace(os.Getenv(string(NomadRegionEnvKey)))
	}

	return val
}

// DefaultOrchestratorRegion gets the coded default region of orchestration
// provider.
func DefaultOrchestratorRegion() string {
	return string(OrchRegionDef)
}

// GetOrchestratorDC gets the not nil datacenter name of orchestrator
func GetOrchestratorDC(profileMap map[string]string) string {
	val := OrchestratorDC(profileMap)
	if val == "" {
		val = DefaultOrchestratorDC()
	}

	return val
}

// OrchestratorDC will fetch the value specified against the
// orchestrator datacenter if available otherwise will return blank.
func OrchestratorDC(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchDCLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(OrchestratorDCEnvVarKey), profileMap)
}

// DefaultOrchestratorDC gets the coded default datacenter of orchestration
// provider.
func DefaultOrchestratorDC() string {
	return string(OrchDCDef)
}

// GetOrchestratorInCluster gets the not nil value of orchestration provider's
// in-cluster flag
func GetOrchestratorInCluster(profileMap map[string]string) string {
	val := OrchestratorInCluster(profileMap)
	if val == "" {
		val = DefaultOrchestratorInCluster()
	}

	return val
}

// OrchestratorInCluster will fetch the value specified against orchestration provider
// in-cluster flag if available otherwise will return blank.
func OrchestratorInCluster(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchInClusterLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(OrchestratorInClusterEnvVarKey), profileMap)
}

// DefaultOrchestratorInCluster will fetch the coded default value of orchestration
// provider in-cluster flag.
func DefaultOrchestratorInCluster() string {
	return string(OrchInClusterDef)
}

// GetOrchestratorNS gets the not nil orchestrator namespace
func GetOrchestratorNS(profileMap map[string]string) string {
	val := OrchestratorNS(profileMap)
	if val == "" {
		val = DefaultOrchestratorNS()
	}

	return val
}

// OrchestratorNS will fetch the value specified against orchestration provider
// namespace if available otherwise will return blank.
func OrchestratorNS(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchNSLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(OrchestratorNSEnvVarKey), profileMap)
}

// DefaultOrchestratorNS will fetch the default value of orchestration provider
// namespace.
func DefaultOrchestratorNS() string {
	return string(OrchNSDef)
}

// GetControllerImage gets the not nil PVP's VSM controller image
func GetControllerImage(profileMap map[string]string) string {
	val := ControllerImage(profileMap)
	if val == "" {
		val = DefaultControllerImage()
	}

	return val
}

// ControllerImage will fetch the value specified against PVP's VSM
// controller image if available otherwise will return blank.
//
// NOTE:
//    This utility function does not validate & just returns if not capable of
// performing
func ControllerImage(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPControllerImageLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(PVPControllerImageEnvVarKey), profileMap)
}

// DefaultControllerImage will fetch the default value for PVP's VSM controller image
func DefaultControllerImage() string {
	return string(PVPControllerImageDef)
}

// GetControllerNodeTaintTolerations gets the node taint tolerations if
// available
func GetControllerNodeTaintTolerations(profileMap map[string]string) (string, error) {
	val, err := ControllerNodeTaintTolerations(profileMap)
	if err != nil {
		return "", err
	}

	if val == "" {
		val, err = DefaultControllerNodeTaintTolerations()
	}

	return val, err
}

// ControllerNodeTaintTolerations extracts the node taint tolerations
func ControllerNodeTaintTolerations(profileMap map[string]string) (string, error) {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPControllerNodeTaintTolerationLbl)])
	}

	if val != "" {
		return val, nil
	}

	// else get from environment variable
	return OSGetEnv(string(PVPControllerNodeTaintTolerationEnvVarKey), profileMap), nil
}

// DefaultControllerNodeTaintTolerations will fetch the default value for node
// taint tolerations
func DefaultControllerNodeTaintTolerations() (string, error) {
	// Controller node taint toleration property is optional. Hence returns blank
	// (i.e. not required) as default.
	return "", nil
}

// GetReplicaNodeTaintTolerations gets the node taint tolerations if
// available
func GetReplicaNodeTaintTolerations(profileMap map[string]string) (string, error) {
	val, err := ReplicaNodeTaintTolerations(profileMap)
	if err != nil {
		return "", err
	}

	if val == "" {
		val, err = DefaultReplicaNodeTaintTolerations()
	}

	return val, err
}

// ReplicaNodeTaintTolerations extracts the node taint tolerations for replica
func ReplicaNodeTaintTolerations(profileMap map[string]string) (string, error) {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPReplicaNodeTaintTolerationLbl)])
	}

	if val != "" {
		return val, nil
	}

	// else get from environment variable
	return OSGetEnv(string(PVPReplicaNodeTaintTolerationEnvVarKey), profileMap), nil
}

// DefaultReplicaNodeTaintTolerations will fetch the default value for node
// taint tolerations
func DefaultReplicaNodeTaintTolerations() (string, error) {
	// Replica node taint toleration property is optional. Hence returns blank
	// (i.e. not required) as default.
	return "", nil
}

// GetOrchestratorNetworkType gets the not nil orchestration provider's network
// type
func GetOrchestratorNetworkType(profileMap map[string]string) string {
	val := OrchestratorNetworkType(profileMap)
	if val == "" {
		val = DefaultOrchestratorNetworkType()
	}

	return val
}

// OrchestratorNetworkType will fetch the value specified  orchestration
// provider's network type if available otherwise will return blank.
func OrchestratorNetworkType(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchCNTypeLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(OrchestratorCNTypeEnvVarKey), profileMap)
}

// DefaultOrchestratorNetworkType will fetch the coded default value for
// orchestration provider's network type
func DefaultOrchestratorNetworkType() string {
	return string(OrchCNTypeDef)
}

// GetOrchestratorNetworkSubnet gets the not nil orchestration provider's network
// subnet
func GetOrchestratorNetworkSubnet(profileMap map[string]string) (string, error) {
	nAddr := GetOrchestratorNetworkAddr(profileMap)

	subnet, err := nethelper.CIDRSubnet(nAddr)
	if err != nil {
		return "", err
	}

	return subnet, nil
}

// GetOrchestratorNetworkInterface gets the not nil orchestration provider's
// network interface
func GetOrchestratorNetworkInterface(profileMap map[string]string) string {
	val := OrchestratorNetworkInterface(profileMap)
	if val == "" {
		val = DefaultOrchestratorNetworkInterface()
	}

	return val
}

// OrchestratorNetworkInterface will fetch the value specified  orchestration
// provider's network interface if available otherwise will return blank.
func OrchestratorNetworkInterface(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchCNInterfaceLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(OrchestratorCNInterfaceEnvVarKey), profileMap)
}

// DefaultOrchestratorNetworkInterface will fetch the coded default value for
// orchestration provider's network interface
func DefaultOrchestratorNetworkInterface() string {
	return string(OrchCNInterfaceDef)
}

// GetOrchestratorNetworkAddr gets the not nil orchestration provider's network
// address in CIDR notation
func GetOrchestratorNetworkAddr(profileMap map[string]string) string {
	val := OrchestratorNetworkAddr(profileMap)
	if val == "" {
		val = DefaultOrchestratorNetworkAddr()
	}

	return val
}

// OrchestratorNetworkAddr will fetch the value specified against orchestration
// provider network address if available otherwise will return blank.
func OrchestratorNetworkAddr(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(OrchCNNetworkAddrLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(OrchestratorCNAddrEnvVarKey), profileMap)
}

// DefaultOrchestratorNetworkAddr will fetch the coded default value of orchestration
// provider network address.
func DefaultOrchestratorNetworkAddr() string {
	return string(OrchNetworkAddrDef)
}

// GetPVPPersistentPathOnly gets the not nil PVP's VSM replica persistent path
// minus the VSM name
func GetPVPPersistentPathOnly(profileMap map[string]string) string {
	val := PVPPersistentPathOnly(profileMap)
	if val == "" {
		val = DefaultPVPPersistentPathOnly()
	}

	return val
}

// PVPPersistentPathOnly will fetch the value specified against PVP's VSM replica
// persistent path if available otherwise will return blank.
func PVPPersistentPathOnly(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPPersistentPathLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(PVPPersistentPathEnvVarKey), profileMap)
}

// DefaultPVPPersistentPathOnly provides the coded default PVP's VSM replica
// persistent path
func DefaultPVPPersistentPathOnly() string {
	return string(PVPPersistentPathDef)
}

// GetPVPPersistentPath gets the not nil PVP's VSM replica persistent path
func GetPVPPersistentPath(profileMap map[string]string, vsmName string, mountPath string) string {
	val := PVPPersistentPath(profileMap, vsmName, mountPath)
	if val == "" {
		val = DefaultPVPPersistentPath(vsmName, mountPath)
	}

	return val
}

// PVPPersistentPath will fetch the value specified against PVP's VSM replica
// persistent path if available otherwise will return blank.
func PVPPersistentPath(profileMap map[string]string, vsmName string, mountPath string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPPersistentPathLbl)])
	}

	if val != "" {
		return val + "/" + vsmName + mountPath
	}

	// else get from environment variable
	val = OSGetEnv(string(PVPPersistentPathEnvVarKey), profileMap)
	if val != "" {
		return val + "/" + vsmName + mountPath
	}

	return ""
}

// DefaultPVPPersistentPath provides the coded default PVP's VSM replica
// persistent path
func DefaultPVPPersistentPath(vsmName string, mountPath string) string {
	return string(PVPPersistentPathDef) + "/" + vsmName + mountPath
}

// GetPVPReplicaImage gets the not nil value of PVP's VSM replica image
func GetPVPReplicaImage(profileMap map[string]string) string {
	val := PVPReplicaImage(profileMap)
	if val == "" {
		val = DefaultPVPReplicaImage()
	}

	return val
}

// PVPReplicaImage will fetch the value specified against PVP's VSM replica image
// if available otherwise will return blank.
func PVPReplicaImage(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPReplicaImageLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(PVPReplicaImageEnvVarKey), profileMap)
}

// DefaultPVPReplicaImage will fetch the coded default value for PVP's VSM
// replica image
func DefaultPVPReplicaImage() string {
	return string(PVPReplicaImageDef)
}

// GetPVPStorageSize gets the not nil PVP's VSM replica size
func GetPVPStorageSize(profileMap map[string]string) string {
	val := PVPStorageSize(profileMap)
	if val == "" {
		val = DefaultPVPStorageSize()
	}

	return val
}

// PVPStorageSize will fetch the value specified against PVP's VSM replica
// size if available otherwise will return blank.
func PVPStorageSize(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPStorageSizeLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(PVPStorageSizeEnvVarKey), profileMap)
}

// DefaultPVPStorageSize provides the coded default PVP's VSM replica size
func DefaultPVPStorageSize() string {
	return string(PVPStorageSizeDef)
}

// GetPVPReplicaCountInt gets the not nil PVP's VSM replica count
func GetPVPReplicaCountInt(profileMap map[string]string) (int, error) {
	return strconv.Atoi(GetPVPReplicaCount(profileMap))
}

// GetPVPReplicaCount gets the not nil PVP's VSM replica count
func GetPVPReplicaCount(profileMap map[string]string) string {
	val := PVPReplicaCount(profileMap)
	if val == "" {
		val = DefaultPVPReplicaCount()
	}

	return val
}

// PVPReplicaCount will fetch the value specified against PVP's VSM replica
// count if available otherwise will return blank.
func PVPReplicaCount(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPReplicaCountLbl)])
	}

	if val != "" {
		return val
	}

	// else get from environment variable
	return OSGetEnv(string(PVPReplicaCountEnvVarKey), profileMap)
}

// DefaultPVPReplicaCount will fetch the coded default value of PVP's VSM
// replica count
func DefaultPVPReplicaCount() string {
	return string(PVPReplicaCountDef)
}

// GetReplicaCount gets the not nil volume replica count
func GetReplicaCount(spec VolumeSpec) *int32 {
	val := ReplicaCount(spec)
	if val == nil {
		val = DefaultReplicaCount()
	}

	return val
}

// ReplicaCount will fetch the value specified against volume replica
// count if available otherwise will return blank.
func ReplicaCount(spec VolumeSpec) *int32 {
	if spec.Replicas != nil {
		return spec.Replicas
	}

	// else get from environment variable
	countStr := OSGetEnv(string(PVPReplicaCountEnvVarKey), nil)
	count, _ := strconv.ParseInt(countStr, 10, 32)
	count32 := int32(count)
	return &count32
}

// DefaultReplicaCount will fetch the coded default value of volume
// replica count
func DefaultReplicaCount() *int32 {
	count, _ := strconv.ParseInt(string(PVPReplicaCountDef), 10, 32)
	count32 := int32(count)
	return &count32
}

// MakeOrDefJivaReplicaArgs will set the placeholders in jiva replica args with
// their appropriate runtime values.
//
// NOTE:
//    The defaults will be set if the replica args are not available
//
// NOTE:
//    This utility function does not validate & just returns if not capable of
// performing
func MakeOrDefJivaReplicaArgs(profileMap map[string]string, clusterIP string) []string {
	if strings.TrimSpace(clusterIP) == "" {
		return nil
	}

	storSize := GetPVPStorageSize(profileMap)

	repArgs := make([]string, len(JivaReplicaArgs))

	for i, rArg := range JivaReplicaArgs {
		rArg = strings.Replace(rArg, string(JivaClusterIPHolder), clusterIP, 1)
		rArg = strings.Replace(rArg, string(JivaStorageSizeHolder), storSize, 1)
		repArgs[i] = rArg
	}

	return repArgs
}

// DefaultJivaISCSIPort will provide the port required to make ISCSI based
// connections
func DefaultJivaISCSIPort() int32 {
	iscsiPort, _ := strconv.Atoi(string(JivaISCSIPortDef))
	return int32(iscsiPort)
}

// DefaultJivaAPIPort will provide the port required for management of
// persistent volume
func DefaultJivaAPIPort() int32 {
	apiPort, _ := strconv.Atoi(string(JivaAPIPortDef))
	return int32(apiPort)
}

// DefaultPersistentPathCount will provide the default count of persistent
// paths required during provisioning.
//func DefaultPersistentPathCount() int {
//	pCount, _ := strconv.Atoi(string(PVPPersistentPathCountDef))
//	return pCount
//}

// PersistentPathCount will fetch the value specified against persistent volume
// persistent path count if available otherwise will return blank.
//
// NOTE:
//    This utility function does not validate & just returns if not capable of
// performing
//func PersistentPathCount(profileMap map[string]string) string {
//if profileMap == nil {
//	return ""
//}

// Extract persistent path count
//return profileMap[string(PVPPersistentPathCountLbl)]
//}

// Replicas returns a pointer to an int32 of a int value
func Replicas(rcount int) *int32 {
	o := int32(rcount)
	return &o
}

//
func MakeOrDefJivaControllerArgs(vsm string, clusterIP string) []string {
	if strings.TrimSpace(vsm) == "" || strings.TrimSpace(clusterIP) == "" {
		return nil
	}

	ctrlArgs := make([]string, len(JivaCtrlArgs))

	for i, cArg := range JivaCtrlArgs {
		cArg = strings.Replace(cArg, string(JivaVolumeNameHolder), vsm, 1)
		cArg = strings.Replace(cArg, string(JivaClusterIPHolder), clusterIP, 1)
		ctrlArgs[i] = cArg
	}

	return ctrlArgs
}

// DefaultJivaMountPath provides the default mount path for jiva based persistent
// volumes
func DefaultJivaMountPath() string {
	return string(JivaPersistentMountPathDef)
}

// DefaultJivaMountName provides the default mount path name for jiva based
// persistent volumes
func DefaultJivaMountName() string {
	return string(JivaPersistentMountNameDef)
}

// DefaultJivaReplicaPort1 provides the default port for jiva based
// persistent volume replicas
func DefaultJivaReplicaPort1() int32 {
	p, _ := strconv.Atoi(string(JivaReplicaPortOneDef))
	return int32(p)
}

// DefaultJivaReplicaPort2 provides the default port for jiva based
// persistent volume replicas
func DefaultJivaReplicaPort2() int32 {
	p, _ := strconv.Atoi(string(JivaReplicaPortTwoDef))
	return int32(p)
}

// DefaultJivaReplicaPort3 provides the default port for jiva based
// persistent volume replicas
func DefaultJivaReplicaPort3() int32 {
	p, _ := strconv.Atoi(string(JivaReplicaPortThreeDef))
	return int32(p)
}

//
func SanitiseVSMName(vsm string) string {
	// Trim the controller suffix if controller based
	v := strings.TrimSuffix(vsm, string(ControllerSuffix))
	// Or Trim the replica suffix if replica based
	v = strings.TrimSuffix(v, string(ReplicaSuffix))

	return v
}

// GetPVPVSMIPs gets not nil values of PVP's VSM Controller IPs & Replica IPs
//
// NOTE:
//    The logic caters to get the VSM IPs i.e. both Controller & Replica IPs.
// It will be error prone to get these IPs separately in cases where maya api
// service gets un-used IPs.
//
// NOTE:
//    Maya api service uses a very naive approach to get un-used IPs. It is
// advised to make use of external networking utilities or orchestrators who
// come up with their own networking tools to get the un-used IPs.
func GetPVPVSMIPs(profileMap map[string]string) (string, string, error) {
	ctrlIPs, repIPs := PVPVSMIPs(profileMap)

	if ctrlIPs != "" && repIPs != "" {
		return ctrlIPs, repIPs, nil
	}

	var err error
	if ctrlIPs == "" && repIPs == "" {
		ctrlIPs, repIPs, err = DefaultPVPVSMIPs(profileMap, true, true)
	} else if ctrlIPs == "" {
		ctrlIPs, repIPs, err = DefaultPVPVSMIPs(profileMap, true, false)
	} else {
		ctrlIPs, repIPs, err = DefaultPVPVSMIPs(profileMap, false, true)
	}

	if err != nil {
		return "", "", err
	}

	return ctrlIPs, repIPs, nil
}

// PVPVSMIPs will fetch the value specified against PVP's VSM Controller
// IPs & Replica IPs if available otherwise will return blank.
func PVPVSMIPs(profileMap map[string]string) (string, string) {
	return ControllerIPs(profileMap), ReplicaIPs(profileMap)
}

// DefaultPVPVSMIPs will fetch the PVP's VSM Controller IPs & Replica IPs based
// on the network address
//
// NOTE:
//    This is a very naive approach to get un-used IPs. It is
// advised to make use of external networking utilities or orchestrators who
// come up with their own networking plugin to get the un-used IPs.
func DefaultPVPVSMIPs(profileMap map[string]string, requestCtrlIPs bool, requestRepIPs bool) (string, string, error) {
	if !requestCtrlIPs && !requestRepIPs {
		return "", "", nil
	}

	cc, err := GetPVPControllerCountInt(profileMap)
	if err != nil {
		return "", "", err
	}
	if cc <= 0 {
		return "", "", fmt.Errorf("Invalid count '%d' w.r.t VSM Controller IPs", cc)
	}

	rc, err := GetPVPReplicaCountInt(profileMap)
	if err != nil {
		return "", "", err
	}
	if rc <= 0 {
		return "", "", fmt.Errorf("Invalid count '%d' w.r.t VSM Replica IPs", rc)
	}

	// IPs will be fetched based on the network address
	nAddr := GetOrchestratorNetworkAddr(profileMap)

	// Logic to get the Controller IPs & Replica IPs
	var uIPs []string
	if requestCtrlIPs && requestRepIPs {
		uIPs, err = GetUnusedIPs(cc+rc, nAddr)
	} else if requestCtrlIPs {
		uIPs, err = GetUnusedIPs(cc, nAddr)
	} else {
		uIPs, err = GetUnusedIPs(rc, nAddr)
	}

	if err != nil {
		return "", "", err
	}

	if uIPs == nil || len(uIPs) == 0 {
		return "", "", fmt.Errorf("Could not find unused IPs")
	}

	if len(uIPs) != cc+rc {
		return "", "", fmt.Errorf("Could not find required '%d' unused IPs, got '%d'.", cc+rc, len(uIPs))
	}

	return strings.Join(uIPs[:cc], ","), strings.Join(uIPs[cc:], ","), nil
}

// ControllerIPs will fetch the value specified against PVP's VSM controller
// IPs if available otherwise will return blank.
func ControllerIPs(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPControllerIPsLbl)])
	}

	return val
}

// ReplicaIPs will fetch the value specified against PVP's VSM replica
// IPs if available otherwise will return blank.
func ReplicaIPs(profileMap map[string]string) string {
	val := ""
	if profileMap != nil {
		val = strings.TrimSpace(profileMap[string(PVPReplicaIPsLbl)])
	}

	return val
}

// GetUnusedIPs gets un-used IPs based on the provided network address.
//
// NOTE:
//    It is advised to make use of external networking utilities or
// orchestrators who come up with their own networking plugin to get the un-used
// IPs.
func GetUnusedIPs(count int, nAddr string) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("Invalid unused IP count '%d' provided", count)
	}

	if nAddr == "" {
		return nil, fmt.Errorf("Blank network address provided")
	}

	ips, err := nethelper.GetAvailableIPs(nAddr, count)
	if err != nil {
		return nil, err
	}

	return ips, nil
}

// OSGetEnv fetches the environment variable value from the machine's
// environment using contextual information
// TODO:
//    Introduce some debug logging for the derived keys & values. Do not log
// the values if they reflect some sensitive info.
func OSGetEnv(envKey string, profileMap map[string]string) string {
	val := ""
	evCtxVal := ""

	if profileMap != nil {
		// derive the context of environment variable
		evCtxVal = strings.ToUpper(strings.TrimSpace(profileMap[string(EnvVariableContextLbl)]))
	}

	if evCtxVal == "" {
		// use the hard coded default context for environment variable
		evCtxVal = string(EnvVariableContextDef)
	}

	val = strings.TrimSpace(os.Getenv(evCtxVal + envKey))
	// TODO
	// Set to DEBUG log
	glog.Infof("Will use env var '%s: %s'", evCtxVal+envKey, val)

	return val
}
