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

package invalidconfig

import (
	"flag"
	"time"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	MaxRetry         = 30
	minNodeCount int = 1
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
	RunSpecs(t, "Test openebs by applying invalid configuration in sc and pvc")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
}

var _ = BeforeSuite(func() {
	var err error

	// set pod client set
	for _, f := range clientBuilderFuncList {
		f()
	}

	By("Waiting for maya-apiserver pod to come into running state")
	podCount := ops.isRunningPodCount(string(artifacts.OpenebsNamespace), string(artifacts.MayaAPIServerLabelSelector), 1)
	Expect(podCount).To(Equal(1))

	By("Waiting for openebs-provisioner pod to come into running state")
	podCount = ops.isRunningPodCount(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSProvisionerLabelSelector), 1)
	Expect(podCount).To(Equal(1))
})

var _ = AfterSuite(func() {
})

var ops = &operations{}

type clientBuilderFunc func() *operations

var clientBuilderFuncList = []clientBuilderFunc{
	ops.newPodClient(),
	ops.newSCClient(),
	ops.newNsClient(),
	ops.newPVCClient(),
}

func (ops *operations) newPodClient() *operations {
	newPodClient := pod.NewKubeClient(pod.WithKubeConfigPath(kubeConfigPath))
	ops.podClient = newPodClient
	return ops
}

func (ops *operations) newNsClient() *operations {
	newNsClient := ns.NewKubeClient(ns.WithKubeConfigPath(kubeConfigPath))
	ops.nsClient = newNsClient
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
