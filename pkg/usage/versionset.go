package usage

import (
	k8sapi "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	openebsversion "github.com/openebs/maya/pkg/version"
)

var (
	clusterUUID    env.ENVKey = "OPENEBS_IO_USAGE_UUID"
	clusterVersion env.ENVKey = "OPENEBS_IO_K8S_VERSION"
	clusterArch    env.ENVKey = "OPENEBS_IO_K8S_ARCH"
	openEBSversion env.ENVKey = "OPENEBS_IO_VERSION_TAG"
	nodeType       env.ENVKey = "OPENEBS_IO_NODE_TYPE"
)

// versionSet is a struct which stores (sort of) fixed information about a
// k8s environment
type versionSet struct {
	id             string // OPENEBS_IO_USAGE_UUID
	k8sVersion     string // OPENEBS_IO_K8S_VERSION
	k8sArch        string // OPENEBS_IO_K8S_ARCH
	openebsVersion string // OPENEBS_IO_VERSION_TAG
	nodeType       string // OPENEBS_IO_NODE_TYPE
}

// fetchAndSetVersion consumes the Kubernetes API to get environment constants
// and returns a versionSet struct
func (v *versionSet) fetchAndSetVersion() error {
	var err error
	v.id, err = getUUIDbyNS("default")
	if err != nil {
		return err
	}
	env.Set(clusterUUID, v.id)

	k8s, err := k8sapi.GetServerVersion()
	if err != nil {
		return err
	}
	// eg. linux/amd64
	v.k8sArch = k8s.Platform
	v.k8sVersion = k8s.GitVersion
	env.Set(clusterArch, v.k8sArch)
	env.Set(clusterVersion, v.k8sVersion)
	v.nodeType, err = k8sapi.GetOSAndKernelVersion()
	env.Set(nodeType, v.nodeType)
	if err != nil {
		return err
	}
	v.openebsVersion = openebsversion.GetVersionDetails()
	env.Set(openEBSversion, v.openebsVersion)
	return nil
}

// getVersion is a wrapper over fetchAndSetVersion
func (v *versionSet) getVersion() error {
	// If ENVs aren't set, fetch the required values from the
	// K8s APIserver
	if _, present := env.Lookup(openEBSversion); !present {
		if err := v.fetchAndSetVersion(); err != nil {
			return err
		}
	}
	// Fetch data from ENV instead
	v.id = env.Get(clusterUUID)
	v.k8sArch = env.Get(clusterArch)
	v.k8sVersion = env.Get(clusterVersion)
	v.nodeType = env.Get(nodeType)
	v.openebsVersion = env.Get(openEBSversion)
	return nil
}
