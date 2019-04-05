package admission

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openebs/maya/integration-tests/artifacts"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	cv "github.com/openebs/maya/pkg/cstorvolume/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	validatehook "github.com/openebs/maya/pkg/kubernetes/webhook/validate/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	cStorPVC artifacts.Artifact = `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: cstor-source-volume
  namespace: default
  labels:
    name: cstor-test
spec:
  storageClassName: openebs-cstor-class
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
  namespace: default
  labels:
    name: test-snap-claim
  annotations:
    snapshot.alpha.kubernetes.io/snapshot: snapshot-cstor
spec:
  storageClassName: openebs-snapshot-promoter
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
  persistentVolumeClaimName: cstor-source-volume
`
)

var _ = Describe("[single-node] [cstor] AdmissionWebhook", func() {
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
		_, err = validatehook.KubeClient().List(metav1.ListOptions{})
		if errors.IsNotFound(err) {
			Fail(fmt.Sprintf("dynamic configuration of webhooks requires the admissionregistration.k8s.io group to be enabled: %v", err))
		}
		Expect(err).ShouldNot(HaveOccurred())

		// Create pvc using storageclass 'cstor-sparse-class'
		By(fmt.Sprintf("Creating pvc %s in default namespace", PVCUnst.GetName()))
		cu = k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(PVCUnst),
			PVCUnst.GetNamespace(),
		)
		_, err = cu.Apply(PVCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		By("verifying pvc to be created and bound with pv")
		Eventually(func() bool {
			pvclaim, err := pvc.
				KubeClient(pvc.WithNamespace(PVCNamespace)).
				Get(PVCUnst.GetName(), metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			return pvc.
				NewForAPIObject(pvclaim).IsBound()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(BeTrue())

		// Check for CVR to get healthy
		Eventually(func() int {
			cvs, err := cv.
				KubeClient(cv.WithNamespace("openebs")).
				List(metav1.ListOptions{LabelSelector: ""})
			Expect(err).ShouldNot(HaveOccurred())
			return cv.
				NewListBuilder().
				WithAPIList(cvs).
				WithFilter(cv.IsHealthy()).
				List().
				Len()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(1), "CVR count should be 1")

	})

	AfterEach(func() {
		SCUnst, err := artifacts.GetArtifactUnstructured(singleReplicaSC)
		Expect(err).ShouldNot(HaveOccurred())

		PVCUnst, err := artifacts.GetArtifactUnstructured(cStorPVC)
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting pvc namespace
		PVCNamespace := PVCUnst.GetNamespace()

		By(fmt.Sprintf("deleting PVC '%s' as part of teardown", PVCUnst.GetName()))
		// Delete the PVC artifacts
		cu := k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(PVCUnst),
			PVCNamespace,
		)
		err = cu.Delete(PVCUnst)
		//Expect(err).ShouldNot(HaveOccurred())
		if err != nil {
			Fail(fmt.Sprintf("could not delete volume %q: %v", PVCUnst.GetName(), err))
		}
		// Verify deletion of pvc instances
		Eventually(func() int {
			pvcs, err := pvc.
				KubeClient(pvc.WithNamespace(PVCNamespace)).
				List(metav1.ListOptions{LabelSelector: "name=cstor-source-volume"})
			Expect(err).ShouldNot(HaveOccurred())
			return len(pvcs.Items)
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(0), "pvc count should be 0")

			// verify deletion of cstorvolume
		Eventually(func() int {
			cvs, err := cv.
				KubeClient(cv.WithNamespace("openebs")).
				List(metav1.ListOptions{LabelSelector: ""})
			Expect(err).ShouldNot(HaveOccurred())
			return len(cvs.Items)
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(0), "CVR count should be 0")

		cu = k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(SCUnst),
			SCUnst.GetNamespace(),
		)
		err = cu.Delete(SCUnst)
		Expect(err).ShouldNot(HaveOccurred())

	})

	Context("Test admission server validation for pvc delete", func() {
		It("should deny the deletion of source volume", func() {

			// Step-1 Create the snapshot
			// Extracting snapshot artifacts unstructured
			By("Creating a snapshot for a given volume")
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
			// Create pvc using storageclass 'cstor-sparse-pool'
			// Extracting PVC artifacts unstructured
			By("Create a clone volume using snapshot")
			ClonePVCUnst, err := artifacts.GetArtifactUnstructured(clonePVCYaml)
			Expect(err).ShouldNot(HaveOccurred())

			// Extracting pvc namespace
			clonePVCNamespace := ClonePVCUnst.GetNamespace()

			cu = k8s.CreateOrUpdate(
				k8s.GroupVersionResourceFromGVK(ClonePVCUnst),
				ClonePVCUnst.GetNamespace(),
			)
			_, err = cu.Apply(ClonePVCUnst)
			Expect(err).ShouldNot(HaveOccurred())

			By(fmt.Sprintf("verifying clone pvc '%s' to be created and bound with pv", ClonePVCUnst.GetName()))
			Eventually(func() bool {
				pvclaim, err := pvc.
					KubeClient(pvc.WithNamespace(clonePVCNamespace)).
					Get(ClonePVCUnst.GetName(), metav1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				return pvc.
					NewForAPIObject(pvclaim).IsBound()
			},
				defaultTimeOut, defaultPollingInterval).
				Should(BeTrue())

			// Step-3 Delete Source-volume
			PVCUnst, err := artifacts.GetArtifactUnstructured(cStorPVC)
			Expect(err).ShouldNot(HaveOccurred())

			By(fmt.Sprintf("Deleting source PVC '%s' should fail with error", PVCUnst.GetName()))
			// Extracting pvc namespace
			PVCNamespace := PVCUnst.GetNamespace()

			del := k8s.DeleteResource(
				k8s.GroupVersionResourceFromGVK(PVCUnst),
				PVCNamespace,
			)
			err = del.Delete(PVCUnst)
			Expect(err).ToNot(BeNil())

			By("Delete clone persistentvolumeclaim ")
			err = del.Delete(ClonePVCUnst)
			if err != nil {
				Fail(fmt.Sprintf("could not delete volume %q: %v", ClonePVCUnst.GetName(), err))
			}
			// Verify deletion of pvc instances
			Eventually(func() int {
				pvcs, err := pvc.
					KubeClient(pvc.WithNamespace(PVCNamespace)).
					List(metav1.ListOptions{LabelSelector: "name=test-snap-claim"})
				Expect(err).ShouldNot(HaveOccurred())
				return len(pvcs.Items)
			},
				defaultTimeOut, defaultPollingInterval).
				Should(Equal(0), "pvc count should be 0")

			By("Delete volume snapshot")
			snap := k8s.DeleteResource(
				k8s.GroupVersionResourceFromGVK(SnapUnst),
				PVCNamespace,
			)
			err = snap.Delete(SnapUnst)
			if err != nil {
				Fail(fmt.Sprintf("could not delete snapshot %q: %v", SnapUnst.GetName(), err))
			}

		})
	})
})
