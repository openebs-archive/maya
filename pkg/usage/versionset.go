/*
Copyright 2021 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package usage

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	clusterUUID    string = "OPENEBS_IO_USAGE_UUID"
	clusterVersion string = "OPENEBS_IO_K8S_VERSION"
	clusterArch    string = "OPENEBS_IO_K8S_ARCH"
	openEBSversion string = "OPENEBS_IO_VERSION_TAG"
	nodeType       string = "OPENEBS_IO_NODE_TYPE"
	installerType  string = "OPENEBS_IO_INSTALLER_TYPE"
)

// versionSet is a struct which stores (sort of) information about a
// k8s environment
type versionSet struct {
	clientset      *kubernetes.Clientset
	id             string // OPENEBS_IO_USAGE_UUID
	k8sVersion     string // OPENEBS_IO_K8S_VERSION
	k8sArch        string // OPENEBS_IO_K8S_ARCH
	openebsVersion string // OPENEBS_IO_VERSION_TAG
	nodeType       string // OPENEBS_IO_NODE_TYPE
	installerType  string // OPENEBS_IO_INSTALLER_TYPE
}

// NewVersion returns a new versionSet struct with given application version
func NewVersion() (*versionSet, error) {
	clientset, err := getK8sClient()
	if err != nil {
		return nil, err
	}

	return &versionSet{
		clientset: clientset,
	}, nil
}

// setAppVersion set openebs version info with given value
func (v *versionSet) setOpenEBSVersion(version string) {
	v.openebsVersion = version
}

// fetchAndSetVersion consumes the Kubernetes API to get environment constants
// and returns a versionSet struct
func (v *versionSet) fetchAndSetVersion() error {
	var err error
	v.id, err = v.getUUIDbyNS("default")
	if err != nil {
		return err
	}
	envSet(clusterUUID, v.id)

	k8s, err := v.getServerVersion()
	if err != nil {
		return err
	}
	// eg. linux/amd64
	v.k8sArch = k8s.Platform
	v.k8sVersion = k8s.GitVersion
	envSet(clusterArch, v.k8sArch)
	envSet(clusterVersion, v.k8sVersion)
	v.nodeType, err = v.getOSAndKernelVersion()
	envSet(nodeType, v.nodeType)
	if err != nil {
		return err
	}
	envSet(openEBSversion, v.openebsVersion)
	return nil
}

// getVersion is a wrapper over fetchAndSetVersion
func (v *versionSet) getVersion(override bool) error {
	// If ENVs aren't set or the override is true, fetch the required
	// values from the K8s APIserver
	if _, present := os.LookupEnv(openEBSversion); !present || override {
		if err := v.fetchAndSetVersion(); err != nil {
			return err
		}
	}
	// Fetch data from ENV
	v.id = envGet(clusterUUID)
	v.k8sArch = envGet(clusterArch)
	v.k8sVersion = envGet(clusterVersion)
	v.nodeType = envGet(nodeType)
	v.openebsVersion = envGet(openEBSversion)
	v.installerType = envGet(installerType)
	return nil
}

// getUUIDbyNS returns the metadata.object.uid of a namespace in Kubernetes
func (v *versionSet) getUUIDbyNS(namespace string) (string, error) {
	NSstruct, err := v.clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if NSstruct != nil {
		return string(NSstruct.GetObjectMeta().GetUID()), nil
	}
	return "", nil
}

func envSet(key string, value string) error {
	return os.Setenv(key, value)
}

func envGet(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

// getServerVersion uses the client-go Discovery client to get the
// kubernetes version struct
func (v *versionSet) getServerVersion() (*version.Info, error) {
	return v.clientset.Discovery().ServerVersion()
}

// getOSAndKernelVersion gets us the OS,Kernel version
func (v *versionSet) getOSAndKernelVersion() (string, error) {
	firstNode, err := v.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return "unknown, unknown", errors.Wrapf(err, "failed to get the os kernel/arch")
	}
	nodedetails := firstNode.Items[0].Status.NodeInfo
	return nodedetails.OSImage + ", " + nodedetails.KernelVersion, nil
}

// GetNumberOfNodes returns the number of nodes registered in a Kubernetes cluster
func (v *versionSet) GetNumberOfNodes() (int, error) {
	nodes, err := v.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get the number of nodes")
	} else {
		return len(nodes.Items), nil
	}
}

// getK8sClient returns a new instance of kubernetes clientset
func getK8sClient() (*kubernetes.Clientset, error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.New("error fetching cluster config")
	}

	return kubernetes.NewForConfig(conf)
}
