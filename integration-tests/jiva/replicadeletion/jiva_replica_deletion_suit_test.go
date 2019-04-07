package replicadeletion

import (
	"flag"
	"os"

	"testing"

	"github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	"github.com/openebs/maya/integration-tests/kubernetes"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// defaultTimeOut is the default time in seconds
	// for Eventually block
	defaultTimeOut int = 5000
	// defaultPollingInterval is the default polling
	// time in seconds for the Eventually block
	defaultPollingInterval int = 10
	// minNodeCount is the minimum number of nodes
	// need to run this test
	minNodeCount int = 3
	// jiva-test namespace to deploy jiva ctrl & replicas
	parentDir     artifacts.ArtifactSource = "../"
	nameSpaceYaml artifacts.Artifact       = `
apiVersion: v1
kind: Namespace
metadata:
  name: jiva-test
`
)

var (
	//Client set
	cl *kube.Clientset
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod")
}

func init() {
	flag.Parse()
}

func checkPodUpandRunning(namespace, lselector string, podCount int) (pods *corev1.PodList) {
	// Verify phase of the pod
	var err error
	Eventually(func() int {
		pods, err = pod.
			KubeClient(pod.WithNamespace(namespace)).
			List(metav1.ListOptions{LabelSelector: lselector})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.
			ListBuilder().
			WithAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(podCount), "Pod count should be "+string(podCount))
	return
}

var _ = BeforeSuite(func() {
	// Fetching the kube config path
	configPath, err := kubernetes.GetConfigPath()
	Expect(err).ShouldNot(HaveOccurred())

	// Setting the path in environemnt variable
	err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	Expect(err).ShouldNot(HaveOccurred())
	// Getting clientset
	cl, err = kubernetes.GetClientSet()
	Expect(err).ShouldNot(HaveOccurred())

	//TODO: Implement the node package in path pkg/kubernetes/v1alpha1/
	// Checking appropriate node numbers. This test is designed to run on a 3 node cluster
	nodes, err := cl.CoreV1().Nodes().List(v1.ListOptions{})
	Expect(nodes.Items).Should(HaveLen(minNodeCount))

	// Fetching the openebs component artifacts
	artifactsOpenEBS, errs := artifacts.GetArtifactsListUnstructuredFromFile(parentDir + artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	// Installing the artifacts to kubernetes cluster
	for _, artifact := range artifactsOpenEBS {
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(artifact),
			artifact.GetNamespace(),
		)
		_, err := cu.Apply(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}

	// Creates jiva-test namespace
	testNameSpaceUnstructured, err := artifacts.GetArtifactUnstructured(
		artifacts.Artifact(nameSpaceYaml),
	)
	Expect(err).ShouldNot(HaveOccurred())
	cu := k8s.CreateOrUpdate(
		k8s.GroupVersionResourceFromGVK(testNameSpaceUnstructured),
		testNameSpaceUnstructured.GetNamespace(),
	)
	_, err = cu.Apply(testNameSpaceUnstructured)
	Expect(err).ShouldNot(HaveOccurred())

	By("Started deploying OpenEBS components")
	// Check for maya-apiserver pod to get created and running
	_ = checkPodUpandRunning(string(artifacts.OpenebsNamespace), string(artifacts.MayaAPIServerLabelSelector), 1)

	// Check for provisioner pod to get created and running
	_ = checkPodUpandRunning(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSProvisionerLabelSelector), 1)

	// Check for snapshot operator to get created and running
	_ = checkPodUpandRunning(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSSnapshotOperatorLabelSelector), 1)

	// Check for admission server to get created and running
	_ = checkPodUpandRunning(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSAdmissionServerLabelSelector), 1)

	// Check for NDM pods to get created and running
	_ = checkPodUpandRunning(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSNDMLabelSelector), minNodeCount)

	// Check for cstor storage pool pods to get created and running
	_ = checkPodUpandRunning(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSCStorPoolLabelSelector), minNodeCount)

	By("OpenEBS components are in running state")
})

var _ = AfterSuite(func() {
	// Fetching the openebs component artifacts
	artifactsOpenEBS, errs := artifacts.GetArtifactsListUnstructuredFromFile(parentDir + artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	// Deleting the artifacts to kubernetes cluster
	for _, artifact := range artifactsOpenEBS {
		cu := k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(artifact),
			artifact.GetNamespace(),
		)
		err := cu.Delete(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}

	// Deletes jiva-test namespace
	testNameSpaceUnstructured, err := artifacts.GetArtifactUnstructured(
		artifacts.Artifact(nameSpaceYaml),
	)
	Expect(err).ShouldNot(HaveOccurred())
	cu := k8s.DeleteResource(
		k8s.GroupVersionResourceFromGVK(testNameSpaceUnstructured),
		testNameSpaceUnstructured.GetNamespace(),
	)
	err = cu.Delete(testNameSpaceUnstructured)
	Expect(err).ShouldNot(HaveOccurred())

	// Unsetting the environment variable
	err = os.Unsetenv(string(v1alpha1.KubeConfigEnvironmentKey))
	Expect(err).ShouldNot(HaveOccurred())
})
