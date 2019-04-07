package replicadeletion

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// auth plugins
	//	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// jivaSCYaml creates jiva-1r storageClass
	jivaSCYaml artifacts.Artifact = `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: jiva-1r
  annotations:
    openebs.io/cas-type: jiva
    cas.openebs.io/config: |
      - name: ReplicaCount
        value: "1"
provisioner: openebs.io/provisioner-iscsi
`
	// pvcYaml create jiva-vol1-1r-claim pvc
	pvcYaml artifacts.Artifact = `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: jiva-vol1-1r-claim
  namespace: jiva-test
spec:
  storageClassName: jiva-1r
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5G
`
)

type Namespace string

const (
	// jivaTestNamespace is the name of the jiva-test namespace
	jivaTestNamespace Namespace = "jiva-test"
	testTimes                   = 20
)

var _ = Describe("[jiva] [node-stickiness] jiva replica pod delete test", func() {
	var (
		// defaultReplicaLabel represents the jiva replica
		defaultReplicaLabel = "openebs.io/replica=jiva-replica"
		// defaultCtrlLabel represents the jiva controller
		defaultCtrlLabel = "openebs.io/controller=jiva-controller"
		// replicaLabel consist of defaultReplicaLabel and coressponding
		// pvcLabel
		replicaLabel string
		// ctrlLabel consist of defaultReplicaLabel and coressponding
		// pvcLabel
		ctrlLabel string
		podObjs   *v1.PodList
	)
	BeforeEach(func() {

		//Extracting storageclass artifacts unstructured
		jivaSCUnstructured, err := artifacts.GetArtifactUnstructured(
			artifacts.Artifact(jivaSCYaml),
		)
		Expect(err).ShouldNot(HaveOccurred())

		//Extracting PVC artifacts unstructured
		pvcUnstructured, err := artifacts.GetArtifactUnstructured(
			artifacts.Artifact(pvcYaml),
		)
		Expect(err).ShouldNot(HaveOccurred())

		//Apply jiva storageClass
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(jivaSCUnstructured),
			jivaSCUnstructured.GetNamespace(),
		)

		By("Deploying jiva-1r storageClass")
		_, err = cu.Apply(jivaSCUnstructured)
		Expect(err).ShouldNot(HaveOccurred())

		//Apply PVC using jiva storageClass in jiva-test namespace
		cu = k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(pvcUnstructured),
			pvcUnstructured.GetNamespace(),
		)

		By("Deploying PVC using jiva-1r storageClass in jiva-test namespace")
		_, err = cu.Apply(pvcUnstructured)
		Expect(err).ShouldNot(HaveOccurred())

		// pvcLabel represents the coressponding pvc
		pvcLabel := "openebs.io/persistent-volume-claim=" + pvcUnstructured.GetName()
		replicaLabel = defaultReplicaLabel + "," + pvcLabel
		ctrlLabel = defaultCtrlLabel + "," + pvcLabel

		// Verify creation of jiva ctrl pod
		_ = checkPodUpandRunning(string(jivaTestNamespace), ctrlLabel, 1)

		// Verify creation of jiva replica pod
		podObjs = checkPodUpandRunning(string(jivaTestNamespace), replicaLabel, 1)
	})

	AfterEach(func() {
		//Extracting PVC artifacts unstructured
		pvcUnstructured, err := artifacts.GetArtifactUnstructured(
			artifacts.Artifact(pvcYaml),
		)
		Expect(err).ShouldNot(HaveOccurred())

		//Extracting storageclass artifacts unstructured
		jivaSCUnstructured, err := artifacts.GetArtifactUnstructured(
			artifacts.Artifact(jivaSCYaml),
		)
		Expect(err).ShouldNot(HaveOccurred())

		//Delete the PVC using it's unstructured format
		cu := k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(pvcUnstructured),
			string(jivaTestNamespace),
		)
		By("Deleting PVC in jiva-test namespace")
		err = cu.Delete(pvcUnstructured)
		Expect(err).ShouldNot(HaveOccurred())

		//Delete jiva1-r storageClass
		cu = k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(jivaSCUnstructured),
			jivaSCUnstructured.GetNamespace(),
		)
		By("Deleting jiva-1r storageClass")
		err = cu.Delete(jivaSCUnstructured)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("node stickiness with jiva replica pod deletion", func() {
		var nodeName, podName string

		It("should verify jiva replica pod sticks to one node", func() {

			for i := 0; i < testTimes; i++ {
				By("fetching node name and podName of jiva replica pod")
				//nodeName holds name of the node where the replica pod deployed
				nodeName = podObjs.Items[0].Spec.NodeName
				podName = podObjs.Items[0].ObjectMeta.Name

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
				podObjs = checkPodUpandRunning(string(jivaTestNamespace), replicaLabel, 1)

				By("verifying jiva replica pod node matches with its old instance node")
				Expect(podObjs.Items[0].Spec.NodeName).Should(Equal(nodeName))
			}
		})
	})
})
