/*
Copyright 2018 The OpenEBS Authors

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

package cstorpoolit

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	citf "github.com/openebs/CITF"
	citfoptions "github.com/openebs/CITF/citf_options"
	apis "github.com/openebs/CITF/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestIntegrationCstorPool function instantiate the cstor pool test suite.
func TestIntegrationCstorPool(t *testing.T) {
	// RegisterFailHandler is used to register failed test cases and produce readable output.
	RegisterFailHandler(Fail)
	// RunSpecs runs all the test cases in the suite.
	RunSpecs(t, "Cstor pool integration test suite")
}

// Create an instance of CITF to use the inbuilt functions that will help
// communicating with the kube-apiserver.
var citfInstance, err = citf.NewCITF(&citfoptions.CreateOptions{
	// K8SInclude is true to get the kube-config from the machine where the suite is running.
	// Kube-config is a config file that establishes communication to the k8s cluster.
	K8SInclude: true,
})

// ToDo: Set up cluster environment before runninng all test cases ( i.e. BeforeSuite)
// The environment set up by BeforeSuite is going to persist for all
// the test cases under run

//var _ = BeforeSuite(func() {
//	//var err error
//	//
//	//Expect(err).NotTo(HaveOccurred())
//})

// ToDo: Set up tear down of cluster environment ( i.e Aftersuite)

// ToDo: Set up cluster environment before every test cases that will be run (i.e. preRunHook)
// ToDo: Reset cluster environment after every test cases that will be run  ( i.e postRunHook)
var _ = Describe("Integration Test", func() {
	// Test Case #1 (sparse-striped-auto-spc). Type : Positive
	When("We apply sparse-striped-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and online status", func() {
			// TODO: Create a generic util function in utils.go to convert yaml into go object.
			// ToDo: More POC regarding this util converter function.
			// Functions generic to both cstor-pool and cstor-vol should go inside common directory

			// 1.Read SPC yaml form a file.
			// 2.Convert SPC yaml to json.
			// 3.Marshall json to SPC go object.

			// Create a storage pool claim object
			spcObject := &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "disk-claim-auto",
				},
				Spec: apis.StoragePoolClaimSpec{
					Name:     "sparse-claim-auto",
					Type:     "sparse",
					MaxPools: 3,
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "striped",
					},
				},
			}
			// Call CITF to create StoragePoolClaim in k8s.
			spcGot, err := citfInstance.K8S.CreateStoragePoolClaim(spcObject)
			Expect(err).To(BeNil())
			// We expect nil error.

			// We expect 3 cstorPool objects.
			var maxRetry int
			var cspCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				cspCount, err = getCstorPoolCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if cspCount == 3 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(cspCount).To(Equal(3))
			// We expect 3 pool deployments.
			var deployCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				deployCount, err = getPoolDeployCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if deployCount == 3 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(cspCount).To(Equal(3))
			// We expect 3 storagePool objects.
			var spCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				spCount, err = getStoragePoolCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if spCount == 3 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(spCount).To(Equal(3))

			// We expect 'online' status on all the three cstorPool objects(i.e. 3 online counts)
			var onlineCspCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				onlineCspCount, err = getCstorPoolStatus(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if onlineCspCount == 3 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(onlineCspCount).To(Equal(3))

		})
	})

	// Test Case #2 (sparse-mirrored-auto-spc). Type : Negative
	When("We apply sparse-mirrored-auto spc yaml with maxPool count equal to 0 on a k8s cluster", func() {
		It("pool resources count should be 0 with no error and online status", func() {
			// TODO: Create a generic util function in utils.go to convert yaml into go object.
			// ToDo: More POC regarding this util converter function.
			// Functions generic to both cstor-pool and cstor-vol should go inside common directory

			// 1.Read SPC yaml form a file.
			// 2.Convert SPC yaml to json.
			// 3.Marshall json to SPC go object.

			// Create a storage pool claim object
			spcObject := &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "disk-claim-auto",
				},
				Spec: apis.StoragePoolClaimSpec{
					Name:     "sparse-claim-auto",
					Type:     "sparse",
					MaxPools: 0,
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "mirrored",
					},
				},
			}
			// Call CITF to create StoragePoolClaim in k8s.
			spcGot, err := citfInstance.K8S.CreateStoragePoolClaim(spcObject)
			Expect(err).To(BeNil())
			// We expect nil error.

			// We expect 0 cstorPool objects.
			var maxRetry int
			var cspCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				cspCount, err = getCstorPoolCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if cspCount == 0 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(cspCount).To(Equal(0))
			// We expect 0 pool deployments.
			var deployCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				deployCount, err = getPoolDeployCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if deployCount == 0 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(cspCount).To(Equal(0))
			// We expect 0 storagePool objects.
			var spCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				spCount, err = getStoragePoolCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if spCount == 0 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(spCount).To(Equal(0))

			// We don't expect 'online' status on any of the cstorPool objects.
			var onlineCspCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				onlineCspCount, err = getCstorPoolStatus(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if onlineCspCount == 0 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(onlineCspCount).To(Equal(0))

		})
	})

	// TODo: Add more test cases. Refer to following design doc
	// https://docs.google.com/document/d/1QAYK-Bsehc7v66kscXCiMJ7_pTIjzNmwyl43tF92gWA/edit

	// Test Case #5 (sparse-mirrored-auto-spc). Type : Positive
	When("We apply sparse-mirrored-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and online status", func() {
			// TODO: Create a generic util function in utils.go to convert yaml into go object.
			// ToDo: More POC regarding this util converter function.
			// Functions generic to both cstor-pool and cstor-vol should go inside common directory

			// 1.Read SPC yaml form a file.
			// 2.Convert SPC yaml to json.
			// 3.Marshall json to SPC go object.

			// Create a storage pool claim object
			spcObject := &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "disk-claim-auto",
				},
				Spec: apis.StoragePoolClaimSpec{
					Name:     "sparse-claim-auto",
					Type:     "sparse",
					MaxPools: 3,
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "mirrored",
					},
				},
			}
			// Call CITF to create StoragePoolClaim in k8s.
			spcGot, err := citfInstance.K8S.CreateStoragePoolClaim(spcObject)
			Expect(err).To(BeNil())
			// We expect nil error.

			// We expect 3 cstorPool objects.
			var maxRetry int
			var cspCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				cspCount, err = getCstorPoolCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if cspCount == 3 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(cspCount).To(Equal(3))
			// We expect 3 pool deployments.
			var deployCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				deployCount, err = getPoolDeployCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if deployCount == 3 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(cspCount).To(Equal(3))
			// We expect 3 storagePool objects.
			var spCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				spCount, err = getStoragePoolCount(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if spCount == 3 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(spCount).To(Equal(3))

			// We expect 'online' status on all the three cstorPool objects(i.e. 3 online counts)
			var onlineCspCount int
			maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				onlineCspCount, err = getCstorPoolStatus(spcGot.Name, citfInstance)
				if err != nil {
					break
				}
				if onlineCspCount == 3 {
					break
				}
				time.Sleep(time.Second * 5)
			}
			Expect(onlineCspCount).To(Equal(3))

		})
	})
})
