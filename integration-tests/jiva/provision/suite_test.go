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

package provision

import (
	"flag"
	"time"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	MaxRetry = 30
)

var (
	kubeConfigPath string
)

type operations struct {
	podClient *pod.KubeClient
	scClient  *sc.Kubeclient
	pvcClient *pvc.Kubeclient
	nsClient  *ns.Kubeclient
}

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test volume provisioning in openebs namespace")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
}

var _ = BeforeSuite(func() {
	// set pod client set
	for _, f := range clientBuilderFuncList {
		f()
	}

	By("Waiting for maya-apiserver pod to come into running state")
	podCount := ops.getPodCountRunningEventually(string(artifacts.OpenebsNamespace), string(artifacts.MayaAPIServerLabelSelector), 1)
	Expect(podCount).To(Equal(1))

	By("Waiting for openebs-provisioner pod to come into running state")
	podCount = ops.getPodCountRunningEventually(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSProvisionerLabelSelector), 1)
	Expect(podCount).To(Equal(1))
})

var _ = AfterSuite(func() {
})

var ops = &operations{}

type clientBuilderFunc func() *operations

var clientBuilderFuncList = []clientBuilderFunc{
	ops.newNsClient,
	ops.newPodClient,
	ops.newSCClient,
	ops.newPVCClient,
}

func (ops *operations) newNsClient() *operations {
	newNsClient := ns.NewKubeClient(ns.WithKubeConfigPath(kubeConfigPath))
	ops.nsClient = newNsClient
	return ops
}

func (ops *operations) newPodClient() *operations {
	newPodClient := pod.NewKubeClient(pod.WithKubeConfigPath(kubeConfigPath))
	ops.podClient = newPodClient
	return ops
}

func (ops *operations) newSCClient() *operations {
	newSCClient := sc.NewKubeClient(sc.WithKubeConfigPath(kubeConfigPath))
	ops.scClient = newSCClient
	return ops
}

func (ops *operations) newPVCClient() *operations {
	newPVCClient := pvc.NewKubeClient(pvc.WithKubeConfigPath(kubeConfigPath))
	ops.pvcClient = newPVCClient
	return ops
}

func (ops *operations) getPodCountRunningEventually(namespace, lselector string, expectedPodCount int) int {
	var maxRetry int
	var podCount int
	maxRetry = MaxRetry
	for i := 0; i < maxRetry; i++ {
		podCount = ops.getRunningPodCount(namespace, lselector)
		if podCount == expectedPodCount {
			return podCount
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}

func (ops *operations) getRunningPodCount(namespace, lselector string) int {
	pods, err := ops.podClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		WithFilter(pod.IsRunning()).
		List().
		Len()
}

func (ops *operations) isBound(pvcName string) bool {
	volume, err := ops.pvcClient.
		Get(pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return pvc.NewForAPIObject(volume).IsBound()
}

// checkDeletedPVC tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (ops *operations) checkDeletedPVC(pvcName string) bool {
	_, err := ops.pvcClient.
		Get(pvcName, metav1.GetOptions{})
	Expect(err).Should(HaveOccurred())
	if isNotFound(err) {
		return true
	}
	return false
}

// isNotFound returns true if the original
// cause of error was due to castemplate's
// not found error or kubernetes not found
// error
func isNotFound(err error) bool {
	switch err := errors.Cause(err).(type) {
	case *templatefuncs.NotFoundError:
		return true
	default:
		return k8serrors.IsNotFound(err)
	}
}
