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

package clone

import (
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	snap "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/cstor"

	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[cstor] TEST VOLUME CLONE PROVISIONING", func() {
	var (
		err          error
		pvcName      = "test-cstor-clone-pvc"
		snapName     = "test-cstor-clone-snapshot"
		clonepvcName = "test-cstor-clone-pvc-cloned"
		appName      = "busybox-cstor"
		cloneAppName = "busybox-cstor-clone"
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
			Expect(cspCount).To(Equal(cstor.PoolCount), "while checking healthy cstorpool count")

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
			Expect(err).To(BeNil(), "while deleting the spc {%s}", spcObj.Name)

			By("deleting storageclass")
			err = ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass", scObj.Name)

		})
	})

	When("cstor pvc with replicacount 1 is created", func() {
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
			sourcetargetLabel := pvcLabel + pvcObj.Name
			controllerPodCount := ops.GetPodRunningCountEventually(openebsNamespace, sourcetargetLabel, 1)
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

			By("verifying cstorVolume status as healthy")
			CstorVolumeLabel := pvLabel + pvcObj.Spec.VolumeName
			cvCount := ops.GetCstorVolumeCountEventually(openebsNamespace, CstorVolumeLabel, 1)
			Expect(cvCount).To(Equal(true), "while checking cstorvolume count")

			By("building a busybox app pod using above cstor volume")
			podObj, err = pod.NewBuilder().
				WithName(appName).
				WithNamespace(nsObj.Name).
				WithContainerBuilder(
					container.NewBuilder().
						WithName("busybox").
						WithImage("busybox").
						WithCommand(
							[]string{
								"sh",
								"-c",
								"date > /mnt/cstore1/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeMounts(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "datavol1",
									MountPath: "/mnt/cstore1",
								},
							},
						),
				).
				WithVolumeBuilder(
					volume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcName),
				).
				Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building pod {%s}", appName)

			By("creating pod with above pvc as volume")
			podObj, err = ops.PodClient.WithNamespace(nsObj.Name).Create(podObj)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while creating pod {%s} in namespace {%s}",
				appName,
				nsObj.Name,
			)

			By("verifying app pod is running")
			status = ops.IsPodRunningEventually(nsObj.Name, appName)
			Expect(status).To(Equal(true), "while checking status of pod {%s}", appName)

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

			snaptype := ops.GetSnapshotTypeEventually(snapName)
			Expect(snaptype).To(Equal("Ready"), "while checking snapshot type")

			By("builing clone persistentvolumeclaim")
			cloneAnnotations := map[string]string{
				"snapshot.alpha.kubernetes.io/snapshot": snapName,
			}

			cloneObj, err := pvc.NewBuilder().
				WithName(clonepvcName).
				WithAnnotations(cloneAnnotations).
				WithNamespace(nsObj.Name).
				WithStorageClass(clonescName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building clone pvc {%s} in namespace {%s}",
				clonepvcName,
				nsObj.Name,
			)

			By("creating clone persistentvolumeclaim")
			_, err = ops.PVCClient.WithNamespace(nsObj.Name).Create(cloneObj)
			Expect(err).To(
				BeNil(),
				"while creating clone pvc {%s} in namespace {%s}",
				clonepvcName,
				nsObj.Name,
			)

			By("verifying clone volume target pod count")

			clonetargetLabel := pvcLabel + clonepvcName
			clonePodCount := ops.GetPodRunningCountEventually(openebsNamespace, clonetargetLabel, 1)
			Expect(clonePodCount).To(Equal(1), "while checking clone pvc pod count")

			By("verifying clone volumeereplica count")
			clonepvcObj, err := ops.PVCClient.WithNamespace(nsObj.Name).Get(clonepvcName, metav1.GetOptions{})
			Expect(err).To(
				BeNil(),
				"while getting pvc {%s} in namespace {%s}",
				clonepvcName,
				nsObj.Name,
			)

			clonecvrLabel := pvLabel + clonepvcObj.Spec.VolumeName
			cvrCount = ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, clonecvrLabel, cstor.ReplicaCount)
			Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")

			By("verifying clone pvc status as bound")
			status = ops.IsPVCBoundEventually(clonepvcName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

			By("building a busybox app pod with above clone pvc as volume")
			clonePodObj, err = pod.NewBuilder().
				WithName(cloneAppName).
				WithNamespace(nsObj.Name).
				WithContainerBuilder(
					container.NewBuilder().
						WithName("busybox").
						WithImage("busybox").
						WithCommand(
							[]string{
								"sh",
								"-c",
								"tail -f /dev/null",
							},
						).
						WithVolumeMounts(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "datavol2",
									MountPath: "/mnt/cstore1",
								},
							},
						),
				).
				WithVolumeBuilder(
					volume.NewBuilder().
						WithName("datavol2").
						WithPVCSource(clonepvcName),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pod {%s}",
				cloneAppName,
			)

			By("creating pod with above clone pvc as volume")
			clonePodObj, err = ops.PodClient.WithNamespace(nsObj.Name).
				Create(clonePodObj)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while creating pod {%s} in namespace {%s}",
				cloneAppName,
				nsObj.Name,
			)

			By("verifying pod is running")
			status = ops.IsPodRunningEventually(nsObj.Name, cloneAppName)
			Expect(status).To(
				Equal(true),
				"while checking status of pod {%s}",
				cloneAppName,
			)

			By("verifying data consistency in pvc and clone pvc")
			By("fetching data from original pvc")
			podOutput, err := ops.PodClient.WithNamespace(nsObj.Name).
				Exec(
					podObj.Name,
					&corev1.PodExecOptions{
						Command: []string{
							"sh",
							"-c",
							"md5sum mnt/cstore1/date.txt",
						},
						Container: "busybox",
						Stdin:     false,
						Stdout:    true,
						Stderr:    true,
					},
				)
			Expect(err).ShouldNot(HaveOccurred(), "while exec in application pod")

			By("fetching data from clone pvc")
			clonePodOutput, err := ops.PodClient.WithNamespace(nsObj.Name).
				Exec(
					clonePodObj.Name,
					&corev1.PodExecOptions{
						Command: []string{
							"sh",
							"-c",
							"md5sum mnt/cstore1/date.txt",
						},
						Container: "busybox",
						Stdin:     false,
						Stdout:    true,
						Stderr:    true,
					},
				)
			Expect(err).ShouldNot(HaveOccurred(), "while exec in clone application pod")

			By("veryfing data consistency")
			Expect(podOutput).To(Equal(clonePodOutput), "while checking data consistency")

			By("deleting application pod")
			err = ops.PodClient.WithNamespace(nsObj.Name).
				Delete(podObj.Name, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting application pod")

			By("deleting clone application pod")
			err = ops.PodClient.WithNamespace(nsObj.Name).
				Delete(clonePodObj.Name, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting clone application pod")

			By("deleting clone persistentvolumeclaim")
			err = ops.PVCClient.Delete(clonepvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				nsObj.Name,
			)

			By("verifying clone target pod count as 0")
			controllerPodCount = ops.GetPodRunningCountEventually(openebsNamespace, clonetargetLabel, 0)
			Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

			By("verifying deleted clone pvc")
			clonepvc := ops.IsPVCDeleted(clonepvcName)
			Expect(clonepvc).To(Equal(true), "while trying to get deleted pvc")

			By("verifying if clone cstorvolume is deleted")
			CstorVolumeLabel = pvLabel + clonepvcObj.Spec.VolumeName
			clonecvCount := ops.GetCstorVolumeCountEventually(openebsNamespace, CstorVolumeLabel, 0)
			Expect(clonecvCount).To(Equal(true), "while checking cstorvolume count")

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

			By("deleting source persistentvolumeclaim")
			err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				nsObj.Name,
			)

			By("verifying source volume target pod count as 0")

			sourcetargetLabel = "openebs.io/persistent-volume-claim=" + pvcName
			controllerPodCount = ops.GetPodRunningCountEventually(openebsNamespace, sourcetargetLabel, 0)
			Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

			By("verifying deleted source pvc")
			pvc := ops.IsPVCDeleted(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

			By("verifying if source cstorvolume is deleted")
			CstorVolumeLabel = pvLabel + pvcObj.Spec.VolumeName
			cvCount = ops.GetCstorVolumeCountEventually(openebsNamespace, CstorVolumeLabel, 0)
			Expect(cvCount).To(Equal(true), "while checking cstorvolume count")

		})
	})

})
