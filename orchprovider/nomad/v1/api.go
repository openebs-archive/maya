// This file transforms a Nomad scheduler as an orchestration
// platform for persistent volume placement. OpenEBS calls this as
// placement of storage pod.
package nomad

import (
	"fmt"

	"github.com/hashicorp/nomad/api"
)

// NomadApiInterface provides a means to issue APIs against a Nomad cluster.
// These APIs are futher categorized into Networking & Storage specific APIs.
type NomadApiInterface interface {

	// Name of the Nomad API implementor
	Name() string

	// This returns a concrete implementation of StorageApis
	StorageApis() (StorageApis, bool)
}

// nomadApi is an implementation of
//
//  nomad.NomadApiInterface interface &
//  nomad.StorageApis interface
//
// It composes NomadUtilInterface
type nomadApi struct {
	nUtil NomadUtilInterface
}

// newNomadApi provides a new instance of nomadApi
func newNomadApi() (*nomadApi, error) {

	nUtil, err := newNomadUtil()
	if err != nil {
		return nil, fmt.Errorf("Failed to create nomad api instance")
	}

	return &nomadApi{
		nUtil: nUtil,
	}, nil
}

// This is a plain nomad api implementor & hence the name
func (n *nomadApi) Name() string {
	return "nomadapi"
}

// nomadApi implements StorageApis, hence it returns self.
func (n *nomadApi) StorageApis() (StorageApis, bool) {
	return n, true
}

// StorageApis provides a means to communicate with Nomad Apis
// w.r.t storage.
//
// NOTE:
//    A Nomad job spec is treated as a persistent volume storage
// spec & then submitted to a Nomad deployment.
//
// NOTE:
//    Nomad has no notion of Persistent Volume.
type StorageApis interface {
	// Create makes a request to Nomad to create a storage resource
	CreateStorage(job *api.Job, profileMap map[string]string) (*api.Evaluation, error)

	// Delete makes a request to Nomad to delete the storage resource
	DeleteStorage(job *api.Job, profileMap map[string]string) (*api.Evaluation, error)

	// Info provides the storage information w.r.t the provided job name
	StorageInfo(jobName string, profileMap map[string]string) (*api.Job, error)
}

// Fetch info about a particular resource/job in Nomad cluster.
//
// NOTE:
//    Nomad does not have persistent volume as its first class citizen.
// Hence, this resource should exhibit storage characteristics. The validations
// for this should have been done at the volume provisioner.
func (n *nomadApi) StorageInfo(jobName string, profileMap map[string]string) (*api.Job, error) {

	nUtil := n.nUtil
	if nUtil == nil {
		return nil, fmt.Errorf("Nomad utility not initialized")
	}

	nClients, ok := nUtil.NomadClients()
	if !ok {
		return nil, fmt.Errorf("Nomad clients not supported by nomad utility '%s'", nUtil.Name())
	}

	nHttpClient, err := nClients.Http(profileMap)
	if err != nil {
		return nil, err
	}

	// Fetch the job info
	job, _, err := nHttpClient.Jobs().Info(jobName, &api.QueryOptions{})

	if err != nil {
		return nil, err
	}

	return job, nil
}

// Creates a resource/job in Nomad cluster.
//
// NOTE:
//    Nomad does not have persistent volume as its first class citizen.
// Hence, this resource should exhibit storage characteristics. The validations
// for this should have been done at the volume provisioner.
func (n *nomadApi) CreateStorage(job *api.Job, profileMap map[string]string) (*api.Evaluation, error) {

	nUtil := n.nUtil
	if nUtil == nil {
		return nil, fmt.Errorf("Nomad utility not initialized")
	}

	nClients, ok := nUtil.NomadClients()
	if !ok {
		return nil, fmt.Errorf("Nomad clients not supported by nomad utility '%s'", nUtil.Name())
	}

	nHttpClient, err := nClients.Http(profileMap)
	if err != nil {
		return nil, err
	}

	// Register a job & get its evaluation id
	evalID, _, err := nHttpClient.Jobs().Register(job, &api.WriteOptions{})

	if err != nil {
		return nil, err
	}

	// Get the evaluation details
	eval, _, err := nHttpClient.Evaluations().Info(evalID, &api.QueryOptions{})

	if err != nil {
		return nil, err
	}

	return eval, nil
}

// Remove a resource/job in Nomad cluster.
//
// NOTE:
//    Nomad does not have persistent volume as its first class citizen.
// Hence, this resource should exhibit storage characteristics. The validations
// for this should have been done at the volume provisioner.
func (n *nomadApi) DeleteStorage(job *api.Job, profileMap map[string]string) (*api.Evaluation, error) {

	nUtil := n.nUtil
	if nUtil == nil {
		return nil, fmt.Errorf("Nomad utility not initialized")
	}

	nClients, ok := nUtil.NomadClients()
	if !ok {
		return nil, fmt.Errorf("Nomad clients not supported by nomad utility '%s'", nUtil.Name())
	}

	nHttpClient, err := nClients.Http(profileMap)
	if err != nil {
		return nil, err
	}

	evalID, _, err := nHttpClient.Jobs().Deregister(*job.Name, &api.WriteOptions{})

	if err != nil {
		return nil, err
	}

	eval, _, err := nHttpClient.Evaluations().Info(evalID, &api.QueryOptions{})
	if err != nil {
		return nil, err
	}

	return eval, nil
}
