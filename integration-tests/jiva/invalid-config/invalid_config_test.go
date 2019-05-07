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

// TODO: Used third party library to get maya-apiserver logs(change upon review
// comments)

import (
	"fmt"

	script "github.com/bitfield/script"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// defaultReplicaLabel represents the jiva replica
	defaultReplicaLabel = "openebs.io/replica=jiva-replica"
	// defaultCtrlLabel represents the jiva controller
	defaultCtrlLabel      = "openebs.io/controller=jiva-controller"
	openebsProvisioner    = "openebs.io/provisioner-iscsi"
	accessModes           = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity              = "5G"
	scObj                 *storagev1.StorageClass
	pvcObj                *corev1.PersistentVolumeClaim
	nsObj                 *corev1.Namespace
	openebsCASConfigValue = "- name: ReplicaCount:\n  Value: 1"
	infraNamespace        = "openebs"
)

var _ = Describe("TEST INVALID CONFIGURATIONS IN SC", func() {
	var (
		namespaceName = "validation-ns1"
		scName        = "jiva-invalid-config-sc"
		pvcName       = "jiva-volume-claim"
		//errMsg        = "invalid sc cas config: " + openebsCASConfigValue + ": error converting YAML to JSON:"
		errMsg = "invalid sc cas config: - name: ReplicaCount:"
	)
	BeforeEach(func() {
		annotations := map[string]string{
			string(apis.CASTypeKey):   string(apis.JivaVolume),
			string(apis.CASConfigKey): openebsCASConfigValue,
		}
		var err error
		cOps = cOps.newPodClient(namespaceName).
			newSCClient().
			newPVCClient(namespaceName).
			newNsClient()

		nsObj, err = ns.NewBuilder().
			WithName(namespaceName).
			Build()
		Expect(err).ShouldNot(HaveOccurred())

		scObj, err = sc.NewBuilder().
			WithName(scName).
			WithAnnotations(annotations).
			WithProvisioner(openebsProvisioner).Build()
		Expect(err).ShouldNot(HaveOccurred())

		pvcObj, err = pvc.NewBuilder().
			WithName(pvcName).
			WithNamespace(namespaceName).
			WithStorageClass(scName).
			WithAccessModes(accessModes).
			WithCapacity(capacity).Build()
		Expect(err).ShouldNot(HaveOccurred())
	})

	When("We apply valid pvc and invalid sc yaml in k8s cluster", func() {
		It("Maya-apiserver log should show error message related to invalid configurations", func() {
			By(fmt.Sprintf("Create test specific namespace {%s}", namespaceName))
			_, err := cOps.nsClient.Create(nsObj)
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Create storageclass named {%s}", scName))
			_, err = cOps.scClient.Create(scObj)
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Create PVC named {%s} in Namespace: {%s}", pvcName, namespaceName))
			_, err = cOps.pvcClient.Create(pvcObj)
			Expect(err).To(BeNil())

			pvcLabel := string(apis.PersistentVolumeClaimKey) + "=" + pvcName
			replicaLabel := defaultReplicaLabel + "," + pvcLabel
			ctrlLabel := defaultCtrlLabel + "," + pvcLabel

			By("jiva-ctrl pod should not come to running state")
			// Check jiva controller pod should not create
			podCount := cOps.isRunningPodCount(namespaceName, ctrlLabel, 1)
			Expect(podCount).To(Equal(0))

			By("jiva-replica pod should not come to running state")
			// Check jiva controller pod should not create
			podCount = cOps.isRunningPodCount(namespaceName, replicaLabel, 3)
			Expect(podCount).To(Equal(0))

			// TODO: used pipe to get the logs of maya-apiserver and verify
			// the error logs are showing about invalid configurations(but
			// change based on review comments)
			// TODO: Refactor below snippet in builder pattern
			openebsPodKubeClient := pod.NewKubeClient(pod.WithNamespace(infraNamespace))
			pods, err := openebsPodKubeClient.
				List(metav1.ListOptions{LabelSelector: string(artifacts.MayaAPIServerLabelSelector)})
			podName := pods.Items[0].ObjectMeta.Name
			cmd := "kubectl logs -n " + infraNamespace + " " + podName
			lines, err := script.Exec(cmd).Match(errMsg).CountLines()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(lines).NotTo(Equal(0))
		})
	})

	AfterEach(func() {
		By(fmt.Sprintf("Delete PVC named {%s} in Namespace: {%s}", pvcName, namespaceName))
		err := cOps.pvcClient.Delete(pvcName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Delete storageclass named {%s}", scName))
		err = cOps.scClient.Delete(scName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Delete {%s} namespace", namespaceName))
		err = cOps.nsClient.Delete(namespaceName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())
	})

})

var _ = Describe("TEST INVALID CONFIGURATIONS IN PVC", func() {
	var (
		namespaceName         = "validation-ns2"
		scName                = "jiva-valid-config-sc"
		pvcName               = "jiva-ivalid-config-volume-claim"
		openebsCASConfigValue = "- name: ReplicaCount\n  Value: 1"
		invalidPVCLabel       = map[string]string{"name": "jiva-ivalid-config-volume-claim:"}
	)
	BeforeEach(func() {
		annotations := map[string]string{
			string(apis.CASTypeKey):   string(apis.JivaVolume),
			string(apis.CASConfigKey): openebsCASConfigValue,
		}
		var err error
		cOps = cOps.newPodClient(namespaceName).
			newSCClient().
			newPVCClient(namespaceName).
			newNsClient()

		nsObj, err = ns.NewBuilder().
			WithName(namespaceName).
			Build()
		Expect(err).ShouldNot(HaveOccurred())

		scObj, err = sc.NewBuilder().
			WithName(scName).
			WithAnnotations(annotations).
			WithProvisioner(openebsProvisioner).Build()
		Expect(err).ShouldNot(HaveOccurred())

		pvcObj, err = pvc.NewBuilder().
			WithName(pvcName).
			WithNamespace(namespaceName).
			WithLabels(invalidPVCLabel).
			WithStorageClass(scName).
			WithAccessModes(accessModes).
			WithCapacity(capacity).Build()
		Expect(err).ShouldNot(HaveOccurred())
	})

	When("We apply valid sc and invalid pvc yaml in k8s cluster", func() {
		It("PVC creation should give error because of invalid pvc yaml", func() {
			By(fmt.Sprintf("Create test specific namespace {%s}", namespaceName))
			_, err := cOps.nsClient.Create(nsObj)
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Create storageclass named {%s}", scName))
			_, err = cOps.scClient.Create(scObj)
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Create PVC named {%s} in Namespace: {%s}", pvcName, namespaceName))
			_, err = cOps.pvcClient.Create(pvcObj)
			Expect(err).NotTo(BeNil())
		})
	})

	AfterEach(func() {
		By(fmt.Sprintf("Delete PVC named {%s} in Namespace: {%s}", pvcName, namespaceName))
		err := cOps.pvcClient.Delete(pvcName, &metav1.DeleteOptions{})
		Expect(err).NotTo(BeNil())

		By(fmt.Sprintf("Delete storageclass named {%s}", scName))
		err = cOps.scClient.Delete(scName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())

		err = cOps.nsClient.Delete(namespaceName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())
	})

})
