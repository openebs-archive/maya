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

package nodeselector

import (
	"strconv"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/artifacts"
	"github.com/openebs/maya/tests/jiva"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	namespace             = "jiva-volume-ns"
	scName                = "jiva-volume-sc"
	openebsProvisioner    = "openebs.io/provisioner-iscsi"
	namespaceObj          *corev1.Namespace
	scObj                 *storagev1.StorageClass
	nodeList              *corev1.NodeList
	annotations           = map[string]string{}
	appNode, storageNode  string
	openebsCASConfigValue = `- name: ReplicaNodeSelector
  value: |-
    nodetype: storage
- name: TargetNodeSelector
  value: |-
    nodetype: app
- name: ReplicaCount
  value: `
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test jiva volume node selector")
}

func init() {
	jiva.ParseFlags()
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(jiva.KubeConfigPath))

	annotations[string(apis.CASTypeKey)] = string(apis.JivaVolume)
	annotations[string(apis.CASConfigKey)] = openebsCASConfigValue + strconv.Itoa(jiva.ReplicaCount)

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

	By("building a storageclass")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithAnnotations(annotations).
		WithProvisioner(openebsProvisioner).Build()
	Expect(err).ShouldNot(HaveOccurred(), "while building storageclass {%s}", scName)

	By("creating above namespace")
	namespaceObj, err = ops.NSClient.Create(namespaceObj)
	Expect(err).To(BeNil(), "while creating namespace {%s}", namespaceObj.GenerateName)

	By("creating above storageclass")
	scObj, err = ops.SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating storageclass {%s}", scObj.GenerateName)

	By("listing ready nodes")
	nodeList = ops.GetReadyNodes()

	By("verifying minimum node count to be 2")
	Expect(len(nodeList.Items)).Should(BeNumerically(">=", 2))

	appNode = nodeList.Items[0].Name
	By("labeling node 1 as app")
	_, err = ops.NodeClient.Patch(appNode,
		types.MergePatchType,
		[]byte(`{"metadata":{"labels":{"nodetype":"app"}}}`),
	)
	Expect(err).ShouldNot(HaveOccurred(), "while patching app node")

	storageNode = nodeList.Items[1].Name
	By("labeling node 2 as storage")
	_, err = ops.NodeClient.Patch(storageNode,
		types.MergePatchType,
		[]byte(`{"metadata":{"labels":{"nodetype":"storage"}}}`),
	)
	Expect(err).ShouldNot(HaveOccurred(), "while patching storage node")

})

var _ = AfterSuite(func() {

	By("removing label from node 1")
	_, err = ops.NodeClient.Patch(appNode,
		types.MergePatchType,
		[]byte(`{"metadata":{"labels":{"nodetype":null}}}`),
	)
	Expect(err).ShouldNot(HaveOccurred(), "while removing app node label")

	By("removing label from node 2")
	_, err = ops.NodeClient.Patch(storageNode,
		types.MergePatchType,
		[]byte(`{"metadata":{"labels":{"nodetype":null}}}`),
	)
	Expect(err).ShouldNot(HaveOccurred(), "while removing storage node label")

	By("deleting storageclass")
	err = ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting storageclass {%s}", scObj.Name)

	By("deleting namespace")
	err = ops.NSClient.Delete(namespaceObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", namespaceObj.Name)

})
