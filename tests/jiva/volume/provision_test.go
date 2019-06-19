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
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	jivaClient "github.com/openebs/maya/pkg/client/jiva"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	"github.com/openebs/maya/tests/jiva"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity    = "5G"
	pvcObj      *corev1.PersistentVolumeClaim
	podList     *corev1.PodList
)

var _ = Describe("[jiva] TEST VOLUME PROVISIONING", func() {
	var (
		pvcName      = "jiva-volume-claim"
		pvcLabel     = "openebs.io/persistent-volume-claim=" + pvcName
		ctrlLabel    = jiva.CtrlLabel + "," + pvcLabel
		replicaLabel = jiva.ReplicaLabel + "," + pvcLabel
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
			controllerPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				ctrlLabel,
				1,
			)
			Expect(controllerPodCount).To(
				Equal(1),
				"while checking controller pod count",
			)

			By("verifying replica pod count ")
			replicaPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				replicaLabel,
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
	})

	When("contoller pod is restarted", func() {
		It("should register replica pods successfully", func() {

			By("deleting controller pod")
			podList, err = ops.PodClient.List(
				metav1.ListOptions{LabelSelector: ctrlLabel},
			)
			Expect(err).ShouldNot(HaveOccurred(), "while listing controller pod")
			err = ops.PodClient.WithNamespace(namespaceObj.Name).
				Delete(podList.Items[0].Name, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting controller pod")

			By("verifying deleted pod is terminated")
			status := ops.IsPodDeletedEventually(
				namespaceObj.Name,
				podList.Items[0].Name,
			)
			Expect(status).To(Equal(true), "while checking for deleted pod")

			By("verifying controller pod count")
			controllerPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				ctrlLabel,
				1,
			)
			Expect(controllerPodCount).To(
				Equal(1),
				"while checking controller pod count",
			)

			By("verifying registered replica count and replication factor")
			podList, err = ops.PodClient.List(
				metav1.ListOptions{LabelSelector: ctrlLabel},
			)
			Expect(err).ShouldNot(HaveOccurred(), "while listing controller pod")

			status = areReplicasRegisteredEventually(
				&podList.Items[0],
				jiva.ReplicaCount,
			)
			Expect(status).To(
				Equal(true),
				"while verifying registered replica count as replication factor",
			)
		})
	})

	When("replica pods is restarted", func() {
		It("should register replica pods to controller pod successfully", func() {
			podList, err = ops.PodClient.List(
				metav1.ListOptions{LabelSelector: ctrlLabel},
			)
			Expect(err).ShouldNot(HaveOccurred(), "while listing controller pod")

			ctrlPod := podList.Items[0]

			podList, err = ops.PodClient.List(
				metav1.ListOptions{LabelSelector: replicaLabel},
			)
			Expect(err).ShouldNot(HaveOccurred(), "while listing replica pods")

			By("deleting replica pods")
			err = ops.PodClient.WithNamespace(namespaceObj.Name).
				DeleteCollection(metav1.ListOptions{LabelSelector: replicaLabel},
					&metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting replica pods")

			By("verifying deleted pods are terminated")
			for _, p := range podList.Items {
				status := ops.IsPodDeletedEventually(namespaceObj.Name, p.Name)
				Expect(status).To(
					Equal(true),
					"while checking for deleted pod {%s}",
					p.Name,
				)
			}
			By("verifying replica pod count")
			replicaPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				replicaLabel,
				jiva.ReplicaCount,
			)
			Expect(replicaPodCount).To(
				Equal(jiva.ReplicaCount),
				"while checking replica pod count",
			)

			By("verifying registered replica count and replication factor")

			status := areReplicasRegisteredEventually(&ctrlPod, jiva.ReplicaCount)
			Expect(status).To(
				Equal(true),
				"while verifying registered replica count as replication factor",
			)
		})
	})

	When("jiva pvc is deleted", func() {
		It("should not have any jiva controller and replica pods", func() {

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
				ctrlLabel,
				0,
			)
			Expect(controllerPodCount).To(
				Equal(0),
				"while checking controller pod count",
			)

			By("verifying replica pod count as 0")
			replicaPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				replicaLabel,
				0,
			)
			Expect(replicaPodCount).To(Equal(0), "while checking replica pod count")

			By("verifying deleted pvc")
			pvc := ops.IsPVCDeleted(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

		})
	})

})

func areReplicasRegisteredEventually(ctrlPod *corev1.Pod, replicationFactor int) bool {
	return Eventually(func() int {
		out, err := ops.PodClient.WithNamespace(namespaceObj.Name).
			Exec(
				ctrlPod.Name,
				&corev1.PodExecOptions{
					Command: []string{
						"/bin/bash",
						"-c",
						"curl http://localhost:9501/v1/volumes",
					},
					Container: ctrlPod.Spec.Containers[0].Name,
					Stdin:     false,
					Stdout:    true,
					Stderr:    true,
				},
			)
		Expect(err).ShouldNot(HaveOccurred(), "while exec in application pod")

		volumes := jivaClient.VolumeCollection{}
		err = json.Unmarshal([]byte(out.Stdout), &volumes)
		Expect(err).To(BeNil(), "while unmarshalling volumes %s", out.Stdout)

		return volumes.Data[0].ReplicaCount
	},
		300, 10).Should(Equal(replicationFactor))
}
