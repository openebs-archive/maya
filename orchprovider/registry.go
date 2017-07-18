// This file provides orchestrator provider's registry related features.
//
// NOTE:
//    This is the new file w.r.t the deprecated plugins.go file
package orchprovider

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
)

type OrchProviderFactory func(label v1.NameLabel, name v1.OrchProviderRegistry) (OrchestratorInterface, error)

// Registration is managed in a safe manner via these variables
var (
	orchProviderRegMutex sync.Mutex
	orchProviderRegistry = make(map[v1.OrchProviderRegistry]OrchProviderFactory)
)

// HasOrchestrator returns true if name corresponds to an already
// registered orchestration provider.
func HasOrchestrator(name v1.OrchProviderRegistry) bool {
	orchProviderRegMutex.Lock()
	defer orchProviderRegMutex.Unlock()

	_, found := orchProviderRegistry[name]
	return found
}

// RegisterOrchestrator registers a orchestration provider by the provider's name.
// This registers the orchestrator provider name with the provider's instance
// creating function i.e. a Factory.
//
// NOTE:
//    Each implementation of orchestrator plugin need to call
// RegisterOrchestrator inside their init() function.
func RegisterOrchestrator(name v1.OrchProviderRegistry, oInstFactory OrchProviderFactory) {
	orchProviderRegMutex.Lock()
	defer orchProviderRegMutex.Unlock()

	_, found := orchProviderRegistry[name]
	if found {
		glog.Fatalf("Duplicate orchestration provider '%s' registration", name)
	}

	//glog.V(1).Infof("Registered '%s' as orchestration provider", name)
	glog.Infof("Registered '%s' as orchestration provider", name)
	orchProviderRegistry[name] = oInstFactory
}

// GetOrchestrator creates a new instance of the named orchestration provider,
// or nil if the name is unknown.
func GetOrchestrator(name v1.OrchProviderRegistry) (OrchestratorInterface, error) {
	orchProviderRegMutex.Lock()
	defer orchProviderRegMutex.Unlock()

	oInstFactory, found := orchProviderRegistry[name]
	if !found {
		return nil, fmt.Errorf("'%s' is not a registered orchestration provider", name)
	}

	// Orchestration provider's instance creating function is invoked here
	// The orchestration provider label is decided here. The label is common to all
	// orchestration provider implementors.
	return oInstFactory(v1.OrchestratorNameLbl, name)
}
