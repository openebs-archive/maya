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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"

	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	openebsNamespace       = "openebs"
	nsName                 = "cstor-provision"
	scName                 = "cstor-volume"
	openebsProvisioner     = "cstor.csi.openebs.io"
	cspcName               = "cspc-sparse"
	pvcName                = "cstor-volume-claim"
	cvcVolumeAnnotationKey = "openebs.io/volumeID"

	nsObj           *corev1.Namespace
	scObj           *storagev1.StorageClass
	cspcObj         *apis.CStorPoolCluster
	deployObj       *appsv1.Deployment
	bdList          *ndmapis.BlockDeviceList
	pvcObj          *corev1.PersistentVolumeClaim
	appPod          *corev1.PodList
	serviceObj      *corev1.Service
	targetLabel     = "openebs.io/target=cstor-target"
	pvLabel         = "openebs.io/persistent-volume="
	accessModes     = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity        = "5G"
	updatedCapacity = "10G"
	targetPort      = 30031
	hostLabel       = "kubernetes.io/hostname"
	svcName         = "cvc-service-injector"
	hostIPPort      string
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test PodDisruptionBudget For CStor Volume")
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
	Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", nsObj.Name)

	By("Creating above namespace")
	nsObj, err = ops.NSClient.Create(nsObj)
	Expect(err).To(BeNil(), "while creating namespace {%s}", nsObj.Name)
	bdList, err = ops.BDClient.List(metav1.ListOptions{})
	Expect(err).To(BeNil(), "while gettting blockdevices")
})

var _ = AfterSuite(func() {

	By("deleting namespace")
	err := ops.NSClient.Delete(nsObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", nsObj.Name)
})
