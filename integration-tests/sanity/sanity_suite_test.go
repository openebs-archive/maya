package sanity

import (
	"flag"
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

const (
	waitTime time.Duration = 10
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sanity")
}

var namespace string

func init() {
	flag.StringVar(&namespace, "namespace", "openebs", "namespace for performing the test")
	flag.Parse()

}

var _ = BeforeSuite(func() {

	// Fetching the kube config path
	configPath, err := kubernetes.GetConfigPath()
	Expect(err).ShouldNot(HaveOccurred())

	// Setting the path in environemnt variable
	err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	Expect(err).ShouldNot(HaveOccurred())

	// Fetching the openebs component artifacts
	artifacts, err := artifacts.GetArtifactsListUnstructured(artifacts.OpenEBSArtifacts)
	Expect(err).ShouldNot(HaveOccurred())

	// Installing the artifacts to kubernetes cluster
	for _, artifact := range artifacts {
		cu := k8s.CreateOrUpdate(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
		_, err := cu.Apply(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}

	// Waiting for pods to be ready
	clientset, err := kubernetes.GetClientSet()
	Expect(err).NotTo(HaveOccurred())

	status := false
	for i := 0; i < 300; i++ {
		pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(pods).NotTo(BeNil())
		expectedStoragePoolPods, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		if kubernetes.CheckPodsRunning(*pods, 4+len(expectedStoragePoolPods.Items)) {
			status = true
			break
		}
		time.Sleep(waitTime * time.Second)
	}
	if !status {
		Fail("Pods were not ready in expected time")
	}
})

var _ = AfterSuite(func() {
	// Fetching the openebs component artifacts
	artifacts, err := artifacts.GetArtifactsListUnstructured(artifacts.OpenEBSArtifacts)
	Expect(err).ShouldNot(HaveOccurred())

	// Deleting artifacts
	for _, artifact := range artifacts {
		d := k8s.DeleteResource(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
		err := d.Delete(artifact)
		Expect(err).NotTo(HaveOccurred())
	}

	// Unsetting the environment variable
	err = os.Unsetenv(string(v1alpha1.KubeConfigEnvironmentKey))
	Expect(err).ShouldNot(HaveOccurred())

	// Waiting for openebs namespace to get terminated
	clientset, err := kubernetes.GetClientSet()
	Expect(err).NotTo(HaveOccurred())

	status := false
	for i := 0; i < 100; i++ {
		namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(namespaces).NotTo(BeNil())

		if kubernetes.CheckForNamespace(*namespaces, namespace) {
			status = true
			break
		}
		time.Sleep(waitTime * time.Second)
	}
	if !status {
		Fail("Pods were not ready in expected time")
	}
})
