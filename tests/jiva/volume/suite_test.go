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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/artifacts"
	"github.com/openebs/maya/tests/jiva"

	jivaClient "github.com/openebs/maya/pkg/client/jiva"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	namespace             = "jiva-volume-ns"
	openebsCASConfigValue = "- name: ReplicaCount\n  Value: "
	openebsProvisioner    = "openebs.io/provisioner-iscsi"
	namespaceObj          *corev1.Namespace
	err                   error
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test jiva volume provisioning")
}

func init() {
	jiva.ParseFlags()
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(jiva.KubeConfigPath))

	By("waiting for maya-apiserver pod to come into running state")
	podCount := ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.MayaAPIServerLabelSelector),
		1,
	)
	Expect(podCount).To(Equal(1))

	By("waiting for openebs-provisioner pod to come into running state")
	podCount = ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSProvisionerLabelSelector),
		1,
	)
	Expect(podCount).To(Equal(1))

	By("building a namespace")
	namespaceObj, err = ns.NewBuilder().
		WithGenerateName(namespace).
		APIObject()
	Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", namespace)

	By("creating a namespace")
	namespaceObj, err = ops.NSClient.Create(namespaceObj)
	Expect(err).To(BeNil(), "while creating namespace {%s}", namespaceObj.GenerateName)

})

var _ = AfterSuite(func() {

	By("deleting namespace")
	err = ops.NSClient.Delete(namespaceObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", namespaceObj.Name)

})

func areReplicasRegisteredEventually(ctrlPod *corev1.Pod, replicationFactor int) bool {
	return Eventually(func() int {
		out, err := ops.PodClient.WithNamespace(ctrlPod.Namespace).
			Exec(
				ctrlPod.Name,
				&corev1.PodExecOptions{
					Command: []string{
						"/bin/bash",
						"-c",
						"curl http://localhost:9501/v1/volumes",
					},
					Container: ctrlPod.Spec.Containers[0].Name,
					Stdin:     false,
					Stdout:    true,
					Stderr:    true,
				},
			)
		Expect(err).ShouldNot(HaveOccurred(), "while exec in application pod")

		volumes := jivaClient.VolumeCollection{}
		err = json.Unmarshal([]byte(out.Stdout), &volumes)
		Expect(err).To(BeNil(), "while unmarshalling volumes %s", out.Stdout)

		return volumes.Data[0].ReplicaCount
	},
		300, 10).Should(Equal(replicationFactor))
}
