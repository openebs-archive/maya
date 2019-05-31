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
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[cstor] [-ve] TEST INVALID STORAGECLASS", func() {
	var (
		generatedSPCName      string
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

			generatedSPCName = tests.GenerateName(spcName)

			By("building storagepoolclaim")
			spcObj = spc.NewBuilder().
				WithName(generatedSPCName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating above storagepoolclaim")
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating storagepoolclaim {%s}", generatedSPCName)

			By("verifying healthy csp count")
			cspCount := ops.GetHealthyCSPCountEventually(generatedSPCName, cstor.PoolCount)
			Expect(cspCount).To(Equal(true), "while checking cstorpool health status")

		})
	})

	AfterEach(func() {
		By("deleting storagepoolclaim")
		_, err = ops.SPCClient.Delete(generatedSPCName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting the storagepoolclaim {%s}", generatedSPCName)

		time.Sleep(5 * time.Second)
	})

	When("creating storageclass with invalid CASConfig", func() {
		It("should not create any pvc pods", func() {

			By("building a CAS Config")
			CASConfig := strings.Replace(
				openebsCASConfigValue,
				"$count",
				strconv.Itoa(cstor.ReplicaCount),
				1,
			)
			annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
			// adding invalid character to casconfig
			annotations[string(apis.CASConfigKey)] = CASConfig + ":"

			generatedSCName := tests.GenerateName(scName)

			By("building storageclass with invalid CASConfig")
			scObj, err = sc.NewBuilder().
				WithName(generatedSCName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building storageclass obj for storageclass {%s}",
				generatedSCName,
			)

			By("creating above storageclass")
			_, err = ops.SCClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass {%s}", scName)

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

			By("deleting above pvc")
			err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while delete=ing pvc {%s}", pvcName)

			By("deleting storageclass")
			err = ops.SCClient.Delete(generatedSCName, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass {%s}", scName)

		})
	})

})
