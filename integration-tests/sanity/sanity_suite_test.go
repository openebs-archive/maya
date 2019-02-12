package sanity

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sanity")
}

var _ = BeforeSuite(func() {
	Describe("Setup", func() {
		PIt("should deploy the openebs operator", func() {

		})

		PIt("should deploy the openebs provisioner", func() {

		})

		PIt("should deploy the ndm", func() {

		})

		PIt("should deploy the cstor storage pool CRDs", func() {

		})

		PIt("should deploy the cstor pools", func() {

		})

		PIt("should deploy the cstor replica CRDs", func() {

		})

		PIt("should deploy cstor volume replica CRDs", func() {

		})
	})
})

var _ = AfterSuite(func() {
	Describe("Setup", func() {
		PIt("should delete the cstor pools", func() {

		})

		PIt("should delete ndm pods", func() {

		})

		PIt("should delete the openebs provisioner", func() {

		})

		PIt("should delete the cstor replica CRDs", func() {

		})

		PIt("should delete the cstor volume CRDs", func() {

		})

		PIt("should delete the cstor storage pool CRDs", func() {

		})

		PIt("shoud delete the openebs operator", func() {

		})
	})
})
