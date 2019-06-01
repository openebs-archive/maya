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
	con "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	deployName    = "busybox-deploy"
	label         = "demo=deployment"
	deployObj     *deploy.Deploy
	conObj        corev1.Container
	command       []string
	mounts        []corev1.VolumeMount
	volumes       []corev1.Volume
	labelselector = map[string]string{
		"demo": "deployment",
	}
)

var _ = Describe("TEST LOCAL PV", func() {

	When("deployment with busybox image is created", func() {
		It("should create a deployment and a running pod", func() {

			command = append(
				command,
				"sh",
				"-c",
				"date > /mnt/store1/date.txt; hostname >> /mnt/store1/hostname.txt; sync; sleep 5; sync; tail -f /dev/null;",
			)

			mounts = append(
				mounts,
				corev1.VolumeMount{
					Name:      "demo-vol1",
					MountPath: "/mnt/store1",
				},
			)

			volumes = append(
				volumes,
				corev1.Volume{
					Name: "demo-vol1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
			)

			By("building a container")
			conObj, err = con.Builder().
				WithName("busybox").
				WithImage("busybox").
				WithCommand(command).
				WithVolumeMounts(mounts).
				Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building a container")

			By("building a deployment")
			deployObj, err = deploy.NewBuilder().
				WithName(deployName).
				WithNamespace(namespace).
				WithLabelsAndSelector(labelselector).
				WithContainer(&conObj).
				WithVolumes(volumes).
				Build()
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building delpoyment {%s} in namespace {%s}",
				deployName,
				namespace,
			)

			By("creating above deployment")
			_, err = ops.DeployClient.WithNamespace(namespace).Create(deployObj.Object)
			Expect(err).To(
				BeNil(),
				"while creating deployment {%s} in namespace {%s}",
				deployName,
				namespace,
			)

			By("verifying pod count as 1")
			podCount := ops.GetPodRunningCountEventually(namespace, label, 1)
			Expect(podCount).To(Equal(1), "while verifying pod count")
		})
	})

	When("deployment is deleted", func() {
		It("should not have any deployment or running pod", func() {

			By("deleting above deployment")
			err := ops.DeployClient.WithNamespace(namespace).Delete(deployName, &metav1.DeleteOptions{})
			Expect(err).To(
				BeNil(),
				"while deleting deployment {%s} in namespace {%s}",
				deployName,
				namespace,
			)

			By("verifying pod count as 0")
			podCount := ops.GetPodRunningCountEventually(namespace, label, 0)
			Expect(podCount).To(Equal(0), "while verifying pod count")

		})
	})

})
