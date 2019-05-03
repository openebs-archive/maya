// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package replicadeletion

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	clientpvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	clientsc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	testTimes = 20
)

var _ = Describe("[jiva] [node-stickiness] jiva replica pod node-stickiness test", func() {
	var (
		// replicaLabel consist of defaultReplicaLabel and coressponding
		// pvcLabel
		replicaLabel string
		// ctrlLabel consist of defaultReplicaLabel and coressponding
		// pvcLabel
		ctrlLabel string
		//podListObj holds the PodList instance
		podListObj    *corev1.PodList
		podKubeClient *pod.KubeClient
		// defaultReplicaLabel represents the jiva replica
		defaultReplicaLabel = "openebs.io/replica=jiva-replica"
		// defaultCtrlLabel represents the jiva controller
		defaultCtrlLabel = "openebs.io/controller=jiva-controller"
		// defaultPVCLabel represents the default OpenEBS PVC label key
		defaultPVCLabel       = "openebs.io/persistent-volume-claim="
		storageEngine         = "jiva"
		replicaCount          = "1"
		openebsCASConfigValue = "- name: ReplicaCount\n  Value: " + replicaCount
		scName                = "jiva-single-replica"
		pvcName               = "jiva-vol1-1r-claim"
		testNamespace         = "jiva-rep-delete-ns"
		accessModes           = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity              = "5G"
		//TODO: following variables should be moved in framework or openebs-artifacts
		openebsCASType     = "cas.openebs.io/cas-type"
		openebsCASConfig   = "cas.openebs.io/config"
		openebsProvisioner = "openebs.io/provisioner-iscsi"
	)
	BeforeEach(func() {
		var err error
		By("Build jiva-single-replica storageclass and deploy it")
		annotations := map[string]string{
			openebsCASType:   storageEngine,
			openebsCASConfig: openebsCASConfigValue,
		}
		buildSCObj, err := clientsc.NewBuilder().
			WithName(scName).
			WithAnnotations(annotations).
			WithProvisioner(openebsProvisioner).Build()
		Expect(err).ShouldNot(HaveOccurred())

		_, err = clientsc.NewKubeClient(clientsc.WithKubeConfigPath(kubeConfigPath)).Create(buildSCObj)
		Expect(err).ShouldNot(HaveOccurred())

		By("Build and deploy PVC using jiva-single-replica storageClass in jiva-rep-delete-ns namespace")
		buildPVCObj, err := clientpvc.NewBuilder().
			WithName(pvcName).
			WithNamespace(testNamespace).
			WithStorageClass(scName).
			WithAccessModes(accessModes).
			WithCapacity(capacity).Build()
		Expect(err).ShouldNot(HaveOccurred())

		_, err = clientpvc.
			NewKubeClient(
				clientpvc.WithNamespace(testNamespace),
				clientpvc.WithKubeConfigPath(kubeConfigPath)).
			Create(buildPVCObj)
		Expect(err).ShouldNot(HaveOccurred())

		podKubeClient = pod.
			NewKubeClient(
				pod.WithNamespace(string(testNamespace)),
				pod.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig"))

		// pvcLabel represents the coressponding pvc
		pvcLabel := defaultPVCLabel + pvcName
		replicaLabel = defaultReplicaLabel + "," + pvcLabel
		ctrlLabel = defaultCtrlLabel + "," + pvcLabel
		// Verify creation of jiva ctrl pod
		_ = getPodList(podKubeClient, string(testNamespace), ctrlLabel, 1)

		// Verify creation of jiva replica pod
		podListObj = getPodList(podKubeClient, string(testNamespace), replicaLabel, 1)
	})

	AfterEach(func() {
		By("Uninstall test artifacts")
		err := clientpvc.
			NewKubeClient(
				clientpvc.WithNamespace(testNamespace),
				clientpvc.WithKubeConfigPath(kubeConfigPath)).
			Delete(pvcName, &metav1.DeleteOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		err = clientsc.
			NewKubeClient(
				clientsc.WithKubeConfigPath(kubeconfigPath)).
			Delete(scName, &metav1.DeleteOptions{})
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("node stickiness with jiva replica pod deletion", func() {
		//	var nodeName, podName string

		It("should verify jiva replica pod sticks to one node", func() {

			for i := 0; i < testTimes; i++ {
				By("fetching node name and podName of jiva replica pod")
				//nodeName holds name of the node where the replica pod deployed
				nodeName = podListObj.Items[0].Spec.NodeName
				podName = podListObj.Items[0].ObjectMeta.Name

				By(fmt.Sprintf("deleting the running jiva replica pod: '%s'", podName))
				err := podKubeClient.Delete(podName, &metav1.DeleteOptions{})
				Expect(err).ShouldNot(HaveOccurred())

				// Makesure that pod is deleted successfully
				Eventually(func() bool {
					_, err := podKubeClient.Get(podName, metav1.GetOptions{})
					if k8serror.IsNotFound(err) {
						return true
					}
					return false
				},
					defaultTimeOut, defaultPollingInterval).
					Should(BeTrue(), "Pod not found")

				By("waiting till jiva replica pod starts running")
				podListObj = getPodList(podKubeClient, string(testNamespace), replicaLabel, 1)

				By("verifying jiva replica pod node matches with its old instance node")
				Expect(podListObj.Items[0].Spec.NodeName).Should(Equal(nodeName))
			}
		})
	})
})
