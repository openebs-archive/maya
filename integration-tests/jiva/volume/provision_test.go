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

package volume

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	replicaLabel = "openebs.io/replica=jiva-replica"
	ctrlLabel    = "openebs.io/controller=jiva-controller"
	accessModes  = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity     = "5G"
	pvcObj       *corev1.PersistentVolumeClaim
	err          error
)

var _ = Describe("[jiva] TEST VOLUME PROVISIONING", func() {
	var (
		pvcName = "jiva-volume-claim"
	)

	BeforeEach(func() {

	})

	When("jiva pvc with replicacount 1 is created", func() {
		It("1 controller and 1 replica pod should be running", func() {

			By("Building a persistentvolumeclaim")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building persistentvolumeclaim {%s} in namespace {default}", pvcName)

			By("Creating a jiva pvc")
			_, err = ops.pvcClient.WithNamespace(nsName).Create(pvcObj)
			Expect(err).To(BeNil(), "while creating persistentvolumeclaim {%s} in namespace {default}", pvcObj.Name)

			By("Checking pod counts of controller and replica")
			controllerPodCount := ops.getPodCountRunningEventually(nsName, ctrlLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking jiva controller pod count")

			replicaPodCount := ops.getPodCountRunningEventually(nsName, replicaLabel, 1)
			Expect(replicaPodCount).To(Equal(1), "while checking jiva replica pod count")

			By("Checking status of jiva pvc")
			status := ops.isBound(pvcName)
			Expect(status).To(Equal(true), "while checking status")

		})
	})

	When("jiva pvc is deleted", func() {
		It("no pods should be running", func() {

			By("deleting jiva pvc")
			err := ops.pvcClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting persistentvolumeclaim {%s} in namespace {default}", pvcObj.Name)

			By("Checking pod counts of controller and replica")
			controllerPodCount := ops.getPodCountRunningEventually(nsName, ctrlLabel, 0)
			Expect(controllerPodCount).To(Equal(0), "while checking jiva controller pod count")

			replicaPodCount := ops.getPodCountRunningEventually(nsName, replicaLabel, 0)
			Expect(replicaPodCount).To(Equal(0), "while checking jiva replica pod count")

			By("Trying to get deleted jiva pvc")
			pvc := ops.checkDeletedPVC(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

		})
	})

	AfterEach(func() {

	})

})
