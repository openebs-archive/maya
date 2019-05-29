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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[cstor] [sparse] TEST VOLUME PROVISIONING", func() {
	var (
		err     error
		pvcName = "cstor-volume-claim"
	)

	BeforeEach(func() {
		When(" creating a cstor based volume", func() {
			By("building a storageclass")
			scObj, err = sc.NewBuilder().
				WithName(scName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building storageclass obj for storageclass {%s}", scName)

			By("creating above storageclass")
			_, err = ops.SCClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass {%s}", scName)

			By("building a storagepoolclaim")
			spcObj = spc.NewBuilder().
				WithName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(1).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating above storagepoolclaim")
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc {%s}", spcName)

			By("verifying healthy cstorpool count")
			cspCount := ops.GetHealthyCSPCount(spcName, 1)
			Expect(cspCount).To(Equal(1), "while checking cstorpool health count")

		})
	})

	AfterEach(func() {
		By("deleting resources created for testing cstor volume provisioning", func() {
			It("should delete storageclass", func() {
				By("deleting storageclass")
				err = ops.SCClient.Delete(scName, &metav1.DeleteOptions{})
				Expect(err).To(BeNil(), "while deleting storageclass {%s}", scName)
			})
			It("should delete storagepoolclaim", func() {
				By("deleting storagepoolclaim")
				_, err = ops.SPCClient.Delete(spcName, &metav1.DeleteOptions{})
				Expect(err).To(BeNil(), "while deleting spc {%s}", spcName)
			})
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

			By("verifying pvc status as bound")
			status := ops.IsPVCBound(pvcName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

			It("should delete pvc", func() {
				By("deleting above pvc")
				err := ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
				Expect(err).To(
					BeNil(),
					"while deleting pvc {%s} in namespace {%s}",
					pvcName,
					nsName,
				)
			})

			By("verifying target pod count as 0")
			controllerPodCount = ops.GetPodRunningCountEventually(openebsNamespace, targetLabel, 0)
			Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

			By("verifying deleted pvc")
			pvc := ops.IsPVCDeleted(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

			By("verifying if cstorvolume is deleted")
			CstorVolumeLabel := "openebs.io/persistent-volume=" + pvcObj.Spec.VolumeName
			cvCount := ops.GetCstorVolumeCountEventually(openebsNamespace, CstorVolumeLabel, 0)
			Expect(cvCount).To(Equal(0), "while checking cstorvolume count")
		})
	})

})
