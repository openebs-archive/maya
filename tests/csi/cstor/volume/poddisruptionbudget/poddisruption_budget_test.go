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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cvc "github.com/openebs/maya/pkg/cstorvolumeclaim/v1alpha1"
	"github.com/openebs/maya/pkg/debug"
)

var _ = Describe("[csi] [cstor] TEST VOLUME PROVISIONING WITH POD DISRUPTION BUDGET", func() {
	BeforeEach(func() {
		By("Creating and verifying cstorpoolcluster", createAndVerifyCstorPoolCluster)
		By("Creating storage class", createStorageClass)
	})
	AfterEach(func() {
		By("Deleting cstorpoolcluster", deleteCstorPoolCluster)
		By("Deleting storage class", deleteStorageClass)
	})

	Context("When HA volume is created PodDisruptionBudget should be created", func() {
		It("Volume Should be in Running and PDB should be Created", volumeCreationTest)
	})
})

var _ = Describe("[csi] [cstor-negative] TEST VOLUME PROVISIONING BY INJECTING ERRORS IN PODDISRUPTIONBUDGET OPERATIONS", func() {
	BeforeEach(func() {
		By("Creating and verifying cstorpoolcluster", createAndVerifyCstorPoolCluster)
		By("Creating storage class", createStorageClass)
	})
	AfterEach(func() {
		By("Deleting cstorpoolcluster", deleteCstorPoolCluster)
		By("Deleting storage class", deleteStorageClass)
	})

	Context("When HA volume is created by injecting errors in PodDisruptionBudget operations", func() {
		It("CStorVolumeConfig should be claimed after removing errors", func() {
			By("Create service for CVC Operator", buildAndCreateService)
			By("Inject errors in PDB Operations", func() { injectOrEjectPDBErrors(debug.Inject) })
			By("creating PVC", func() { pvcObj = createPVC() })
			// wait for some random time
			time.Sleep(20)
			By("Verify CStorVolumeClaim Pending Status", func() {
				ops.VerifyCVCStatusEventually(pvcObj.Spec.VolumeName, openebsNamespace, 1, cvc.PredicateList{cvc.IsCVCPending()})
			})
			By("Eject errors in PDB Operations", func() { injectOrEjectPDBErrors(debug.Eject) })
			By("Verifying the presence of components related to volume", verifyVolumeComponents)
			By("Verifying the poddisruption budget of volume", func() {
				err := ops.VerifyPodDisruptionBudget(pvcObj.Spec.VolumeName, openebsNamespace)
				Expect(err).To(BeNil(), "error occuered while checking the pod disruption budget")
			})
			By("Deleteing service", deleteSVC)
			By("Deleting pvc", deletePVC)
			By("Verifying deletion of components related to volume", verifyVolumeComponentsDeletion)
		})
	})
})

func volumeCreationTest() {
	By("creating and verifying PVC bound status", createAndVerifyPVC)
	By("Verifying the presence of components related to volume", verifyVolumeComponents)
	By("Verifying the poddisruption budget of volume", func() {
		err := ops.VerifyPodDisruptionBudget(pvcObj.Spec.VolumeName, openebsNamespace)
		Expect(err).To(BeNil(), "error occuered while checking the pod disruption budget")
	})
	By("Deleting pvc", deletePVC)
	By("Verifying deletion of components related to volume", verifyVolumeComponentsDeletion)
}
