package validation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	install "github.com/openebs/maya/integration-tests/artifacts/installer/v1alpha1"
	file "github.com/openebs/maya/pkg/file"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func installTestNamespace() {
	var err error
	By("Creating validation-ns test namespace")
	namespaceComponent := install.BuilderForYaml(string(nameSpaceYaml))
	namespaceInstaller, err = namespaceComponent.Build()
	Expect(err).ShouldNot(HaveOccurred())
	err = namespaceInstaller.Install()
	Expect(err).ShouldNot(HaveOccurred())
}

func clearTestRelatedArtifacts() {
	By("Clearing validation test related artifacts")
	for _, tcInstaller := range testComponentsInstaller {
		err := tcInstaller.UnInstall()
		Expect(err).ShouldNot(HaveOccurred())
	}
	testComponentsInstaller = nil
	err := namespaceInstaller.UnInstall()
	Expect(err).ShouldNot(HaveOccurred())
}

func deployTestArtifacts(testArtifactsUn []*unstructured.Unstructured) {
	for _, artifact := range testArtifactsUn {
		buildTestComponents := install.BuilderForObject(artifact)
		testComponentInstaller, err := buildTestComponents.Build()
		Expect(err).ShouldNot(HaveOccurred())
		err = testComponentInstaller.Install()
		if len(invalidChar) == 0 {
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			Expect(err).Should(HaveOccurred())
		}
		testComponentsInstaller = append(testComponentsInstaller, testComponentInstaller)
	}
}

const (
	crctValueSC                            = "      - name: TargetResourceLimits"
	crctValuePVC                           = "   name: label-validation-pvc"
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
	namespaceInstaller      *install.DefaultInstaller
	testComponentsInstaller []*install.DefaultInstaller
	index                   int
)

var _ = Describe("[SC] Test to validate Yamls", func() {
	BeforeEach(func() {
		FileOperatorVar = file.RealFileOperator{}
		var err error
		installTestNamespace()

		if len(invalidChar) > 0 {
			By("Updating storage class artifact with invalid character")
			testValue := crctValueSC + invalidChar
			index, _, err = FileOperatorVar.GetLineDetails(string(testArtifacts), crctValueSC)
			Expect(err).ShouldNot(HaveOccurred())
			err = FileOperatorVar.Updatefile(string(testArtifacts), testValue, index, 0644)
			Expect(err).ShouldNot(HaveOccurred())
		}
	})
	AfterEach(func() {
		if len(invalidChar) > 0 {
			By("Updating storage class artifact with original string")
			err := FileOperatorVar.Updatefile(string(testArtifacts), crctValueSC, index, 0644)
			Expect(err).ShouldNot(HaveOccurred())
		}

		clearTestRelatedArtifacts()
	})

	Context("Test to validate Yamls", func() {
		It("should show appropriate error in logs of maya-apiserver", func() {
			// Fetching the openebs component artifacts
			testArtifactsUn, errs := artifacts.GetArtifactsListUnstructuredFromFile(testArtifacts)
			Expect(errs).Should(HaveLen(0))

			By("Deploying test related artifacts(sc, pvc) by injecting error in SC artifact")
			deployTestArtifacts(testArtifactsUn)
		})
	})
})

var _ = Describe("[PVC] Test to validate Yamls", func() {
	BeforeEach(func() {
		FileOperatorVar = file.RealFileOperator{}
		var err error
		installTestNamespace()

		if len(invalidChar) > 0 {
			By("Updating PVC artifact with invalid character")
			testValue := crctValuePVC + invalidChar
			index, _, err = FileOperatorVar.GetLineDetails(string(testArtifacts), crctValuePVC)
			Expect(err).ShouldNot(HaveOccurred())
			err = FileOperatorVar.Updatefile(string(testArtifacts), testValue, index, 0644)
			Expect(err).ShouldNot(HaveOccurred())
		}
	})

	Context("Test to validate the Yamls", func() {
		It("should show appropriate error in logs of maya-apiserver", func() {
			// Fetching the openebs component artifacts
			testArtifactsUn, errs := artifacts.GetArtifactsListUnstructuredFromFile(testArtifacts)
			Expect(errs).Should(HaveLen(0))

			By("Deploying test related artifacts(sc, pvc) by injecting error in PVC artifact")
			deployTestArtifacts(testArtifactsUn)
		})
	})

	AfterEach(func() {
		if len(invalidChar) > 0 {
			By("Updating PVC artifact with original string")
			err := FileOperatorVar.Updatefile(string(testArtifacts), crctValuePVC, index, 0644)
			Expect(err).ShouldNot(HaveOccurred())
		}

		clearTestRelatedArtifacts()
	})
})
