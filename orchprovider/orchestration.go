// This file defines interfaces that determines an orchestrator w.r.t maya api
// server. All the features that maya api server wants from an orchestrator is
// defined in these set of interfaces.
package orchprovider

import (
	oe_api_v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/types/v1"
	volProfile "github.com/openebs/maya/volume/profiles"
	k8sApiV1 "k8s.io/api/core/v1"
)

// OrchestrationInterface is an interface abstraction of a real orchestrator.
// It represents an abstraction that serves operations feasible from an
// orchestrator.
type OrchestratorInterface interface {
	// Label assigned against the orchestration provider
	Label() string

	// Name of the orchestration provider
	Name() string

	// Region where this orchestration provider is running/deployed
	Region() string

	// StorageOps gets the instance that deals with storage related operations.
	// Will return false if not supported.
	StorageOps() (StorageOps, bool)

	// PolicyOps gets the instance that deals with volume policy related operations.
	// Will return false if not supported.
	PolicyOps(vol *v1.Volume) (PolicyOps, bool, error)
}

// StorageOps exposes various storage related operations that deals with
// storage placements, scheduling, etc. The low level work is in turn delegated
// to the respective orchestrator.
type StorageOps interface {

	// AddStorage will add persistent volume running as containers
	AddStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.Volume, error)

	// DeleteStorage will remove the persistent volume
	DeleteStorage(volProProfile volProfile.VolumeProvisionerProfile) (bool, error)

	// ReadStorage will fetch information about the persistent volume
	ReadStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.Volume, error)

	// ListStorage will list a collection of VSMs in a given context e.g. namespace
	// if working in a K8s setup, etc.
	ListStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.VolumeList, error)
}

// PolicyOps exposes various volume policy related operations. Volume policies
// influence volume placements, provisioning, backup, etc. decisions.
type PolicyOps interface {

	// SCPolicies will fetch volume policies from a particular StorageClass
	SCPolicies() (map[string]string, error)

	// PVCPolicies will fetch volume policies from a particular
	// Persistent Volume Claim
	PVCPolicies() (k8sApiV1.PersistentVolumeClaimSpec, error)

	// SPPolicies will fetch volume policies from a particular StoragePool
	SPPolicies() (oe_api_v1alpha1.StoragePoolSpec, error)
}
