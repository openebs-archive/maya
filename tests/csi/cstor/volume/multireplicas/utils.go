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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	cvc "github.com/openebs/maya/pkg/cstorvolumeclaim/v1alpha1"
	debug "github.com/openebs/maya/pkg/debug"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"replicaCount":     strconv.Itoa(cstor.ReplicaCount),
		"cstorPoolCluster": cspcObj.Name,
		"cas-type":         "cstor",
	}

	By("building a storageclass")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(openebsProvisioner).
		WithVolumeExpansion(true).Build()
	Expect(err).ShouldNot(HaveOccurred(),
		"while building storageclass obj with prefix {%s}", scName)

	By("creating above storageclass")
	scObj, err = ops.SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating storageclass with prefix {%s}", scName)
}

func deleteCstorPoolCluster() {
	By("deleting cstorpoolcluster")
	err := ops.CSPCClient.Delete(
		cspcObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting cspc {%s}", cspcObj.Name)
	By("verifying deleted cspc")
	status := ops.IsCSPCDeletedEventually(cspcObj.Name)
	Expect(status).To(Equal(true), "while trying to get deleted cspc")
}

func deleteStorageClass() {
	By("deleting storageclass")
	err := ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(),
		"while deleting storageclass {%s}", scObj.Name)
}

func createAndVerifyCstorPoolCluster() {
	var err error
	cspcConfig := tests.CSPCConfig{
		Name:      cspcName,
		PoolType:  cstor.PoolType,
		PoolCount: cstor.PoolCount,
		Namespace: openebsNamespace,
	}
	ops.Config = &cspcConfig
	cspcObj, err = ops.BuildAndCreateCSPC()
	Expect(err).To(BeNil(), "while building and creating cstorpoolcluster {%s}", cspcName)

	By("verifying healthy cstorpool count")
	ops.NameSpace = openebsNamespace
	cspCount := ops.GetCSPICountWithCSPCName(
		cspcObj.Name, cstor.PoolCount, []cspi.Predicate{cspi.IsStatus("ONLINE")})
	Expect(cspCount).To(Equal(cstor.PoolCount),
		"while checking healthy cstor pool count")
}

func createPVC() *corev1.PersistentVolumeClaim {
	pvcConfig := tests.PVCConfig{
		Name:        pvcName,
		Namespace:   nsObj.Name,
		SCName:      scObj.Name,
		AccessModes: accessModes,
		Capacity:    capacity,
	}
	ops.Config = &pvcConfig
	return ops.BuildAndCreatePVC()
}

func createAndVerifyPVCStatus() {
	var err error
	pvcObj = createPVC()

	By("verifying pvc status as bound")
	status := ops.IsPVCBoundEventually(pvcName)
	Expect(status).To(Equal(true),
		"while checking status equal to bound")

	pvcObj, err = ops.PVCClient.WithNamespace(nsObj.Name).Get(pvcObj.Name, metav1.GetOptions{})
	Expect(err).To(
		BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		nsObj.Name,
	)
}

func verifyVolumeComponents(replicaCount int) {
	By("Verify CStorVolume target pod status", func() { verifyTargetPodCount(1) })
	By("Verify CStorVolumeReplica count", func() { verifyCstorVolumeReplicaCount(replicaCount) })
	By("Verify CStorVolumeClaim Status", func() {
		ops.VerifyCVCStatusEventually(cspcObj.Name, openebsNamespace, 1,
			cvc.PredicateList{cvc.IsCVCBounded(), cvc.HasAnnotation(cvcVolumeAnnotationKey, pvcObj.Spec.VolumeName)})
	})

	Eventually(func() bool {
		cvcObj, err := ops.CVCClient.WithNamespace(openebsNamespace).Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
		Expect(err).To(BeNil())
		return len(cvcObj.Status.PoolInfo) == replicaCount
        },
                120, 10).Should(BeTrue())
}

func verifyTargetPodCount(count int) {
	By("verifying target pod count as 1 once the app has been deployed")
	targetVolumeLabel := pvLabel + pvcObj.Spec.VolumeName
	controllerPodCount := ops.GetPodRunningCountEventually(
		openebsNamespace, targetVolumeLabel, count)
	Expect(controllerPodCount).To(Equal(count),
		"while checking controller pod count")
}

func verifyCstorVolumeReplicaCount(count int) {
	targetVolumeLabel := pvLabel + pvcObj.Spec.VolumeName
	isReqCVRCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, targetVolumeLabel, count, cvr.IsHealthy())
	Expect(isReqCVRCount).To(Equal(true), "while checking cstorvolume replica count")
}

func deletePVC() {
	err := ops.PVCClient.WithNamespace(nsObj.Name).Delete(pvcName, &metav1.DeleteOptions{})
	Expect(err).To(
		BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcName,
		nsObj.Name,
	)
	By("verifying deleted pvc")
	status := ops.IsPVCDeletedEventually(pvcName)
	Expect(status).To(Equal(true), "while trying to get deleted pvc")

}

func verifyVolumeComponentsDeletion() {
	By("verifying target pod count as 0")
	controllerPodCount := ops.GetPodRunningCountEventually(
		openebsNamespace, targetLabel, 0)
	Expect(controllerPodCount).To(Equal(0),
		"while checking controller pod count")

	By("verifying if cstorvolume is deleted")
	CstorVolumeLabel := "openebs.io/persistent-volume=" + pvcObj.Spec.VolumeName
	cvCount := ops.GetCstorVolumeCountEventually(
		openebsNamespace, CstorVolumeLabel, 0)
	Expect(cvCount).To(Equal(true), "while checking cstorvolume count")
	By("Verifying cstorvolume replica count", func() { verifyCstorVolumeReplicaCount(0) })
}

func buildAndCreateService() {
	var nodeIP string
	cvcLabel := "openebs.io/component-name=cvc-operator"
	poolPodList, err := ops.PodClient.
		WithNamespace(openebsNamespace).
		List(metav1.ListOptions{LabelSelector: cvcLabel})
	Expect(err).To(BeNil())
	Expect(len(poolPodList.Items)).Should(BeNumerically("==", 1),
		"Mismatch count of CVC operator pod")
	servicePort := []corev1.ServicePort{
		corev1.ServicePort{
			Name:     "injector",
			Port:     int32(8080),
			Protocol: "TCP",
			NodePort: int32(targetPort),
		},
	}
	nodeName := poolPodList.Items[0].Spec.NodeName
	nodeObj, err := ops.NodeClient.Get(nodeName, metav1.GetOptions{})
	Expect(err).To(BeNil())

	//GetNode Ip
	for _, address := range nodeObj.Status.Addresses {
		if address.Type == corev1.NodeExternalIP {
			nodeIP = address.Address
			break
		}
	}
	hostIPPort = nodeIP + ":" + strconv.Itoa(targetPort)
	serviceConfig := &tests.ServiceConfig{
		Name:        svcName,
		Namespace:   openebsNamespace,
		Selectors:   poolPodList.Items[0].Labels,
		ServicePort: servicePort,
		ServiceType: corev1.ServiceTypeNodePort,
	}
	ops.Config = serviceConfig
	serviceObj = ops.BuildAndCreateService()
}

func injectOrEjectPDBErrors(injectOrEject string) {
	//Injecting Errors During PDB Creation Time
	injectError := debug.NewClient(hostIPPort)
	err := injectError.PostInject(
		debug.NewErrorInjection().
			WithPDBCreateError(injectOrEject))
	Expect(err).To(BeNil())
}

func deleteSVC() {
	_ = ops.SVCClient.WithNamespace(openebsNamespace).Delete(serviceObj.Name, &metav1.DeleteOptions{})
}
