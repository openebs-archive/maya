// This file handles jiva storage logic related to mayaserver's orchestration
// provider.
//
// NOTE:
//    jiva storage delegates the provisioning, placement & other operational
// aspects to an orchestration provider. Some of the orchestration providers
// can be Kubernetes, Nomad, etc.
package jiva

import (
	"fmt"

	"github.com/openebs/maya/orchprovider"
	"github.com/openebs/maya/types/v1"
	vProfile "github.com/openebs/maya/volume/profiles"
)

type JivaInterface interface {
	// Name provides the name of the JivaInterface implementor
	Name() string

	// JivaProProfile sets an instance of VolumeProvisionerProfile.
	//
	// Note:
	//    It returns false if setting not supported
	JivaProProfile(vProfile.VolumeProvisionerProfile) (bool, error)

	// StorageOps provides an instance of StorageOps. It will return false
	// if Storage operations is not supported.
	StorageOps() (StorageOps, bool)
}

// StorageOps abstracts the storage specific operations of jiva persistent
// volume provisioner.
type StorageOps interface {
	// Info / Read operation
	ReadStorage(*v1.Volume) (*v1.Volume, error)

	// Add operation
	AddStorage(*v1.Volume) (*v1.Volume, error)

	// ListStorage will list a collection of persistent volumes
	ListStorage() (*v1.VolumeList, error)

	// Delete operation
	RemoveStorage() (bool, error)
}

// jivaUtil is the concrete implementation for
//
//  JivaInterface interface &
//  StorageOps interface
type jivaUtil struct {
	// jivaProProfile holds persistent volume provisioner's profile
	// This can be set lazily.
	jivaProProfile vProfile.VolumeProvisionerProfile
}

// newJivaProUtil provides a new instance of JivaInterface that can execute jiva
// persistent volume provisioner's low level tasks.
func newJivaProUtil() (JivaInterface, error) {
	return &jivaUtil{}, nil
}

// Name provides the name assigned to this instance of JivaInterface
//
// Note:
//    There can be multiple instances (due to unique requests) which will have
// this provided name. Name is not required to be unique.
func (j *jivaUtil) Name() string {
	return "JivaProvisionerUtil"
}

// JivaProProfile sets the persistent volume provisioner's profile. This returns
// true as its first argument as jiva supports volume provisioner profile.
func (j *jivaUtil) JivaProProfile(volProProfile vProfile.VolumeProvisionerProfile) (bool, error) {

	if volProProfile == nil {
		return true, fmt.Errorf("Nil persistent volume provisioner profile was provided to '%s'", j.Name())
	}

	j.jivaProProfile = volProProfile
	return true, nil
}

// StorageOps method provides an instance of StorageOps interface
//
// NOTE:
//  jivaUtil implements StorageOps interface. Hence it returns self.
func (j *jivaUtil) StorageOps() (StorageOps, bool) {
	return j, true
}

// ListStorage fetches a collection of jiva persistent volumes.
// It gets the appropriate orchestration provider to delegate further execution.
func (j *jivaUtil) ListStorage() (*v1.VolumeList, error) {
	// TODO
	// Move the below set of validations to StorageOps()
	if j.jivaProProfile == nil {
		return nil, fmt.Errorf("Volume provisioner profile not set in '%s'", j.Name())
	}

	oName, supported, err := j.jivaProProfile.Orchestrator()
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, fmt.Errorf("No orchestrator support in '%s:%s'", j.jivaProProfile.Label(), j.jivaProProfile.Name())
	}

	orchestrator, err := orchprovider.GetOrchestrator(oName)
	if err != nil {
		return nil, err
	}

	storageOrchestrator, ok := orchestrator.StorageOps()

	if !ok {
		return nil, fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return storageOrchestrator.ListStorage(j.jivaProProfile)
}

// TODO
// Remove the use of pvc
//
// ReadStorage fetches details of a jiva persistent volume.
// It gets the appropriate orchestration provider to delegate further execution.
func (j *jivaUtil) ReadStorage(pvc *v1.Volume) (*v1.Volume, error) {

	// TODO
	// Move the below set of validations to StorageOps()
	if j.jivaProProfile == nil {
		return nil, fmt.Errorf("Volume provisioner profile not set in '%s'", j.Name())
	}

	oName, supported, err := j.jivaProProfile.Orchestrator()
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, fmt.Errorf("No orchestrator support in '%s:%s'", j.jivaProProfile.Label(), j.jivaProProfile.Name())
	}

	orchestrator, err := orchprovider.GetOrchestrator(oName)
	if err != nil {
		return nil, err
	}

	storageOrchestrator, ok := orchestrator.StorageOps()

	if !ok {
		return nil, fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return storageOrchestrator.ReadStorage(j.jivaProProfile)
}

// AddStorage adds a jiva persistent volume
func (j *jivaUtil) AddStorage(pvc *v1.Volume) (*v1.Volume, error) {

	// TODO
	// Move the below set of validations to StorageOps() method
	if j.jivaProProfile == nil {
		return nil, fmt.Errorf("Provisioner profile not found for '%s'", j.Name())
	}

	oName, supported, err := j.jivaProProfile.Orchestrator()
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, fmt.Errorf("No orchestrator support in '%s:%s'", j.jivaProProfile.Label(), j.jivaProProfile.Name())
	}

	orchestrator, err := orchprovider.GetOrchestrator(oName)
	if err != nil {
		return nil, err
	}

	storageOrchestrator, ok := orchestrator.StorageOps()

	if !ok {
		return nil, fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return storageOrchestrator.AddStorage(j.jivaProProfile)
}

// RemoveStorage removes the peristent storage
func (j *jivaUtil) RemoveStorage() (bool, error) {
	// TODO
	// Move the below set of validations to StorageOps()
	if j.jivaProProfile == nil {
		return false, fmt.Errorf("Volume provisioner profile not set in '%s'", j.Name())
	}

	oName, supported, err := j.jivaProProfile.Orchestrator()
	if err != nil {
		return false, err
	}

	if !supported {
		return false, fmt.Errorf("No orchestrator support in '%s:%s'", j.jivaProProfile.Label(), j.jivaProProfile.Name())
	}

	orchestrator, err := orchprovider.GetOrchestrator(oName)
	if err != nil {
		return false, err
	}

	storageOrchestrator, ok := orchestrator.StorageOps()

	if !ok {
		return false, fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return storageOrchestrator.DeleteStorage(j.jivaProProfile)
}
