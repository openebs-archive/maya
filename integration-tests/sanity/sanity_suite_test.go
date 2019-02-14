package sanity

import (
	"os"
	"testing"
	"time"

	"github.com/openebs/maya/pkg/client/k8s/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	"github.com/openebs/maya/integration-tests/kubernetes"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sanity")
}

var _ = BeforeSuite(func() {
	// Fetching the kube config path
	configPath, err := kubernetes.GetConfigPath()
	Expect(err).ShouldNot(HaveOccurred())

	// Setting the path in environemnt variable
	err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	Expect(err).ShouldNot(HaveOccurred())

	// Fetching the openebs component artifacts
	al, err := artifacts.GetArtifactsUnstructured()
	Expect(err).ShouldNot(HaveOccurred())

	// Installing the artifacts to kubernetes cluster
	for _, a := range al {
		cu := k8s.CreateOrUpdate(k8s.GroupVersionResourceFromGVK(a), a.GetNamespace())
		_, err := cu.Apply(a)
		Expect(err).ShouldNot(HaveOccurred())
	}

	// Waiting for pods to be ready
	cl, err := kubernetes.GetClientSet()
	Expect(err).NotTo(HaveOccurred())

	pods, err := cl.CoreV1().Pods("openebs").List(metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(pods).NotTo(BeNil())

	status := false
	for i := 0; i < 100; i++ {
		if kubernetes.CheckPodsRunning(*pods, 4) {
			status = true
			break
		}
		time.Sleep(10 * time.Second)
	}
	if !status {
		Fail("Pods were not ready in expected time")
	}
})

var _ = AfterSuite(func() {
	// Unsetting the environment variable
	err := os.Unsetenv(string(v1alpha1.KubeConfigEnvironmentKey))
	Expect(err).ShouldNot(HaveOccurred())
})
