// This file registers jiva as maya api server's persistent volume provisioner plugin.
package jiva

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
	volProfile "github.com/openebs/maya/volume/profiles"
	"github.com/openebs/maya/volume/provisioners"
)

// TODO
// Rename to jivaProvisioner ??
//
// jivaStor is the concrete implementation that implements
// following interfaces:
//
//  1. provisioners.VolumeInterface interface
//  2. provisioners.Provisioner interface
//  3. provisioners.Deleter interface
type jivaStor struct {
	// label assigned against this jiva persistent volume provisioner
	label string

	// name is the name of this jiva persistent volume provisioner.
	name string

	// isProfileSet flags if the volume provisioner profile is set
	isProfileSet bool

	// jivaProUtil enables all low level jiva persistent volume provisioner features.
	jivaProUtil JivaInterface
}

// NewJivaProvisioner generates a new instance of jiva based persistent volume
// provisioner plugin.
//
// Note:
//    This function aligns with the callback function signature
func NewJivaProvisioner(label, name string) (provisioners.VolumeInterface, error) {

	if label == "" {
		return nil, fmt.Errorf("Label not provided for jiva persistent volume provisioner instance")
	}

	if name == "" {
		return nil, fmt.Errorf("Name not provided for jiva persistent volume provisioner instance")
	}

	jProUtil, err := newJivaProUtil()
	if err != nil {
		return nil, err
	}

	glog.Infof("Building new instance of jiva persistent volume provisioner '%s:%s'", label, name)

	// build the provisioner instance
	jivaStor := &jivaStor{
		label:       label,
		name:        name,
		jivaProUtil: jProUtil,
	}

	return jivaStor, nil
}

// Label returns the label assigned against this jiva persistent volume provisioner
//
// NOTE:
//    This is a contract implementation of volume.VolumeInterface
func (j *jivaStor) Label() string {
	return j.label
}

// Name returns the namespaced name of this volume
//
// NOTE:
//    This is a contract implementation of volume.VolumeInterface
func (j *jivaStor) Name() string {
	return j.name
}

// Profile sets the persistent volume provisioner profile against this jiva
// volume provisioner.
//
// NOTE:
//    This method is expected to be invoked when pvc is available. In other
// words this is lazily invoked after the creation of jivaStor.
func (j *jivaStor) Profile(pvc *v1.Volume) (bool, error) {
	// Get the persistent volume provisioner profile
	vProfl, err := volProfile.GetVolProProfileByPVC(pvc)
	if err != nil {
		return true, err
	}

	// Set the above persistent volume provisioner profile
	supported, err := j.jivaProUtil.JivaProProfile(vProfl)
	if err == nil && supported {
		j.isProfileSet = true
	}

	return supported, err
}

// isProfile indicates if volume provisioner profile was set earlier
func (j *jivaStor) isProfile() bool {
	return j.isProfileSet
}

// Reader provides a instance of volume.Reader interface.
// Since jivaStor implements volume.Reader, it returns self.
//
// NOTE:
//    This is one of the concrete implementations of volume.VolumeInterface
func (j *jivaStor) Reader() (provisioners.Reader, bool) {
	return j, true
}

// Adder provides a instance of volume.Adder interface.
// Since jivaStor implements volume.Adder, it returns self.
//
// NOTE:
//    This is one of the concrete implementations of volume.VolumeInterface
func (j *jivaStor) Adder() (provisioners.Adder, bool) {
	return j, true
}

// Lister provides a instance of volume.Lister interface.
// Since jivaStor implements volume.Lister, it returns self.
//
// NOTE:
//    This is one of the concrete implementations of volume.VolumeInterface
func (j *jivaStor) Lister() (provisioners.Lister, bool, error) {
	if j.jivaProUtil == nil {
		return nil, true, fmt.Errorf("Jiva provisioner util is not set at 'jiva provisioner: %s:%s'", j.Label(), j.Name())
	}

	if !j.isProfile() {
		return nil, true, fmt.Errorf("Jiva provisioner profile is not set at 'jiva provisioner: %s:%s' with 'provisioner util: %s'", j.Label(), j.Name(), j.jivaProUtil.Name())
	}

	// Lister depends on jiva provisioner util's StorageOps
	_, supported := j.jivaProUtil.StorageOps()
	if !supported {
		return nil, true, fmt.Errorf("Storage operations not supported by 'jiva provisioner: %s:%s' with 'provisioner util: %s'", j.Label(), j.Name(), j.jivaProUtil.Name())
	}

	return j, true, nil
}

// Remover provides a instance of volume.Remover interface.
// Since jivaStor implements volume.Remover, it returns self.
//
// NOTE:
//    This is one of the concrete implementations of volume.VolumeInterface
func (j *jivaStor) Remover() (provisioners.Remover, bool, error) {
	if j.jivaProUtil == nil {
		return nil, true, fmt.Errorf("Jiva provisioner util is not set at 'jiva provisioner: %s:%s'", j.Label(), j.Name())
	}

	if !j.isProfile() {
		return nil, true, fmt.Errorf("Jiva provisioner profile is not set at 'jiva provisioner: %s:%s' with 'provisioner util: %s'", j.Label(), j.Name(), j.jivaProUtil.Name())
	}

	// Remover depends on jiva provisioner util's StorageOps
	_, supported := j.jivaProUtil.StorageOps()
	if !supported {
		return nil, true, fmt.Errorf("Storage operations not supported by 'jiva provisioner: %s:%s' with 'provisioner util: %s'", j.Label(), j.Name(), j.jivaProUtil.Name())
	}

	return j, true, nil
}

// List provides a collection of jiva persistent volumes
//
// NOTE:
//    This is expected to be invoked after setting the volume provisioner
// profile
//
// NOTE:
//    This is a concrete implementation of volume.Lister interface
func (j *jivaStor) List() (*v1.VolumeList, error) {
	// Delegate to the storage util
	storOps, _ := j.jivaProUtil.StorageOps()

	return storOps.ListStorage()
}

// TODO
// pvc need not be passed at all as it should have been set via Profile()
//
// Read provides information about a jiva persistent volume
//
// NOTE:
//    This is expected to be invoked after setting the volume provisioner
// profile
//
// NOTE:
//    This is a concrete implementation of volume.Informer interface
func (j *jivaStor) Read(pvc *v1.Volume) (*v1.Volume, error) {
	// TODO
	// Move the validations to j.Reader()

	// Delegate to the storage util
	storOps, supported := j.jivaProUtil.StorageOps()
	if !supported {
		return nil, fmt.Errorf("Storage operations not supported in '%s:%s' '%s'", j.Label(), j.Name(), j.jivaProUtil.Name())
	}

	return storOps.ReadStorage(pvc)
}

// TODO
// pvc need not be passed at all as it should have been set via Profile()
//
// Add creates a new jiva persistent volume
//
// NOTE:
//    This is expected to be invoked after setting the volume provisioner
// profile
//
// NOTE:
//    This is a concrete implementation of volume.Adder interface
func (j *jivaStor) Add(pvc *v1.Volume) (*v1.Volume, error) {
	// TODO
	// Move the validations to j.Adder()

	// Delegate to the storage util
	storOps, supported := j.jivaProUtil.StorageOps()
	if !supported {
		return nil, fmt.Errorf("Storage operations not supported in '%s:%s' '%s'", j.Label(), j.Name(), j.jivaProUtil.Name())
	}

	return storOps.AddStorage(pvc)
}

// Remove removes a jiva volume
//
// NOTE:
//    This is a concrete implementation of volume.Remover interface
func (j *jivaStor) Remove() (bool, error) {

	// Delegate to the storage util
	storOps, _ := j.jivaProUtil.StorageOps()

	return storOps.RemoveStorage()
}
