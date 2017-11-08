package profiles

import (
	"fmt"
	"strings"

	"github.com/openebs/maya/types/v1"
)

// VolumeProvisionerProfile abstracts & exposes a persistent volume provisioner's
// runtime features.
//
// NOTE:
//    A persistent volume provisioner can align to a specific implementation of
// this profile & hence change its execution strategy at runtime.
type VolumeProvisionerProfile interface {
	// Label assigned against the persistent volume provisioner profile.
	Label() v1.VolumeProvisionerProfileLabel

	// Registered volume provisioner profile name.
	Name() v1.VolumeProvisionerProfileRegistry

	// Get the volume associated with this provisioner profile
	Volume() (*v1.Volume, error)

	// Gets the orchestration provider name.
	// A persistent volume provisioner plugin may be linked with a orchestrator
	// e.g. K8s, Nomad, Mesos, Swarm, etc. It can be Docker engine as well.
	//
	// Note:
	//    It can return false in its second return argument if orchestrator support
	// is not applicable.
	//
	// Note:
	//    OpenEBS believes in running storage software in containers & hence
	// these container specific orchestrators.
	Orchestrator() (v1.OrchProviderRegistry, bool, error)

	// Copy returns a copy of this volume provisioner profile
	//
	// NOTE:
	//    In certain cases, name of VSM is derived much later & hence this
	// method provides the option to set VSM name against a copy of this volume
	// provisioner profile.
	Copy(vsm string) (VolumeProvisionerProfile, error)

	// Get the name of the VSM
	VSMName() (string, error)

	// Get the number of controllers
	ControllerCount() (*int32, error)

	// Gets the controller's image e.g. docker image version. The second return value
	// indicates if image based replica is supported or not.
	ControllerImage() (string, bool, error)

	// Get the IP addresses that needs to be assigned against the controller(s)
	ControllerIPs() ([]string, error)

	// Gets the replica's image e.g. docker image version.
	ReplicaImage() (string, error)

	// Get the storage size for each replica(s)
	StorageSize() (string, error)

	// Get the number of replicas
	ReplicaCount() (*int32, error)

	// Get the IP addresses that needs to be assigned against the replica(s)
	ReplicaIPs() ([]string, error)

	// Get the storage backend i.e. a persistent path of the replica.
	PersistentPath() (string, error)

	// Verify if node level taint tolerations are required for controller?
	IsControllerNodeTaintTolerations() ([]string, bool, error)

	// Verify if node level taint tolerations are required for replica?
	IsReplicaNodeTaintTolerations() ([]string, bool, error)
}

// GetVolProProfile will return a specific persistent volume provisioner
// profile. It will decide first based on the provided specifications failing
// which will ensure a default profile is returned.
func GetVolProProfile(vol *v1.Volume) (VolumeProvisionerProfile, error) {
	if vol == nil {
		return nil, fmt.Errorf("Volume is required to create a volume provisioner profile")
	}

	return GetDefaultVolProProfile(vol)
}

// GetDefaultVolProProfile will return the default volume provisioner
// profile.
func GetDefaultVolProProfile(vol *v1.Volume) (VolumeProvisionerProfile, error) {

	if vol == nil {
		return nil, fmt.Errorf("Volume is required to create default volume provisioner profile")
	}

	return newDefVolProProfile(vol)
}

// defVolProProfile is a persistent volume provisioner profile that is based on
// volume.
//
// NOTE:
//    This will use defaults in-case the values are not set in persistent volume
// claim.
//
// NOTE:
//    This is a concrete implementation of
// volumeprovisioner.VolumeProvisionerProfile
type defVolProProfile struct {
	vol     *v1.Volume
	vsmName string
}

// newDefVolProProfile provides a new instance of VolumeProvisionerProfile that is
// based on volume
func newDefVolProProfile(vol *v1.Volume) (VolumeProvisionerProfile, error) {
	return &defVolProProfile{
		vol: vol,
	}, nil
}

// Label provides the label assigned against the persistent volume provisioner
// profile.
//
// NOTE:
//    There can be many persistent volume provisioner profiles with this same label.
// This is used along with Name() method.
func (pp *defVolProProfile) Label() v1.VolumeProvisionerProfileLabel {
	return v1.PVPProfileNameLbl
}

// Name provides the name assigned to the persistent volume provisioner profile.
//
// NOTE:
//    Name provides the uniqueness among various variants of persistent volume
// provisioner profiles.
func (pp *defVolProProfile) Name() v1.VolumeProvisionerProfileRegistry {
	return v1.VolumeProvisionerProfile
}

// Volume provides the volume associated with this profile.
//
// NOTE:
//    This method provides a convinient way to access volume. In other words
// volume provisioner profile acts as a wrapper over volume.
func (pp *defVolProProfile) Volume() (*v1.Volume, error) {
	return pp.vol, nil
}

// Orchestrator gets the suitable orchestration provider.
// A persistent volume provisioner plugin may be linked with a orchestrator
// e.g. K8s, Nomad, Mesos, Swarm, etc. It can be Docker engine as well.
func (pp *defVolProProfile) Orchestrator() (v1.OrchProviderRegistry, bool, error) {
	// Extract the name of orchestration provider
	//oName := v1.GetOrchestratorName(pp.vol.Labels)
	oName := v1.GetOrchestratorName(nil)

	// Get the orchestrator instance
	return oName, true, nil
}

// Copy returns a copy of this volume provisioner profile
//
// NOTE:
//    In certain cases, name of VSM is derived much later & hence this
// method provides the runtime option to set VSM name against a copy of this
// volume provisioner profile.
func (pp *defVolProProfile) Copy(vsm string) (VolumeProvisionerProfile, error) {

	if pp == nil {
		return nil, nil
	}

	t := strings.TrimSpace(vsm)
	if t == "" {
		return nil, fmt.Errorf("Volume name can not be empty")
	}

	// copy
	n := new(defVolProProfile)
	*n = *pp
	n.vsmName = t

	return n, nil
}

// VSMName gets the name of the VSM
// Operator must provide this.
func (pp *defVolProProfile) VSMName() (string, error) {
	// Extract the VSM name from Volume
	vsmName := strings.TrimSpace(v1.VSMName(pp.vol.Name))

	// This might be a case where VSM name was set at runtime
	// during the life span of the request & not during the initiation of
	// this volume provisioner profile. Hence, check the runtime VSM name as well.
	if vsmName == "" {
		vsmName = pp.vsmName
	}

	if vsmName == "" {
		return "", fmt.Errorf("Missing VSM name in '%s:%s'", pp.Label(), pp.Name())
	}

	return vsmName, nil
}

// ControllerCount gets the number of controllers
func (pp *defVolProProfile) ControllerCount() (*int32, error) {
	var rCount *int32
	specs := pp.vol.Specs

	for _, spec := range specs {
		if spec.Context == v1.ControllerVolumeContext {
			rCount = spec.Replicas
		}
	}

	if rCount == nil {
		return nil, fmt.Errorf("Volume controller count is missing")
	}

	return rCount, nil
}

// ControllerImage gets the controller's image currently its docker image label.
func (pp *defVolProProfile) ControllerImage() (string, bool, error) {
	// Extract the controller image
	// Extract the replica image
	specs := pp.vol.Specs
	rImg := ""

	for _, spec := range specs {
		if spec.Context == v1.ControllerVolumeContext {
			rImg = spec.Image
			break
		}
	}

	if rImg == "" {
		return "", true, fmt.Errorf("Volume controller image is missing")
	}

	return rImg, true, nil
}

// ReplicaImage gets the replica's image currently its docker image label.
func (pp *defVolProProfile) ReplicaImage() (string, error) {
	// Extract the replica image
	specs := pp.vol.Specs
	rImg := ""

	for _, spec := range specs {
		if spec.Context == v1.ReplicaVolumeContext {
			rImg = spec.Image
			break
		}
	}

	if rImg == "" {
		return "", fmt.Errorf("Volume replica image is missing")
	}

	return rImg, nil
}

// IsControllerNodeTaintTolerations provides the node level taint tolerations.
// Since node level taint toleration for controller is an optional feature, it
// can return false.
func (pp *defVolProProfile) IsControllerNodeTaintTolerations() ([]string, bool, error) {
	// Extract the node taint toleration for controller
	//nTTs, err := v1.GetControllerNodeTaintTolerations(pp.vol.Labels)
	nTTs, err := v1.GetControllerNodeTaintTolerations(nil)
	if err != nil {
		return nil, false, err
	}

	if strings.TrimSpace(nTTs) == "" {
		return nil, false, nil
	}

	// nTTs is expected of below form
	// key=value:effect, key1=value1:effect1
	// __or__
	// key=value:effect
	return strings.Split(nTTs, ","), true, nil
}

// IsReplicaNodeTaintTolerations provides the node level taint tolerations.
// Since node level taint toleration for replica is an optional feature, it
// can return false.
func (pp *defVolProProfile) IsReplicaNodeTaintTolerations() ([]string, bool, error) {
	// Extract the node taint toleration for replica
	//nTTs, err := v1.GetReplicaNodeTaintTolerations(pp.vol.Labels)
	nTTs, err := v1.GetReplicaNodeTaintTolerations(nil)
	if err != nil {
		return nil, false, err
	}

	if strings.TrimSpace(nTTs) == "" {
		return nil, false, nil
	}

	// nTTs is expected of below form
	// key=value:effect, key1=value1:effect1
	// __or__
	// key=value:effect
	return strings.Split(nTTs, ","), true, nil
}

// StorageSize gets the storage size for each persistent volume replica(s)
func (pp *defVolProProfile) StorageSize() (string, error) {
	// Extract the storage size
	sSize := pp.vol.Capacity

	if len(sSize) == 0 {
		return "", fmt.Errorf("Volume capacity is missing")
	}

	return sSize, nil
}

// ReplicaCount get the number of replicas required
func (pp *defVolProProfile) ReplicaCount() (*int32, error) {
	var rCount *int32
	specs := pp.vol.Specs

	for _, spec := range specs {
		if spec.Context == v1.ReplicaVolumeContext {
			rCount = spec.Replicas
		}
	}

	if rCount == nil {
		return nil, fmt.Errorf("Volume replica count is missing")
	}

	return rCount, nil
}

// ControllerIPs gets the IP addresses that needs to be assigned against the
// controller(s)
//
// NOTE:
//    There is no default assignment of IPs
func (pp *defVolProProfile) ControllerIPs() ([]string, error) {
	// Extract the controller IPs
	//cIPs := v1.ControllerIPs(pp.vol.Labels)
	cIPs := v1.ControllerIPs(nil)

	if cIPs == "" {
		return nil, nil
	}

	cIPsArr := strings.Split(cIPs, ",")

	if len(cIPsArr) == 0 {
		return nil, fmt.Errorf("Invalid controller IPs in '%s:%s'", pp.Label(), pp.Name())
	}

	return cIPsArr, nil
}

// ReplicaIPs gets the IP addresses that needs to be assigned against the
// replica(s)
//
// NOTE:
//    There is no default assignment of IPs
func (pp *defVolProProfile) ReplicaIPs() ([]string, error) {
	// Extract the controller IPs
	//rIPs := v1.ReplicaIPs(pp.vol.Labels)
	rIPs := v1.ReplicaIPs(nil)

	if rIPs == "" {
		return nil, nil
	}

	rIPsArr := strings.Split(rIPs, ",")

	if len(rIPsArr) == 0 {
		return nil, fmt.Errorf("Invalid replica IPs in '%s:%s'", pp.Label(), pp.Name())
	}

	return rIPsArr, nil
}

// PersistentPath gets the persistent path based on the replica position.
//
// NOTE:
//    `position` is just a positional value that determines a particular replica
// out of the total replica count i.e. rCount.
func (pp *defVolProProfile) PersistentPath() (string, error) {
	vsm, err := pp.VSMName()
	if err != nil {
		return "", err
	}

	// Extract the persistent path
	//pPath := v1.GetPVPPersistentPath(pp.vol.Labels, vsm, string(v1.JivaPersistentMountPathDef))
	pPath := v1.GetPVPPersistentPath(nil, vsm, string(v1.JivaPersistentMountPathDef))

	return pPath, nil
}

// etcdVolProProfile represents a generic volume provisioner profile whose
// properties are stored in etcd database.
//
// NOTE:
//    There can be multiple persistent volume provisioner profiles stored in
// etcd
//
// NOTE:
//    Properties specified in persistent volume claim will override the ones
// specified in etcd
//
// NOTE:
//    This is a concrete implementation of volume.VolumeProvisionerProfile
//type etcdVolProProfile struct {
//	pvc *v1.Volume
//}
