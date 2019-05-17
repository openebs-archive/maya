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
	"flag"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/artifacts"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	kubeConfigPath        string
	nsName                = "provision-ns"
	scName                = "jiva-pods-in-openebs-ns"
	openebsCASConfigValue = "- name: ReplicaCount\n  Value: 1"
	openebsProvisioner    = "openebs.io/provisioner-iscsi"
	nsObj                 *corev1.Namespace
	scObj                 *storagev1.StorageClass
	annotations           = map[string]string{
		string(apis.CASTypeKey):   string(apis.JivaVolume),
		string(apis.CASConfigKey): openebsCASConfigValue,
	}
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test jiva volume provisioning ")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(kubeConfigPath))

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
	nsObj, err = ns.NewBuilder().
		WithName(nsName).
		APIObject()
	Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", nsName)

	By("building a storageclass")
	scObj, err = sc.NewBuilder().
		WithName(scName).
		WithAnnotations(annotations).
		WithProvisioner(openebsProvisioner).Build()
	Expect(err).ShouldNot(HaveOccurred(), "while building storageclass {%s}", scName)

	By("creating a namespace")
	_, err = ops.NsClient.Create(nsObj)
	Expect(err).To(BeNil(), "while creating storageclass {%s}", nsObj.Name)

	By("creating a storageclass")
	_, err = ops.ScClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating storageclass {%s}", scObj.Name)

})

var _ = AfterSuite(func() {

	By("deleting storageclass")
	err := ops.ScClient.Delete(scName, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting storageclass {%s}", scObj.Name)

	By("deleting namespace")
	err = ops.NsClient.Delete(nsName, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", nsObj.Name)

})
