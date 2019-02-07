package bootstrap

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getting Kubernetes version", func() {
	Context("using kubeconfig", func() {
		It("should return the version of kubernetes", func() {
			KubernetesVersion, err := kubeConfig.Discovery().ServerVersion()
			Expect(err).To(BeNil())
			fmt.Printf("Kubernets Version: %v", KubernetesVersion)
		})
	})
})
