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

package replicareplace

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

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	openebsNamespace      = "openebs"
	nsName                = "test-cstor-replica-replace"
	scName                = "test-cstor-replica-replace-sc"
	CstorPoolNameLabel    = "cstorpool.openebs.io/name"
	openebsCASConfigValue = `
- name: ReplicaCount
  value: $count
- name: StoragePoolClaim
  value: $spcName
`
	openebsProvisioner = "openebs.io/provisioner-iscsi"
	spcName            = "test-cstor-sparse-pool-auto"
	pvcName            = "replica-replace-pvc"
	nsObj              *corev1.Namespace
	scObj              *storagev1.StorageClass
	spcObj             *apis.StoragePoolClaim
	targetCVRObjList   *apis.CStorVolumeReplicaList
	poolPodObjList     *corev1.PodList
	//TODO:Need to remove
	migratingCSPObj        *apis.CStorPool
	cvObj                  *apis.CStorVolume
	newCVRObj              *apis.CStorVolumeReplica
	pvcObj                 *corev1.PersistentVolumeClaim
	pvLabel                = "openebs.io/persistent-volume="
	cspLabel               = "openebs.io/cstor-pool="
	CstorPoolUIDLabel      = "cstorpool.openebs.io/uid"
	PoolPrefix             = "cstor-"
	CstorPoolMgmtContainer = "cstor-pool-mgmt"
	accessModes            = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	annotations            = map[string]string{}
	MaxRetry               = 10
	ReplicaCount           int
	FailureReplicaCount    int
	ZvolGUID               string
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test cstor replica replace")
}

func init() {
	cstor.ParseFlags()
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(cstor.KubeConfigPath)).VerifyOpenebs(1)
	var err error

	By("Building a namespace")
	nsObj, err = ns.NewBuilder().
		WithGenerateName(nsName).
		APIObject()
	Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", nsName)

	By("creating  namespace")
	nsObj, err = ops.NSClient.Create(nsObj)
	Expect(err).To(BeNil(), "while creating namespace {%s}", nsName)
})

var _ = AfterSuite(func() {

	By("deleting namespace")
	err := ops.NSClient.Delete(nsObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", nsObj.Name)

})
