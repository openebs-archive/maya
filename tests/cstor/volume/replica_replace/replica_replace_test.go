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

// Included replicareplace and replicamigration in same test
package replicareplace

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cv "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"

	// auth plugins
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("[REPLICA REPLACE] CSTOR REPLICA REPLACE", func() {
	When("SPC is created", func() {
		It("cStor Pools Should be Provisioned ", func() {

			By("Build And Create StoragePoolClaim object")
			// Populate configurations and create
			spcConfig := &tests.SPCConfig{
				Name:                spcName,
				DiskType:            "sparse",
				PoolCount:           cstor.PoolCount,
				IsThickProvisioning: true,
				PoolType:            "striped",
			}
			ops.Config = spcConfig
			spcObj = ops.BuildAndCreateSPC()
			By("Creating SPC, Desired Number of CSP Should Be Created", func() {
				ops.VerifyDesiredCSPCount(spcObj, cstor.PoolCount)
			})
		})
	})
	When("PersistentVolumeClaim Is Created", func() {
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
				ops.VerifyVolumeStatus(pvcObj,
					cstor.ReplicaCount,
					cvr.PredicateList{cvr.IsHealthy()},
					cv.PredicateList{cv.IsHealthy()},
				)
			})
			// GetLatest PVC object
			var err error
			pvcObj, err = ops.PVCClient.
				WithNamespace(nsObj.Name).
				Get(pvcObj.Name, metav1.GetOptions{})
			Expect(err).To(BeNil())
		})
	})
	When("More Than Quorum Replicas Replaced In CStor", func() {
		It("Volume Replica Should Become Healthy and CStor Volume Configurations Should Be Updated Accordingly", func() {
			By("Destroy Quorum Count of Volume Datasets", deleteZFSDataSets)
			By("Update CStorVolume Configurations to Start Rebuild", updateCVConfigurationsAndVerifyStatus)
			By("Restart CStor-Pool-Mgmt Container Pods Doesn't Have Volume DataSets", restartPoolPods)
			By("Verify Volume Status after Performing Replica Replace", func() {
				ops.VerifyVolumeStatus(pvcObj,
					cstor.ReplicaCount,
					cvr.PredicateList{cvr.IsHealthy()},
					cv.PredicateList{cv.IsHealthy()},
				)
			})
			By("Verify Volume configurations from cstor volume", verifyCVConfigForReplicaReplaceEventually)
		})
	})

	When("Replica Moved To Different Pool", func() {
		It("Volume Replica Should Become Healthy and CStor Volume Configurations Should Be Updated Accordingly", func() {
			By("Build And Create CstorVolumeReplica then Delete Which Has To Migrate Replica", migrateReplica)
			By("Verify Volume configurations from cstorvolume", verifyCVConfigForReplicaMigrationEventually)
			By("Verify Volume Status after Performing Replica Movement", func() {
				ops.VerifyVolumeStatus(pvcObj,
					cstor.ReplicaCount,
					cvr.PredicateList{cvr.IsHealthy()},
					cv.PredicateList{cv.IsHealthy()},
				)
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
