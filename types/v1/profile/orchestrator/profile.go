package orchestrator

import (
	"fmt"

	"github.com/openebs/maya/pkg/nethelper"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
)

// VolumeProvisionerProfile abstracts & exposes a persistent volume provisioner's
// runtime features.
//
// NOTE:
//    A persistent volume provisioner can align to a specific implementation of
// this profile & hence change its execution strategy at runtime.
type OrchProviderProfile interface {
	// Label assigned against the orchestration provider profile.
	Label() v1.OrchProviderProfileLabel

	// Registered orchestration provider profile name.
	Name() v1.OrchProviderProfileRegistry

	// Get the persistent volume claim associated with this orchestration provider
	PVC() (*v1.Volume, error)

	// Get the network address in CIDR format
	NetworkAddr() (string, error)

	// Get the network subnet
	NetworkSubnet() (string, error)

	// Get the namespace used at the orchestrator, where the request needs to be
	// operated on
	NS() (string, error)

	// InCluster indicates if the request to the orchestrator is scoped to the
	// cluster where this request originated
	//
	// TODO
	// Should this be termed as InDC ? Is a cluster same as a DataCenter ?
	// Cluster vs. DC vs. Region ?
	InCluster() (bool, error)
}

// GetOrchProviderProfile will return a specific orchestration provider profile.
//
// TODO
//  It will decide first based on the provided specifications failing which will
// ensure a default profile is returned.
func GetOrchProviderProfile(pvc *v1.Volume) (OrchProviderProfile, error) {
	var profileMap map[string]string

	if pvc != nil && pvc.Labels != nil {
		profileMap = pvc.Labels
	} else {
		profileMap = nil
	}

	// TODO
	// This is hard coded to pvcOrchProviderProfile struct
	// It should be based on inputs/env vars
	return &pvcOrchProviderProfile{
		pvc:        pvc,
		profileMap: profileMap,
	}, nil
}

// GetDefaultOrchProviderProfile will return the default orchestration provider
// profile.
//
// NOTE:
//    PVC based orchestration provider profile is considered as default
//func GetDefaultOrchProviderProfile() (OrchProviderProfile, error) {
//	return &pvcOrchProviderProfile{}, nil
//}

// TODO
//
// GetOrchProviderProfileByName will return a orchestration provider profile by
// looking up from the provided profile name.
//func GetOrchProviderProfileByName(name string) (OrchProviderProfile, error) {
// TODO
// Search from the in-memory registry

// TODO
// Alternatively, search from external discoverable DBs if any

//return nil, fmt.Errorf("GetOrchProviderProfileByName is not yet implemented")
//}

// pvcOrchProviderProfile is a orchestration provider profile that is based on
// persistent volume claim.
//
// NOTE:
//    This is a concrete implementation of orchprovider.VolumeProvisionerProfile
type pvcOrchProviderProfile struct {
	pvc        *v1.Volume
	profileMap map[string]string
}

// newPvcOrchProviderProfile provides a new instance of OrchProviderProfile that
// is based on pvc (i.e. persistent volume claim).
//func newPvcOrchProviderProfile(pvc *v1.PersistentVolumeClaim) (OrchProviderProfile, error) {
// This does not care if pvc instance is nil
//return &pvcOrchProviderProfile{
//	pvc: pvc,
//}, nil
//}

// Label provides the label assigned against the persistent volume provisioner
// profile.
//
// NOTE:
//    There can be many persistent volume provisioner profiles with this same label.
// This is used along with Name() method.
func (op *pvcOrchProviderProfile) Label() v1.OrchProviderProfileLabel {
	return v1.OrchProfileNameLbl
}

// Name provides the name assigned to the orchestration provider profile.
//
// NOTE:
//    Name provides the uniqueness among various variants of orchestration
// provider profiles.
func (op *pvcOrchProviderProfile) Name() v1.OrchProviderProfileRegistry {
	return v1.PVCOrchestratorProfile
}

// PVC provides the persistent volume claim associated with this profile.
//
// NOTE:
//    This method provides a convinient way to access pvc. In other words
// orchestration provider profile acts as a wrapper over pvc.
func (op *pvcOrchProviderProfile) PVC() (*v1.Volume, error) {
	return op.pvc, nil
}

// NetworkAddr gets the network address in CIDR format
func (op *pvcOrchProviderProfile) NetworkAddr() (string, error) {
	nAddr := v1.GetOrchestratorNetworkAddr(op.profileMap)

	if !nethelper.IsCIDR(nAddr) {
		return "", fmt.Errorf("Network address not in CIDR format in '%s:%s'", op.Label(), op.Name())
	}

	return nAddr, nil
}

// NetworkSubnet gets the network's subnet in decimal format
func (op *pvcOrchProviderProfile) NetworkSubnet() (string, error) {
	nAddr, err := op.NetworkAddr()
	if err != nil {
		return "", err
	}

	subnet, err := nethelper.CIDRSubnet(nAddr)
	if err != nil {
		return "", err
	}

	return subnet, nil
}

// Get the namespace used at the orchestrator, where the request needs to be
// operated on
func (op *pvcOrchProviderProfile) NS() (string, error) {
	ns := v1.GetOrchestratorNS(op.profileMap)

	return ns, nil
}

// InCluster indicates if the request to the orchestrator is scoped to the
// cluster where this request originated
func (op *pvcOrchProviderProfile) InCluster() (bool, error) {
	inCluster := v1.GetOrchestratorInCluster(op.profileMap)

	return util.CheckTruthy(inCluster), nil
}

// TODO
//
// etcdOrchProviderProfile represents a generic orchestration provider profile
// whose properties are stored in etcd database.
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
//    Properties missing in etcd & persistent volume claim will make use of the
// defaults provided by maya api service
//
// NOTE:
//    This is a concrete implementation of volume.VolumeProvisionerProfile
type etcdOrchProviderProfile struct {
}
