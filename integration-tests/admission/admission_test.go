package admission

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openebs/maya/integration-tests/artifacts"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstorvolumereplica/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	vwebhook "github.com/openebs/maya/pkg/kubernetes/webhook/validate/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	cStorPVC artifacts.Artifact = `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: cstor-test
spec:
  storageClassName: openebs-cstor-class
  selector:
    matchLabels:
      openebs.io/casType: cstor
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: "2G"
`
	// singleReplicaSC holds the yaml spec
	// for pool with single replica
	singleReplicaSC artifacts.Artifact = `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-cstor-class
  annotations:
    cas.openebs.io/config: |
      - name: StoragePoolClaim
        value: "cstor-sparse-pool"
      - name: ReplicaCount
        value: "1"
    openebs.io/cas-type: cstor
provisioner: openebs.io/provisioner-iscsi
reclaimPolicy: Delete
`
	// clonePVCYaml holds the yaml spec
	// for clone persistentvolumeclaim
	clonePVCYaml artifacts.Artifact = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-snap-claim
  annotations:
    snapshot.alpha.kubernetes.io/snapshot: fastfurious
spec:
  storageClassName: openebs-snapshot-promoter
  selector:
    matchLabels:
      openebs.io/casType: cstor
  accessModes: [ "ReadWriteOnce" ]
  resources:
    requests:
      storage: 2G`
	// cstorSnapshotYaml holds the yaml spec
	// for volume snapshot
	cstorSnapshotYaml artifacts.Artifact = `
apiVersion: volumesnapshot.external-storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: snapshot-cstor
  namespace: default
spec:
  persistentVolumeClaimName: cstor-test
`
)

var _ = Describe("AdmissionWebhook", func() {
	BeforeEach(func() {
		// Extracting storageclass artifacts unstructured
		SCUnst, err := artifacts.GetArtifactUnstructured(singleReplicaSC)
		Expect(err).ShouldNot(HaveOccurred())

		// Apply  single replica storageclass
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(SCUnst),
			SCUnst.GetNamespace(),
		)

		_, err = cu.Apply(SCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting PVC artifacts unstructured
		PVCUnst, err := artifacts.GetArtifactUnstructured(cStorPVC)
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting pvc namespace
		PVCNamespace := PVCUnst.GetNamespace()

		// Webhook stuffs
		client, err := k8s.Clientset().Get()
		_, err = client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().List(metav1.ListOptions{})
		if errors.IsNotFound(err) {
			Skip("dynamic configuration of webhooks requires the admissionregistration.k8s.io group to be enabled")
		}
		Expect(err).ShouldNot(HaveOccurred())

		_, err = vwebhook.KubeClient().List(metav1.ListOptions{})
		if errors.IsNotFound(err) {
			Skip("dynamic configuration of webhooks requires the admissionregistration.k8s.io group to be enabled")
		}
		Expect(err).ShouldNot(HaveOccurred())

		// Create pvc using storageclass 'cstor-sparse-pool'
		cu = k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(PVCUnst),
			PVCUnst.GetNamespace(),
		)
		_, err = cu.Apply(PVCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		By("verifying pvc to be created and bound with pv")
		//	Eventually(func() string {
		//		pvclaim, err := pvc.
		//			KubeClient(pvc.WithNamespace(PVCNamespace)).Get(PVCNamespace, PVCUnst.GetName(), metav1.GetOptions{})
		//		Expect(err).ShouldNot(HaveOccurred())
		//		return string(pvclaim.Status.Phase)
		//	},
		//		defaultTimeOut, defaultPollingInterval).
		//		Should(Receive((ContainSubstring("Bound")), "PVC phase should bound"))

		// Generating label selector for stsResources
		PVCLabel := "openebs.io/casType=" + "cstor"

		Eventually(func() int {
			pvcs, err := pvc.
				KubeClient(pvc.WithNamespace(PVCNamespace)).
				List(metav1.ListOptions{LabelSelector: PVCLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return pvc.
				ListBuilder().
				WithAPIList(pvcs).
				WithFilter(pvc.IsBound()).
				List().
				Len()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(1), "PVC count should be "+string(1))

		// Check for CVR to get healthy
		Eventually(func() int {
			cvrs, err := cvr.
				KubeClient(cvr.WithNamespace("openebs")).
				List(metav1.ListOptions{LabelSelector: PVCLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return cvr.
				ListBuilder().
				WithAPIList(cvrs).
				WithFilter(cvr.IsHealthy()).
				List().
				Len()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(1), "CVR count should be 1")

	})

	AfterEach(func() {

		PVCUnst, err := artifacts.GetArtifactUnstructured(cStorPVC)
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
				List(metav1.ListOptions{LabelSelector: PVCLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return len(pvcs.Items)
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(0), "pvc count should be 0")

	})

	Context("Test admission server validation for pvc delete", func() {
		It("should deny the deletion of source volume", func() {

			// Step-1 Create the snapshot
			//CreateSnapshot()
			// Extracting snapshot artifacts unstructured
			By("Create a snapshot for a given volume")
			SnapUnst, err := artifacts.GetArtifactUnstructured(cstorSnapshotYaml)
			Expect(err).ShouldNot(HaveOccurred())
			// Apply volume snapshot
			cu := k8s.CreateOrUpdate(
				k8s.GroupVersionResourceFromGVK(SnapUnst),
				SnapUnst.GetNamespace(),
			)

			_, err = cu.Apply(SnapUnst)
			Expect(err).ShouldNot(HaveOccurred())

			// Step-2 Create Clone PVC
			//CreateClonePVC()
			// Create pvc using storageclass 'cstor-sparse-pool'
			// Extracting PVC artifacts unstructured
			By("Create a clone volume using snapshot")
			ClonePVCUnst, err := artifacts.GetArtifactUnstructured(clonePVCYaml)
			Expect(err).ShouldNot(HaveOccurred())

			// Extracting pvc namespace
			//PVCNamespace := ClonePVCUnst.GetNamespace()

			cu = k8s.CreateOrUpdate(
				k8s.GroupVersionResourceFromGVK(ClonePVCUnst),
				ClonePVCUnst.GetNamespace(),
			)
			_, err = cu.Apply(ClonePVCUnst)
			Expect(err).ShouldNot(HaveOccurred())

			// Step-3 Delete Source-volume
			//DeletePVC()
			By("Deleting source PVC should fail with error")
			PVCUnst, err := artifacts.GetArtifactUnstructured(cStorPVC)
			Expect(err).ShouldNot(HaveOccurred())

			// Extracting pvc namespace
			PVCNamespace := PVCUnst.GetNamespace()

			_ = k8s.DeleteResource(
				k8s.GroupVersionResourceFromGVK(PVCUnst),
				PVCNamespace,
			)
			err = cu.Delete(PVCUnst)
			Expect(err).Should(HaveOccurred())

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
