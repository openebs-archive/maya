// This file plugs the following:
//
//    1. Generic orchprovider &
//    2. Nomad orchprovider
package nomad

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/openebs/maya/orchprovider"
	"github.com/openebs/maya/types/v1"
	volProfile "github.com/openebs/maya/volume/profiles"
)

// NomadOrchestrator is a concrete representation of following
// interfaces:
//
//  1. orchprovider.OrchestratorInterface, &
//  2. orchprovider.StoragePlacements
type NomadOrchestrator struct {

	// label assigned to this orchestrator
	label string

	// Name of this orchestrator
	name string

	// The region where this orchestrator is deployed
	// This is set during the initilization time.
	region string

	// nStorApis represents an instance capable of invoking
	// storage related APIs
	nStorApis StorageApis
}

// NewNomadOrchestrator provides a new instance of NomadOrchestrator. This is
// invoked during binary startup.
//func NewNomadOrchestrator(name v1.OrchProviderRegistry, region string, config io.Reader) (orchprovider.OrchestratorInterface, error) {
func NewNomadOrchestrator(label v1.NameLabel, name v1.OrchProviderRegistry) (orchprovider.OrchestratorInterface, error) {

	glog.Infof("Building nomad orchestration provider")

	if label == "" {
		return nil, fmt.Errorf("Label is missing while building nomad orchestrator")
	}

	if name == "" {
		return nil, fmt.Errorf("Name is missing while building nomad orchestrator")
	}

	// Get a new instance of Nomad API
	nApi, err := newNomadApi()
	if err != nil {
		return nil, err
	}

	// Get Nomad's storage specific API provider
	nStorApis, ok := nApi.StorageApis()
	if !ok {
		return nil, fmt.Errorf("Storage APIs not supported in nomad api instance '%s'", nApi.Name())
	}

	// build the orchestrator instance
	nOrch := &NomadOrchestrator{
		label:     string(label),
		nStorApis: nStorApis,
		name:      string(name),
	}

	return nOrch, nil
}

// Label provides the label assigned against this orchestrator. This is used
// along with Name() method.
// This is an implementation of the orchprovider.OrchestratorInterface interface.
func (n *NomadOrchestrator) Label() string {
	return n.label
}

// Name provides the name of this orchestrator.
// This is an implementation of the orchprovider.OrchestratorInterface interface.
func (n *NomadOrchestrator) Name() string {
	return n.name
}

// Region provides the region where this orchestrator is running.
// This is an implementation of the orchprovider.OrchestratorInterface interface.
func (n *NomadOrchestrator) Region() string {
	return n.region
}

// StorageOps deals with storage related operations e.g. scheduling, placements,
// removal, etc. of persistent volume containers. The low level workings are
// delegated to the orchestration provider.
//
// NOTE:
//    This is orchestration provider's implementation of
// orchprovider.OrchestratorInterface interface.
func (n *NomadOrchestrator) StorageOps() (orchprovider.StorageOps, bool) {
	return n, true
}

// ReadStorage will fetch information about the persistent volume
func (n *NomadOrchestrator) ReadStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.Volume, error) {
	pvc, err := volProProfile.PVC()
	if err != nil {
		return nil, err
	}

	jobName, err := PvcToJobName(pvc)
	if err != nil {
		return nil, err
	}

	job, err := n.nStorApis.StorageInfo(jobName, pvc.Labels)
	if err != nil {
		return nil, err
	}

	return JobToPv(job)
}

// AddStorage will add persistent volume running as containers. In OpenEBS
// terms AddStorage will add a VSM.
func (n *NomadOrchestrator) AddStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.Volume, error) {
	// TODO
	// This is jiva specific
	// Move this entire logic to a separate package that will couple jiva
	// provisioner with nomad orchestrator

	pvc, err := volProProfile.PVC()
	if err != nil {
		return nil, err
	}

	job, err := PvcToJob(pvc)
	if err != nil {
		return nil, err
	}

	eval, err := n.nStorApis.CreateStorage(job, pvc.Labels)
	if err != nil {
		return nil, err
	}

	glog.Infof("Volume '%s' was placed for provisioning with eval '%v'", *job.Name, eval)

	return JobEvalToPv(*job.Name, eval)
}

// DeleteStorage will remove the VSM.
func (n *NomadOrchestrator) DeleteStorage(volProProfile volProfile.VolumeProvisionerProfile) (bool, error) {
	pvc, err := volProProfile.PVC()
	if err != nil {
		return false, err
	}

	job, err := MakeJob(pvc.Name)
	if err != nil {
		return false, err
	}

	eval, err := n.nStorApis.DeleteStorage(job, pvc.Labels)

	if err != nil {
		return false, err
	}

	glog.Infof("Volume '%s' was placed for removal with eval '%v'", pvc.Name, eval)

	_, err = JobEvalToPv(*job.Name, eval)

	if err != nil {
		return false, err
	}

	return true, nil
}

// ListStorage will list a collections of VSMs
func (n *NomadOrchestrator) ListStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.VolumeList, error) {
	return nil, fmt.Errorf("ListStorage is not implemented by '%s: %s'", n.Label(), n.Name())
}
