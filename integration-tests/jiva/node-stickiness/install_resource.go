package nodestickiness

import (
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	installer "github.com/openebs/maya/integration-tests/artifacts/installer/v1alpha1"
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
func (j *TestInstaller) WithArtifact(artifact artifacts.ArtifactSource) *TestInstaller {
	j.Artifact = artifact
	return j
}

// GetUnstructObj builds the TestInstaller instance with unstructured object
func (j *TestInstaller) GetUnstructObj() *TestInstaller {
	// Extracting artifact unstructured
	artifactUnstruct, err := artifacts.GetArtifactUnstructuredFromFile(j.Artifact)
	Expect(err).ShouldNot(HaveOccurred())
	j.ComponentUnstructured = artifactUnstruct
	return j
}

// GetInstallerObj builds the TestInstaller instance with installer object
func (j *TestInstaller) GetInstallerObj() *TestInstaller {
	i, err := installer.BuilderForObject(j.ComponentUnstructured).Build()
	Expect(err).ShouldNot(HaveOccurred())
	j.ComponentInstaller = i
	return j
}

// Install installs the resource based on the installer
func (j *TestInstaller) Install() *TestInstaller {
	err := j.ComponentInstaller.Install()
	Expect(err).ShouldNot(HaveOccurred())
	return j
}
