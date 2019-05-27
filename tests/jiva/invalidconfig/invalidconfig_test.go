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

package invalidconfig

import (
	"fmt"

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
	// defaultReplicaLabel represents the jiva replica
	defaultReplicaLabel = "openebs.io/replica=jiva-replica"
	// defaultCtrlLabel represents the jiva controller
	defaultCtrlLabel   = "openebs.io/controller=jiva-controller"
	openebsProvisioner = "openebs.io/provisioner-iscsi"
	accessModes        = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity           = "5G"
	scObj              *storagev1.StorageClass
	pvcObj             *corev1.PersistentVolumeClaim
	nsObj              *corev1.Namespace
	// openebsCASConfigValue contains invalid name ReplicaCount:
	openebsCASConfigValue = "- name: ReplicaCount:\n  Value: " + replicaCount
)

var _ = Describe("[jiva] [-ve] TEST INVALID CAS CONFIGURATIONS IN SC", func() {
	var (
		nsName       = "validation-ns1"
		scName       = "jiva-invalid-config-sc"
		pvcName      = "jiva-volume-claim"
		pvcLabel     = "openebs.io/persistent-volume-claim" + "=" + pvcName
		replicaLabel = defaultReplicaLabel + "," + pvcLabel
		ctrlLabel    = defaultCtrlLabel + "," + pvcLabel
	)
	BeforeEach(func() {
		var err error
		annotations := map[string]string{
			string(apis.CASTypeKey):   string(apis.JivaVolume),
			string(apis.CASConfigKey): openebsCASConfigValue,
		}

		By("building a namespace")
		nsObj, err = ns.NewBuilder().
			WithName(nsName).
			APIObject()
		Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", nsName)

		By("building a storageclass")
		scObj, err = sc.NewBuilder().
			WithName(scName).
			WithAnnotations(annotations).
			WithProvisioner(openebsProvisioner).Build()
		Expect(err).ShouldNot(HaveOccurred(), "while building storageclass {%s}", scName)

		By("building a persistentvolumeclaim")
		pvcObj, err = pvc.NewBuilder().
			WithName(pvcName).
			WithNamespace(nsName).
			WithStorageClass(scName).
			WithAccessModes(accessModes).
			WithCapacity(capacity).Build()
		Expect(err).ShouldNot(HaveOccurred(), "while building persistentvolumeclaim {%s} in namespace {%s}", pvcName, nsName)

		By("creating a namespace")
		_, err = ops.NSClient.Create(nsObj)
		Expect(err).To(BeNil(), "while creating namespace {%s}", nsObj.Name)

		By("creating a storageclass")
		_, err = ops.SCClient.Create(scObj)
		Expect(err).To(BeNil(), "while creating storageclass {%s}", scObj.Name)

	})

	When("jiva persistentvolumeclaim referring to invalid storageclass is applied", func() {
		It("should not create Jiva controller and replica pods", func() {

			By("creating a persistentvolumeclaim")
			_, err := ops.PVCClient.WithNamespace(nsName).Create(pvcObj)
			Expect(err).To(BeNil(), "while creating persistentvolumeclaim {%s} in namespace {%s}", pvcObj.Name, nsName)

			By("verifying controller pod count")
			controllerPodCount := ops.GetPodRunningCountEventually(nsName, ctrlLabel, 1)
			Expect(controllerPodCount).To(Equal(0), "while checking jiva controller pod count")

			By("verifying replica pod count")
			replicaPodCount := ops.GetPodRunningCountEventually(nsName, replicaLabel, repCountInt)
			Expect(replicaPodCount).To(Equal(0), "while checking jiva replica pod count")
		})
	})

	AfterEach(func() {
		By("deleting persistentvolumeclaim")
		err := ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting persistentvolumeclaim {%s} in namespace {%s}", pvcObj.Name, nsName)

		By("deleting storageclass")
		err = ops.SCClient.Delete(scName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting storrageclass {%s}", scObj.Name)

		By("deleting namespace")
		err = ops.NSClient.Delete(nsName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting namespace {%s}", nsName)
	})

})

var _ = Describe("[jiva] [-ve] TEST INVALID CONFIGURATIONS IN persistentvolumeclaim", func() {
	var (
		nsName                = "validation-ns2"
		scName                = "jiva-valid-config-sc"
		pvcName               = "jiva-invalid-config-volume-claim"
		openebsCASConfigValue = "- name: ReplicaCount\n  Value: " + replicaCount
		// invalidPVCLabel contains invalid label value
		invalidPVCLabel = map[string]string{"name": "jiva-invalid-config-volume-claim:"}
	)
	BeforeEach(func() {
		annotations := map[string]string{
			string(apis.CASTypeKey):   string(apis.JivaVolume),
			string(apis.CASConfigKey): openebsCASConfigValue,
		}
		var err error

		By("building a namespace")
		nsObj, err = ns.NewBuilder().
			WithName(nsName).
			APIObject()
		Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", nsName)

		By("building a storageclass")
		scObj, err = sc.NewBuilder().
			WithName(scName).
			WithAnnotations(annotations).
			WithProvisioner(openebsProvisioner).Build()
		Expect(err).ShouldNot(HaveOccurred(), "while building storageclass {%s}", scName)

		By("building a persistentvolumeclaim")
		pvcObj, err = pvc.NewBuilder().
			WithName(pvcName).
			WithNamespace(nsName).
			WithLabels(invalidPVCLabel).
			WithStorageClass(scName).
			WithAccessModes(accessModes).
			WithCapacity(capacity).Build()
		Expect(err).ShouldNot(HaveOccurred(), "while building persistentvolumeclaim {%s} in namespace {%s}", pvcName, nsName)

		By("creating a namespace")
		_, err = ops.NSClient.Create(nsObj)
		Expect(err).To(BeNil(), "while creating namespace {%s}", nsObj.Name)

		By("createing a storageclass")
		_, err = ops.SCClient.Create(scObj)
		Expect(err).To(BeNil(), "while creating storageclass {%s}", scObj.Name)
	})

	When("We apply invalid persistentvolumeclaim yaml in k8s cluster", func() {
		It("PVC creation should give error because of invalid pvc yaml", func() {
			By(fmt.Sprintf("create PVC named {%s} in Namespace: {%s}", pvcName, nsName))
			_, err := ops.PVCClient.WithNamespace(nsName).Create(pvcObj)
			Expect(err).NotTo(BeNil(), "while creating persistentvolumeclaim {%s} in namespace {%s}", pvcObj.Name, nsName)
		})
	})

	AfterEach(func() {
		By("deleting persistentvolumeclaim")
		err := ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
		Expect(err).NotTo(BeNil(), "while deleting persistentvolumeclaim {%s} in namespace {%s}", pvcName, nsName)

		By("deleting storageclass")
		err = ops.SCClient.Delete(scName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting storageclass {%s}", scName)

		By("deleting namespace")
		err = ops.NSClient.Delete(nsName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting namespace {%s}", nsName)
	})

})
