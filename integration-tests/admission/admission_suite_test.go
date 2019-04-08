package admission

import (
	"flag"
	"os"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	"github.com/openebs/maya/integration-tests/kubernetes"
	"github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	validatehook "github.com/openebs/maya/pkg/kubernetes/webhook/validate/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// defaultTimeOut is the default time in seconds
	// for Eventually block
	defaultTimeOut int = 600
	// defaultPollingInterval is the default polling
	// time in seconds for the Eventually block
	defaultPollingInterval int = 10
	// minNodeCount is the minimum number of nodes
	// needed to run this test
	minNodeCount int = 1
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[single-node] Admission server")
}

func init() {
	flag.Parse()
}

var _ = BeforeSuite(func() {
	// Fetching the kube config path
	configPath, err := kubernetes.GetConfigPath()
	Expect(err).ShouldNot(HaveOccurred())

	// Setting the path in environemnt variable
	err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	Expect(err).ShouldNot(HaveOccurred())
	// Getting clientset
	cl, err := kubernetes.GetClientSet()
	Expect(err).ShouldNot(HaveOccurred())

	// Checking appropriate node numbers
	nodes, err := cl.CoreV1().Nodes().List(v1.ListOptions{})
	Expect(nodes.Items).Should(HaveLen(minNodeCount))

	// Fetching the openebs component artifacts
	artifactsOpenEBS, errs := artifacts.GetArtifactsListUnstructuredFromFile(artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	By("Deploying OpenEBS components in openebs namespace")
	for _, artifact := range artifactsOpenEBS {
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(artifact),
			artifact.GetNamespace(),
		)
		_, err := cu.Apply(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}

	By("Verifying 'maya-apiserver' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace), string(artifacts.MayaAPIServerLabelSelector), 1)

	By("Verifying 'openebs-provisioner' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSProvisionerLabelSelector), 1)

	By("Verifying 'snapshot-operator' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSSnapshotOperatorLabelSelector), 1)

	By("Verifying 'admission-server' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSAdmissionServerLabelSelector), 1)

	By("Verifying 'Node-device-manager' pods status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSNDMLabelSelector), minNodeCount)

	By("Verifying 'cstor-pool' pods status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSCStorPoolLabelSelector), minNodeCount)

	// check validationWebhookConfiguration API enable in cluster
	_, err = validatehook.KubeClient().List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())

	By("OpenEBS control-plane components are in running state")
})

var _ = AfterSuite(func() {
	// Fetching the openebs component artifacts
	artifactsOpenEBS, errs := artifacts.GetArtifactsListUnstructuredFromFile(artifacts.OpenEBSArtifacts)
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
})

func checkComponentStatus(namespace, lselector string, podCount int) (pods *corev1.PodList) {
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
