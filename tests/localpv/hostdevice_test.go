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

package localpv

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	localpv_app "github.com/openebs/maya/cmd/provisioner-localpv/app"
	blockdeviceclaim "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pts "github.com/openebs/maya/pkg/kubernetes/podtemplatespec/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TEST HOSTDEVICE LOCAL PV", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "2Gi"
		deployName    = "busybox-device"
		label         = "demo=hostdevice-deployment"
		pvcName       = "pvc-hd"
		deployObj     *appsv1.Deployment
		labelselector = map[string]string{
			"demo": "hostdevice-deployment",
		}
	)

	When("pvc with storageclass openebs-device is created", func() {
		It("should create a pvc ", func() {
			var (
				scName = "openebs-device"
			)

			By("building a pvc")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above pvc")
			pvcObj, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)
		})
	})

	When("deployment with busybox image is created", func() {
		It("should create a deployment and a running pod", func() {

			By("building a deployment")
			deployObj, err = deploy.NewBuilder().
				WithName(deployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(labelselector).
				WithSelectorMatchLabelsNew(labelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(labelselector).
						WithContainerBuildersNew(
							container.NewBuilder().
								WithName("busybox").
								WithImage("busybox").
								WithCommandNew(
									[]string{
										"sleep",
										"3600",
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
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-vol2").
								WithPVCSource(pvcName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building delpoyment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(deployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 1")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")

		})
	})
	When("remove finalizer", func() {
		It("finalizer should come back after provisioner restart", func() {
			bdcName := "bdc-pvc-" + string(pvcObj.GetUID())
			bdcObj, err := ops.BDCClient.WithNamespace(string(artifacts.OpenebsNamespace)).Get(bdcName,
				metav1.GetOptions{})
			Expect(err).To(BeNil())

			_, err = blockdeviceclaim.BuilderForAPIObject(bdcObj).WithConfigPath(ops.KubeConfigPath).
				BDC.RemoveFinalizer(localpv_app.LocalPVFinalizer)
			Expect(err).To(BeNil())

			podList, err := ops.PodClient.
				WithNamespace(string(artifacts.OpenebsNamespace)).
				List(metav1.ListOptions{LabelSelector: LocalPVProvisionerLabelSelector})
			Expect(err).To(BeNil())
			err = ops.PodClient.Delete(podList.Items[0].Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())

			Expect(ops.IsFinalizerExistsOnBDC(bdcName, localpv_app.LocalPVFinalizer)).To(BeTrue())
		})
	})
	When("deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(deployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("pvc with storageclass openebs-device is deleted ", func() {
		It("should delete the pvc", func() {

			By("deleting above pvc")
			err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

		})
	})

})

var _ = Describe("TEST HOSTDEVICE LOCAL PV WITH VOLUMEMODE AS BLOCK", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "2Gi"
		deployName    = "busybox-device"
		label         = "demo=hostdevice-deployment"
		pvcName       = "pvc-hd-block"
		deployObj     *appsv1.Deployment
		labelselector = map[string]string{
			"demo": "hostdevice-deployment",
		}
	)

	When("pvc with storageclass openebs-device, and volumeMode as Block, is created", func() {
		It("should create a pvc ", func() {
			var (
				scName          = "openebs-device"
				blockVolumeMode = corev1.PersistentVolumeBlock
			)

			By("building a pvc")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithVolumeMode(blockVolumeMode).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above pvc")
			pvcObj, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)
		})
	})

	When("deployment with busybox image is created", func() {
		It("should create a deployment and a running pod", func() {

			By("building a deployment")
			deployObj, err = deploy.NewBuilder().
				WithName(deployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(labelselector).
				WithSelectorMatchLabelsNew(labelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(labelselector).
						WithContainerBuildersNew(
							container.NewBuilder().
								WithName("busybox").
								WithImage("busybox").
								WithCommandNew(
									[]string{
										"sleep",
										"3600",
									},
								).
								WithVolumeDevices(
									[]corev1.VolumeDevice{
										corev1.VolumeDevice{
											Name:       "demo-block-vol1",
											DevicePath: "/dev/sdc",
										},
									},
								),
						).
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-block-vol1").
								WithPVCSource(pvcName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building delpoyment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(deployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 1")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")

		})
	})

	When("remove finalizer", func() {
		It("finalizer should come back after provisioner restart", func() {
			bdcName := "bdc-pvc-" + string(pvcObj.GetUID())
			bdcObj, err := ops.BDCClient.WithNamespace(string(artifacts.OpenebsNamespace)).Get(bdcName,
				metav1.GetOptions{})
			Expect(err).To(BeNil())

			_, err = blockdeviceclaim.BuilderForAPIObject(bdcObj).WithConfigPath(ops.KubeConfigPath).
				BDC.RemoveFinalizer(localpv_app.LocalPVFinalizer)
			Expect(err).To(BeNil())

			podList, err := ops.PodClient.
				WithNamespace(string(artifacts.OpenebsNamespace)).
				List(metav1.ListOptions{LabelSelector: LocalPVProvisionerLabelSelector})
			Expect(err).To(BeNil())
			err = ops.PodClient.Delete(podList.Items[0].Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())

			Expect(ops.IsFinalizerExistsOnBDC(bdcName, localpv_app.LocalPVFinalizer)).To(BeTrue())
		})
	})
	When("deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(deployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("pvc with storageclass openebs-device is deleted ", func() {
		It("should delete the pvc", func() {

			By("deleting above pvc")
			err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

		})
	})

})
var _ = Describe("[-ve] TEST HOSTDEVICE LOCAL PV", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "2Gi"
		deployName    = "busybox-device"
		label         = "demo=hostdevice-deployment"
		pvcName       = "pvc-hd"
		deployObj     *appsv1.Deployment
		labelselector = map[string]string{
			"demo": "hostdevice-deployment",
		}
		scName                = "openebs-device"
		existingPVCObj        *corev1.PersistentVolumeClaim
		existingDeployName    = "existing-busybox-device"
		existinglabel         = "demo=existing-hostdevice-deployment"
		existingPVCName       = "existing-pvc-hd"
		existingDeployObj     *appsv1.Deployment
		existingLabelselector = map[string]string{
			"demo": "existing-hostdevice-deployment",
		}
	)
	When("existing pvc with storageclass openebs-device is created", func() {
		It("should create a pvc", func() {

			By("building a pvc")
			existingPVCObj, err = pvc.NewBuilder().
				WithName(existingPVCName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)

			By("creating above pvc")
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(existingPVCObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)
		})
	})

	When("existing deployment with busybox image is created", func() {
		It("should create a deployment and a running pod", func() {

			By("building a deployment")
			existingDeployObj, err = deploy.NewBuilder().
				WithName(existingDeployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(existingLabelselector).
				WithSelectorMatchLabelsNew(existingLabelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(existingLabelselector).
						WithContainerBuildersNew(
							container.NewBuilder().
								WithName("busybox").
								WithImage("busybox").
								WithCommandNew(
									[]string{
										"sleep",
										"3600",
									},
								).
								WithVolumeMountsNew(
									[]corev1.VolumeMount{
										corev1.VolumeMount{
											Name:      "demo-vol3",
											MountPath: "/mnt/store1",
										},
									},
								),
						).
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-vol3").
								WithPVCSource(existingPVCName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(existingDeployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 1")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, existinglabel, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")

		})
	})

	When("another pvc with storageclass openebs-device is created", func() {
		It("should create a pvc ", func() {

			By("building a pvc")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

			By("creating above pvc")
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(pvcObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)
		})
	})

	When("another deployment with busybox image and above pvc is created", func() {
		It("should not create a deployment and a running pod", func() {

			By("building a deployment")
			deployObj, err = deploy.NewBuilder().
				WithName(deployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(labelselector).
				WithSelectorMatchLabelsNew(labelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(labelselector).
						WithContainerBuildersNew(
							container.NewBuilder().
								WithName("busybox").
								WithImage("busybox").
								WithCommandNew(
									[]string{
										"sleep",
										"3600",
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
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-vol2").
								WithPVCSource(pvcName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building delpoyment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(deployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("above deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(deployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("above pvc with storageclass openebs-device is deleted ", func() {
		It("should delete the pvc", func() {

			By("deleting above pvc")
			err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				pvcName,
				namespaceObj.Name,
			)

		})
	})

	When("existing deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Delete(existingDeployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, existinglabel, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")

		})
	})

	When("existing pvc with storageclass openebs-device is deleted ", func() {
		It("should delete the pvc", func() {

			By("deleting above pvc")
			err = ops.PVCClient.Delete(existingPVCName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)

		})
	})
})

var _ = Describe("[-ve] TEST HOSTDEVICE LOCAL PV WITH VOLUMEMODE AS BLOCK ", func() {
	var (
		//pvcObj          *corev1.PersistentVolumeClaim
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity    = "2Gi"
		//deployName      = "busybox-device"
		//label           = "demo=hostdevice-deployment"
		//pvcName         = "pvc-hd-block"
		//deployObj       *appsv1.Deployment
		blockVolumeMode = corev1.PersistentVolumeBlock
		//labelselector   = map[string]string{
		//	"demo": "hostdevice-deployment",
		//}
		scName                = "openebs-device"
		existingPVCObj        *corev1.PersistentVolumeClaim
		existingDeployName    = "existing-busybox-device"
		existinglabel         = "demo=existing-hostdevice-deployment"
		existingPVCName       = "existing-pvc-hd-block"
		existingDeployObj     *appsv1.Deployment
		existingLabelselector = map[string]string{
			"demo": "existing-hostdevice-deployment",
		}
	)
	When("existing pvc with storageclass openebs-device is created", func() {
		It("should create a pvc", func() {

			By("building a pvc")
			existingPVCObj, err = pvc.NewBuilder().
				WithName(existingPVCName).
				WithNamespace(namespaceObj.Name).
				WithStorageClass(scName).
				WithAccessModes(accessModes).
				WithVolumeMode(blockVolumeMode).
				WithCapacity(capacity).Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pvc {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)

			By("creating above pvc")
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(existingPVCObj)
			Expect(err).To(
				BeNil(),
				"while creating pvc {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)
		})
	})
	When("existing deployment with busybox image is created", func() {
		It("should create a deployment but should be unable to get a running pod, with PVC volumeMode set to Block,but added as volumeMount in Deployment", func() {

			By("building a deployment, with volume Mount for a Block volumeMode PVC")
			existingDeployObj, err = deploy.NewBuilder().
				WithName(existingDeployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsNew(existingLabelselector).
				WithSelectorMatchLabelsNew(existingLabelselector).
				WithPodTemplateSpecBuilder(
					pts.NewBuilder().
						WithLabelsNew(existingLabelselector).
						WithContainerBuildersNew(
							container.NewBuilder().
								WithName("busybox").
								WithImage("busybox").
								WithCommandNew(
									[]string{
										"sleep",
										"3600",
									},
								).
								WithVolumeMountsNew(
									[]corev1.VolumeMount{
										corev1.VolumeMount{
											Name:      "demo-block-vol2",
											MountPath: "/mnt/store1",
										},
									},
								),
						).
						WithVolumeBuildersNew(
							volume.NewBuilder().
								WithName("demo-block-vol2").
								WithPVCSource(existingPVCName),
						),
				).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).
				Create(existingDeployObj)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("verifying pvc status as bound")
			status := ops.IsPVCBoundEventually(existingPVCName)
			Expect(status).To(Equal(true), "while checking status equal to bound")

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, existinglabel, 1)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})
	When("above deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err = ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(existingDeployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				existingDeployName,
				namespaceObj.Name,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespaceObj.Name, existinglabel, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

	When("existing pvc with storageclass openebs-device is deleted ", func() {
		It("should delete the pvc", func() {

			By("deleting above pvc")
			err = ops.PVCClient.Delete(existingPVCName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting pvc {%s} in namespace {%s}",
				existingPVCName,
				namespaceObj.Name,
			)

		})
	})
})
