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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests/cstor"
)

/* To run positive test please run by following commang
 * ginkgo -v -focus="\[csi\]\ \[cstor-scaling\]" -- -kubeconfig=<path_to_kube_config>
 * -cstor-maxpools=<no.of_pool> -cstor-replicas=<replica_count>
 * -cstor-pool-type=<type of pool>
 */
var _ = Describe("[csi] [cstor-scaling] TEST SCALING OF CSTORVOLUMEREPLICAS", func() {
	BeforeEach(func() {
		By("Creating and verifying cstorpoolcluster", createAndVerifyCstorPoolCluster)
		By("Creating storage class", createStorageClass)
	})
	AfterEach(func() {
		By("Deleting cstorpoolcluster", deleteCstorPoolCluster)
		By("Deleting storage class", deleteStorageClass)
	})

	Context("Creating CSI Volumes and Scaling VolumeReplicas", func() {
		It("Should create cstor volume and should perform scaling of application", func() {
			By("creating and verifying PVC bound status", createAndVerifyPVCStatus)
			By("Verifying the presence of components related to volume", func() { verifyVolumeComponents(cstor.ReplicaCount) })
			By("Verifying the poddisruption budget of volume", func() {
				err := ops.VerifyPodDisruptionBudget(pvcObj.Spec.VolumeName, openebsNamespace)
				Expect(err).To(BeNil(), "error occuered while checking the pod disruption budget")
			})
			currentReplicaCount := cstor.ReplicaCount + 1
			By(fmt.Sprintf("Scale the CStorVolumeReplicas From %d to %d", cstor.ReplicaCount, currentReplicaCount), func() {
				err := ops.ScaleUpCVC(pvcObj.Spec.VolumeName, openebsNamespace, 1)
				Expect(err).To(BeNil(), "error occuered while scaling the volume replicas")
			})
			By("Verifying the volume components after scaling volume replicas", func() { verifyVolumeComponents(currentReplicaCount) })

			replicaCount := currentReplicaCount - 1
			By(fmt.Sprintf("Scale the CStorVolumeReplicas From %d to %d", currentReplicaCount, replicaCount), func() {
				err := ops.ScaleDownCVC(pvcObj.Spec.VolumeName, openebsNamespace, 1)
				Expect(err).To(BeNil(), "error occuered while scaling the volume replicas")
			})
			By("Verifying the volume components after scaling volume replicas", func() { verifyVolumeComponents(currentReplicaCount) })

			By("Deleting pvc", deletePVC)
			By("Verifying deletion of components related to volume", verifyVolumeComponentsDeletion)
		})
	})
})
