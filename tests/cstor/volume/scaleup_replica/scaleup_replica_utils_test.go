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

package replicascaleup

import (
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	hash "github.com/openebs/maya/pkg/hash"
	"github.com/openebs/maya/tests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func verifyVolumeConfigurationEventually() {
	var err error
	consistencyFactor := (ReplicaCount / 2) + 1
	for i := 0; i < MaxRetry; i++ {
		cvObj, err = ops.CVClient.WithNamespace(openebsNamespace).
			Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
		Expect(err).To(BeNil())
		if cvObj.Spec.ReplicationFactor == ReplicaCount {
			break
		}
		time.Sleep(5 * time.Second)
	}
	Expect(cvObj.Spec.ConsistencyFactor).To(Equal(consistencyFactor), "mismatch of consistencyFactor")
	_, isReplicaIDExist := cvObj.Status.ReplicaDetails.KnownReplicas[ReplicaID]
	Expect(isReplicaIDExist).To(Equal(true), "replicaId should exist in known replicas of cstorvolume")
	Expect(cvObj.Status.Phase).To(Equal(apis.CStorVolumePhase("Healthy")))
}

func buildAndCreateSC() {
	casConfig := strings.Replace(
		openebsCASConfigValue, "$spcName", spcObj.Name, 1)
	casConfig = strings.Replace(
		casConfig, "$count", strconv.Itoa(ReplicaCount), 1)
	annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
	annotations[string(apis.CASConfigKey)] = casConfig
	scConfig := &tests.SCConfig{
		Name:        scName,
		Annotations: annotations,
		Provisioner: openebsProvisioner,
	}
	ops.Config = scConfig
	scObj = ops.CreateStorageClass()
}

func updateDesiredReplicationFactor() {
	var err error
	cvObj, err = ops.CVClient.WithNamespace(openebsNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	Expect(err).To(BeNil())
	cvObj.Spec.DesiredReplicationFactor = cvObj.Spec.DesiredReplicationFactor + 1
	// Namespace is already set to CVClient in above step
	cvObj, err = ops.CVClient.Update(cvObj)
	Expect(err).To(BeNil())
}

func buildAndCreateCVR() {
	var err, getErr error
	retryUpdate := 3
	volumeLabel := pvLabel + pvcObj.Spec.VolumeName
	cvrObjList, err := ops.CVRClient.
		WithNamespace(openebsNamespace).
		List(metav1.ListOptions{LabelSelector: volumeLabel})
	Expect(err).To(BeNil())

	cvrObj = &cvrObjList.Items[0]
	poolLabel := string(apis.StoragePoolClaimCPK) + "=" + spcObj.Name
	cspObj = ops.GetUnUsedCStorPool(cvrObjList, poolLabel)
	cvrConfig := &tests.CVRConfig{
		VolumeName: pvcObj.Spec.VolumeName,
		PoolObj:    cspObj,
		Namespace:  openebsNamespace,
		TargetIP:   cvrObj.Spec.TargetIP,
		Phase:      "Recreate",
		Capacity:   cvrObj.Spec.Capacity,
	}
	ops.Config = cvrConfig
	newCVRObj = ops.BuildAndCreateCVR()

	cvrName := pvcObj.Spec.VolumeName + "-" + cspObj.Name
	hashUID, err := hash.Hash(newCVRObj.UID)
	Expect(err).To(BeNil())
	ReplicaID = strings.ToUpper(hashUID)
	for i := 0; i < retryUpdate; i++ {
		newCVRObj.Spec.ReplicaID = ReplicaID
		newCVRObj, err = ops.CVRClient.
			WithNamespace(openebsNamespace).
			Update(newCVRObj)
		if err == nil {
			break
		}
		time.Sleep(time.Second * 5)
		newCVRObj, getErr = ops.CVRClient.Get(cvrName, metav1.GetOptions{})
		Expect(getErr).To(BeNil())
	}
	Expect(err).To(BeNil())
	//TODO: Need to fix bug in cvr during creation time
	podLabel := cspLabel + cspObj.Name
	podObjList, err := ops.PodClient.
		WithNamespace(openebsNamespace).
		List(metav1.ListOptions{LabelSelector: podLabel})
	Expect(err).To(BeNil())
	err = ops.PodClient.Delete(podObjList.Items[0].Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil())
	isPodDeleted := ops.IsPodDeletedEventually(
		podObjList.Items[0].Namespace,
		podObjList.Items[0].Name)
	Expect(isPodDeleted).To(Equal(true))
}
