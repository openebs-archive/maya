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

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/cstor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[cstor] TEST VOLUME PROVISIONING", func() {
	var (
		err     error
		pvcName = "cstor-volume-claim"
	)

	BeforeEach(func() {
		When(" creating a cstor based volume", func() {
			By("building object of storageclass")
			scObj, err = sc.NewBuilder().
				WithName(scName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building storageclass obj for storageclass {%s}", scName)

			By("creating storageclass")
			_, err = ops.SCClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass", scName)

			By("building spc object")
			spcObj = spc.NewBuilder().
				WithName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating storagepoolclaim")
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc", spcName)

			By("verifying healthy csp count")
			cspCount := ops.GetHealthyCSPCountEventually(spcName, cstor.PoolCount)
			Expect(cspCount).To(Equal(true), "while checking cstorpool health status")
			//		Expect(cspCount).To(Equal(1), "while checking cstorpool health count")

		})
	})

	AfterEach(func() {
		By("deleting resources created for testing cstor volume provisioning", func() {
			By("deleting storageclass")
			err = ops.SCClient.Delete(scName, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass", scName)

			By("listing spc")
			spcList, err = ops.SPCClient.List(metav1.ListOptions{})
			Expect(err).To(BeNil(), "while listing spc clients", spcList)

			By("deleting spc")
			for _, spc := range spcList.Items {
				_, err = ops.SPCClient.Delete(spc.Name, &metav1.DeleteOptions{})
				Expect(err).To(BeNil(), "while deleting the spc's", spc)
			}
		})
	})

	When("cstor pvc with replicacount 1 is created", func() {
		It("should create cstor volume target pod", func() {

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

			By("verifying target pod count as 1")
			controllerPodCount := ops.GetPodRunningCountEventually(openebsNamespace, targetLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			pvcObj, err = ops.PVCClient.WithNamespace(nsName).Get(pvcName, metav1.GetOptions{})
			Expect(err).To(
				BeNil(),
				"while getting pvc {%s} in namespace {%s}",
				pvcName,
				nsName,
			)

			By("verifying cstorvolume replica count")
			cvrLabel := "openebs.io/persistent-volume=" + pvcObj.Spec.VolumeName
			cvrCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, cvrLabel, cstor.ReplicaCount)
			Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")

			By("verifying pvc status as bound")
			status := ops.IsPVCBound(pvcName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

			By("deleting above pvc")
			err := ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				nsName,
			)

			By("verifying target pod count as 0")
			controllerPodCount = ops.GetPodRunningCountEventually(openebsNamespace, targetLabel, 0)
			Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

			By("verifying deleted pvc")
			pvc := ops.IsPVCDeleted(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

			By("verifying if cstorvolume is deleted")
			cvLabel := "openebs.io/persistent-volume=" + pvcObj.Spec.VolumeName
			cvCount := ops.GetCstorVolumeCountEventually(openebsNamespace, cvLabel, 0)
			Expect(cvCount).To(Equal(true), "while checking cstorvolume count")
		})
	})

})
