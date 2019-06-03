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

package negative

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/cstor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[cstor] [-ve] TEST INVALID STORAGEPOOLCLAIM", func() {
	var (
		err error
	)

	AfterEach(func() {

		By("deleting storagepoolclaim")
		_, err = ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting the storagepoolclaim {%s}", spcObj.Name)

		time.Sleep(5 * time.Second)
	})

	When("creating storagepoolclaim with invalid disk type", func() {
		It("should not create any cstorpool", func() {

			By("building storagepoolclaim with invalid disk type")
			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType("invalid-disk-type").
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating above storagepoolclaim")
			spcObj, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating storagepoolclaim {%s}", spcObj.Name)

			By("verifying cstorpool count as 0")
			cspCount := ops.GetCSPCount(spcObj.Name, cstor.PoolCount)
			Expect(cspCount).To(Equal(0), "while checking cstorpool count")

		})
	})

	When("creating storagepoolclaim with invalid pool type", func() {
		It("should not create any cstorpool", func() {

			By("building a storagepoolclaim with invalid pool type")
			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string("invalid-pool-type")).
				Build().Object

			By("creating above storagepoolclaim")
			spcObj, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating storagepoolclaim {%s}", spcObj.Name)

			By("verifying cstorpool count as 0")
			cspCount := ops.GetCSPCount(spcObj.Name, cstor.PoolCount)
			Expect(cspCount).To(Equal(0), "while checking cstorpool count")

		})
	})

	When("creating storagepoolclaim with invalid pool count", func() {
		It("should not create any cstorpool", func() {

			By("building storagepoolclaim with invalid pool count")
			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(-1).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating above storagepoolclaim")
			spcObj, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating storagepoolclaim {%s}", spcObj.Name)

			By("verifying cstorpool count as 0")
			cspCount := ops.GetCSPCount(spcObj.Name, cstor.PoolCount)
			Expect(cspCount).To(Equal(0), "while checking cstorpool count")

		})
	})

})
