package server

import (
	"io"
	"log"
	"sync"

	"github.com/openebs/maya/cmd/maya-apiserver/app/config"
	"github.com/openebs/maya/orchprovider"
	"github.com/openebs/maya/orchprovider/k8s/v1"
	"github.com/openebs/maya/orchprovider/nomad/v1"
	"github.com/openebs/maya/types/v1"
	"github.com/openebs/maya/volume/provisioners"
	"github.com/openebs/maya/volume/provisioners/jiva"
)

// MayaApiServer is a long running stateless daemon that runs
// at openebs maya master(s)
type MayaApiServer struct {
	config    *config.MayaConfig
	logger    *log.Logger
	logOutput io.Writer

	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

// NewMayaApiServer is used to create a new maya api server
// with the given configuration
func NewMayaApiServer(config *config.MayaConfig, logOutput io.Writer) (*MayaApiServer, error) {

	ms := &MayaApiServer{
		config:     config,
		logger:     log.New(logOutput, "", log.LstdFlags|log.Lmicroseconds),
		logOutput:  logOutput,
		shutdownCh: make(chan struct{}),
	}

	err := ms.BootstrapPlugins()
	if err != nil {
		return nil, err
	}

	return ms, nil
}

// TODO
// Create a Bootstrap interface that facilitates initialization
// Create another Bootstraped interface that provides the initialized instances
// Perhaps at lib/bootstrap
// MayaServer struct will make use of above interfaces & hence specialized
// structs that cater to bootstraping & bootstraped features.
//
// NOTE:
//    The current implementation is tightly coupled & cannot be unit tested.
//
// NOTE:
//    Mayaserver should be entrusted to registering all possible variants of
// volume plugins.
//
// A volume plugin variant is composed of:
//    volume plugin + orchestrator of volume plugin + region of orchestrator
//
// In addition, Mayaserver should initialize the `default volume plugin`
// instance with its `default orchestrator` & `default region` of the
// orchestrator. User initiated requests requiring specific variants should be
// initialized at runtime.
func (ms *MayaApiServer) BootstrapPlugins() error {
	// Register persistent volume provisioner(s)
	isJivaProvisionerReg := provisioners.HasVolumeProvisioner(v1.JivaVolumeProvisioner)
	if !isJivaProvisionerReg {
		provisioners.RegisterVolumeProvisioner(
			// Registration entry when Jiva is a persistent volume provisioner
			v1.JivaVolumeProvisioner,

			// Below is a callback function that creates a new instance of jiva as persistent
			// volume provisioner
			func(label, name string) (provisioners.VolumeInterface, error) {
				return jiva.NewJivaProvisioner(label, name)
			})
	}

	// Register orchestrator(s)
	isK8sOrchReg := orchprovider.HasOrchestrator(v1.K8sOrchestrator)
	if !isK8sOrchReg {
		orchprovider.RegisterOrchestrator(
			// Registration entry when Kubernetes is the orchestrator provider
			v1.K8sOrchestrator,
			// Below is a callback function that creates a new instance of Kubernetes
			// orchestration provider
			func(label v1.NameLabel, name v1.OrchProviderRegistry) (orchprovider.OrchestratorInterface, error) {
				return k8s.NewK8sOrchestrator(label, name)
			})
	}

	isNomadOrchReg := orchprovider.HasOrchestrator(v1.NomadOrchestrator)
	if !isNomadOrchReg {
		orchprovider.RegisterOrchestrator(
			// Registration entry when Nomad is the orchestrator provider
			v1.NomadOrchestrator,
			// Below is a callback function that creates a new instance of Nomad
			// orchestration provider
			func(label v1.NameLabel, name v1.OrchProviderRegistry) (orchprovider.OrchestratorInterface, error) {
				return nomad.NewNomadOrchestrator(label, name)
			})
	}

	return nil
}

// Shutdown is used to terminate MayaServer.
func (ms *MayaApiServer) Shutdown() error {

	ms.shutdownLock.Lock()
	defer ms.shutdownLock.Unlock()

	ms.logger.Println("[INFO] maya api server: requesting shutdown")

	if ms.shutdown {
		return nil
	}

	ms.logger.Println("[INFO] maya api server: shutdown complete")
	ms.shutdown = true

	close(ms.shutdownCh)

	return nil
}

// Leave is used gracefully exit.
func (ms *MayaApiServer) Leave() error {

	ms.logger.Println("[INFO] maya api server: exiting gracefully")

	// Nothing as of now
	return nil
}
