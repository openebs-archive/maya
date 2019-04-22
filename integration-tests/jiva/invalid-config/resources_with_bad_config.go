package invalidconfig

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	installer "github.com/openebs/maya/integration-tests/artifacts/installer/v1alpha1"
	clientsc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestInstaller holds the required objeccts for installation of resource
type TestInstaller struct {
	Artifact              artifacts.ArtifactSource
	ComponentInstaller    *installer.DefaultInstaller
	ComponentUnstructured *unstructured.Unstructured
}

// NewTestInstaller defines new instance of TestInstaller
func NewTestInstaller() *TestInstaller {
	return &TestInstaller{}
}

// WithArtifact builds the TestInstaller instance with artifact
func (t *TestInstaller) WithArtifact(artifact artifacts.ArtifactSource) *TestInstaller {
	t.Artifact = artifact
	return t
}

// GetUnstructObj builds the TestInstaller instance with unstructured obtect
func (t *TestInstaller) GetUnstructObj() *TestInstaller {
	// Extracting artifact unstructured
	artifactUnstruct, err := artifacts.GetArtifactUnstructuredFromFile(t.Artifact)
	Expect(err).ShouldNot(HaveOccurred())
	t.ComponentUnstructured = artifactUnstruct
	return t
}

// GetInstallerObj builds the TestInstaller instance with installer object
func (t *TestInstaller) GetInstallerObj() *TestInstaller {
	i, err := installer.BuilderForObject(t.ComponentUnstructured).Build()
	Expect(err).ShouldNot(HaveOccurred())
	t.ComponentInstaller = i
	return t
}

// Install installs the resource based on the installer
func (t *TestInstaller) Install() *TestInstaller {
	err := t.ComponentInstaller.Install()
	Expect(err).ShouldNot(HaveOccurred())
	return t
}

func (t *TestInstaller) isSCDeployed() bool {
	//	scClient := sc.StorageV1Client{restClient: client}
	name := t.ComponentUnstructured.GetName()
	By(fmt.Sprintf("Check whether the '%s' storageclass is available in cluster", name))
	Eventually(func() int {
		storageclass, err := clientsc.KubeClient().Get(name, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return clientsc.
			NewListBuilder().
			WithAPIObject(*storageclass).
			List().
			Len()
	},
		100, 5).
		Should(Equal(1), "StorageClass count should be 1")
	return true
}
