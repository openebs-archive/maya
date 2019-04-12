package validation

import (
	"flag"
	"os"

	"testing"

	"github.com/openebs/maya/pkg/client/k8s/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	install "github.com/openebs/maya/integration-tests/artifacts/installer/v1alpha1"
	"github.com/openebs/maya/integration-tests/kubernetes"
	node "github.com/openebs/maya/pkg/kubernetes/nodes/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kube "k8s.io/client-go/kubernetes"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// defaultTimeOut is the default time in seconds
	// for Eventually block
	defaultTimeOut int = 500
	// defaultPollingInterval is the default polling
	// time in seconds for the Eventually block
	defaultPollingInterval int = 10
	// minNodeCount is the minimum number of nodes
	// need to run this test
	minNodeCount int = 1
)

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

var (
	//Client set
	cl                    *kube.Clientset
	artifactsOpenEBS      []*unstructured.Unstructured
	defaultoebsComponents []*install.DefaultInstaller
	invalidChar           string
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod")
}

func init() {
	flag.StringVar(&invalidChar, "injectErr", "", "Used to insert invalid charecter in yaml")
}

var _ = BeforeSuite(func() {
	var errs []error
	// Fetching the kube config path
	configPath, err := kubernetes.GetConfigPath()
	Expect(err).ShouldNot(HaveOccurred())

	// Setting the path in environemnt variable
	err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	Expect(err).ShouldNot(HaveOccurred())

	// Check the running node count
	nodes, err := node.
		KubeClient().List(metav1.ListOptions{})
	nodeCnt := node.
		ListBuilderFunc().
		WithAPIList(nodes).
		WithFilter(node.IsReady()).
		List().
		Len()
	Expect(nodeCnt).Should(Equal(minNodeCount), "Running node count should be "+string(nodeCnt))

	//builder := installer.BuilderForYaml(artifacts.OpenEBSArtifacts)

	// Fetch openebs component artifacts
	artifactsOpenEBS, errs = artifacts.GetArtifactsListUnstructuredFromFile(artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	By("Installing OpenEBS components")
	// Installing the artifacts to kubernetes cluster
	for _, artifact := range artifactsOpenEBS {
		buildOpenebsComponents := install.BuilderForObject(artifact)
		oebsComponentInstaller, err := buildOpenebsComponents.Build()
		Expect(err).ShouldNot(HaveOccurred())
		err = oebsComponentInstaller.Install()
		Expect(err).ShouldNot(HaveOccurred())
		defaultoebsComponents = append(defaultoebsComponents, oebsComponentInstaller)
	}

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
	By("Uinstalling OpenEBS Components")
	for _, oebsComponent := range defaultoebsComponents {
		err := oebsComponent.UnInstall()
		Expect(err).ShouldNot(HaveOccurred())
	}
})
