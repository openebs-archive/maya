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
)

var _ = Describe("[csi] [cstor] TEST VOLUME PROVISIONING WITH APP POD RESTART", func() {
	BeforeEach(prepareForVolumeCreationTest)
	AfterEach(cleanupAfterVolumeCreationTest)

	Context("App is deployed and restarted on pvc with replica count 1", func() {
		It("Should run Volume Creation Test", volumeCreationTest)
	})
})

func volumeCreationTest() {
	By("creating and verifying PVC bound status", createAndVerifyPVC)
	By("Creating and deploying app pod", createDeployVerifyApp)
	By("Verifying the presence of components related to volume", verifyVolumeComponents)
	By("Restarting app pod and verifying app pod running status", restartAppPodAndVerifyRunningStatus)
	By("Deleting application deployment", deleteAppDeployment)
	By("Deleting pvc", deletePVC)
	By("Verifying deletion of components related to volume", verifyVolumeComponentsDeletion)
}

func prepareForVolumeCreationTest() {
	By("Creating and verifying cstorpoolcluster", createAndVerifyCstorPoolCluster)
	By("Creating storage class", createStorageClass)
}

func cleanupAfterVolumeCreationTest() {
	By("Deleting cstorpoolcluster", deleteCstorPoolCluster)
	By("Deleting storage class", deleteStorageClass)
}
