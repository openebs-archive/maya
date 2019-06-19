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

package snapshot

import (
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	snap "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/cstor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[cstor] TEST SNAPSHOT PROVISIONING", func() {
	var (
		err      error
		pvcName  = "test-cstor-snap-pvc"
		snapName = "test-cstor-snap-snapshot"
	)

	BeforeEach(func() {
		When("deploying cstor sparse pool", func() {
			By("building spc object")
			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating storagepoolclaim")
			spcObj, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc", spcName)

			By("verifying healthy csp count")
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, cstor.PoolCount)
			Expect(cspCount).To(Equal(cstor.ReplicaCount), "while checking cstorpool health count")

			By("building a CAS Config with generated SPC name")
			CASConfig := strings.Replace(openebsCASConfigValue, "$spcName", spcObj.Name, 1)
			CASConfig = strings.Replace(CASConfig, "$count", strconv.Itoa(cstor.ReplicaCount), 1)
			annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
			annotations[string(apis.CASConfigKey)] = CASConfig

			By("building storageclass object")
			scObj, err = sc.NewBuilder().
				WithGenerateName(scName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building storageclass obj for storageclass {%s}", scName)

			By("creating storageclass")
			scObj, err = ops.SCClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass", scName)

		})
	})

	AfterEach(func() {
		By("deleting resources created for cstor volume snapshot provisioning", func() {
			By("deleting storagepoolclaim")
			_, err = ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting the spc's {%s}", spcName)

			By("deleting storageclass")
			err = ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass", scName)

		})
	})

	When("cstor pvc with replicacount n is created", func() {
		It("should create cstor volume target pod", func() {

			By("building a persistentvolumeclaim")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(nsObj.Name).
				WithStorageClass(scObj.Name).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				pvcName,
				nsObj.Name,
			)

			By("creating cstor persistentvolumeclaim")
			pvcObj, err = ops.PVCClient.WithNamespace(nsObj.Name).Create(pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				pvcName,
				nsObj.Name,
			)

			By("verifying volume target pod count as 1")
			targetVolumeLabel := pvcLabel + pvcObj.Name
			controllerPodCount := ops.GetPodRunningCountEventually(openebsNamespace, targetVolumeLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			By("verifying cstorvolume replica count")
			pvcObj, err = ops.PVCClient.WithNamespace(nsObj.Name).Get(pvcName, metav1.GetOptions{})
			Expect(err).To(
				BeNil(),
				"while getting pvc {%s} in namespace {%s}",
				pvcName,
				nsObj.Name,
			)
			cvrLabel := pvLabel + pvcObj.Spec.VolumeName
			cvrCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, cvrLabel, cstor.ReplicaCount)
			Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")

			By("verifying pvc status as bound")
			status := ops.IsPVCBoundEventually(pvcName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

			By("building a cstor volume snapshot")
			snapObj, err = snap.NewBuilder().
				WithName(snapName).
				WithNamespace(nsObj.Name).
				WithPVC(pvcName).
				Build()
			Expect(err).To(
				BeNil(),
				"while building snapshot {%s} in namespace {%s}",
				snapName,
				nsObj.Name,
			)

			By("creating cstor volume snapshot")
			_, err = ops.SnapClient.WithNamespace(nsObj.Name).Create(snapObj)
			Expect(err).To(
				BeNil(),
				"while creating snapshot {%s} in namespace {%s}",
				snapName,
				nsObj.Name,
			)

			By("verifying snapshot status as ready")
			snaptype := ops.GetSnapshotTypeEventually(snapName)
			Expect(snaptype).To(Equal("Ready"), "while checking snapshot type")

			By("deleting cstor volume snapshot")
			err = ops.SnapClient.Delete(snapName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting snapshot {%s} in namespace {%s}",
				snapName,
				nsObj.Name,
			)

			By("verifying deleted snapshot")
			snap := ops.IsSnapshotDeleted(snapName)
			Expect(snap).To(Equal(true), "while checking for deleted snapshot")

			By("deleting above pvc")
			err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				nsObj.Name,
			)

			By("verifying target pod count as 0")
			controllerPodCount = ops.GetPodRunningCountEventually(openebsNamespace, targetVolumeLabel, 0)
			Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

			By("verifying deleted pvc")
			pvc := ops.IsPVCDeleted(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

			By("verifying if cstorvolume is deleted")
			CstorVolumeLabel := pvLabel + pvcObj.Spec.VolumeName
			cvCount := ops.GetCstorVolumeCountEventually(openebsNamespace, CstorVolumeLabel, 0)
			Expect(cvCount).To(Equal(true), "while checking cstorvolume count")
		})
	})

})
