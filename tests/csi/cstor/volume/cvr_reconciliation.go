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
	"github.com/openebs/maya/tests/cstor"
)

var _ = Describe("[cstor] [sparse] TEST VOLUME PROVISIONING WITH APP POD RESTART", func() {
	BeforeEach(prepareForCVRReconcilationTest)
	AfterEach(cleanupAfterCVRReconcilationTest)

	Context("App is deployed and restarted on pvc with replica count 1", func() {
		It("Should run Volume Creation Test", CVRReconcilationTest)
	})
})

func CVRReconcilationTest() {
	By("creating and verifying PVC bound status", createAndVerifyPVC)
	By("Creating and deploying app pod", createAndDeployAppPod)
	By("should verify target pod count as 1", func() { verifyTargetPodCount(1) })
	By("Verifying cstorvolume replica count", func() { verifyCstorVolumeReplicaCount(0) })
	By("Creating and verifying cstorpoolcluster", createAndVerifyCstorPoolCluster)
	By("Verifying cstorvolume replica count", func() { verifyCstorVolumeReplicaCount(cstor.ReplicaCount) })
	By("Creating and deploying app pod", verifyAppPodRunning)
	By("Deleting application deployment", deleteAppDeployment)
	By("Deleting pvc", deletePVC)
	By("Verifying deletion of components related to volume", verifyVolumeComponentsDeletion)
	By("Deleting cstorpoolcluster", deleteCstorPoolCluster)
}

func prepareForCVRReconcilationTest() {
	By("Creating storage class", createStorageClass)
}

func cleanupAfterCVRReconcilationTest() {
	By("Deleting storage class", deleteStorageClass)
}
