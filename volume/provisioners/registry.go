// This file provides persistent volume provisioner's registry related features.
//
// NOTE:
//    This is the new file w.r.t the deprecated volume/plugins.go file
package provisioners

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
)

type VolumeProvisionerFactory func(label, name string) (VolumeInterface, error)

// Registration is managed in a safe manner via these variables
var (
	volProvisionerRegMutex sync.Mutex
	volProvisionerRegistry = make(map[v1.VolumeProvisionerRegistry]VolumeProvisionerFactory)
)

// HasVolumeProvisioner returns true if name corresponds to an already
// registered volume provisioner.
func HasVolumeProvisioner(name v1.VolumeProvisionerRegistry) bool {
	volProvisionerRegMutex.Lock()
	defer volProvisionerRegMutex.Unlock()

	_, found := volProvisionerRegistry[name]
	return found
}

// RegisterVolumeProvisioner registers a persistent volume provisioner by the
// provisioner's name. This registers the provisioner name with the provisioner's
// instance creating function i.e. a Factory.
//
// NOTE:
//    Each implementation of persistent volume provisioner plugin need to call
// RegisterVolumeProvisioner inside their init() function.
func RegisterVolumeProvisioner(name v1.VolumeProvisionerRegistry, vpInstFactory VolumeProvisionerFactory) {
	volProvisionerRegMutex.Lock()
	defer volProvisionerRegMutex.Unlock()

	if _, found := volProvisionerRegistry[name]; found {
		glog.Fatalf("Persistent volume provisioner '%s' was registered twice", name)
	}

	glog.V(1).Infof("Registered '%s' as persistent volume provisioner", name)
	volProvisionerRegistry[name] = vpInstFactory
}

// GetVolumeProvisioner gets a new instance of the persistent volume
// provisioner as requested in the profileMap.
func GetVolumeProvisioner(profileMap map[string]string) (VolumeInterface, error) {
	pvp := v1.VolumeProvisionerName(profileMap)
	return GetVolumeProvisionerByName(v1.VolumeProvisionerRegistry(pvp))
}

// GetVolumeProvisionerByName gets a new instance of the named persistent volume
// provisioner or nil if the name is unknown.
func GetVolumeProvisionerByName(name v1.VolumeProvisionerRegistry) (VolumeInterface, error) {

	volProvisionerRegMutex.Lock()
	defer volProvisionerRegMutex.Unlock()

	if string(name) == "" {
		// Get the hard coded default
		name = v1.DefaultVolumeProvisionerName()
	}

	// Look it up in the registry
	vpInstFactory, found := volProvisionerRegistry[name]
	if !found {
		return nil, fmt.Errorf("'%s' is not registered as a persistent volume provisioner", name)
	}

	// TODO
	// Do not typecast to string. Maintain the type safety throughout.
	//
	// Persistent volume provisioner's instance creating function is invoked here
	// The persistent volume provisioner label is decided here. This label is common
	// to all persistent volume provisioner implementors.
	return vpInstFactory(string(v1.VolumeProvisionerNameLbl), string(name))
}
