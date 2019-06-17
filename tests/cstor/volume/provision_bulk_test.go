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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/cstor"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[cstor] TEST BULK VOLUME PROVISIONING", func() {
	var (
		err           error
		pvcNamePrefix = "cstor-volume-claim-"
		bulkCount     = 10
		pvcObjList    *v1.PersistentVolumeClaimList
		pvcObjTpl     *v1.PersistentVolumeClaim
	)

	When("creating resources required by test", func() {
		It("should create a spc and storageclass", func() {

			By("building spc")
			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(cstor.PoolCount).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating above spc")
			spcObj, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc {%s}", spcName)

			By("verifying healthy csp count")
			cspCount := ops.GetHealthyCSPCountEventually(spcObj.Name, cstor.PoolCount)
			Expect(cspCount).To(Equal(true), "while checking csp health status")

			By("building cas config with above spc")
			CASConfig := strings.Replace(
				openebsCASConfigValue, "$spcName", spcObj.Name, 1,
			)
			CASConfig = strings.Replace(
				CASConfig, "$count", strconv.Itoa(cstor.ReplicaCount), 1,
			)
			annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
			annotations[string(apis.CASConfigKey)] = CASConfig

			By("building a storageclass")
			scObj, err = sc.NewBuilder().
				WithGenerateName(scName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building storageclass {%s}", scName,
			)

			By("creating above storageclass")
			scObj, err = ops.SCClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass {%s}", scName)
		})
	})

	// Actual tests begin here !!!
	When("cstor pvcs are created", func() {
		It("should create cstor volumes", func() {

			By("building a pvc template")
			pvcObjTpl, err = pvc.NewBuilder().
				WithGenerateName(pvcNamePrefix).
				WithNamespace(nsObj.Name).
				WithStorageClass(scObj.Name).
				WithAccessModes(accessModes).
				WithCapacity(capacity).
				WithLabels(
					map[string]string{
						"bulk.delete": nsObj.Name,
					},
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc template {%s} in namespace {%s}",
				pvcNamePrefix,
				nsObj.Name,
			)

			By("building a list of pvcs")
			pvcObjList, err = pvc.ListBuilderFromTemplate(pvcObjTpl).
				WithCount(bulkCount).
				APIList()

			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building a list of pvcs {%s} in namespace {%s}",
				pvcNamePrefix,
				nsObj.Name,
			)

			By("creating above pvcs in a bulk")
			pvcObjList, err = ops.PVCClient.
				WithNamespace(nsObj.Name).
				CreateCollection(pvcObjList)
			Expect(err).To(
				BeNil(),
				"while creating pvcs {%s} in namespace {%s}", pvcNamePrefix, nsObj.Name,
			)

			By("verifying target pod count for all pvcs")
			podCount := ops.GetPodRunningCountEventually(
				openebsNamespace,
				"openebs.io/storage-class="+scObj.Name,
				bulkCount,
			)
			Expect(podCount).To(Equal(bulkCount), "while checking target pod count")

			By("fetching pvc list")
			pvcObjList, err = ops.PVCClient.WithNamespace(nsObj.Name).
				List(metav1.ListOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while fetching pvcs")

			for _, pvcObj := range pvcObjList.Items {

				By("verifying pvc status as bound")
				status := ops.IsPVCBoundEventually(pvcObj.Name)
				Expect(status).To(Equal(true), "while checking status equal to bound")

				By("verifying target pod count")
				targetLabel := pvcLabel + pvcObj.Name
				controllerPodCount := ops.GetPodRunningCountEventually(
					openebsNamespace, targetLabel, 1,
				)
				Expect(controllerPodCount).To(
					Equal(1),
					"while checking target pod count",
				)

				By("verifying if cstorvolume is created")
				cvLabel := pvLabel + pvcObj.Spec.VolumeName
				cvCount := ops.GetCstorVolumeCountEventually(
					openebsNamespace, cvLabel, 1,
				)
				Expect(cvCount).To(
					Equal(true),
					"while checking if cstorvolume is created",
				)

				By("verifying cvr count")
				cvrLabel := pvLabel + pvcObj.Spec.VolumeName
				cvrCount := ops.GetCstorVolumeReplicaCountEventually(
					openebsNamespace, cvrLabel, cstor.ReplicaCount,
				)
				Expect(cvrCount).To(
					Equal(true),
					"while checking cvr count",
				)

			}
		})
	})

	When("eventually pool pods are deleted", func() {
		It("should delete cstor pools pod before pvc delete", func() {

			lopts := metav1.ListOptions{
				LabelSelector: "openebs.io/cstor-pool=" + spcObj.Name,
			}

			err = ops.PodDeleteCollection(nsObj.Name, lopts)
			Expect(err).To(
				BeNil(),
				"while deleting pools {%s} in namespace {%s}", spcObj.Name, nsObj.Name,
			)
		})
	})

	When("cstor pvcs are deleted", func() {
		It("should delete cstor volumes", func() {

			By("deleting above pvcs in a bulk")
			lopts := metav1.ListOptions{
				LabelSelector: "bulk.delete=" + nsObj.Name,
			}

			deletePolicy := metav1.DeletePropagationForeground
			dopts := &metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			}
			By("deleting the pool pods again")
			err = ops.PodDeleteCollection(nsObj.Name, lopts)
			Expect(err).To(
				BeNil(),
				"while deleting pools {%s} in namespace {%s}", spcObj.Name, nsObj.Name,
			)

			err = ops.PVCClient.
				WithNamespace(nsObj.Name).
				DeleteCollection(lopts, dopts)
			Expect(err).To(
				BeNil(),
				"while deleting pvcs {%s} in namespace {%s}", pvcNamePrefix, nsObj.Name,
			)

			for _, pvcObj := range pvcObjList.Items {

				By("verifying target pod count as 0")
				targetLabel := pvcLabel + pvcObj.Name
				controllerPodCount := ops.GetPodRunningCountEventually(
					openebsNamespace, targetLabel, 0)
				Expect(controllerPodCount).To(
					Equal(0),
					"while checking controller pod count",
				)

				By("verifying deleted pvc")
				pvc := ops.IsPVCDeleted(pvcObj.Name)
				Expect(pvc).To(
					Equal(true),
					"while trying to get deleted pvc",
				)

				By("verifying if cstorvolume is deleted")
				cvLabel := pvLabel + pvcObj.Spec.VolumeName
				cvCount := ops.GetCstorVolumeCountEventually(
					openebsNamespace, cvLabel, 0,
				)
				Expect(cvCount).To(
					Equal(true),
					"while checking if cstorvolume is deleted",
				)

				By("verifying if cvr is deleted")
				cvrLabel := pvLabel + pvcObj.Spec.VolumeName
				cvrCount := ops.GetCstorVolumeReplicaCountEventually(
					openebsNamespace, cvrLabel, 0,
				)
				Expect(cvrCount).To(
					Equal(true),
					"while checking if cvr is deleted",
				)

			}
		})
	})

	When("cleaning up test resources", func() {
		It("should delete resources created during this test", func() {

			By("deleting spc")
			_, err = ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting spc {%s}", spcObj.Name)

			By("deleting storageclass")
			err = ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass {%s}", scName)

		})
	})
})
