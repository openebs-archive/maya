package kubernetes

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Connect", func() {
	Context("with kubeconfig", func() {
		It("should fetch kubernetes version", func() {

			clientset, err := GetClientSet()
			Expect(err).To(BeNil())

			KubernetesVersion, err := clientset.Discovery().ServerVersion()
			Expect(err).To(BeNil())

			Expect(KubernetesVersion.GitVersion).ToNot(BeEmpty())
		})
	})

})
