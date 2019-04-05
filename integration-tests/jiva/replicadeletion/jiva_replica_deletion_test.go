package replicadeletion

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	"github.com/stretchr/testify/assert"
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

var _ = Describe("jiva replica pod delete test", func() {
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
		_ = checkPodUpandRunning(string(jivaTestNamespace), replicaLabel, 1)
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

	Context("jiva replica deletion test", func() {
		It("Deletion of jiva replica pod", func() {
			By("Get the jiva replica pod details and perform test")
			var nodeName, podName string
			for i := 0; i < testTimes; i++ {
				pods := checkPodUpandRunning(string(jivaTestNamespace), replicaLabel, 1)
				// Deployed volume using single replica
				podName = pods.Items[0].ObjectMeta.Name
				if i == 0 {
					// nodeName consist where the replica pod deployed
					nodeName = pods.Items[0].Spec.NodeName
				} else {
					assert.Equal(GinkgoT(), pods.Items[0].Spec.NodeName, nodeName)
				}

				fmt.Printf("Delete pod: '%s' count: %d\n", podName, i)
				// Delete the jiva replica pod
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
			}
		})
	})
})
