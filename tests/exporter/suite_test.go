// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter

import (
	"flag"
	"fmt"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/artifacts"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	kubeConfigPath        string
	nsName                = "exporter"
	scName                = "test-pool-exporter"
	spcName               = "sparse-striped-auto"
	pvcName               = "exporter-volume"
	openebsCASConfigValue = "- name: ReplicaCount\n  value: 1\n- name: StoragePoolClaim\n  value: sparse-striped-auto"
	openebsProvisioner    = "openebs.io/provisioner-iscsi"
	nsObj                 *corev1.Namespace
	scObj                 *storagev1.StorageClass
	spcObj                *apis.StoragePoolClaim
	pvcObj                *corev1.PersistentVolumeClaim
	cspAPIList            *apis.CStorPoolList
	pvcs                  *corev1.PersistentVolumeClaimList
	cvs                   *apis.CStorVolumeList
	csv                   *apis.CStorVolume
	annotations           = map[string]string{
		string(apis.CASTypeKey):   string(apis.CstorVolume),
		string(apis.CASConfigKey): openebsCASConfigValue,
	}
)

var (
	ops *tests.Operations
)

func TestExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Maya exporter test suite")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
}

var _ = BeforeSuite(func() {

	var err error
	ops = tests.NewOperations(tests.WithKubeConfigPath(kubeConfigPath))

	When("we have already deployed openebs components", func() {
		By("waiting for maya-apiserver pod to come into running state")
		podCount := ops.GetPodRunningCountEventually(
			string(artifacts.OpenebsNamespace),
			string(artifacts.MayaAPIServerLabelSelector),
			1,
		)
		Expect(podCount).To(Equal(1), "while checking maya-apiserver pod count")
		By("waiting for openebs-provisioner pod to come into running state")
		podCnt := ops.GetPodRunningCountEventually(
			string(artifacts.OpenebsNamespace),
			string(artifacts.OpenEBSProvisionerLabelSelector),
			1,
		)
		Expect(podCnt).To(Equal(1), "while checking openebs-provisioner pod count")
	})

	When("we are creating namespace", func() {
		By("building namespace object")
		nsObj, err = ns.NewBuilder().
			WithName(nsName).
			APIObject()
		Expect(err).ShouldNot(HaveOccurred(), "while building namespace object for namespace {%s}", nsName)

		By(fmt.Sprintf("creating namespace {%s}", nsName))
		_, err = ops.NSClient.Create(nsObj)
		Expect(err).ShouldNot(HaveOccurred(), "while creating namespace {%s}", nsName)
	})

})

var _ = AfterSuite(func() {

	When("we are deleting namespace", func() {
		err := ops.NSClient.Delete(nsName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil(), "while deleting namespace {%s}", nsObj.Name)
	})
})
