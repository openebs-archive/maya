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

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	poolspec "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/cstorpoolspecs"
	rgrp "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/raidgroups"
	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pts "github.com/openebs/maya/pkg/kubernetes/podtemplatespec/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	k8svolume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	"github.com/openebs/maya/tests/cstor"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"replicaCount":     strconv.Itoa(cstor.ReplicaCount),
		"cstorPoolCluster": cspcName,
		"cas-type":         "cstor",
	}

	By("building a storageclass")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(openebsProvisioner).
		WithVolumeExpansion(true).Build()
	Expect(err).ShouldNot(HaveOccurred(),
		"while building storageclass obj with prefix {%s}", scName)

	By("creating above storageclass")
	scObj, err = ops.SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating storageclass with prefix {%s}", scName)
}

func deleteCstorPoolCluster() {
	By("deleting cstorpoolcluster")
	err := ops.CSPCClient.Delete(
		cspcObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting cspc {%s}", cspcObj.Name)
	By("verifying deleted cspc")
	status := ops.IsCSPCDeletedEventually(cspcObj.Name)
	Expect(status).To(Equal(true), "while trying to get deleted cspc")
}

func deleteStorageClass() {
	By("deleting storageclass")
	err := ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(),
		"while deleting storageclass {%s}", scObj.Name)
}

const (
	getSizeCmd = "df -h | grep \"/mnt/cstore1\" | awk '{print $2}'"
)

func verifyIncreasedSizeInAppPod() {
	size := ops.ExecuteCMDEventually(
		&appPod.Items[0],
		"busybox",
		getSizeCmd,
		true,
	)
	actualSizeInAppPod, _ := resource.ParseQuantity(size)
	previousSize, _ := resource.ParseQuantity(capacity)

	Expect((actualSizeInAppPod).Cmp(previousSize) < 0).To(Equal(true),
		"while verifying updated size in app pod")
}

func createAndVerifyCstorPoolCluster() {
	var (
		err      error
		nodeName string
	)
	var cspcBDList []*apis.CStorPoolClusterBlockDevice
	for _, bd := range bdList.Items {
		if string(bd.Status.ClaimState) != "Unclaimed" {
			continue
		}
		if nodeName != "" && nodeName != bd.Labels[string(apis.HostNameCPK)] {
			continue
		}
		nodeName = bd.Labels[string(apis.HostNameCPK)]
		cspcBDList = append(cspcBDList, &apis.CStorPoolClusterBlockDevice{
			BlockDeviceName: bd.Name,
		})
	}
	nodeSelector := map[string]string{string(apis.HostNameCPK): nodeName}
	poolspec := poolspec.NewBuilder().
		WithNodeSelector(nodeSelector).
		WithCompression("off").
		WithRaidGroupBuilder(
			rgrp.NewBuilder().
				WithType("stripe").
				WithCSPCBlockDeviceList(cspcBDList),
		)
	cspcObj, err = cspc.NewBuilder().
		WithName(cspcName).
		WithNamespace("openebs").
		WithPoolSpecBuilder(poolspec).
		GetObj()
	Expect(err).To(BeNil(), "while creating cstorpoolcluster {%s}", cspcName)

	By("creating above cstorpoolcluster")
	cspcObj, err = ops.CSPCClient.WithNamespace("openebs").Create(cspcObj)
	Expect(err).To(BeNil(),
		"while creating cspc with prefix {%s}", cspcName)

	By("verifying healthy cstorpool count")
	cspCount := ops.GetHealthyCSPICount(cspcObj.Name, cstor.PoolCount)
	Expect(cspCount).To(Equal(cstor.PoolCount),
		"while checking healthy cstor pool count")
}

func createAndVerifyPVC() {
	var (
		err     error
		pvcName = "cstor-volume-claim"
	)
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

	By("verifying pvc status as bound")

	status := ops.IsPVCBoundEventually(pvcName)
	Expect(status).To(Equal(true),
		"while checking status equal to bound")

	pvcObj, err = ops.PVCClient.WithNamespace(nsObj.Name).Get(pvcObj.Name, metav1.GetOptions{})
	Expect(err).To(
		BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		nsObj.Name,
	)
}

func createDeployVerifyApp() {
	By("creating and deploying app pod", createAndDeployAppPod)
	time.Sleep(30 * time.Second)
	By("verifying app pod is running", verifyAppPodRunning)
}

func createAndDeployAppPod() {
	var err error
	By("building a busybox app pod deployment using above csi cstor volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appName).
		WithNamespace(nsObj.Name).
		WithLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithSelectorMatchLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(
					map[string]string{
						"app": "busybox",
					},
				).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("busybox").
						WithName("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/cstore1/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "datavol1",
									MountPath: "/mnt/cstore1",
								},
							},
						),
				).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcObj.Name),
				),
		).
		Build()

	Expect(err).ShouldNot(HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = ops.DeployClient.WithNamespace(nsObj.Name).Create(deployObj)
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		nsObj.Name,
	)

}
func verifyAppPodRunning() {
	var err error
	appPod, err = ops.PodClient.WithNamespace(nsObj.Name).
		List(metav1.ListOptions{
			LabelSelector: "app=busybox",
		},
		)
	Expect(err).ShouldNot(HaveOccurred(), "while verifying application pod")

	status := ops.IsPodRunningEventually(nsObj.Name, appPod.Items[0].Name)
	Expect(status).To(Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

func verifyPDB() {
	ops.VerifyPodDisruptionBudget(pvcObj.Spec.VolumeName, openebsNamespace)
}

func verifyVolumeComponents() {
	By("should verify target pod count as 1", func() { verifyTargetPodCount(1) })
	By("should verify cstorvolume replica count", func() { verifyCstorVolumeReplicaCount(cstor.ReplicaCount) })
	By("Should verify PodDisruptionBudget for HA volume", func() {verifyPDB()})
}

func verifyTargetPodCount(count int) {
	var err error
	By("verifying target pod count as 1 once the app has been deployed")
	pvcObj, err = ops.PVCClient.WithNamespace(nsObj.Name).
		Get(pvcObj.Name, metav1.GetOptions{})
	Expect(err).To(
		BeNil(),
		"while getting pvc {%s} in namespace {%s}",
		pvcName,
		nsObj.Name,
	)
	targetVolumeLabel := pvLabel + pvcObj.Spec.VolumeName
	controllerPodCount := ops.GetPodRunningCountEventually(
		openebsNamespace, targetVolumeLabel, count)
	Expect(controllerPodCount).To(Equal(count),
		"while checking controller pod count")
}

func verifyCstorVolumeReplicaCount(count int) {
	targetVolumeLabel := pvLabel + pvcObj.Spec.VolumeName
	isReqCVRCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, targetVolumeLabel, count)
	Expect(isReqCVRCount).To(Equal(true), "while checking cstorvolume replica count")
}

func verifyCstorVolumeClaimCount(count int) {
	isReqCVCCount := ops.GetCstorVolumeClaimCountEventually(openebsNamespace, pvcObj.Spec.VolumeName, count)
	Expect(isReqCVCCount).To(Equal(true), "while checking cstorvolume claim count")
}

func restartAppPodAndVerifyRunningStatus() {
	By("restarting application to remount the volume again")
	err := ops.PodClient.WithNamespace(nsObj.Name).
		Delete(appPod.Items[0].Name, &metav1.DeleteOptions{})
	Expect(err).ShouldNot(HaveOccurred(), "while restarting application pod")

	By("verifying app pod is terminated properly")
	status := ops.IsPodDeletedEventually(nsObj.Name, appPod.Items[0].Name)
	Expect(status).To(Equal(true), "while checking termination of pod {%s}", appPod.Items[0].Name)

	By("verifying app pod is running again")
	appPod, err = ops.PodClient.WithNamespace(nsObj.Name).
		List(metav1.ListOptions{
			LabelSelector: "app=busybox",
		},
		)
	Expect(err).ShouldNot(HaveOccurred(), "while verifying application pod")
	status = ops.IsPodRunningEventually(nsObj.Name, appPod.Items[0].Name)
	Expect(status).To(Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

func deleteAppDeployment() {
	err := ops.DeployClient.WithNamespace(nsObj.Name).
		Delete(deployObj.Name, &metav1.DeleteOptions{})
	Expect(err).ShouldNot(HaveOccurred(), "while deleting application pod")
}

func verifyTargetDeploymentReconciliation() {
	currTime := metav1.Now()
	targetDeploymentName := pvcObj.Spec.VolumeName + "-" + "target"
	err := ops.DeployClient.WithNamespace(openebsNamespace).
		Delete(targetDeploymentName, &metav1.DeleteOptions{})
	Expect(err).ShouldNot(HaveOccurred(), "while deleting target deployment")

	isRecreated := Eventually(func() bool {
		targetDeploy, err := ops.DeployClient.WithNamespace(openebsNamespace).
			Get(targetDeploymentName)
		if k8serror.IsNotFound(err) {
			return false
		}
		Expect(err).ShouldNot(HaveOccurred(), "while fetching target deployment")
		return currTime.Before(&targetDeploy.CreationTimestamp)
	},
		120, 10).
		Should(Equal(true))
	Expect(isRecreated).To(Equal(true), "while verifying target deployment recreation")
}

func deletePVC() {
	err := ops.PVCClient.WithNamespace(nsObj.Name).Delete(pvcName, &metav1.DeleteOptions{})
	Expect(err).To(
		BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcName,
		nsObj.Name,
	)
	By("verifying deleted pvc")
	status := ops.IsPVCDeletedEventually(pvcName)
	Expect(status).To(Equal(true), "while trying to get deleted pvc")

}

func verifyVolumeComponentsDeletion() {
	By("verifying target pod count as 0")
	controllerPodCount := ops.GetPodRunningCountEventually(
		openebsNamespace, targetLabel, 0)
	Expect(controllerPodCount).To(Equal(0),
		"while checking controller pod count")

	By("verifying if cstorvolume is deleted")
	CstorVolumeLabel := "openebs.io/persistent-volume=" + pvcObj.Spec.VolumeName
	cvCount := ops.GetCstorVolumeCountEventually(
		openebsNamespace, CstorVolumeLabel, 0)
	Expect(cvCount).To(Equal(true), "while checking cstorvolume count")
	By("Verifying cstorvolume replica count", func() { verifyCstorVolumeReplicaCount(0) })
}

func expandPVC() {
	var err error
	By("updating size in above pvc")
	pvcObj, err = pvc.BuildFrom(pvcObj).
		WithCapacity(updatedCapacity).Build()
	Expect(err).ShouldNot(HaveOccurred(), "while building pvc for update {%s}", pvcObj.Name)
	_, err = ops.PVCClient.WithNamespace(nsObj.Name).Update(pvcObj)
	Expect(err).To(
		BeNil(),
		"while updating size in pvc {%s} in namespace {%s}, size {%s}",
		pvcName,
		nsObj.Name,
	)

	By("verifying updated pvc capacity")
	status := ops.VerifyCapacity(pvcName, updatedCapacity)
	Expect(status).To(Equal(true),
		"while verifying updated pvc size")
}
