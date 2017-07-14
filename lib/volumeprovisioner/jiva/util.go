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

	"github.com/openebs/mayaserver/lib/api/v1"
	"github.com/openebs/mayaserver/lib/orchprovider"
	vProfile "github.com/openebs/mayaserver/lib/profile/volumeprovisioner"
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
	ReadStorage(*v1.PersistentVolumeClaim) (*v1.PersistentVolume, error)

	// Add operation
	AddStorage(*v1.PersistentVolumeClaim) (*v1.PersistentVolume, error)

	// ListStorage will list a collection of persistent volumes
	ListStorage() (*v1.PersistentVolumeList, error)

	// Delete operation
	RemoveStorage() error
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
func (j *jivaUtil) ListStorage() (*v1.PersistentVolumeList, error) {
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
func (j *jivaUtil) ReadStorage(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolume, error) {

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
func (j *jivaUtil) AddStorage(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolume, error) {

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
func (j *jivaUtil) RemoveStorage() error {
	// TODO
	// Move the below set of validations to StorageOps()
	if j.jivaProProfile == nil {
		return fmt.Errorf("Volume provisioner profile not set in '%s'", j.Name())
	}

	oName, supported, err := j.jivaProProfile.Orchestrator()
	if err != nil {
		return err
	}

	if !supported {
		return fmt.Errorf("No orchestrator support in '%s:%s'", j.jivaProProfile.Label(), j.jivaProProfile.Name())
	}

	orchestrator, err := orchprovider.GetOrchestrator(oName)
	if err != nil {
		return err
	}

	storageOrchestrator, ok := orchestrator.StorageOps()

	if !ok {
		return fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return storageOrchestrator.DeleteStorage(j.jivaProProfile)
}
