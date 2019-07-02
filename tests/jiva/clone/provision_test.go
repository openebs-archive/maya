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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests/jiva"

	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	snap "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	accessModes         = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity            = "5G"
	pvcObj              *corev1.PersistentVolumeClaim
	cloneObj            *corev1.PersistentVolumeClaim
	cloneLable          = "openebs.io/persistent-volume-claim=jiva-clone"
	podObj, clonePodObj *corev1.Pod
)

var _ = Describe("[jiva] TEST JIVA CLONE CREATION", func() {
	var (
		pvcName      = "jiva-pvc"
		snapName     = "jiva-snapshot"
		cloneName    = "jiva-clone"
		appName      = "busybox-jiva"
		cloneAppName = "busybox-jiva-clone"
	)

	When("pvc with replicacount n is created", func() {
		It("should create 1 controller pod and n replica pod", func() {

			By("building a pvc")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scObj.Name).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above pvc")
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).
				Create(pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("verifying controller pod count")
			controllerPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				jiva.CtrlLabel,
				1,
			)
			Expect(controllerPodCount).To(
				Equal(1),
				"while checking controller pod count",
			)

			By("verifying replica pod count")
			replicaPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				jiva.ReplicaLabel,
				jiva.ReplicaCount,
			)
			Expect(replicaPodCount).To(
				Equal(jiva.ReplicaCount),
				"while checking replica pod count",
			)

			By("verifying status as bound")
			status := ops.IsPVCBoundEventually(pvcName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

		})
	})

	When("creating application pod with above pvc as volume", func() {
		It("should create a running pod", func() {
			podObj, err = pod.NewBuilder().
				WithName(appName).
				WithNamespace(namespaceObj.Name).
				WithContainerBuilder(
					container.NewBuilder().
						WithName("busybox").
						WithImage("busybox").
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/store1/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "demo-vol1",
									MountPath: "/mnt/store1",
								},
							},
						),
				).
				WithVolumeBuilder(
					volume.NewBuilder().
						WithName("demo-vol1").
						WithPVCSource(pvcName),
				).
				Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building pod {%s}", appName)

			By("creating pod with above pvc as volume")
			podObj, err = ops.PodClient.WithNamespace(namespaceObj.Name).Create(podObj)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while creating pod {%s} in namespace {%s}",
				appName,
				namespaceObj.Name,
			)

			By("verifying pod is running")
			status := ops.IsPodRunningEventually(namespaceObj.Name, appName)
			Expect(status).To(Equal(true), "while checking status of pod {%s}", appName)

		})
	})

	When("jiva snapshot is created", func() {
		It("should create a snapshot with type ready", func() {

			By("building a snapshot")
			snapObj, err = snap.NewBuilder().
				WithName(snapName).
				WithNamespace(namespaceObj.Name).
				WithPVC(pvcName).
				Build()
			Expect(err).To(
				BeNil(),
				"while building snapshot {%s} in namespace {%s}",
				snapName,
				namespaceObj.Name,
			)

			By("creating above snapshot")
			_, err = ops.SnapClient.WithNamespace(namespaceObj.Name).Create(snapObj)
			Expect(err).To(
				BeNil(),
				"while creating snapshot{%s} in namespace {%s}",
				snapName,
				namespaceObj.Name,
			)

			By("verifying type as ready")
			snaptype := ops.GetSnapshotTypeEventually(snapName)
			Expect(snaptype).To(Equal("Ready"), "while checking snapshot status")

		})
	})

	When("jiva clone pvc is created", func() {
		It("should create same number of pods as above pvc", func() {

			cloneAnnotations := map[string]string{
				"snapshot.alpha.kubernetes.io/snapshot": snapName,
			}

			By("building a clone pvc")
			cloneObj, err = pvc.NewBuilder().
				WithName(cloneName).
				WithAnnotations(cloneAnnotations).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(openebsCloneStorageclass).
				WithAccessModes(accessModes).
				WithCapacity(capacity).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building clone pvc {%s} in namespace {%s}",
				cloneName,
				namespaceObj.Name,
			)

			By("creating above clone pvc")
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(cloneObj)
			Expect(err).To(
				BeNil(),
				"while creating clone pvc {%s} in namespace {%s}",
				cloneName,
				namespaceObj.Name,
			)

			By("verifying clone pod count")
			clonePodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				cloneLable,
				jiva.ReplicaCount+1,
			)
			Expect(clonePodCount).To(
				Equal(jiva.ReplicaCount+1),
				"while checking clone pvc pod count",
			)

			By("verifying status as bound")
			status := ops.IsPVCBound(cloneName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

		})
	})

	When("creating application pod with above clone pvc as volume", func() {
		It("should create a running pod", func() {
			clonePodObj, err = pod.NewBuilder().
				WithName(cloneAppName).
				WithNamespace(namespaceObj.Name).
				WithContainerBuilder(
					container.NewBuilder().
						WithName("busybox").
						WithImage("busybox").
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"tail -f /dev/null",
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "demo-vol2",
									MountPath: "/mnt/store1",
								},
							},
						),
				).
				WithVolumeBuilder(
					volume.NewBuilder().
						WithName("demo-vol2").
						WithPVCSource(cloneName),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pod {%s}",
				cloneAppName,
			)

			By("creating pod with above pvc as volume")
			clonePodObj, err = ops.PodClient.WithNamespace(namespaceObj.Name).
				Create(clonePodObj)
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while creating pod {%s} in namespace {%s}",
				cloneAppName,
				namespaceObj.Name,
			)

			By("verifying pod is running")
			status := ops.IsPodRunningEventually(namespaceObj.Name, cloneAppName)
			Expect(status).To(
				Equal(true),
				"while checking status of pod {%s}",
				cloneAppName,
			)

		})
	})

	When("verifying data consistency in pvc and clone pvc", func() {
		It("should have consistent data between the two pvcs", func() {
			By("fetching data from original pvc")
			podOutput, err := ops.PodClient.WithNamespace(namespaceObj.Name).
				Exec(
					podObj.Name,
					&corev1.PodExecOptions{
						Command: []string{
							"sh",
							"-c",
							"md5sum mnt/store1/date.txt",
						},
						Container: "busybox",
						Stdin:     false,
						Stdout:    true,
						Stderr:    true,
					},
				)
			Expect(err).ShouldNot(HaveOccurred(), "while exec in application pod")

			By("fetching data from clone pvc")
			clonePodOutput, err := ops.PodClient.WithNamespace(namespaceObj.Name).
				Exec(
					clonePodObj.Name,
					&corev1.PodExecOptions{
						Command: []string{
							"sh",
							"-c",
							"md5sum mnt/store1/date.txt",
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
			err = ops.PodClient.WithNamespace(namespaceObj.Name).
				Delete(podObj.Name, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting application pod")

			By("deleting clone application pod")
			err = ops.PodClient.WithNamespace(namespaceObj.Name).
				Delete(clonePodObj.Name, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting clone application pod")

		})
	})

	When("deleting clone pvc", func() {
		It("should remove clone pvc pods", func() {

			By("deleting above clone pvc")
			err := ops.PVCClient.WithNamespace(namespaceObj.Name).
				Delete(cloneName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting clone pvc {%s} in namespace {%s}",
				cloneName,
				namespaceObj.Name,
			)

			By("verifying clone pvc pods as 0")
			clonePodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				cloneLable,
				0,
			)
			Expect(clonePodCount).To(Equal(0), "while checking clone pvc pod count")

			By("verifying deleted clone pvc")
			pvc := ops.IsPVCDeleted(cloneName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

		})
	})

	When("jiva snapshot is deleted", func() {
		It("should remove above snapshot", func() {

			By("deleting above snapshot")
			err := ops.SnapClient.Delete(snapName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting snapshot {%s} in namespace {%s}",
				snapName,
				namespaceObj.Name,
			)

			By("verifying deleted snapshot")
			snap := ops.IsSnapshotDeleted(snapName)
			Expect(snap).To(Equal(true), "while checking for deleted snapshot")

		})
	})

	When("jiva pvc is deleted", func() {
		It("should not have any jiva controller and replica pods", func() {

			By("deleting above pvc")
			err := ops.PVCClient.WithNamespace(namespaceObj.Name).
				Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("verifying controller pod count as 0")
			controllerPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				jiva.CtrlLabel,
				0,
			)
			Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

			By("verifying replica pod count as 0")
			replicaPodCount := ops.GetPodRunningCountEventually(
				namespaceObj.Name,
				jiva.ReplicaLabel,
				0,
			)
			Expect(replicaPodCount).To(Equal(0), "while checking replica pod count")

			By("verifying deleted pvc")
			pvc := ops.IsPVCDeleted(pvcName)
			Expect(pvc).To(Equal(true), "while trying to get deleted pvc")

		})
	})

})
