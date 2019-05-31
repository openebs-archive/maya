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
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/cstor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[cstor] [-ve] TEST INVALID STORAGEPOOLCLAIM", func() {
	var (
		err error
	)

	AfterEach(func() {
		By("deleting resources created for testing cstor volume provisioning", func() {

			By("listing spc")
			spcList, err = ops.SPCClient.List(metav1.ListOptions{})
			Expect(err).To(BeNil(), "while listing spc clients", spcList)

			By("deleting spc")
			for _, spc := range spcList.Items {
				_, err = ops.SPCClient.Delete(spc.Name, &metav1.DeleteOptions{})
				Expect(err).To(BeNil(), "while deleting the spc's", spc)
			}
			time.Sleep(5 * time.Second)
		})
	})

	When("creating storagepoolclaim with invalid disk type", func() {
		It("should not create any cstorpool", func() {

			By("building spc object")
			spcObj = spc.NewBuilder().
				WithName(spcName).
				WithDiskType("invalid-disk-type").
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating storagepoolclaim")
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc", spcName)

			By("verifying healthy csp count as 0")
			cspCount := ops.GetHealthyCSPCount(spcName, cstor.PoolCount)
			Expect(cspCount).To(Equal(0), "while checking cstorpool health status")

		})
	})

	When("creating storagepoolclaim with invalid pool type", func() {
		It("should not create any cstorpool", func() {

			By("building spc object")
			spcObj = spc.NewBuilder().
				WithName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string("invalid-pool-type")).
				Build().Object

			By("creating storagepoolclaim")
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc", spcName)

			By("verifying healthy csp count as 0")
			cspCount := ops.GetHealthyCSPCount(spcName, cstor.PoolCount)
			Expect(cspCount).To(Equal(0), "while checking cstorpool health status")

		})
	})

	When("creating storagepoolclaim with invalid pool count", func() {
		It("should not create any cstorpool", func() {

			By("building spc object")
			spcObj = spc.NewBuilder().
				WithName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(-1).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating storagepoolclaim")
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc", spcName)

			By("verifying healthy csp count as 0")
			cspCount := ops.GetHealthyCSPCount(spcName, cstor.PoolCount)
			Expect(cspCount).To(Equal(0), "while checking cstorpool health status")

		})
	})

})

var _ = Describe("[cstor] TEST INVALID STORAGECLASS", func() {
	var (
		err                   error
		pvcName               = "cstor-volume-claim"
		openebsCASConfigValue = `
- name: ReplicaCount
  value: $count
- name: StoragePoolClaim
  value: test-cstor-provision-sparse-pool-auto
`
	)

	BeforeEach(func() {
		When(" creating a cstor based volume", func() {

			By("building spc object")
			spcObj = spc.NewBuilder().
				WithName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating storagepoolclaim")
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc", spcName)

			By("verifying healthy csp count")
			cspCount := ops.GetHealthyCSPCountEventually(spcName, cstor.PoolCount)
			Expect(cspCount).To(Equal(true), "while checking cstorpool health status")

		})
	})

	AfterEach(func() {
		By("deleting resources created for testing cstor volume provisioning", func() {
			By("deleting storageclass")
			err = ops.SCClient.Delete(scName, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass", scName)

			By("listing spc")
			spcList, err = ops.SPCClient.List(metav1.ListOptions{})
			Expect(err).To(BeNil(), "while listing spc clients", spcList)

			By("deleting spc")
			for _, spc := range spcList.Items {
				_, err = ops.SPCClient.Delete(spc.Name, &metav1.DeleteOptions{})
				Expect(err).To(BeNil(), "while deleting the spc's", spc)
			}
			time.Sleep(5 * time.Second)
		})
	})

	When("creating storageclass with invalid CASConfig", func() {
		It("should not create any pvc pods", func() {

			By("building a CAS Config")
			CASConfig := strings.Replace(openebsCASConfigValue, "$count", strconv.Itoa(cstor.ReplicaCount), 1)
			annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
			// adding invalid character to casconfig
			annotations[string(apis.CASConfigKey)] = CASConfig + ":"

			By("building object of storageclass")
			scObj, err = sc.NewBuilder().
				WithName(scName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building storageclass obj for storageclass {%s}", scName)

			By("creating storageclass")
			_, err = ops.SCClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass", scName)

			By("building a pvc")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespace).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				pvcName,
				namespace,
			)

			By("creating above pvc")
			_, err = ops.PVCClient.WithNamespace(namespace).Create(pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				pvcName,
				namespace,
			)

			By("verifying target pod count as 0")
			controllerPodCount := ops.GetPodRunningCountEventually(openebsNamespace, targetLabel, 1)
			Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

		})
	})

})
