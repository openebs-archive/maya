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

import . "github.com/onsi/ginkgo"

var _ = Describe("[cstor] [sparse] TEST VOLUME PROVISIONING WITH APP POD RESTART", func() {
	BeforeEach(prepareForVolumeResizeTest)
	AfterEach(cleanupAfterVolumeResizeTest)

	When("app is deployed and restarted on pvc with replica count 1", volumeResizeTest)
})

func volumeResizeTest() {
	When("volumeResizeTest", func() {
		It("should crete and verify PVC bound status", CreateAndVerifyPVC)
		It("should crete and deploy app pod", CreateAndDeployApp)
		It("should verify presence of components related to volume", VerifyVolumeComponents)
		It("should expand PVC", expandPVC)
		It("should delete application deployment", deleteAppDeployment)
		It("should delete pvc", deletePVC)
		It("should verify volume components deletion", verifyVolumeComponentsDeletion)
	})
}

func prepareForVolumeResizeTest() {
	When("prepareForVolumeResizeTest", func() {
		By("should create and verify cstorpoolcluster", createAndVerifyCstorPoolCluster)
		By("should create storage class", createStorageClass)
	})
}

func cleanupAfterVolumeResizeTest() {
	When("cleanupAfterVolumeResizeTest", func() {
		By("should delete cstorpoolcluster", deleteCstorPoolCluster)
		By("should delete storage class", deleteStorageClass)
	})
}
