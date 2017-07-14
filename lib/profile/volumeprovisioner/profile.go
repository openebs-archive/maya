package volumeprovisioner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openebs/mayaserver/lib/api/v1"
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

	// Get the persistent volume claim associated with this provisioner profile
	PVC() (*v1.PersistentVolumeClaim, error)

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
	ControllerCount() (int, error)

	// Gets the controller's image e.g. docker image version. The second return value
	// indicates if image based replica is supported or not.
	ControllerImage() (string, bool, error)

	// Get the IP addresses that needs to be assigned against the controller(s)
	ControllerIPs() ([]string, error)

	// Gets the replica's image e.g. docker image version. The second return value
	// indicates if image based replica is supported or not.
	ReplicaImage() (string, bool, error)

	// Get the storage size for each replica(s)
	StorageSize() (string, error)

	// Get the number of replicas
	ReplicaCount() (int, error)

	// Get the IP addresses that needs to be assigned against the replica(s)
	ReplicaIPs() ([]string, error)

	// Get the count of persistent paths required for all the replicas
	PersistentPathCount() (int, error)

	// Get the persistent path based on the replica position.
	//
	// NOTE:
	//    `position` is just a parameter that determines a particular replica out
	// of the total replica count i.e. `rCount`.
	PersistentPath(position int, rCount int) (string, error)
}

// GetVolProProfileByPVC will return a specific persistent volume provisioner
// profile. It will decide first based on the provided specifications failing
// which will ensure a default profile is returned.
func GetVolProProfileByPVC(pvc *v1.PersistentVolumeClaim) (VolumeProvisionerProfile, error) {
	if pvc == nil {
		return nil, fmt.Errorf("PVC is required to create a volume provisioner profile")
	}

	// Extract the name of volume provisioner profile
	volProflName := v1.VolumeProvisionerProfileName(pvc.Labels)

	if volProflName == "" {
		return GetDefaultVolProProfile(pvc)
	}

	return GetVolProProfileByName(volProflName, pvc)
}

// GetDefaultVolProProfile will return the default volume provisioner
// profile.
//
// NOTE:
//    PVC based volume provisioner profile is considered as default
func GetDefaultVolProProfile(pvc *v1.PersistentVolumeClaim) (VolumeProvisionerProfile, error) {

	if pvc == nil {
		return nil, fmt.Errorf("PVC is required to create default volume provisioner profile")
	}

	return newPvcVolProProfile(pvc)
}

// TODO
//
// GetVolProProfileByName will return a volume provisioner profile by
// looking up from the provided profile name.
func GetVolProProfileByName(name string, pvc *v1.PersistentVolumeClaim) (VolumeProvisionerProfile, error) {
	// TODO
	// Search from the in-memory registry

	// TODO
	// Alternatively, search from external discoverable DBs if any

	return nil, fmt.Errorf("GetVolProProfileByName is not yet implemented")
}

// pvcVolProProfile is a persistent volume provisioner profile that is based on
// persistent volume claim.
//
// NOTE:
//    This will use defaults in-case the values are not set in persistent volume
// claim.
//
// NOTE:
//    This is a concrete implementation of
// volumeprovisioner.VolumeProvisionerProfile
type pvcVolProProfile struct {
	pvc     *v1.PersistentVolumeClaim
	vsmName string
}

// newPvcVolProProfile provides a new instance of VolumeProvisionerProfile that is
// based on pvc (i.e. persistent volume claim).
func newPvcVolProProfile(pvc *v1.PersistentVolumeClaim) (VolumeProvisionerProfile, error) {
	return &pvcVolProProfile{
		pvc: pvc,
	}, nil
}

// Label provides the label assigned against the persistent volume provisioner
// profile.
//
// NOTE:
//    There can be many persistent volume provisioner profiles with this same label.
// This is used along with Name() method.
func (pp *pvcVolProProfile) Label() v1.VolumeProvisionerProfileLabel {
	return v1.PVPProfileNameLbl
}

// Name provides the name assigned to the persistent volume provisioner profile.
//
// NOTE:
//    Name provides the uniqueness among various variants of persistent volume
// provisioner profiles.
func (pp *pvcVolProProfile) Name() v1.VolumeProvisionerProfileRegistry {
	return v1.PVCProvisionerProfile
}

// PVC provides the persistent volume claim associated with this profile.
//
// NOTE:
//    This method provides a convinient way to access pvc. In other words
// volume provisioner profile acts as a wrapper over pvc.
func (pp *pvcVolProProfile) PVC() (*v1.PersistentVolumeClaim, error) {
	return pp.pvc, nil
}

// Orchestrator gets the suitable orchestration provider.
// A persistent volume provisioner plugin may be linked with a orchestrator
// e.g. K8s, Nomad, Mesos, Swarm, etc. It can be Docker engine as well.
func (pp *pvcVolProProfile) Orchestrator() (v1.OrchProviderRegistry, bool, error) {
	// Extract the name of orchestration provider
	oName := v1.GetOrchestratorName(pp.pvc.Labels)

	// Get the orchestrator instance
	return oName, true, nil
}

// Copy returns a copy of this volume provisioner profile
//
// NOTE:
//    In certain cases, name of VSM is derived much later & hence this
// method provides the runtime option to set VSM name against a copy of this
// volume provisioner profile.
func (pp *pvcVolProProfile) Copy(vsm string) (VolumeProvisionerProfile, error) {

	if pp == nil {
		return nil, nil
	}

	t := strings.TrimSpace(vsm)
	if t == "" {
		return nil, fmt.Errorf("VSM name can not be empty")
	}

	// copy
	n := new(pvcVolProProfile)
	*n = *pp
	n.vsmName = t

	return n, nil
}

// VSMName gets the name of the VSM
// Operator must provide this.
func (pp *pvcVolProProfile) VSMName() (string, error) {
	// Extract the VSM name from PVC
	// Name of PVC is the name of VSM
	vsmName := strings.TrimSpace(v1.VSMName(pp.pvc.Name))

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
func (pp *pvcVolProProfile) ControllerCount() (int, error) {
	// Extract the controller count from pvc
	return v1.GetPVPControllerCountInt(pp.pvc.Labels)
}

// ControllerImage gets the controller's image currently its docker image label.
func (pp *pvcVolProProfile) ControllerImage() (string, bool, error) {
	// Extract the controller image from pvc
	cImg := v1.GetControllerImage(pp.pvc.Labels)

	return cImg, true, nil
}

// ReplicaImage gets the replica's image currently its docker image label.
func (pp *pvcVolProProfile) ReplicaImage() (string, bool, error) {
	// Extract the replica image from pvc
	rImg := v1.GetPVPReplicaImage(pp.pvc.Labels)

	return rImg, true, nil
}

// StorageSize gets the storage size for each persistent volume replica(s)
func (pp *pvcVolProProfile) StorageSize() (string, error) {
	// Extract the storage size from pvc
	sSize := v1.GetPVPStorageSize(pp.pvc.Labels)

	if sSize == "" {
		return "", fmt.Errorf("Missing storage size in '%s:%s'", pp.Label(), pp.Name())
	}

	return sSize, nil
}

// ReplicaCount get the number of replicas required
func (pp *pvcVolProProfile) ReplicaCount() (int, error) {
	// Extract the replica count from pvc
	return v1.GetPVPReplicaCountInt(pp.pvc.Labels)
}

// ControllerIPs gets the IP addresses that needs to be assigned against the
// controller(s)
//
// NOTE:
//    There is no default assignment of IPs
func (pp *pvcVolProProfile) ControllerIPs() ([]string, error) {
	// Extract the controller IPs from pvc
	cIPs := v1.ControllerIPs(pp.pvc.Labels)

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
func (pp *pvcVolProProfile) ReplicaIPs() ([]string, error) {
	// Extract the controller IPs from pvc
	rIPs := v1.ReplicaIPs(pp.pvc.Labels)

	if rIPs == "" {
		return nil, nil
	}

	rIPsArr := strings.Split(rIPs, ",")

	if len(rIPsArr) == 0 {
		return nil, fmt.Errorf("Invalid replica IPs in '%s:%s'", pp.Label(), pp.Name())
	}

	return rIPsArr, nil
}

// TODO
// Return int32 ??
//
// PersistentPathCount gets the count of persistent paths required for all the
// replicas.
//
// NOTE:
//    The count needs to be equal to no of replicas.
func (pp *pvcVolProProfile) PersistentPathCount() (int, error) {
	// Extract the persistent path count from pvc
	pCount := v1.PersistentPathCount(pp.pvc.Labels)

	if pCount == "" {
		return v1.DefaultPersistentPathCount(), nil
	}

	iPCount, err := strconv.Atoi(pCount)
	if err != nil {
		return 0, err
	}

	return iPCount, nil
}

// PersistentPath gets the persistent path based on the replica position.
//
// NOTE:
//    `position` is just a positional value that determines a particular replica
// out of the total replica count i.e. rCount.
func (pp *pvcVolProProfile) PersistentPath(position int, rCount int) (string, error) {
	if rCount <= 0 {
		return "", fmt.Errorf("Invalid replica count '%d' provided", rCount)
	}

	if position <= 0 {
		return "", fmt.Errorf("Invalid persistent path index '%d' provided", position)
	}

	vsm, err := pp.VSMName()
	if err != nil {
		return "", err
	}

	// Extract the persistent path from pvc
	pPath := v1.GetPVPPersistentPath(pp.pvc.Labels, vsm, string(v1.JivaPersistentMountPathDef))

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
type etcdVolProProfile struct {
	pvc *v1.PersistentVolumeClaim
}
