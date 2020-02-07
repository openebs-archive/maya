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

package reconciliation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	spc_v1alpha1 "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	zpoolGUIDCmd = "zpool get guid -H | awk '{print $3}'"
)

// This function is local to this package
func getLabelSelector(spc *apis.StoragePoolClaim) string {
	return string(apis.StoragePoolClaimCPK) + "=" + spc.Name
}

var _ = Describe("STRIPED SPARSE SPC", func() {

	When("We apply sparse-striped-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and healthy status", func() {
			Spc = spc_v1alpha1.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(3).
				WithThickProvisioning(true).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build()
			spcObj = Spc.Object

			// Create a storage pool claim
			_, err := ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
			// Check blockdeviceclaim count after creating the cstor pool
			bdcCount := ops.GetBDCCount(getLabelSelector(spcObj), string(artifacts.OpenebsNamespace))
			Expect(bdcCount).To(
				Equal(spc_v1alpha1.DefaultDiskCount[string(apis.PoolTypeStripedCPV)]*3),
				"mismatch of blockdeviceclaim count",
			)
			// Check is there any extra csp's are created
			cspCount = ops.GetCSPCount(getLabelSelector(spcObj))
			Expect(cspCount).To(Equal(3), "mismatch of csp count")

			Expect(ops.IsSPCFinalizerExistsOnSPC(spcObj.Name, spc_v1alpha1.SPCFinalizer)).To(BeTrue())

			Expect(ops.IsSPCFinalizerExistsOnBDCs(metav1.ListOptions{
				LabelSelector: getLabelSelector(spcObj),
			}, spc_v1alpha1.SPCFinalizer)).To(BeTrue())
		})
	})

	When("Remove finalizer", func() {
		It("make sure finalizer comes back as part of reconcilation", func() {
			err := Spc.RemoveFinalizer(spc_v1alpha1.SPCFinalizer)
			Expect(err).To(BeNil())
			Expect(ops.IsSPCFinalizerExistsOnSPC(spcObj.Name, spc_v1alpha1.SPCFinalizer)).To(BeTrue())
		})
	})

	When("We we delete cstor sparse pool pod", func() {
		It("pool should be imported", func() {
			namespace := string(artifacts.OpenebsNamespace)
			expectOutput := true

			cspList, err := ops.CSPClient.List(
				metav1.ListOptions{
					LabelSelector: "openebs.io/storage-pool-claim=" + spcObj.Name,
				},
			)
			Expect(err).To(BeNil(), "failed to list csp of spc {%s}", spcObj.Name)

			// get pool pod corresponding to above spc
			poolPodList, err := ops.PodClient.WithNamespace(namespace).
				List(metav1.ListOptions{
					LabelSelector: "openebs.io/cstor-pool=" + cspList.Items[0].Name,
				},
				)
			Expect(err).To(
				BeNil(),
				"failed to list cstor pool pod of csp %s",
				cspList.Items[0].Name,
			)
			poolPodObj := poolPodList.Items[0]

			oldGUID := ops.ExecuteCMDEventually(
				&poolPodObj,
				poolPodObj.Spec.Containers[0].Name,
				zpoolGUIDCmd,
				expectOutput,
			)

			By("Restarting cstor pool pod")
			err = ops.RestartPodEventually(&poolPodObj)
			Expect(err).To(BeNil(), "failed to restart cstor pool pod")

			// get pool pod corresponding to above spc
			poolPodList, err = ops.PodClient.WithNamespace(namespace).
				List(metav1.ListOptions{
					LabelSelector: "openebs.io/cstor-pool=" + cspList.Items[0].Name,
				},
				)
			Expect(err).To(
				BeNil(),
				"failed to list cstor pool pod of csp %s",
				cspList.Items[0].Name,
			)
			poolPodObj = poolPodList.Items[0]

			newGUID := ops.ExecuteCMDEventually(
				&poolPodObj,
				poolPodObj.Spec.Containers[0].Name,
				zpoolGUIDCmd,
				expectOutput,
			)

			//Check zpool pool guid before and after restarts
			Expect(oldGUID).To(
				Equal(newGUID),
				"pool is created after restarting the cstor sparse pool pod",
			)
		})
	})

	//Test Case #2 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 1 pool out of 3 by deleting one of the csp", func() {
		It("a new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 1)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #3 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 2 pool out of 3 by deleting 2 csps", func() {
		It("2 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 2)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #4 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 3 pool out of 3 by deleting 3 csps", func() {
		It("3 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 3)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #5 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We update spc maxPool field to 1", func() {
		It("3 cstor pool should be present as down scaling is not supported ", func() {
			// Get the latest spc
			newSPC, err := ops.SPCClient.Get(spcObj.Name, metav1.GetOptions{})
			Expect(err).To(BeNil())

			// update the spc to set maxPool field to 1
			obj := spc_v1alpha1.BuilderForAPIObject(newSPC).WithMaxPool(1)
			_, err = ops.SPCClient.Update(obj.Spc.Object)
			Expect(err).To(BeNil())

			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 1)
			Expect(cspCount).To(Equal(3))
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
			bdcCount := ops.GetBDCCountEventually(
				metav1.ListOptions{
					LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcObj.Name},
				0, string(artifacts.OpenebsNamespace))
			Expect(bdcCount).To(BeZero())
			Expect(ops.IsSPCNotExists(spcObj.Name)).To(BeTrue())
		})
	})
})

var _ = Describe("MIRRORED SPARSE SPC", func() {

	// Test Case #1 (sparse-striped-auto-spc). | TestType : Pool Creation
	When("We apply sparse-mirrored-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and healthy status", func() {
			spcObj = spc_v1alpha1.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(3).
				WithThickProvisioning(true).
				WithPoolType(string(apis.PoolTypeMirroredCPV)).
				Build().Object

			// Create a storage pool claim
			_, err := ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
			bdcCount := ops.GetBDCCount(getLabelSelector(spcObj), string(artifacts.OpenebsNamespace))
			Expect(bdcCount).To(
				Equal(spc_v1alpha1.DefaultDiskCount[string(apis.PoolTypeMirroredCPV)]*3),
				"mismatch of blockdeviceclaim count",
			)
			// Check is there any extra csp's are created
			cspCount = ops.GetCSPCount(getLabelSelector(spcObj))
			Expect(cspCount).To(Equal(3), "mismatch of csp count")
		})
	})

	//Test Case #2 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 1 pool out of 3 by deleting one of the csp", func() {
		It("a new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 1)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #3 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 2 pool out of 3 by deleting 2 csps", func() {
		It("2 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 2)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #4 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 3 pool out of 3 by deleting 3 csps", func() {
		It("3 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 3)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		})
	})

})

var _ = Describe("RAIDZ SPARSE SPC", func() {

	// Test Case #1 (sparse-striped-auto-spc). | TestType : Pool Creation
	When("We apply sparse-raidz-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and healthy status", func() {

			spcObj = spc_v1alpha1.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(3).
				WithThickProvisioning(true).
				WithPoolType(string(apis.PoolTypeRaidzCPV)).
				Build().Object

			// Create a storage pool claim
			_, err := ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
			bdcCount := ops.GetBDCCount(getLabelSelector(spcObj), string(artifacts.OpenebsNamespace))
			Expect(bdcCount).To(
				Equal(spc_v1alpha1.DefaultDiskCount[string(apis.PoolTypeRaidzCPV)]*3),
				"mismatch of blockdeviceclaim count",
			)
			// Check is there any extra csp's are created
			cspCount = ops.GetCSPCount(getLabelSelector(spcObj))
			Expect(cspCount).To(Equal(3), "mismatch of csp count")
		})
	})

	//Test Case #2 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 1 pool out of 3 by deleting one of the csp", func() {
		It("a new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 1)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #3 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 2 pool out of 3 by deleting 2 csps", func() {
		It("2 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 2)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #4 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 3 pool out of 3 by deleting 3 csps", func() {
		It("3 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 3)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		})
	})

})

var _ = Describe("RAIDZ2 SPARSE SPC", func() {

	// Test Case #1 (sparse-striped-auto-spc). | TestType : Pool Creation
	When("We apply sparse-raidz2-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and healthy status", func() {
			mayaPodList, err := ops.PodClient.
				WithNamespace(string(artifacts.OpenebsNamespace)).
				List(metav1.ListOptions{LabelSelector: string(artifacts.MayaAPIServerLabelSelector)})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(mayaPodList.Items)).To(Equal(1))
			mayaPod := mayaPodList.Items[0]

			spcObj = spc_v1alpha1.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(3).
				WithThickProvisioning(true).
				WithPoolType(string(apis.PoolTypeRaidz2CPV)).
				Build().Object

			// Create a storage pool claim
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			By("Restarting maya-apiserver pod")
			// TODO: Need to make sure after partial BDC creation need to delete
			// maya-apiserver pod
			err = ops.RestartPodEventually(&mayaPod)
			Expect(err).To(BeNil(), "failed to restart maya-apiserver pod")

			podCount := ops.GetPodRunningCountEventually(
				string(artifacts.OpenebsNamespace),
				string(artifacts.MayaAPIServerLabelSelector),
				1,
			)
			Expect(podCount).To(Equal(1), "failed to get restarted maya-apiserver pod")

			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
			bdcCount := ops.GetBDCCount(getLabelSelector(spcObj), string(artifacts.OpenebsNamespace))
			Expect(bdcCount).To(
				Equal(spc_v1alpha1.DefaultDiskCount[string(apis.PoolTypeRaidz2CPV)]*3),
				"mismatch of blockdeviceclaim count",
			)
			// Check is there any extra csp's are created
			cspCount = ops.GetCSPCount(getLabelSelector(spcObj))
			Expect(cspCount).To(Equal(3), "mismatch of csp count")
		})
	})

	//Test Case #2 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 1 pool out of 3 by deleting one of the csp", func() {
		It("a new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 1)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #3 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 2 pool out of 3 by deleting 2 csps", func() {
		It("2 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 2)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #4 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 3 pool out of 3 by deleting 3 csps", func() {
		It("3 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 3)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		})
	})

})
