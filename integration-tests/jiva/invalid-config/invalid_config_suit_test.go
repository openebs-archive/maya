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
	"strconv"
	"time"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	installer "github.com/openebs/maya/integration-tests/artifacts/installer/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	node "github.com/openebs/maya/pkg/kubernetes/node/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	MaxRetry                              = 30
	minNodeCount int                      = 1
	parentDir    artifacts.ArtifactSource = "../"
)

type clientOps struct {
	podClient *pod.KubeClient
	scClient  *sc.Kubeclient
	pvcClient *pvc.Kubeclient
	nsClient  *ns.Kubeclient
}

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeConfigPath", "", "Based on arguments test will be triggered on corresponding cluster")
}

var _ = BeforeSuite(func() {
	var errs []error

	// Check the running node count
	nodesClient := node.
		NewKubeClient(node.WithKubeConfigPath(kubeConfigPath))
	nodes, err := nodesClient.List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	nodeCnt := node.
		NewListBuilder().
		WithAPIList(nodes).
		WithFilter(node.IsReady()).
		List().
		Len()
	Expect(nodeCnt).Should(Equal(minNodeCount), "Running node count should be "+strconv.Itoa(int(minNodeCount)))

	// Fetch openebs component artifacts
	artifactsOpenEBS, errs := artifacts.GetArtifactsListUnstructuredFromFile(parentDir + artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	By("Installing OpenEBS components")
	// Installing the artifacts to kubernetes cluster
	for _, artifact := range artifactsOpenEBS {
		buildOpenebsComponents := installer.BuilderForObject(artifact)
		oebsComponentInstaller, err := buildOpenebsComponents.Build()
		Expect(err).ShouldNot(HaveOccurred())
		err = oebsComponentInstaller.Install()
		Expect(err).ShouldNot(HaveOccurred())
		defaultoebsComponents = append(defaultoebsComponents, oebsComponentInstaller)
	}

	// get pod client set
	cOps = cOps.newPodClient("openebs")

	// Check for maya-apiserver pod to get created and running
	podCount := cOps.isRunningPodCount(string(artifacts.OpenebsNamespace), string(artifacts.MayaAPIServerLabelSelector), 1)
	Expect(podCount).To(Equal(1))

	// Check for provisioner pod to get created and running
	podCount = cOps.isRunningPodCount(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSProvisionerLabelSelector), 1)
	Expect(podCount).To(Equal(1))

	// Check for snapshot operator to get created and running
	podCount = cOps.isRunningPodCount(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSSnapshotOperatorLabelSelector), 1)
	Expect(podCount).To(Equal(1))

	// Check for admission server to get created and running
	podCount = cOps.isRunningPodCount(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSAdmissionServerLabelSelector), 1)
	Expect(podCount).To(Equal(1))

	// Check for NDM pods to get created and running
	podCount = cOps.isRunningPodCount(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSNDMLabelSelector), minNodeCount)
	Expect(podCount).To(Equal(minNodeCount))

	// Check for cstor storage pool pods to get created and running
	_ = cOps.isRunningPodCount(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSCStorPoolLabelSelector), 1)
	Expect(podCount).To(Equal(1))

	By("OpenEBS components are in running state")
})

var _ = AfterSuite(func() {
	By("Uinstalling OpenEBS Components")
	for _, oebsComponent := range defaultoebsComponents {
		err := oebsComponent.UnInstall()
		Expect(err).ShouldNot(HaveOccurred())
	}
})

var cOps = &clientOps{}

var (
	defaultoebsComponents []*installer.DefaultInstaller
	kubeConfigPath        string
)

func (cOps *clientOps) newPodClient(namespace string) *clientOps {
	newPodClient := pod.NewKubeClient(pod.WithKubeConfigPath(kubeConfigPath), pod.WithNamespace(namespace))
	cOps.podClient = newPodClient
	return cOps
}

func (cOps *clientOps) newNsClient() *clientOps {
	newNsClient := ns.NewKubeClient(ns.WithKubeConfigPath(kubeConfigPath))
	cOps.nsClient = newNsClient
	return cOps
}

func (cOps *clientOps) newSCClient() *clientOps {
	newSCClient := sc.NewKubeClient(sc.WithKubeConfigPath(kubeConfigPath))
	cOps.scClient = newSCClient
	return cOps
}

func (cOps *clientOps) newPVCClient(namespace string) *clientOps {
	newPVCClient := pvc.NewKubeClient(pvc.WithKubeConfigPath(kubeConfigPath), pvc.WithNamespace(namespace))
	cOps.pvcClient = newPVCClient
	return cOps
}

func (cOps *clientOps) isRunningPodCount(namespace, lselector string, expectedPodCount int) int {
	var maxRetry int
	var podCount int
	maxRetry = MaxRetry
	for i := 0; i < maxRetry; i++ {
		podCount = cOps.getRunningPodCount(namespace, lselector)
		if podCount == expectedPodCount {
			return podCount
		}
		if maxRetry == 0 {
			break
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}

func (cOps *clientOps) getRunningPodCount(namespace, lselector string) int {
	pods, err := cOps.podClient.
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		WithFilter(pod.IsRunning()).
		List().
		Len()
}
