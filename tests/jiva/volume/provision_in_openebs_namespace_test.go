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
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	"github.com/openebs/maya/tests/jiva"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[jiva] TEST VOLUME PROVISIONING IN OPENEBS NAMESPACE", func() {

	var (
		pvcObj                *corev1.PersistentVolumeClaim
		scObj                 *storagev1.StorageClass
		scName                = "jiva-volume-sc-openebs-ns"
		pvcName               = "jiva-volume-claim-openebs-ns"
		pvcLabel              = "openebs.io/persistent-volume-claim=" + pvcName
		ctrlLabel             = jiva.CtrlLabel + "," + pvcLabel
		replicaLabel          = jiva.ReplicaLabel + "," + pvcLabel
		annotations           = map[string]string{}
		openebsCASConfigValue = `
- name: ReplicaCount
  value: $count
- name: DeployInOpenEBSNamespace
  enabled: true
`
	)

	When("jiva pvc with replicacount n is created", func() {
		It("should create 1 controller pod and n replica pod", func() {
			CASConfig := strings.Replace(openebsCASConfigValue, "$count", strconv.Itoa(jiva.ReplicaCount), 1)
			annotations[string(apis.CASTypeKey)] = string(apis.JivaVolume)
			annotations[string(apis.CASConfigKey)] = CASConfig

			By("building a storageclass")
			scObj, err = sc.NewBuilder().
				WithGenerateName(scName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building storageclass {%s}", scName)

			By("creating a storageclass")
			scObj, err = ops.SCClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass {%s}", scObj.GenerateName)

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
				string(artifacts.OpenebsNamespace),
				ctrlLabel,
				1,
			)
			Expect(controllerPodCount).To(
				Equal(1),
				"while checking controller pod count",
			)

			By("verifying replica pod count ")
			replicaPodCount := ops.GetPodRunningCountEventually(
				string(artifacts.OpenebsNamespace),
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
			podList, err = ops.PodClient.WithNamespace(string(artifacts.OpenebsNamespace)).List(
				metav1.ListOptions{LabelSelector: ctrlLabel},
			)
			Expect(err).ShouldNot(HaveOccurred(), "while listing controller pod")
			err = ops.PodClient.WithNamespace(string(artifacts.OpenebsNamespace)).
				Delete(podList.Items[0].Name, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting controller pod")

			By("verifying deleted pod is terminated")
			status := ops.IsPodDeletedEventually(
				string(artifacts.OpenebsNamespace),
				podList.Items[0].Name,
			)
			Expect(status).To(Equal(true), "while checking for deleted pod")

			By("verifying controller pod count")
			controllerPodCount := ops.GetPodRunningCountEventually(
				string(artifacts.OpenebsNamespace),
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
			err = ops.PodClient.WithNamespace(string(artifacts.OpenebsNamespace)).
				DeleteCollection(metav1.ListOptions{LabelSelector: replicaLabel},
					&metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting replica pods")

			By("verifying deleted pods are terminated")
			for _, p := range podList.Items {
				status := ops.IsPodDeletedEventually(string(artifacts.OpenebsNamespace), p.Name)
				Expect(status).To(
					Equal(true),
					"while checking for deleted pod {%s}",
					p.Name,
				)
			}
			By("verifying replica pod count")
			replicaPodCount := ops.GetPodRunningCountEventually(
				string(artifacts.OpenebsNamespace),
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
				string(artifacts.OpenebsNamespace),
			)

			By("verifying controller pod count as 0")
			controllerPodCount := ops.GetPodRunningCountEventually(
				string(artifacts.OpenebsNamespace),
				ctrlLabel,
				0,
			)
			Expect(controllerPodCount).To(
				Equal(0),
				"while checking controller pod count",
			)

			By("verifying replica pod count as 0")
			replicaPodCount := ops.GetPodRunningCountEventually(
				string(artifacts.OpenebsNamespace),
				replicaLabel,
				0,
			)
			Expect(replicaPodCount).To(Equal(0), "while checking replica pod count")

			By("verifying deleted pvc")
			pvc := ops.IsPVCDeleted(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

			By("deleting storageclass")
			err = ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass {%s}", scObj.Name)

		})
	})

})
