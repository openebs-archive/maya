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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	openebsNamespace      = "openebs"
	nsName                = "test-cstor-volume-selector"
	scName                = "test-cstor-volume-selector-sc"
	openebsCASConfigValue = `
- name: ReplicaCount
  value: $count
- name: StoragePoolClaim
  value: $spcName
- name: TargetNodeSelector
  value: |-
    nodetype: storage
`
	openebsProvisioner = "openebs.io/provisioner-iscsi"
	spcName            = "test-cstor-selector-sparse-pool-auto"
	nsObj              *corev1.Namespace
	scObj              *storagev1.StorageClass
	spcObj             *apis.StoragePoolClaim
	pvcObj             *corev1.PersistentVolumeClaim
	nodeList           *corev1.NodeList
	pvLabel            = "openebs.io/persistent-volume="
	pvcLabel           = "openebs.io/persistent-volume-claim="
	accessModes        = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity           = "5G"
	storageNode        string
	annotations        = map[string]string{}
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test cstor volume provisioning using target Node selector")
}

func init() {
	cstor.ParseFlags()
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(cstor.KubeConfigPath)).VerifyOpenebs(1)
	var err error

	By("building a namespace")
	nsObj, err = ns.NewBuilder().
		WithGenerateName(nsName).
		APIObject()
	Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", nsName)

	By("creating a namespace")
	nsObj, err = ops.NSClient.Create(nsObj)
	Expect(err).To(BeNil(), "while creating namespace {%s}", nsObj.Name)

	By("listing ready nodes")
	nodeList = ops.GetReadyNodes()

	By("verifying minimum node count to be 1")
	Expect(len(nodeList.Items)).Should(BeNumerically(">=", 1))

	storageNode = nodeList.Items[0].Name
	By("labeling node as 'nodetype:storage' ")
	_, err = ops.NodeClient.Patch(storageNode,
		types.MergePatchType,
		[]byte(`{"metadata":{"labels":{"nodetype":"storage"}}}`),
	)
	Expect(err).ShouldNot(HaveOccurred(), "while patching storage node")

})

var _ = AfterSuite(func() {

	By("deleting namespace")
	err := ops.NSClient.Delete(nsObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", nsObj.Name)

	By("remove selector label from labeled nodes")
	_, err = ops.NodeClient.Patch(storageNode,
		types.MergePatchType,
		[]byte(`{"metadata":{"labels":{"nodetype": null }}}`),
	)
	Expect(err).ShouldNot(HaveOccurred(), "while patching storage node")

})
