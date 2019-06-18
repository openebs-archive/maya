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

package nodeselector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	"github.com/openebs/maya/tests/jiva"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	podList        *corev1.PodList
	restartCounter = 5
	readWriteOnce  = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity       = "5G"
	pvcObj         *corev1.PersistentVolumeClaim
	err            error
)

var _ = Describe("[jiva] TEST NODE SELECTOR", func() {
	var (
		pvcName = "jiva-volume-claim"
	)

	BeforeEach(func() {

		By("building a pvc")
		pvcObj, err = pvc.NewBuilder().
			WithName(pvcName).
			WithNamespace(namespaceObj.Name).
			WithStorageClass(scObj.Name).
			WithAccessModes(readWriteOnce).
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
		controllerPodCount := ops.GetPodRunningCountEventually(
			namespaceObj.Name,
			jiva.CtrlLabel,
			1,
		)
		Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

		By("verifying replica pod count ")
		replicaPodCount := ops.GetPodRunningCountEventually(
			namespaceObj.Name,
			jiva.ReplicaLabel,
			jiva.ReplicaCount,
		)
		Expect(replicaPodCount).To(
			Equal(jiva.ReplicaCount),
			"while checking replica pod count",
		)

		By("verifying status as bound")
		status := ops.IsPVCBoundEventually(pvcName)
		Expect(status).To(Equal(true), "while checking status equal to bound")

	})

	AfterEach(func() {

		By("deleting above pvc")
		err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
		Expect(err).To(
			BeNil(),
			"while deleting pvc {%s} in namespace {%s}",
			pvcName,
			namespaceObj.Name,
		)

		By("verifying controller pod count as 0")
		controllerPodCount := ops.GetPodRunningCountEventually(
			namespaceObj.Name,
			jiva.CtrlLabel,
			0,
		)
		Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

		By("verifying replica pod count as 0")
		replicaPodCount := ops.GetPodRunningCountEventually(
			namespaceObj.Name,
			jiva.ReplicaLabel,
			0,
		)
		Expect(replicaPodCount).To(Equal(0), "while checking replica pod count")

		By("verifying deleted pvc")
		pvc := ops.IsPVCDeleted(pvcName)
		Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

	})

	When("replica pod of pvc is restarted", func() {
		It("should stick to storage node after reconciliation", func() {
			podList, err = ops.PodClient.
				WithNamespace(namespaceObj.Name).
				List(metav1.ListOptions{LabelSelector: jiva.ReplicaLabel})
			Expect(err).ShouldNot(HaveOccurred(), "while fetching replica pods")

			for i := 0; i < restartCounter; i++ {

				By("deleting a replica pod")
				err = ops.PodClient.Delete(
					podList.Items[0].Name,
					&metav1.DeleteOptions{},
				)
				Expect(err).ShouldNot(HaveOccurred(), "while deleting replica pod")

				By("verifying deleted pod is terminated")
				status := ops.IsPodDeletedEventually(
					namespaceObj.Name,
					podList.Items[0].Name,
				)
				Expect(status).To(Equal(true), "while checking for deleted pod")

				By("verifying running replica pod count ")
				replicaPodCount := ops.GetPodRunningCountEventually(
					namespaceObj.Name,
					jiva.ReplicaLabel,
					jiva.ReplicaCount,
				)
				Expect(replicaPodCount).To(
					Equal(jiva.ReplicaCount),
					"while checking replica pod count",
				)

				By("verifying replica node selector")
				podList, err = ops.PodClient.
					WithNamespace(namespaceObj.Name).
					List(metav1.ListOptions{LabelSelector: jiva.ReplicaLabel})
				Expect(err).ShouldNot(HaveOccurred(), "while fetching replica pods")

				Expect(podList.Items[0].Spec.NodeName).To(
					Equal(storageNode),
					"checking replica node selector",
				)
			}
		})
	})

	When("controller pod of pvc is restarted", func() {
		It("should stick to app node after reconciliation", func() {
			podList, err = ops.PodClient.
				WithNamespace(namespaceObj.Name).
				List(metav1.ListOptions{LabelSelector: jiva.CtrlLabel})
			Expect(err).ShouldNot(HaveOccurred(), "while fetching controller pods")

			for i := 0; i < restartCounter; i++ {

				By("deleting a controller pod")
				err = ops.PodClient.Delete(
					podList.Items[0].Name,
					&metav1.DeleteOptions{},
				)
				Expect(err).ShouldNot(HaveOccurred(), "while deleting controller pod")

				By("verifying deleted pod is terminated")
				status := ops.IsPodDeletedEventually(
					namespaceObj.Name,
					podList.Items[0].Name,
				)
				Expect(status).To(Equal(true), "while checking for deleted pod")

				By("verifying running controller pod count ")
				replicaPodCount := ops.GetPodRunningCountEventually(
					namespaceObj.Name,
					jiva.CtrlLabel,
					1,
				)
				Expect(replicaPodCount).To(
					Equal(1),
					"while checking controller pod count",
				)

				By("verifying target node selector")
				podList, err = ops.PodClient.
					WithNamespace(namespaceObj.Name).
					List(metav1.ListOptions{LabelSelector: jiva.CtrlLabel})
				Expect(err).ShouldNot(HaveOccurred(), "while fetching controller pods")

				Expect(podList.Items[0].Spec.NodeName).To(
					Equal(appNode),
					"checking controller node selector",
				)
			}
		})
	})

})
