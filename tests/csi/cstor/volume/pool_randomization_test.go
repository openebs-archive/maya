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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests/cstor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	resourceCount    = 5
	pvcPrefix        = "cstor-bulk-volume-claim"
	appPrefix        = "bulk-busybox-app"
	buildPVCLabel    = map[string]string{"bulk.pvc": strconv.Itoa(resourceCount)}
	appLabelSelector = "bulk.app=" + strconv.Itoa(resourceCount)
	poolLabel        = "cstorpoolinstance.openebs.io/name"
)

var _ = Describe("[cstor] [sparse] TEST POOL RANDOMIZATION", func() {
	BeforeEach(prepareForPoolRandomizationTest)
	AfterEach(cleanupAfterPoolRandomizationTest)

	Context("Creating bulk volumes and application", func() {
		It("Volumes should be distributed among pools", bulkVolumeCreationTest)
	})
	//When("we create bulkvolumes replica should be distributed among pools", bulkVolumeCreationTest)
})

func bulkVolumeCreationTest() {
	By("Creating and verifying bulk PVC bound status", func() {
		CreateAndVerifyBulkPVC(pvcPrefix, resourceCount, buildPVCLabel)
	})
	By("Creating and verify bulk application", func() {
		CreateAndDeployBulkApps(appPrefix, pvcPrefix, resourceCount)
	})
	// Since this test is regarding single replica with bulk volumes no need to
	// verify volume components
	By("Verify whether volume are created in different pools", verifyPoolRandomization)
	By("Deleting bulk applications", func() {
		DeleteBulkApplications(appLabelSelector)
	})
	By("Deleting bulk persistent volume claim", func() {
		DeleteBulkPVCs(buildPVCLabel)
	})
}

func prepareForPoolRandomizationTest() {
	cacheFile := "/tmp/pool1.cache"
	poolType := "stripe"
	By("should create and verify cstorpoolcluster", func() {
		CreateAndVerifyCStorPoolCluster(cacheFile, poolType, cstor.PoolCount)
	})
	By("should create storage class", func() { createStorageClass(1) })
}

func cleanupAfterPoolRandomizationTest() {
	By("Deleting cstorpoolcluster", deleteCstorPoolCluster)
	By("Deleting storageclass", deleteStorageClass)
}

func verifyPoolRandomization() {
	poolsUsed := map[string]bool{}
	for i := 0; i < resourceCount; i++ {
		pvcName := pvcPrefix + "-" + strconv.Itoa(i)
		pvcObj, err := ops.PVCClient.WithNamespace(nsObj.Name).Get(pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "while getting pvc {%s} object", pvcName)
		targetLabel := pvLabel + pvcObj.Spec.VolumeName
		cvrObjList, err := ops.CVRClient.WithNamespace("openebs").List(metav1.ListOptions{LabelSelector: targetLabel})
		Expect(err).ShouldNot(HaveOccurred(), "while listing cvr objects")
		for _, cvrObj := range cvrObjList.Items {
			poolsUsed[cvrObj.GetLabels()[poolLabel]] = true
		}
	}
	Expect(len(poolsUsed)).Should(Equal(cstor.PoolCount), "Volumes should be created on all pools")
}
