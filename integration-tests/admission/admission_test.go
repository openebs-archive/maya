package admission

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	cv "github.com/openebs/maya/pkg/cstorvolume/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// namespaceYaml holds the yaml spec
	// to create admission namespace
	namespaceYaml artifacts.Artifact = `
apiVersion: v1
kind: Namespace
metadata:
  name: admission
`
	// cStorPVC holds the yaml spec
	// for source persistentvolumeclaim
	cStorPVC artifacts.Artifact = `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: cstor-source-volume
  namespace: admission
  labels:
    name: cstor-source-volume
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
  namespace: admission
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
  namespace: admission
spec:
  persistentVolumeClaimName: cstor-source-volume
`
)

var _ = Describe("[single-node] [cstor] AdmissionWebhook", func() {

	var (
		NSUnst, SCUnst, PVCUnst *unstructured.Unstructured
		pvclaim                 *corev1.PersistentVolumeClaim
	)
	BeforeEach(func() {
		// Extracting storageclass artifacts unstructured
		var err error
		SCUnst, err = artifacts.GetArtifactUnstructured(singleReplicaSC)
		Expect(err).ShouldNot(HaveOccurred())

		// Apply  single replica storageclass
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(SCUnst),
			SCUnst.GetNamespace(),
		)

		_, err = cu.Apply(SCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		// Creates admission namespace
		NSUnst, err = artifacts.GetArtifactUnstructured(
			artifacts.Artifact(namespaceYaml),
		)
		Expect(err).ShouldNot(HaveOccurred())

		cu = k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(NSUnst),
			NSUnst.GetNamespace(),
		)
		_, err = cu.Apply(NSUnst)
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting PVC artifacts unstructured
		PVCUnst, err = artifacts.GetArtifactUnstructured(cStorPVC)
		Expect(err).ShouldNot(HaveOccurred())

		// Create pvc using storageclass 'cstor-sparse-class'
		By(fmt.Sprintf("Creating pvc '%s' in '%s' namespace", PVCUnst.GetName(), PVCUnst.GetNamespace()))
		cu = k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(PVCUnst),
			PVCUnst.GetNamespace(),
		)
		_, err = cu.Apply(PVCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		By("verifying pvc to be created and bound with pv")
		Eventually(func() bool {
			pvclaim, err = pvc.
				NewKubeClient(pvc.WithNamespace(PVCUnst.GetNamespace())).
				Get(PVCUnst.GetName(), metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			return pvc.
				NewForAPIObject(pvclaim).IsBound()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(BeTrue())

		// Check for cstorvolume to get healthy
		Eventually(func() bool {
			cstorvolume, err := cv.
				NewKubeclient(cv.WithNamespace("openebs")).
				Get(pvclaim.Spec.VolumeName, metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			return cv.
				NewForAPIObject(cstorvolume).IsHealthy()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(BeTrue())
	})

	AfterEach(func() {
		By(fmt.Sprintf("deleting PVC '%s' as part of teardown", PVCUnst.GetName()))
		// Delete the PVC artifacts
		cu := k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(PVCUnst),
			PVCUnst.GetNamespace(),
		)
		err := cu.Delete(PVCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify deletion of pvc instances
		Eventually(func() int {
			pvcs, err := pvc.
				NewKubeClient(pvc.WithNamespace(PVCUnst.GetNamespace())).
				List(metav1.ListOptions{LabelSelector: "name=cstor-source-volume"})
			Expect(err).ShouldNot(HaveOccurred())
			return len(pvcs.Items)
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(0), "pvc count should be 0")

		CstorVolumeLabel := "openebs.io/persistent-volume=" + pvclaim.Spec.VolumeName

		// verify deletion of cstorvolume
		Eventually(func() int {
			cvs, err := cv.
				NewKubeclient(cv.WithNamespace("openebs")).
				List(metav1.ListOptions{LabelSelector: CstorVolumeLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return len(cvs.Items)
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(0), "cStorvolume count should be 0")

		cu = k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(SCUnst),
			SCUnst.GetNamespace(),
		)
		err = cu.Delete(SCUnst)
		Expect(err).ShouldNot(HaveOccurred())

		cu = k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(NSUnst),
			"",
		)
		err = cu.Delete(NSUnst)
		Expect(err).ShouldNot(HaveOccurred())

	})

	Context("Test admission server validation for pvc delete", func() {
		It("should deny the deletion of source volume", func() {

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

			// Extracting clone PVC artifacts unstructured
			By("Creating a clone volume using snapshot")
			ClonePVCUnst, err := artifacts.GetArtifactUnstructured(clonePVCYaml)
			Expect(err).ShouldNot(HaveOccurred())

			cu = k8s.CreateOrUpdate(
				k8s.GroupVersionResourceFromGVK(ClonePVCUnst),
				ClonePVCUnst.GetNamespace(),
			)
			_, err = cu.Apply(ClonePVCUnst)
			Expect(err).ShouldNot(HaveOccurred())

			By(fmt.Sprintf("verifying clone pvc '%s' to be created and bound with pv", ClonePVCUnst.GetName()))
			Eventually(func() bool {
				pvclone, err := pvc.
					NewKubeClient(pvc.WithNamespace(ClonePVCUnst.GetNamespace())).
					Get(ClonePVCUnst.GetName(), metav1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				return pvc.
					NewForAPIObject(pvclone).IsBound()
			},
				defaultTimeOut, defaultPollingInterval).
				Should(BeTrue())

			By(fmt.Sprintf("Deleting source PVC '%s' should fail with error", PVCUnst.GetName()))

			del := k8s.DeleteResource(
				k8s.GroupVersionResourceFromGVK(PVCUnst),
				PVCUnst.GetNamespace(),
			)
			err = del.Delete(PVCUnst)
			Expect(err).ToNot(BeNil())

			By(fmt.Sprintf("Deleting clone persistentvolumeclaim '%s'", ClonePVCUnst.GetName()))
			err = del.Delete(ClonePVCUnst)
			Expect(err).ShouldNot(HaveOccurred())

			// Verify deletion of pvc instances
			Eventually(func() int {
				pvcs, err := pvc.
					NewKubeClient(pvc.WithNamespace(ClonePVCUnst.GetNamespace())).
					List(metav1.ListOptions{LabelSelector: "name=test-snap-claim"})
				Expect(err).ShouldNot(HaveOccurred())
				return len(pvcs.Items)
			},
				defaultTimeOut, defaultPollingInterval).
				Should(Equal(0), "pvc count should be 0")

			By("Deleting volume snapshot")
			snap := k8s.DeleteResource(
				k8s.GroupVersionResourceFromGVK(SnapUnst),
				SnapUnst.GetNamespace(),
			)
			err = snap.Delete(SnapUnst)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
