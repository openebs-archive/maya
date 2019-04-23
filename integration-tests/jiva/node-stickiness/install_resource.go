package nodestickiness

import (
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	installer "github.com/openebs/maya/integration-tests/artifacts/installer/v1alpha1"
	clientsc "github.com/openebs/maya/pkg/kubernetes/storageclasses/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// NodeSticky holds the required objects for installation of test related resource
type NodeSticky struct {
	Artifact              artifacts.ArtifactSource
	ComponentInstaller    *installer.DefaultInstaller
	ComponentUnstructured *unstructured.Unstructured
}

const (
	// defaultTimeOut is the default time in seconds
	// for Eventually block
	defaultTimeOut int = 500
	// defaultPollingInterval is the default polling
	// time in seconds for the Eventually block
	defaultPollingInterval int = 10
)

// NewNodeSticky defines new instance of NodeSticky
func NewNodeSticky() *NodeSticky {
	return &NodeSticky{}
}

// WithArtifact builds the NodeSticky instance with artifact
func (n *NodeSticky) WithArtifact(artifact artifacts.ArtifactSource) *NodeSticky {
	n.Artifact = artifact
	return n
}

// GetUnstructObj builds the NodeSticky instance with unstructured object
func (n *NodeSticky) GetUnstructObj() *NodeSticky {
	// Extracting artifact unstructured
	artifactUnstruct, err := artifacts.GetArtifactUnstructuredFromFile(n.Artifact)
	Expect(err).ShouldNot(HaveOccurred())
	n.ComponentUnstructured = artifactUnstruct
	return n
}

// GetInstallerObj builds the NodeSticky instance with installer object
func (n *NodeSticky) GetInstallerObj() *NodeSticky {
	i, err := installer.BuilderForObject(n.ComponentUnstructured).Build()
	Expect(err).ShouldNot(HaveOccurred())
	n.ComponentInstaller = i
	return n
}

// Install installs the resource based on the installer
func (n *NodeSticky) Install() *NodeSticky {
	err := n.ComponentInstaller.Install()
	Expect(err).ShouldNot(HaveOccurred())
	return n
}

// IsSCDeployed checks whether sc is present in the cluster or not
func (n *NodeSticky) IsSCDeployed() bool {
	//	scClient := sc.StorageV1Client{restClient: client}
	name := n.ComponentUnstructured.GetName()
	Eventually(func() int {
		storageclass, err := clientsc.KubeClient().Get(name, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return clientsc.
			NewListBuilder().
			WithAPIObject(*storageclass).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(1), "StorageClass count should be 1")
	return true
}
