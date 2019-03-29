package webhook

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openebs/maya/integration-tests/artifacts"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstorvolumereplica/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// stsYaml holds the yaml spec
	// for statefulset application
	clonePVCYaml artifacts.Artifact = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: demo-snap-vol-claim
  annotations:
    snapshot.alpha.kubernetes.io/snapshot: fastfurious
spec:
  storageClassName: openebs-snapshot-promoter
  accessModes: [ "ReadWriteOnce" ]
  resources:
    requests:
      storage: 2G`
)

var _ = Describe("AdmissionWebhook", func() {
	BeforeEach(func() {
		// Extracting storageclass artifacts unstructured
		_, err := artifacts.GetArtifactUnstructured(
			artifacts.Artifact(artifacts.SingleReplicaSC),
		)
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting PVC artifacts unstructured
		PVCUnst, err := artifacts.GetArtifactUnstructured(artifacts.Artifact(artifacts.CStorPVCArtifacts))
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting statefulset application namespace
		PVCNamespace := PVCUnst.GetNamespace()

		// Webhook stuffs
		client, err := k8s.Clientset().Get()
		client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().List(metav1.ListOptions{})
		if errors.IsNotFound(err) {
			ginkgo.Skip("dynamic configuration of webhooks requires the admissionregistration.k8s.io group to be enabled")
		}
		Expect(err).ShouldNot(HaveOccurred())

		// Create pvc using storageclass 'cstor-sparse-pool'
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(PVCUnst),
			PVCUnst.GetNamespace(),
		)
		_, err = cu.Apply(PVCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		// Check for pvc to get created and bound
		Eventually(func() string {
			pvclaim, err := pvc.
				KubeClient(pvc.WithNamespace(PVCNamespace)).Get(PVCNamespace, PVCUnst.GetName())
			Expect(err).ShouldNot(HaveOccurred())
			return string(pvclaim.Status.Phase)
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Receive((ContainSubstring("Bound")), "Pvc phase should bound"))

		PVLabel := "openebs.io/casType=" + "cstor"
		// Check for CVR to get healthy
		Eventually(func() int {
			cvrs, err := cvr.
				KubeClient(cvr.WithNamespace("")).
				List(metav1.ListOptions{LabelSelector: PVLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return cvr.
				ListBuilder().
				WithAPIList(cvrs).
				WithFilter(cvr.IsHealthy()).
				List().
				Len()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(1), "CVR count should be "+string(1))

	})

	AfterEach(func() {

		// Extracting statefulset artifacts unstructured
		PVCUnst, err := artifacts.GetArtifactUnstructured(artifacts.Artifact(artifacts.CStorPVCArtifacts))
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting pvc namespace
		PVCNamespace := PVCUnst.GetNamespace()

		// Generating label selector for stsResources
		PVCLabel := "openebs.io/casType=" + "cstor"

		// Fetch PVCs to be deleted
		pvcs, err := pvc.KubeClient(pvc.WithNamespace(PVCNamespace)).
			List(metav1.ListOptions{LabelSelector: PVCLabel})
		Expect(err).ShouldNot(HaveOccurred())
		// Delete PVCs
		for _, p := range pvcs.Items {
			err = pvc.KubeClient(pvc.WithNamespace(PVCNamespace)).
				Delete(p.GetName(), &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred())
		}

		// Delete the PVC artifacts
		cu := k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(PVCUnst),
			PVCNamespace,
		)
		err = cu.Delete(PVCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify deletion of pvc instances
		Eventually(func() int {
			pvcs, err := pvc.
				KubeClient(pvc.WithNamespace(PVCNamespace)).
				List(metav1.ListOptions{LabelSelector: ""})
			Expect(err).ShouldNot(HaveOccurred())
			return len(pvcs.Items)
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(0), "pvc count should be 0")

	})

	Context("Test admission server validation for pvc delete", func() {
		It("should deny the deletion of source volume", func() {

			// Step-1 Create the snapshot
			//
			//
			// Step-2 Create Clone PVC
			//
			//
			// Step-3 Delete Source-volume

		})

		PIt("should delete volume b/c clone volume not exists", func() {

			// Step-1 Create the snapshot
			//
			//
			// Step-2 Create Clone PVC
			//
			//
			// Step-3 Delete Source-volume

		})
	})
})
