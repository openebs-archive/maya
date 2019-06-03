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

package app

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
	err           error
	command       []string
	labelselector = map[string]string{
		"demo": "deployment",
	}
)

var _ = Describe("TEST DEPLOYMENT CREATION ", func() {

	When("deployment with busybox image is created", func() {
		It("should create a deployment and a running pod", func() {

			command = append(command, "sleep", "3600")

			By("building a deployment")
			deployObj, err = deploy.NewBuilder().
				WithName(deployName).
				WithNamespace(namespaceObj.Name).
				WithLabelsAndSelector(labelselector).
				WithContainerBuilder(
					con.NewBuilder().
						WithName("busybox").
						WithImage("busybox").
						WithCommand(command),
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
			err := ops.DeployClient.WithNamespace(namespaceObj.Name).Delete(deployName, &metav1.DeleteOptions{})
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

})
