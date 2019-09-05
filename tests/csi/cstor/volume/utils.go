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
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//TODO: Below changes need to make for csi changes
//1) Change package name and move to operations.go
//2) Remove dependency with current package
//3) Make structures to create cspc, pvc and deployment resources

func createStorageClass(replicaCount int) {
	var (
		err error
	)
	parameters := map[string]string{
		"replicaCount":     strconv.Itoa(replicaCount),
		"cstorPoolCluster": cspcObj.Name,
		"cas-type":         "cstor",
	}

	By("building a storageclass")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(openebsProvisioner).Build()
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
}

func deleteStorageClass() {
	By("deleting storageclass")
	err := ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(),
		"while deleting storageclass {%s}", scObj.Name)
}

// CreateAndVerifyCStorPoolCluster creates cspc and
// verify whether pools are ONLINE or not
func CreateAndVerifyCStorPoolCluster(cacheFile, poolType string, poolCount int) {
	minRequiredBD := cspc.DefaultBDCount[poolType]
	blockDeviceList, err := blockdevice.GetBlockDeviceList(
		openebsNamespace,
		metav1.ListOptions{},
		blockdevice.WithKubeConfigPath(cstor.KubeConfigPath),
	)
	Expect(err).To(BeNil(), "while getting blockdeviceList")
	// Create pools on sparse blockdevices
	filteredBlockDevices := blockDeviceList.Filter(
		blockdevice.IsActive(),
		blockdevice.IsSparse(),
		blockdevice.IsNonFSType(),
		blockdevice.IsClaimStateMatched(ndmapis.BlockDeviceUnclaimed),
	)
	nodeBlockDevices := filteredBlockDevices.NodeBlockDeviceTopology()
	Expect(len(nodeBlockDevices)).Should(BeNumerically(">=", poolCount), "while segrigating blockdevices per node")
	cspcBuild := cspc.NewBuilder().
		WithGenerateName(cspcName).
		WithNamespace(openebsNamespace)

	//TODO: Move to some package other than test
	// Build pool spec configurations
	poolSpecCount := 0
	for hostName, blockdevices := range nodeBlockDevices {
		if len(blockdevices) < minRequiredBD {
			continue
		}
		if poolSpecCount == poolCount {
			break
		}
		cspcBDList := []*apis.CStorPoolClusterBlockDevice{}
		for i, bd := range blockdevices {
			cspcBDList = append(cspcBDList, &apis.CStorPoolClusterBlockDevice{
				BlockDeviceName: bd,
			})
			if i == minRequiredBD-1 {
				break
			}
		}
		nodeSelector := map[string]string{string(apis.HostNameCPK): hostName}
		poolSpecBuilder := poolspec.NewBuilder().
			WithNodeSelector(nodeSelector).
			WithCompression("off").
			WithDefaultRaidGroupType(poolType).
			WithRaidGroupBuilder(
				rgrp.NewBuilder().
					WithCSPCBlockDeviceList(cspcBDList),
			)
		if cacheFile != "" {
			poolSpecBuilder = poolSpecBuilder.WithCacheFilePath(cacheFile)
		}
		cspcBuild = cspcBuild.WithPoolSpecBuilder(poolSpecBuilder)
		poolSpecCount++
	}
	cspcBuildObj, err := cspcBuild.GetObj()
	Expect(err).To(BeNil(), "while building cstorpoolcluster")

	By("creating above cstorpoolcluster")
	cspcObj, err = ops.CSPCClient.WithNamespace(openebsNamespace).Create(cspcBuildObj)
	Expect(err).To(BeNil(),
		"while creating cspc with prefix {%s}", cspcName)

	By("verifying healthy cstorpoolinstance count")
	cspCount := ops.GetHealthyCSPICount(cspcObj.Name, poolCount)
	Expect(cspCount).To(Equal(poolCount),
		"while checking healthy cstorpoolinstance count")
}

// CreateAndVerifyPVC creates the pvc in provided namespace and
// verifies pvc bound status
func CreateAndVerifyPVC(pvcName string, labels map[string]string) {
	var (
		err error
	)
	By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithLabelsNew(labels).
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
	_, err = ops.PVCClient.WithNamespace(nsObj.Name).Create(pvcObj)
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
}

// CreateAndVerifyBulkPVC creates multiple pvcs based on count
func CreateAndVerifyBulkPVC(pvcPrefixName string, pvcCount int, labels map[string]string) {
	for i := 0; i < pvcCount; i++ {
		pvcName := pvcPrefixName + "-" + strconv.Itoa(i)
		CreateAndVerifyPVC(pvcName, labels)
	}
}

// CreateAndDeployApp creates deployment based on arguments and verifies application status
func CreateAndDeployApp(appName, pvcName string, appLabels map[string]string) {
	var err error
	By("Building busybox app deployment using above csi cstor volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appName).
		WithNamespace(nsObj.Name).
		WithLabelsNew(appLabels).
		WithSelectorMatchLabelsNew(appLabels).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(appLabels).
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
						WithPVCSource(pvcName),
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
	// Waiting for pod to be spawn by deployment controller
	_ = ops.IsDeploymentSuccessEventually(deployObj.Namespace, deployObj.Name, *deployObj.Spec.Replicas)
	By("verifying app pod is running")
	labelSelector := getLabelSelector(appLabels)
	appPod, err = ops.PodClient.WithNamespace(nsObj.Name).
		List(metav1.ListOptions{
			LabelSelector: labelSelector,
		},
		)
	Expect(err).ShouldNot(HaveOccurred(), "while verifying application pod")

	status := ops.IsPodRunningEventually(nsObj.Name, appPod.Items[0].Name)
	Expect(status).To(Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

// CreateAndDeployBulkApps build and deploy required number of applications
// deployment
func CreateAndDeployBulkApps(appPrefixName, pvcPrefixName string, appCount int) {
	for i := 0; i < appCount; i++ {
		appName := appPrefixName + "-" + strconv.Itoa(i)
		pvcName := pvcPrefixName + "-" + strconv.Itoa(i)
		appLabel := map[string]string{
			"bulk.app": strconv.Itoa(appCount),
			"app":      appName,
		}
		CreateAndDeployApp(appName, pvcName, appLabel)
	}
}

// VerifyVolumeComponents verifies volume related resources
func VerifyVolumeComponents() {
	By("should verify target pod count as 1", func() { verifyTargetPodCount(1) })
	By("should verify cstorvolume replica count", func() { verifyCstorVolumeReplicaCount(cstor.ReplicaCount) })
}

func verifyTargetPodCount(count int) {
	By("verifying target pod count as 1 once the app has been deployed")
	pvcObj, err := ops.PVCClient.WithNamespace(nsObj.Name).
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
	By("verifying cstorvolume replica count")
	cvrCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, targetVolumeLabel, count)
	Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")
}

func restartAppPodAndVerifyRunningStatus() {
	var err error
	By("restarting application to remount the volume again")
	err = ops.PodClient.WithNamespace(nsObj.Name).
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
	var err error
	By("deleting application deployment")
	err = ops.DeployClient.WithNamespace(nsObj.Name).
		Delete(deployObj.Name, &metav1.DeleteOptions{})
	Expect(err).ShouldNot(HaveOccurred(), "while deleting application pod")
}

// DeleteBulkApplications delete bulk application
// based on application label
func DeleteBulkApplications(appLabel string) {
	lopts := metav1.ListOptions{
		LabelSelector: appLabel,
	}
	err := ops.DeployClient.
		WithNamespace(nsObj.Name).
		DeleteCollection(lopts, &metav1.DeleteOptions{})
	Expect(err).ShouldNot(HaveOccurred(), "while deleting bulk applications")
}

// DeleteBulkPVCs delete bulk pvcs based on pvc's label
func DeleteBulkPVCs(bulkPVCLabel map[string]string) {
	lopts := metav1.ListOptions{
		LabelSelector: getLabelSelector(bulkPVCLabel),
	}
	err := ops.PVCClient.
		WithNamespace(nsObj.Name).
		DeleteCollection(lopts, &metav1.DeleteOptions{})
	Expect(err).ShouldNot(HaveOccurred(), "while deleting bulk pvc")
}

func deletePVC() {
	var err error
	By("deleting above pvc")
	err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
	Expect(err).To(
		BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcName,
		nsObj.Name,
	)
	By("verifying deleted pvc")
	pvc := ops.IsPVCDeleted(pvcName)
	Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

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
}

func expandPVC() {
	var err error
	By("updating size in above pvc")
	pvcObj, err = pvc.BuildFrom(pvcObj).
		WithCapacity(updatedCapacity).Build()
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

func getLabelSelector(labels map[string]string) string {
	var labelSelector string
	count := 0
	labelsLength := len(labels)
	for key, value := range labels {
		labelSelector = labelSelector + key + "=" + value
		if count < labelsLength-1 {
			labelSelector = labelSelector + ","
		}
		count++
	}
	return labelSelector
}
