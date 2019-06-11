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
	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TEST HOSTPATH LOCAL PV", func() {
	var (
		pvcObj        *corev1.PersistentVolumeClaim
		accessModes   = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		capacity      = "2Gi"
		deployName    = "busybox-hostpath"
		label         = "demo=hostpath-deployment"
		pvcName       = "pvc-hp"
		deployObj     *deploy.Deploy
		labelselector = map[string]string{
			"demo": "hostpath-deployment",
		}
	)

	When("pvc with storageclass openebs-hostpath is created", func() {
		It("should create a pvc ", func() {
			var (
				scName = "openebs-hostpath"
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
			_, err = ops.PVCClient.WithNamespace(namespaceObj.Name).Create(pvcObj)
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
				WithLabelSelector(labelselector).
				WithContainerBuilder(
					container.NewBuilder().
						WithName("busybox").
						WithImage("busybox").
						WithCommand(
							[]string{
								"sleep",
								"3600",
							},
						).
						WithVolumeMounts(
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
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building delpoyment {%s} in namespace {%s}",
				deployName,
				namespaceObj.Name,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespaceObj.Name).Create(deployObj.Object)
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

	When("pvc with storageclass openebs-hostpath is deleted ", func() {
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
