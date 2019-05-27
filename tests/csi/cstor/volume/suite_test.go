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
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	kubeConfigPath        string
	openebsNamespace      = "openebs"
	nsName                = "cstor-provision"
	scName                = "cstor-volume"
	openebsCASConfigValue = "- name: ReplicaCount\n  value: 1\n- name: StoragePoolClaim\n  value: sparse-pool-auto"
	openebsProvisioner    = "openebs-csi.openebs.io"
	spcName               = "sparse-pool-auto"
	nsObj                 *corev1.Namespace
	scObj                 *storagev1.StorageClass
	spcObj                *apis.StoragePoolClaim
	pvcObj                *corev1.PersistentVolumeClaim
	spcList               *apis.StoragePoolClaimList
	targetLabel           = "openebs.io/target=cstor-target"
	accessModes           = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity              = "5G"
	annotations           = map[string]string{
		string(apis.CASTypeKey):   string(apis.CstorVolume),
		string(apis.CASConfigKey): openebsCASConfigValue,
	}
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test cstor volume provisioning")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(kubeConfigPath))
	var err error
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

	By("creating a namespace")
	_, err = ops.NSClient.Create(nsObj)
	Expect(err).To(BeNil(), "while creating storageclass {%s}", nsObj.Name)
})

var _ = AfterSuite(func() {

	By("deleting namespace")
	err := ops.NSClient.Delete(nsName, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", nsObj.Name)

})
