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

package replicascaleup

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"

	// auth plugins
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("[REPLICA SCALEUP/SCALEDOWN] CSTOR REPLICA SCALEUP And SCALEDOWN", func() {
	When("SPC is created", func() {
		It("cStor Pools Should be Provisioned ", func() {

			By("Build And Create StoragePoolClaim object")
			// Populate configurations and create
			spcConfig := &tests.SPCConfig{
				Name:               spcName,
				DiskType:           "sparse",
				PoolCount:          cstor.PoolCount,
				IsOverProvisioning: false,
				PoolType:           "striped",
			}
			ops.Config = spcConfig
			spcObj = ops.BuildAndCreateSPC()
			By("Creating SPC, Desired Number of CSP Should Be Created", func() {
				ops.VerifyDesiredCSPCount(spcObj, cstor.PoolCount)
			})
		})
	})
	When("Persistent Volume Claim Is Created", func() {
		It("Volume Should be Created and Provisioned", func() {
			By("Build And Create StorageClass", buildAndCreateSC)
			pvcConfig := &tests.PVCConfig{
				Name:        pvcName,
				Namespace:   nsObj.Name,
				SCName:      scObj.Name,
				Capacity:    "5G",
				AccessModes: accessModes,
			}
			ops.Config = pvcConfig
			pvcObj = ops.BuildAndCreatePVC()
			By("Creating PVC, Desired Number of CVR Should Be Created", func() {
				// ReplicaCount is initilized as 1
				ops.VerifyVolumeStatus(pvcObj, ReplicaCount)
			})
			// GetLatest PVC object
			var err error
			pvcObj, err = ops.PVCClient.
				WithNamespace(nsObj.Name).
				Get(pvcObj.Name, metav1.GetOptions{})
			Expect(err).To(BeNil())
		})
	})
	When("CStor Replica ScaledUp", func() {
		It("Volume Replica Should Become Healthy and CStor Volume Configurations Should Be Updated", func() {
			By("Update DesiredReplicationFactor", updateDesiredReplicationFactor)
			By("Build and Create CStorVolumeReplica", buildAndCreateCVR)
			ReplicaCount = ReplicaCount + 1
			By("Verify Volume Status after ScaleUp Replica", func() {
				ops.VerifyVolumeStatus(pvcObj, ReplicaCount)
			})
			By("Verify Volume configurations from cstor volume", verifyVolumeConfigurationEventually)
		})
	})
	When("CStor Replica ScaleDown", func() {
		It("Volume Replica Should be disconnected and CStor Volume Configurations Should Be Updated", func() {
			ReplicaCount = ReplicaCount - 1
			replicaIDList := []string{ReplicaID}
			By("Update cStor Volume Configurations", func() {
				updateCStorVolumeConfigurations(ReplicaCount, replicaIDList)
			})
			By("Verify CStorVolume Configurations After Performing Scaledown", func() {
				verifyCVConfigForReplicaScaleDownEventually(ReplicaCount)
			})
			By("Verify Volume Status after ScaleDown Replica", func() {
				ops.VerifyVolumeStatus(pvcObj, ReplicaCount)
			})
		})
	})

	When("Clean up test resources", func() {
		It("Test related resources should be cleanedup", func() {
			By("Delete persistentVolumeClaim then volume resources should be deleted", func() {
				ops.DeleteVolumeResources(pvcObj, scObj)
			})
			By("Delete StoragePoolClaim then pool related resources should be deleted", func() {
				ops.DeleteStoragePoolClaim(spcObj.Name)
			})
		})
	})
})
