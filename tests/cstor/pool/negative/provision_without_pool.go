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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

var _ = Describe("[cstor] [-ve] TEST PROVISION WITHOUT POOL", func() {
	var (
		err error
	)

	BeforeEach(func() {
		By("creating storagepoolclaim")
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

	AfterEach(func() {
		By("deleting storagepoolclaim")

		err = ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting the storagepoolclaim {%s}", spcObj.Name)

		cspCount := ops.GetCSPCount(getLabelSelector(spcObj))
		Expect(cspCount).To(Equal(0), "stale CSP")
	})

	When("Deleting SPC if the pool does not exist", func() {
		It("Should delete the SPC", func() {
			spcLabel := getLabelSelector(spcObj)
			poolPodList, err := ops.PodClient.
				WithNamespace(openebsNamespace).
				List(metav1.ListOptions{LabelSelector: spcLabel})
			Expect(err).To(BeNil())
			// Since spc is created successfully, pool pod count should be >= 1
			Expect(len(poolPodList.Items)).Should(BeNumerically(">=", 1))

			tPod := poolPodList.Items[0]
			var puid string
			for _, e := range tPod.Spec.Containers[0].Env {
				if e.Name == "OPENEBS_IO_CSTOR_ID" {
					puid = e.Value
				}
			}
			Expect(len(puid)).Should(BeNumerically(">=", 1))

			cspList, err := ops.CSPClient.List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcObj.Name})
			Expect(err).To(BeNil())
			Expect(len(cspList.Items)).Should(BeNumerically(">=", 1))
			var cspObj apis.CStorPool
			for _, l := range cspList.Items {
				if l.GetUID() == types.UID(puid) {
					cspObj = l
					break
				}
			}

			poolName := string(pool.PoolPrefix) + string(cspObj.ObjectMeta.UID)
			zfsCmd := "zpool destroy -f " + poolName
			cmd := []string{"/bin/bash", "-c", zfsCmd}

			opts := tests.NewOptions().
				WithPodName(tPod.Name).
				WithNamespace(openebsNamespace).
				WithContainer("cstor-pool").
				WithCommand(cmd...)

			out, err := ops.ExecPod(opts)
			Expect(err).To(BeNil())
			Expect(len(out)).Should(BeNumerically("==", 0))
		})
	})
})

func getLabelSelector(spc *apis.StoragePoolClaim) string {
	return string(apis.StoragePoolClaimCPK) + "=" + spc.Name
}
