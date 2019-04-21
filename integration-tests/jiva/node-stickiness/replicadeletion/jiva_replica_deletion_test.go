package replicadeletion

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	stickiness "github.com/openebs/maya/integration-tests/jiva/node-stickiness"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	testTimes                            = 20
	pvcArtifact artifacts.ArtifactSource = "../jiva_pvc_resource.yaml"
	scArtifact  artifacts.ArtifactSource = "../jiva_sc_resource.yaml"
)

var _ = Describe("[jiva] [node-stickiness] jiva replica pod node-stickiness test", func() {
	var (
		// defaultReplicaLabel represents the jiva replica
		defaultReplicaLabel = "openebs.io/replica=jiva-replica"
		// defaultCtrlLabel represents the jiva controller
		defaultCtrlLabel = "openebs.io/controller=jiva-controller"
		// defaultPVCLabel represents the default OpenEBS PVC label key
		defaultPVCLabel = "openebs.io/persistent-volume-claim="
		// replicaLabel consist of defaultReplicaLabel and coressponding
		// pvcLabel
		replicaLabel string
		// ctrlLabel consist of defaultReplicaLabel and coressponding
		// pvcLabel
		ctrlLabel                 string
		jivaTestNamespace         string
		podListObj                *v1.PodList
		scInstaller, pvcInstaller *stickiness.TestInstaller
	)
	BeforeEach(func() {

		By("Deploying jiva-single-replica storageclass")
		scInstaller = stickiness.NewTestInstaller().
			WithArtifact(scArtifact).
			GetUnstructObj().
			GetInstallerObj().
			Install()

		By("Deploying PVC using jiva-single-replica storageClass in jiva-test namespace")
		pvcInstaller = stickiness.NewTestInstaller().
			WithArtifact(pvcArtifact).
			GetUnstructObj().
			GetInstallerObj().
			Install()

		// pvcLabel represents the coressponding pvc
		pvcLabel := defaultPVCLabel + pvcInstaller.ComponentUnstructured.GetName()
		replicaLabel = defaultReplicaLabel + "," + pvcLabel
		ctrlLabel = defaultCtrlLabel + "," + pvcLabel
		jivaTestNamespace = pvcInstaller.ComponentUnstructured.GetNamespace()
		// Verify creation of jiva ctrl pod
		_ = getPodList(string(jivaTestNamespace), ctrlLabel, 1)

		// Verify creation of jiva replica pod
		podListObj = getPodList(string(jivaTestNamespace), replicaLabel, 1)
	})

	AfterEach(func() {
		By("Uninstall test artifacts")
		err := scInstaller.ComponentInstaller.UnInstall()
		Expect(err).ShouldNot(HaveOccurred())
		err = pvcInstaller.ComponentInstaller.UnInstall()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("node stickiness with jiva replica pod deletion", func() {
		var nodeName, podName string

		It("should verify jiva replica pod sticks to one node", func() {

			for i := 0; i < testTimes; i++ {
				By("fetching node name and podName of jiva replica pod")
				//nodeName holds name of the node where the replica pod deployed
				nodeName = podListObj.Items[0].Spec.NodeName
				podName = podListObj.Items[0].ObjectMeta.Name

				By(fmt.Sprintf("deleting the running jiva replica pod: '%s'", podName))
				err := pod.
					KubeClient(pod.WithNamespace(string(jivaTestNamespace))).
					Delete(podName, &metav1.DeleteOptions{})
				Expect(err).ShouldNot(HaveOccurred())

				// Makesure that pod is deleted successfully
				Eventually(func() error {
					_, err := pod.
						KubeClient(pod.WithNamespace(string(jivaTestNamespace))).
						Get(podName, metav1.GetOptions{})
					return err
				},
					defaultTimeOut, defaultPollingInterval).
					Should(HaveOccurred(), "Pod not found")

				By("waiting till jiva replica pod starts running")
				podListObj = getPodList(string(jivaTestNamespace), replicaLabel, 1)

				By("verifying jiva replica pod node matches with its old instance node")
				Expect(podListObj.Items[0].Spec.NodeName).Should(Equal(nodeName))
			}
		})
	})
})
