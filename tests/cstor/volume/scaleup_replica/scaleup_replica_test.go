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
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("[REPLICA SCALEUP] CSTOR REPLICA SCALEUP", func() {
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
			By("Creating SPC, Desired Number of CSP Should Be Created", verifyDesiredCSPCount)
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
			By("Creating PVC, Desired Number of CVR Should Be Created", verifyVolumeStatus)
		})
	})
	When("CStor Replica ScaledUp", func() {
		It("Volume Replica Should Become Healthy and CStor Volume Configurations Should Be Updated", func() {
			By("Update DesiredReplicationFactor", updateDesiredReplicationFactor)
			By("Build and Create CStorVolumeReplica", buildAndCreateCVR)
			ReplicaCount = ReplicaCount + 1
			By("Verify Volume Status after ScaleUp Replica", verifyVolumeStatus)
			By("Verify Volume configurations from cstor volume", verifyVolumeConfigurationEventually)
		})
	})

	When("Clean up test resources", func() {
		It("Test related resources should be cleanedup", func() {
			By("Delete persistentVolumeClaim then volume resources should be deleted", deleteVolumeResources)
			By("Delete StoragePoolClaim then pool related resources should be deleted", deletePoolResources)
		})
	})
})
