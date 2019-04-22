package invalidconfig

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
)

var (
	scArtifacts = map[string]artifacts.ArtifactSource{
		"bad_sc_with_semicolumn": artifacts.ArtifactSource("resources_with_bad_sc.yaml"),
	}
	pvcArtifacts = map[string]artifacts.ArtifactSource{
		"bad_pvc_with_semicolumn": artifacts.ArtifactSource("resources_with_bad_pvc.yaml"),
	}
	pvcsInstaller, scsInstaller []*TestInstaller
)

var _ = Describe("Debuging proper error logs", func() {
	BeforeEach(func() {
		By("Deploy required storageclasses")
		for _, artifact := range scArtifacts {
			installerObj := NewTestInstaller().
				WithArtifact(artifact).
				GetUnstructObj().
				GetInstallerObj().
				Install()
			isPresent := installerObj.isSCDeployed()
			Expect(isPresent).Should(BeTrue())
			scsInstaller = append(scsInstaller, installerObj)
		}
	})
	AfterEach(func() {
		for _, installer := range scsInstaller {
			err := installer.ComponentInstaller.UnInstall()
			Expect(err).ShouldNot(HaveOccurred())
		}
	})

	Context("test debug logs in maya-apiserver", func() {
		It("should show perfect logs regarding error", func() {
			By("Deploy persistent volume claim")
			for _, artifact := range pvcArtifacts {
				installerObj := NewTestInstaller().
					WithArtifact(artifact).
					GetUnstructObj().
					GetInstallerObj().
					Install()
				pvcsInstaller = append(pvcsInstaller, installerObj)
			}

			for _, installer := range pvcsInstaller {
				err := installer.ComponentInstaller.UnInstall()
				Expect(err).ShouldNot(HaveOccurred())
			}
		})
	})
})
