/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package snapshot

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests/jiva"

	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	snap "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	replicaLabel = "openebs.io/replica=jiva-replica"
	ctrlLabel    = "openebs.io/controller=jiva-controller"
	accessModes  = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity     = "5G"
	pvcObj       *corev1.PersistentVolumeClaim
)

var _ = Describe("[jiva] TEST JIVA SNAPSHOT CREATION", func() {
	var (
		pvcName  = "jiva-pvc"
		snapName = "jivasnapshot"
	)

	When("jiva pvc with replicacount n is created", func() {
		It("should create 1 controller pod and n replica pod", func() {

			By("building a pvc")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scObj.Name).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above pvc")
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("verifying controller pod count")
			controllerPodCount := ops.GetPodRunningCountEventually(namespaceObj.Name, ctrlLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			By("verifying replica pod count")
			replicaPodCount := ops.GetPodRunningCountEventually(namespaceObj.Name, replicaLabel, jiva.ReplicaCount)
			Expect(replicaPodCount).To(Equal(jiva.ReplicaCount), "while checking replica pod count")

			By("verifying status as bound")
			status := ops.IsPVCBound(pvcName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

		})
	})

	When("jiva snapshot is created", func() {
		It("should create a snapshot with type ready", func() {

			By("building a snapshot")
			snapObj, err = snap.NewBuilder().
				WithName(snapName).
				WithNamespace(namespaceObj.Name).
				WithPVC(pvcName).
				Build()
			Expect(err).To(
				BeNil(),
				"while building snapshot {%s} in namespace {%s}",
				snapName,
				namespaceObj.Name,
			)

			By("creating above snapshot")
			_, err = ops.SnapClient.WithNamespace(namespaceObj.Name).Create(snapObj)
			Expect(err).To(
				BeNil(),
				"while creating snapshot {%s} in namespace {%s}",
				snapName,
				namespaceObj.Name,
			)

			By("verifying type as ready")
			snaptype := ops.GetSnapshotTypeEventually(snapName)
			Expect(snaptype).To(Equal("Ready"), "while checking snapshot type")

		})
	})

	When("jiva snapshot is deleted", func() {
		It("should remove above snapshot", func() {

			By("deleting above snapshot")
			err := ops.SnapClient.Delete(snapName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting snapshot {%s} in namespace {%s}",
				snapName,
				namespaceObj.Name,
			)

			By("verifying deleted snapshot")
			snap := ops.IsSnapshotDeleted(snapName)
			Expect(snap).To(Equal(true), "while checking for deleted snapshot")

		})
	})

	When("jiva pvc is deleted", func() {
		It("should not have any jiva controller and replica pods", func() {

			By("deleting above pvc")
			err := ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("verifying controller pod count as 0")
			controllerPodCount := ops.GetPodRunningCountEventually(namespaceObj.Name, ctrlLabel, 0)
			Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

			By("verifying replica pod count as 0")
			replicaPodCount := ops.GetPodRunningCountEventually(namespaceObj.Name, replicaLabel, 0)
			Expect(replicaPodCount).To(Equal(0), "while checking replica pod count")

			By("verifying deleted pvc")
			pvc := ops.IsPVCDeleted(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

		})
	})

})
