package profiles

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
	k8sClientV1Beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type VolToK8sDeployTransType string

const (
	// NodeTaintVolToK8sDeploy is a transformer type used while
	// transforming a OpenEBS Volume to K8s Deployment
	TaintTolerationVolToK8sDeploy VolToK8sDeployTransType = "voltok8sdeploy/node-taint"
)

type DefaultTransformerVersion string

const (
  // TaintTolerationVolToK8sDeployVer represents the default version
  // for TaintTolerationVolToK8sDeploy transformer
  TaintTolerationVolToK8sDeployVer DefaultTransformerVersion = "1.0"
)

type VolToK8sDeployFactory func(vol *v1.Volume, deploy *k8sClientV1Beta1.Deployment) (K8sDeployTransformer, error)

// Registration is managed in a safe manner via these variables
var (
	volToK8sDeployMutex    sync.Mutex
	volToK8sDeployRegistry = make(map[VolToK8sDeployTransType]VolToK8sDeployFactory)
)

// HasVolToK8sDeployFactory returns true if transformer type corresponds to
// an already registered VolToK8sDeployFactory object
func HasVolToK8sDeployFactory(tt VolToK8sDeployTransType) bool {
	volToK8sDeployMutex.Lock()
	defer volToK8sDeployMutex.Unlock()

	_, found := volToK8sDeployRegistry[tt]
	return found
}

// RegisterVolToK8sDeployFactory registers a VolToK8sDeployFactory
// by the transformer type.
func RegisterVolToK8sDeployFactory(tt VolToK8sDeployTransType, tFactory VolToK8sDeployFactory) {
	volToK8sDeployMutex.Lock()
	defer volToK8sDeployMutex.Unlock()

	_, found := volToK8sDeployRegistry[tt]
	if found {
		glog.Fatalf("Duplicate registration attempt for VolToK8sDeployTransformer '%s' ", tt)
	}

	glog.Infof("Registered '%s' as transformer", tt)
	volToK8sDeployRegistry[tt] = tFactory
}

// GetVolToK8sDeployTrans creates a new instance of VolToK8sDeployTransformer
func GetVolToK8sDeployTrans(vol *v1.Volume, deploy *k8sClientV1Beta1.Deployment, tt VolToK8sDeployTransType) (K8sDeployTransformer, error) {
	volToK8sDeployMutex.Lock()
	defer volToK8sDeployMutex.Unlock()

	tFactory, found := volToK8sDeployRegistry[tt]
	if !found {
		return nil, fmt.Errorf("'%s' is not a registered VolToK8sDeployTransformer", tt)
	}

	// This functional invocation should result in creation of a new instance of
	// VolToK8sDeployTransformer
	return tFactory(vol, deploy)
}
