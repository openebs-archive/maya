package replicadeletion

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	clientpvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	clientsc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	testTimes = 20
)

var _ = Describe("[jiva] [node-stickiness] jiva replica pod node-stickiness test", func() {
	var (
		// replicaLabel consist of defaultReplicaLabel and coressponding
		// pvcLabel
		replicaLabel string
		// ctrlLabel consist of defaultReplicaLabel and coressponding
		// pvcLabel
		ctrlLabel string
		//podListObj holds the PodList instance
		podListObj    *corev1.PodList
		scObj         *clientsc.StorageClass
		pvcObj        *clientpvc.PVC
		podKubeClient *pod.Kubeclient
		// defaultReplicaLabel represents the jiva replica
		defaultReplicaLabel = "openebs.io/replica=jiva-replica"
		// defaultCtrlLabel represents the jiva controller
		defaultCtrlLabel = "openebs.io/controller=jiva-controller"
		// defaultPVCLabel represents the default OpenEBS PVC label key
		defaultPVCLabel       = "openebs.io/persistent-volume-claim="
		storageEngine         = "jiva"
		replicaCount          = "1"
		openebsCASConfigValue = "- name: ReplicaCount\n  Value: " + replicaCount
		scName                = "jiva-single-replica"
		pvcName               = "jiva-vol1-1r-claim"
		testNamespace         = "jiva-rep-delete-ns"
		accessModes           = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity              = "5G"
		//TODO: following variables should be moved in framework
		openebsCASType     = "cas.openebs.io/cas-type"
		openebsCASConfig   = "cas.openebs.io/config"
		openebsProvisioner = "openebs.io/provisioner-iscsi"
	)
	BeforeEach(func() {
		var err error
		By("Build jiva-single-replica storageclass and deploy it")
		annotations := map[string]string{
			openebsCASType:   storageEngine,
			openebsCASConfig: openebsCASConfigValue,
		}
		buildSCObj := clientsc.NewStorageClass().
			WithName(scName).
			WithAnnotations(annotations).
			WithProvisioner(openebsProvisioner)
		Expect(buildSCObj.Err).ShouldNot(HaveOccurred())

		scObj = &clientsc.StorageClass{}
		scObj.Object, err = clientsc.KubeClient().Create(buildSCObj.Object)
		Expect(err).ShouldNot(HaveOccurred())

		By("Build and deploy PVC using jiva-single-replica storageClass in jiva-rep-delete-ns namespace")
		buildPVCObj := clientpvc.NewPVC().
			WithName(pvcName).
			WithNamespace(testNamespace).
			WithStorageClass(scName).
			WithAccessModes(accessModes).
			WithCapacity(capacity)
		Expect(buildPVCObj.Err).ShouldNot(HaveOccurred())

		pvcObj = &clientpvc.PVC{}
		pvcObj.Object, err = clientpvc.NewKubeClient(clientpvc.WithNamespace(testNamespace)).Create(buildPVCObj.Object)
		Expect(err).ShouldNot(HaveOccurred())

		podKubeClient = pod.KubeClient(pod.WithNamespace(string(testNamespace)))

		// pvcLabel represents the coressponding pvc
		pvcLabel := defaultPVCLabel + pvcName
		replicaLabel = defaultReplicaLabel + "," + pvcLabel
		ctrlLabel = defaultCtrlLabel + "," + pvcLabel
		// Verify creation of jiva ctrl pod
		_ = getPodList(podKubeClient, string(testNamespace), ctrlLabel, 1)

		// Verify creation of jiva replica pod
		podListObj = getPodList(podKubeClient, string(testNamespace), replicaLabel, 1)
	})

	AfterEach(func() {
		By("Uninstall test artifacts")
		err := clientpvc.NewKubeClient(clientpvc.WithNamespace(testNamespace)).Delete(pvcName, &metav1.DeleteOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		err = clientsc.KubeClient().Delete(scName, &metav1.DeleteOptions{})
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
				err := podKubeClient.Delete(podName, &metav1.DeleteOptions{})
				Expect(err).ShouldNot(HaveOccurred())

				// Makesure that pod is deleted successfully
				Eventually(func() bool {
					_, err := podKubeClient.Get(podName, metav1.GetOptions{})
					if k8serror.IsNotFound(err) {
						return true
					}
					return false
				},
					defaultTimeOut, defaultPollingInterval).
					Should(BeTrue(), "Pod not found")

				By("waiting till jiva replica pod starts running")
				podListObj = getPodList(podKubeClient, string(testNamespace), replicaLabel, 1)

				By("verifying jiva replica pod node matches with its old instance node")
				Expect(podListObj.Items[0].Spec.NodeName).Should(Equal(nodeName))
			}
		})
	})
})
