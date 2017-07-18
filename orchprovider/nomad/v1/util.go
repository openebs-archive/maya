package nomad

import (
	"github.com/golang/glog"
	"github.com/hashicorp/nomad/api"
	"github.com/openebs/maya/types/v1"
)

// NomadUtilInterface is an abstraction over Hashicorp's Nomad properties &
// communication utilities.
type NomadUtilInterface interface {

	// Name of nomad utility
	Name() string

	// This is a builder for NomadClients interface. Will return
	// false if not supported.
	NomadClients() (NomadClients, bool)
}

// NomadClients is an abstraction over various connection modes (http, rpc)
// to Nomad. Http client is currently supported.
//
// NOTE:
//    This abstraction makes use of Nomad's api package. Nomad's api
// package provides:
//
// 1. http client abstraction &
// 2. structures that can send http requests to Nomad's APIs.
type NomadClients interface {
	// Http returns the http client that is capable to communicate
	// with Nomad
	Http(profileMap map[string]string) (*api.Client, error)
}

// nomadUtil is the concrete implementation for
//
// 1. nomad.NomadClients interface
// 2. nomad.NomadNetworks interface
type nomadUtil struct {
	// profileMap is a set of key value user entries provided to
	// the current request
	profileMap map[string]string

	caCert     string
	caPath     string
	clientCert string
	clientKey  string
	insecure   bool
}

// newNomadUtil provides a new instance of nomadUtil
func newNomadUtil() (*nomadUtil, error) {
	return &nomadUtil{}, nil
}

// This is a plain nomad utility & hence the name
func (m *nomadUtil) Name() string {
	return "nomadutil"
}

// nomadUtil implements NomadClients interface. Hence it returns
// self
func (m *nomadUtil) NomadClients() (NomadClients, bool) {
	return m, true
}

// Client is used to initialize and return a new API client capable
// of calling Nomad APIs.
func (m *nomadUtil) Http(profileMap map[string]string) (*api.Client, error) {
	// Nomad API client config
	apiCConf := api.DefaultConfig()

	reg := v1.GetOrchestratorRegion(profileMap)
	apiCConf.Region = reg

	addr := v1.GetOrchestratorAddress(profileMap)
	apiCConf.Address = addr

	glog.Infof("Nomad will be reached at 'region: %s' 'address: %s'", apiCConf.Region, apiCConf.Address)

	// If we need custom TLS configuration, then set it
	// TODO
	// Need to check the best way to get these !!
	if m.caCert != "" || m.caPath != "" || m.clientCert != "" || m.clientKey != "" || m.insecure {
		t := &api.TLSConfig{
			CACert:     m.caCert,
			CAPath:     m.caPath,
			ClientCert: m.clientCert,
			ClientKey:  m.clientKey,
			Insecure:   m.insecure,
		}
		apiCConf.TLSConfig = t
	}

	// This has the http address & authentication details
	// required to invoke Nomad APIs
	return api.NewClient(apiCConf)
}
