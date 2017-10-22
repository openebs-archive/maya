// This file abstracts & exposes persistent volume provisioner features. All
// maya api server's persistent volume provisioners need to implement these
// contracts.
package provisioners

import (
	"github.com/openebs/maya/types/v1"
)

// VolumeInterface abstracts the persistent volume features of any persistent
// volume provisioner.
//
// NOTE:
//    maya api server can make use of any persistent volume provisioner & execute
// corresponding volume related operations.
type VolumeInterface interface {
	// Label assigned against this persistent volume provisioner
	Label() string

	// Name of the persistent volume provisioner
	Name() string

	// Profile will set the persistent volume provisioner's profile
	//
	// NOTE:
	//    Will return false if profile is not supported by the persistent
	// volume provisioner.
	//
	// NOTE:
	//    This is used to set the persistent volume provisioner profile lazily
	// i.e. much after the initialization of persistent volume provisioner instance.
	// It is assumed that persistent volume claim will be available at the time of
	// invoking this method.
	Profile(*v1.Volume) (bool, error)

	// Remover gets the instance capable of deleting volumes w.r.t this
	// persistent volume provisioner.
	//
	// Note:
	//    Will return false if deletion of volumes is not supported by the
	// persistent volume provisioner.
	Remover() (Remover, bool, error)

	// Reader gets the instance capable of providing persistent volume information
	// w.r.t this persistent volume provisioner.
	//
	// Note:
	//    Will return false if providing persistent volume information is not
	// supported by this persistent volume provisioner.
	Reader() (Reader, bool)

	// Adder gets the instance capable of creating a persistent volume
	// w.r.t this persistent volume provisioner.
	//
	// Note:
	//    Will return false if creating persistent volume is not
	// supported by this persistent volume provisioner.
	Adder() (Adder, bool)

	// Lister gets the instance capable of listing persistent volumes
	// w.r.t this persistent volume provisioner.
	//
	// Note:
	//    Will return false if listing persistent volumes is not
	// supported by this persistent volume provisioner.
	Lister() (Lister, bool, error)
}

// Lister interface abstracts listing of persistent volumes from a persistent
// volume provisioner.
type Lister interface {
	// List fetches a collection of persistent volumes created by this volume
	// provisioner
	List() (*v1.VolumeList, error)
}

// Reader interface abstracts fetching of persistent volume related information
// from a persistent volume provisioner.
type Reader interface {
	// Read fetches the volume details from the persistent volume
	// provisioner.
	Read(*v1.Volume) (*v1.Volume, error)
}

// Adder interface abstracts creation of persistent volume from a persistent
// volume provisioner.
type Adder interface {
	// Add creates a new persistent volume
	Add(*v1.Volume) (*v1.Volume, error)
}

// Remover interface abstracts deletion of volume of a persistent volume
// provisioner.
type Remover interface {
	// Delete tries to delete a volume of a persistent volume provisioner.
	Remove() (bool, error)
}
