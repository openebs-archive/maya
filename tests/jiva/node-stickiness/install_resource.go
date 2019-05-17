// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nodestickiness

import (
	. "github.com/onsi/gomega"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha2"
	"github.com/openebs/maya/tests/artifacts"
	installer "github.com/openebs/maya/tests/artifacts/installer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeStickyInstaller holds the required objects for installation of test related resource
type NodeStickyInstaller struct {
	*installer.DefaultInstaller
}

const (
	// defaultTimeOut is the default time in seconds
	// for Eventually block
	defaultTimeOut int = 500
	// defaultPollingInterval is the default polling
	// time in seconds for the Eventually block
	defaultPollingInterval int = 10
)

// NewNodeStickyInstallerForArtifacts defines new instance of NodeStickyInstaller
func NewNodeStickyInstallerForArtifacts(artifact artifacts.Artifact, opts ...unstruct.KubeclientBuildOption) *NodeStickyInstaller {
	n := NodeStickyInstaller{}
	// Extracting artifact unstructured
	artifactUnstruct, err := artifacts.GetArtifactUnstructured(artifact)
	Expect(err).ShouldNot(HaveOccurred())
	n.DefaultInstaller, err = installer.
		BuilderForObject(artifactUnstruct).
		WithKubeClient(opts...).
		Build()
	Expect(err).ShouldNot(HaveOccurred())
	return &n
}

// GetInstallerInstance builds the NodeStickyInstaller instance with installer object
func (n *NodeStickyInstaller) GetInstallerInstance() *installer.DefaultInstaller {
	return n.DefaultInstaller
}

// IsSCDeployed checks whether sc is present in the cluster or not
func IsSCDeployed(name string) bool {
	Eventually(func() bool {
		storageClass, err := sc.NewKubeClient().Get(name, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		if storageClass != nil {
			return true
		}
		return false
	},
		defaultTimeOut, defaultPollingInterval).
		Should(BeTrue(), "StorageClass should present")
	return true
}
