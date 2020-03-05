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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	pv "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/cstor"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// mayaAPIServiceLabel is label for Maya-API server Service
	mayaAPIServiceLabel = "openebs.io/component-name=maya-apiserver-svc"

	// mayaAPIPodLabel is label for Maya-API server Pod
	mayaAPIPodLabel = "openebs.io/component-name=maya-apiserver"
)

var _ = Describe("[Cstor Volume Provisioning Positive] TEST VOLUME PROVISIONING", func() {
	var (
		err     error
		pvcName = "cstor-volume-claim"
	)

	BeforeEach(func() {
		When(" creating a cstor based volume", func() {
			By("building spc object")
			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(cstor.PoolCount).
				WithThickProvisioning(true).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating storagepoolclaim")
			spcObj, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc", spcName)

			By("verifying healthy csp count")
			cspCount := ops.GetHealthyCSPCountEventually(spcObj.Name, cstor.PoolCount)
			Expect(cspCount).To(Equal(true), "while checking cstorpool health status")

			By("building a CAS Config with generated SPC name")
			CASConfig := strings.Replace(openebsCASConfigValue, "$spcName", spcObj.Name, 1)
			CASConfig = strings.Replace(CASConfig, "$count", strconv.Itoa(cstor.ReplicaCount), 1)
			annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
			annotations[string(apis.CASConfigKey)] = CASConfig

			By("building object of storageclass")
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
		By("deleting resources created for testing cstor volume provisioning", func() {

			By("deleting storagepoolclaim")
			err = ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting the spc's {%s}", spcName)

			By("deleting storageclass")
			err = ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass", scName)

		})
	})

	When("cstor pvc with replicacount n is created", func() {
		It("should create cstor volume target pod", func() {

			By("building a pvc")
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

			By("creating above pvc")
			pvcObj, err = ops.PVCClient.WithNamespace(nsObj.Name).Create(pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				pvcName,
				nsObj.Name,
			)

			By("verifying target pod count as 1")
			targetVolumeLabel := pvcLabel + pvcObj.Name
			controllerPodCount := ops.GetPodRunningCountEventually(openebsNamespace, targetVolumeLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			pvcObj, err = ops.PVCClient.WithNamespace(nsObj.Name).Get(pvcName, metav1.GetOptions{})
			Expect(err).To(
				BeNil(),
				"while getting pvc {%s} in namespace {%s}",
				pvcName,
				nsObj.Name,
			)

			By("verifying cstorvolume replica count")
			cvrLabel := pvLabel + pvcObj.Spec.VolumeName
			cvrCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, cvrLabel, cstor.ReplicaCount, cvr.IsHealthy())
			Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")

			By("verifying pvc status as bound")
			status := ops.IsPVCBound(pvcName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

			By("should delete cstor pools pod before pvc delete")
			lopts := metav1.ListOptions{
				LabelSelector: "openebs.io/cstor-pool=" + spcObj.Name,
			}

			err = ops.PodDeleteCollection(nsObj.Name, lopts)
			Expect(err).To(
				BeNil(),
				"while deleting pools {%s} in namespace {%s}", spcObj.Name, nsObj.Name,
			)

			//			poolLabel := "openebs.io/cstor-pool=" + spcObj.Name
			//			podList, err := ops.PodClient.
			//				WithNamespace(nsObj.Name).
			//				List(metav1.ListOptions{LabelSelector: poolLabel})
			//			Expect(err).ShouldNot(HaveOccurred(), "while deleting cstor pool pods")
			//			count := len(podList.Items)
			//			for i := 0; i < count; i++ {
			//				By("deleting a pool pods")
			//				err = ops.PodClient.Delete(podList.Items[i].Name, &metav1.DeleteOptions{})
			//			}
			//
			By("deleting above pvc")
			err = ops.PodDeleteCollection(nsObj.Name, lopts)
			Expect(err).To(
				BeNil(),
				"while deleting pools {%s} in namespace {%s}", spcObj.Name, nsObj.Name,
			)

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
			cvLabel := pvLabel + pvcObj.Spec.VolumeName
			cvCount := ops.GetCstorVolumeCountEventually(openebsNamespace, cvLabel, 0)
			Expect(cvCount).To(Equal(true), "while checking cstorvolume count")
		})
	})

	When("cstor PV with replicacount n is created without PVC", func() {
		It("should create cstor volume target pod", func() {
			_ = pvcName
			_ = err
			volname := "pv-created-without-pvc"

			maddr, err := ops.GetSVCClusterIP(openebsNamespace, mayaAPIServiceLabel)
			Expect(err).To(BeNil(), "While fetching maya-apiserver address")
			Expect(len(maddr)).To(Equal(1), "maya-apiserver address")

			podList := ops.GetPodList(openebsNamespace, mayaAPIPodLabel)
			Expect(err).To(BeNil(), "maya-apiserver pod fetch error")
			mPodList := podList.ToAPIList()
			Expect(len(mPodList.Items)).To(Equal(1), "maya-apiserver pod count")

			mpod := mPodList.Items[0]
			capacity := "200M"
			vol := newCASVolumeWithoutPVC(volname, capacity, scObj.Name, false)

			volData, err := json.Marshal(vol)
			Expect(err).To(BeNil())

			command := "curl -XPOST -d '" + string(volData) + "'" + " http://" + maddr[0] + "/latest/volumes/"
			res := ops.ExecuteCMDEventually(&mpod, "", command, true)
			Expect(res).NotTo(BeEmpty())
			Expect(strings.Contains(res, "failed to create volume")).To(Equal(false), fmt.Sprintf("volume creation error=%s", res))

			var rvol apis.CASVolume
			err = json.Unmarshal([]byte(res), &rvol)
			Expect(err).To(BeNil())
			Expect(rvol.Spec.TargetPortal).NotTo(BeEmpty())
			Expect(rvol.Spec.TargetIP).NotTo(BeEmpty())
			Expect(rvol.Spec.TargetPort).NotTo(BeEmpty())

			// create pv object
			pvobj, err := createPVObj(volname, scObj.Name, capacity, rvol)
			Expect(err).To(BeNil())
			pvobj, err = ops.PVClient.Create(pvobj)
			Expect(err).To(BeNil())

			By("verifying target pod count as 1")
			targetVolumeLabel := pvLabel + volname
			controllerPodCount := ops.GetPodRunningCountEventually(openebsNamespace, targetVolumeLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			By("verifying cstorvolume replica count")
			cvrLabel := pvLabel + volname
			cvrCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, cvrLabel, cstor.ReplicaCount, cvr.IsHealthy())
			Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")

			// delete the volume
			command = "curl -XDELETE " + " http://" + maddr[0] + "/latest/volumes/" + volname
			res = ops.ExecuteCMDEventually(&mpod, "", command, true)
			Expect(res).NotTo(BeEmpty())
			Expect(strings.Contains(res, "failed to delete volume")).To(Equal(false), fmt.Sprintf("volume deletion error=%s", res))

			// delete pv object
			err = ops.PVClient.Delete(pvobj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())

			By("verifying target pod count as 0")
			targetVolumeLabel = pvLabel + volname
			controllerPodCount = ops.GetPodRunningCountEventually(openebsNamespace, targetVolumeLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			By("verifying cstorvolume replica count")
			cvrLabel = pvLabel + volname
			cvrCount = ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, cvrLabel, 0, cvr.IsErrored())
			Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")
		})
	})

	When("cstor PV with replicacount n is created without PVC and restore label", func() {
		It("should create cstor volume target pod", func() {
			_ = pvcName
			_ = err
			volname := "restore-pv-created-without-pvc"

			maddr, err := ops.GetSVCClusterIP(openebsNamespace, mayaAPIServiceLabel)
			Expect(err).To(BeNil(), "While fetching maya-apiserver address")
			Expect(len(maddr)).To(Equal(1), "maya-apiserver address")

			podList := ops.GetPodList(openebsNamespace, mayaAPIPodLabel)
			Expect(err).To(BeNil(), "maya-apiserver pod fetch error")
			mPodList := podList.ToAPIList()
			Expect(len(mPodList.Items)).To(Equal(1), "maya-apiserver pod count")

			mpod := mPodList.Items[0]
			capacity := "200M"
			vol := newCASVolumeWithoutPVC(volname, capacity, scObj.Name, true)

			volData, err := json.Marshal(vol)
			Expect(err).To(BeNil())

			command := "curl -XPOST -d '" + string(volData) + "'" + " http://" + maddr[0] + "/latest/volumes/"
			res := ops.ExecuteCMDEventually(&mpod, "", command, true)
			Expect(res).NotTo(BeEmpty())
			Expect(strings.Contains(res, "failed to create volume")).To(Equal(false), fmt.Sprintf("volume creation error=%s", res))

			var rvol apis.CASVolume
			err = json.Unmarshal([]byte(res), &rvol)
			Expect(err).To(BeNil())
			Expect(rvol.Spec.TargetPortal).NotTo(BeEmpty())
			Expect(rvol.Spec.TargetIP).NotTo(BeEmpty())
			Expect(rvol.Spec.TargetPort).NotTo(BeEmpty())

			// create pv object
			pvobj, err := createPVObj(volname, scObj.Name, capacity, rvol)
			Expect(err).To(BeNil())
			pvobj, err = ops.PVClient.Create(pvobj)
			Expect(err).To(BeNil())

			By("verifying target pod count as 1")
			targetVolumeLabel := pvLabel + volname
			controllerPodCount := ops.GetPodRunningCountEventually(openebsNamespace, targetVolumeLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			By("verifying cstorvolume replica count")
			cvrLabel := pvLabel + volname
			cvrCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, cvrLabel, cstor.ReplicaCount, cvr.IsErrored())
			Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")

			// delete the volume
			command = "curl -XDELETE " + " http://" + maddr[0] + "/latest/volumes/" + volname
			res = ops.ExecuteCMDEventually(&mpod, "", command, true)
			Expect(res).NotTo(BeEmpty())
			Expect(strings.Contains(res, "failed to delete volume")).To(Equal(false), fmt.Sprintf("volume deletion error=%s", res))

			// delete pv object
			err = ops.PVClient.Delete(pvobj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())

			By("verifying target pod count as 0")
			targetVolumeLabel = pvLabel + volname
			controllerPodCount = ops.GetPodRunningCountEventually(openebsNamespace, targetVolumeLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			By("verifying cstorvolume replica count")
			cvrLabel = pvLabel + volname
			cvrCount = ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, cvrLabel, 0, cvr.IsErrored())
			Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")
		})

	})
})

func createPVObj(pvname, scname, size string, cas apis.CASVolume) (*v1.PersistentVolume, error) {
	iscsiVolSource := &v1.PersistentVolumeSource{
		ISCSI: &v1.ISCSIPersistentVolumeSource{
			TargetPortal: cas.Spec.TargetPortal,
			IQN:          cas.Spec.Iqn,
			Lun:          cas.Spec.Lun,
			FSType:       cas.Spec.FSType,
			ReadOnly:     false,
		},
	}

	pvObj, err := pv.NewBuilder().
		WithName(pvname).
		WithAnnotations(
			map[string]string{
				"openebs.io/cas-type":             "cstor",
				"pv.kubernetes.io/provisioned-by": "openebs.io/provisioner-iscsi",
			},
		).
		WithLabels(
			map[string]string{
				"openebs.io/cas-type":     "cstor",
				"openebs.io/storageclass": scname,
			},
		).
		WithReclaimPolicy(v1.PersistentVolumeReclaimDelete).
		WithVolumeMode(v1.PersistentVolumeFilesystem).
		WithCapacity(size).
		WithPersistentVolumeSource(iscsiVolSource).
		WithAccessModes([]v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}).
		Build()
	if err != nil {
		return nil, err
	}

	return pvObj, nil
}

func newCASVolumeWithoutPVC(name, capacity, sc string, isRestore bool) *apis.CASVolume {
	casAnn := map[string]string{}

	if isRestore {
		casAnn[apis.PVCreatedByKey] = "restore"
	}
	return &apis.CASVolume{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				string(apis.StorageClassKey): sc,
			},
			Annotations: casAnn,
			Name:        name,
		},
		Spec: apis.CASVolumeSpec{
			Capacity: capacity,
		},
	}
}
