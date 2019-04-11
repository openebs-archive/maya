package validation

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	instal "github.com/openebs/maya/integration-tests/artifacts/installer/v1alpha1"
	file "github.com/openebs/maya/pkg/file"
)

const (
	value                                  = "      - name: TargetResourceLimits"
	testArtifacts artifacts.ArtifactSource = "validation-artifacts.yaml"
	// nameSpaceYaml namespace to deploy volumes in validation-ns namespace
	nameSpaceYaml artifacts.Artifact = `
apiVersion: v1
kind: Namespace
metadata:
  name: validation-ns
`
)

var (
	FileOperatorVar         file.FileOperator
	namespaceInstaller      *instal.DefaultInstaller
	testComponentsInstaller []*instal.DefaultInstaller
)

var _ = Describe("Debuging proper error logs", func() {
	BeforeEach(func() {
		FileOperatorVar = file.RealFileOperator{}
		var err error
		By("Creating test namespace")
		namespaceComponent := instal.BuilderForYaml(string(nameSpaceYaml))
		namespaceInstaller, err = namespaceComponent.Build()
		Expect(err).ShouldNot(HaveOccurred())
		err = namespaceInstaller.Install()
		Expect(err).ShouldNot(HaveOccurred())
	})
	AfterEach(func() {
		time.Sleep(time.Second * 1000)
		By("Deleting test related artifacts")
		err := namespaceInstaller.UnInstall()
		Expect(err).ShouldNot(HaveOccurred())
		for _, tcInstaller := range testComponentsInstaller {
			err := tcInstaller.UnInstall()
			Expect(err).ShouldNot(HaveOccurred())
		}
	})

	Context("test debug logs in maya-apiserver", func() {
		It("should show perfect logs regarding error", func() {
			invalidChar := ":"
			testValue := value + invalidChar
			index, _, err := FileOperatorVar.GetLineDetails(string(testArtifacts), value)
			Expect(err).ShouldNot(HaveOccurred())
			err = FileOperatorVar.Updatefile(string(testArtifacts), testValue, index, 0644)
			Expect(err).ShouldNot(HaveOccurred())

			// Fetching the openebs component artifacts
			testArtifactsUn, errs := artifacts.GetArtifactsListUnstructuredFromFile(testArtifacts)
			Expect(errs).Should(HaveLen(0))

			for _, artifact := range testArtifactsUn {
				buildTestComponents := instal.BuilderForObject(artifact)
				testComponentInstaller, err := buildTestComponents.Build()
				Expect(err).ShouldNot(HaveOccurred())
				err = testComponentInstaller.Install()
				Expect(err).ShouldNot(HaveOccurred())
				testComponentsInstaller = append(testComponentsInstaller, testComponentInstaller)
			}
		})
	})
})
