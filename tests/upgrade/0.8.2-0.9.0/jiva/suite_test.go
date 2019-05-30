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

package jiva

import (
	"flag"
	"strconv"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha2"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/artifacts"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	kubeConfigPath        string
	replicaCount          int
	nsName                = "default"
	scName                = "jiva-upgrade-sc"
	openebsProvisioner    = "openebs.io/provisioner-iscsi"
	replicaLabel          = "openebs.io/replica=jiva-replica"
	ctrlLabel             = "openebs.io/controller=jiva-controller"
	openebsCASConfigValue = "- name: ReplicaCount\n  Value: "
	accessModes           = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity              = "5G"
	pvcObj                *corev1.PersistentVolumeClaim
	pvcName               = "jiva-volume-claim"
	scObj                 *storagev1.StorageClass
	openebsURL            = "https://openebs.github.io/charts/openebs-operator-0.8.2.yaml"
	rbacURL               = "https://raw.githubusercontent.com/openebs/openebs/master/k8s/upgrades/0.8.2-0.9.0/rbac.yaml"
	crURL                 = "https://raw.githubusercontent.com/openebs/openebs/master/k8s/upgrades/0.8.2-0.9.0/jiva/cr.yaml"
	runtaskURL            = "https://raw.githubusercontent.com/openebs/openebs/master/k8s/upgrades/0.8.2-0.9.0/jiva/jiva_upgrade_runtask.yaml"
	jobURL                = "https://raw.githubusercontent.com/openebs/openebs/master/k8s/upgrades/0.8.2-0.9.0/jiva/volume-upgrade-job.yaml"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test jiva volume upgrade")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
	flag.IntVar(&replicaCount, "replicas", 1, "number of replicas to be created")
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(kubeConfigPath))
	openebsCASConfigValue = openebsCASConfigValue + strconv.Itoa(replicaCount)

	By("applying openebs 0.8.2")
	applyFromURL(openebsURL)

	By("waiting for maya-apiserver pod to come into running state")
	podCount := ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.MayaAPIServerLabelSelector),
		1,
	)
	Expect(podCount).To(Equal(1))

	annotations := map[string]string{
		string(apis.CASTypeKey):   string(apis.JivaVolume),
		string(apis.CASConfigKey): openebsCASConfigValue,
	}

	By("building a storageclass")
	scObj, err := sc.NewBuilder().
		WithName(scName).
		WithAnnotations(annotations).
		WithProvisioner(openebsProvisioner).Build()
	Expect(err).ShouldNot(HaveOccurred(), "while building storageclass {%s}", scName)

	By("creating above storageclass")
	_, err = ops.SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating storageclass {%s}", scObj.Name)

	By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(nsName).
		WithStorageClass(scName).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		nsName,
	)

	By("creating above pvc")
	_, err = ops.PVCClient.WithNamespace(nsName).Create(pvcObj)
	Expect(err).To(
		BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		nsName,
	)

	By("verifying controller pod count ")
	controllerPodCount := ops.GetPodRunningCountEventually(nsName, ctrlLabel, 1)
	Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

	By("verifying replica pod count ")
	replicaPodCount := ops.GetPodRunningCountEventually(nsName, replicaLabel, replicaCount)
	Expect(replicaPodCount).To(Equal(replicaCount), "while checking replica pod count")

	By("verifying status as bound")
	status := ops.IsPVCBound(pvcName)
	Expect(status).To(Equal(true), "while checking status equal to bound")

})

var _ = AfterSuite(func() {

	By("deleting above pvc")
	err := ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
	Expect(err).To(
		BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcName,
		nsName,
	)

	By("verifying controller pod count")
	controllerPodCount := ops.GetPodRunningCountEventually(nsName, ctrlLabel, 0)
	Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

	By("verifying replica pod count")
	replicaPodCount := ops.GetPodRunningCountEventually(nsName, replicaLabel, 0)
	Expect(replicaPodCount).To(Equal(0), "while checking replica pod count")

	By("deleting storageclass")
	err = ops.SCClient.Delete(scName, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting storageclass {%s}", scName)

	By("cleanup")
	deleteFromURL(jobURL)
	deleteFromURL(runtaskURL)
	deleteFromURL(crURL)
	deleteFromURL(rbacURL)
	deleteFromURL(openebsURL)
	By("waiting for maya-apiserver pod to terminate")
	podCount := ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.MayaAPIServerLabelSelector),
		0,
	)
	Expect(podCount).To(Equal(0))
	// deleting all completed pods
	podList, err := ops.PodClient.
		WithNamespace("default").
		List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	for _, po := range podList.Items {
		if po.Status.Phase == "Succeeded" {
			err = ops.PodClient.Delete(po.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting completed pods")
		}
	}
})

func applyFromURL(url string) {
	unstructList, err := unstruct.FromURL(url)
	Expect(err).ShouldNot(HaveOccurred())
	// Applying unstructured objects
	for _, us := range unstructList.Items {
		if us.Object.GetName() == "jiva-upgrade-config" {
			unstructured.SetNestedStringMap(us.Object.Object, data, "data")
		}
		if us.Object.GetName() == "jiva-volume-upgrade" {
			us.Object.SetNamespace("default")
		}
		err = ops.UnstructClient.Create(us.Object)
		Expect(err).ShouldNot(HaveOccurred())
	}
}

func deleteFromURL(url string) {
	unstructList, err := unstruct.FromURL(url)
	Expect(err).ShouldNot(HaveOccurred())
	// Deleting unstructured objects
	for _, us := range unstructList.Items {
		if us.Object.GetName() == "jiva-volume-upgrade" {
			us.Object.SetNamespace("default")
		}
		err = ops.UnstructClient.Delete(us.Object)
		Expect(err).ShouldNot(HaveOccurred())
	}
}
