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

package provision

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	replicaLabel          = "openebs.io/replica=jiva-replica"
	ctrlLabel             = "openebs.io/controller=jiva-controller"
	openebsProvisioner    = "openebs.io/provisioner-iscsi"
	accessModes           = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity              = "5G"
	scObj                 *storagev1.StorageClass
	pvcObj                *corev1.PersistentVolumeClaim
	nsObj                 *corev1.Namespace
	openebsCASConfigValue = "- name: ReplicaCount\n  Value: 1"
	err                   error
)

var _ = Describe("[jiva] TEST VOLUME PROVISIONING", func() {
	var (
		nsName      = "provision-ns"
		scName      = "jiva-pods-in-openebs-ns"
		pvcName     = "jiva-volume-claim"
		annotations = map[string]string{
			string(apis.CASTypeKey):   string(apis.JivaVolume),
			string(apis.CASConfigKey): openebsCASConfigValue,
		}
	)
	BeforeEach(func() {

	})

	When("creating namespace and storageclass", func() {
		It("should create namespace and storageclass", func() {

			By("Building a namespace")
			nsObj, err = ns.NewBuilder().
				WithName(nsName).
				APIObject()
			Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", nsName)

			By("Building a storageclass")
			scObj, err = sc.NewBuilder().
				WithName(scName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building storageclass {%s}", scName)

			By("Creating a storageclass")
			_, err = ops.scClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass {%s}", scObj.Name)

			By("Creating a namespace")
			_, err = ops.nsClient.Create(nsObj)
			Expect(err).To(BeNil(), "while creating storageclass {%s}", nsObj.Name)
		})
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

	When("deleting namespace and storageclass", func() {
		It("should delete namespace and storageclass", func() {

			By("deleting storageclass")
			err := ops.scClient.Delete(scName, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storrageclass {%s}", scObj.Name)

			By("deleting namespace")
			err = ops.nsClient.Delete(nsName, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storrageclass {%s}", nsObj.Name)

		})
	})
	AfterEach(func() {

	})

})
