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

package v1alpha1

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	installer "github.com/openebs/maya/tests/artifacts/installer/v1alpha1"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	node "github.com/openebs/maya/pkg/kubernetes/node/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	snap "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	validatehook "github.com/openebs/maya/pkg/kubernetes/webhook/validate/v1alpha1"
	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// auth plugins
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// DefaultTimeOut is the default time in seconds
	// for Eventually block
	DefaultTimeOut int = 600
	// DefaultPollingInterval is the default polling
	// time in seconds for the Eventually block
	DefaultPollingInterval int = 10
	// NodeCountOne is the 1 node
	// needed to run this test
	NodeCountOne int = 1
	// maxRetry is max retry count attempt
	maxRetry = 30
)

var (
	defaultOpenebsComponents []*installer.DefaultInstaller
)

// Framework supports common operations used by tests
type Framework struct {
	// Name is the name of test
	Name string
	// MinNodeCount is no of node count required to run
	// any test
	MinNodeCount   int
	Artifacts      artifacts.ArtifactSource
	PodClient      *pod.KubeClient
	SCClient       *sc.Kubeclient
	PVCClient      *pvc.Kubeclient
	NSClient       *ns.Kubeclient
	SnapClient     *snap.Kubeclient
	kubeConfigPath string
}

//func NewFrameworkDefault(baseName string, MinNodeCount int) *Framework {
//	options := FrameworkOptions{
//		MinNodeCount: MinNodeCount,
//	}
//	return NewFramework(baseName, options)
//}

// Options abstracts creating an
// instance of operations
type Options func(*Framework)

// WithKubeConfigPath sets the kubeConfig path
// against operations instance
func WithKubeConfigPath(path string) Options {
	return func(f *Framework) {
		f.kubeConfigPath = path
	}
}

// Default makes a framework using default values and sets
// up a BeforeEach/AfterEach
func Default(baseName string, options ...Framework) *Framework {
	f := Framework{}
	for _, value := range options {
		f = Framework{
			MinNodeCount: value.MinNodeCount,
			Artifacts:    value.Artifacts,
		}
	}

	f.withDefaults()
	return New(baseName, f)
}

// withDefaults sets the default options
// of operations instance
func (f *Framework) withDefaults() {
	if f.NSClient == nil {
		f.NSClient = ns.NewKubeClient(ns.WithKubeConfigPath(f.kubeConfigPath))
	}
	if f.SCClient == nil {
		f.SCClient = sc.NewKubeClient(sc.WithKubeConfigPath(f.kubeConfigPath))
	}
	if f.PodClient == nil {
		f.PodClient = pod.NewKubeClient(pod.WithKubeConfigPath(f.kubeConfigPath))
	}
	if f.PVCClient == nil {
		f.PVCClient = pvc.NewKubeClient(pvc.WithKubeConfigPath(f.kubeConfigPath))
	}
	if f.SnapClient == nil {
		f.SnapClient = snap.NewKubeClient(snap.WithKubeConfigPath(f.kubeConfigPath))
	}
}

// New creates a test framework.
func New(baseName string, options Framework) *Framework {
	f := &options
	BeforeEach(f.BeforeSuite)
	AfterEach(f.AfterSuite)

	return f
}

// BeforeSuite installs openebs control plane
// components
func (f *Framework) BeforeSuite() {

	// Check the running node count
	nodesClient := node.
		NewKubeClient(node.WithKubeConfigPath(f.kubeConfigPath))
	nodes, err := nodesClient.List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	nodeCnt := node.
		NewListBuilder().
		WithAPIList(nodes).
		WithFilter(node.IsReady()).
		List().
		Len()
	Expect(nodeCnt).Should(Equal(f.MinNodeCount), "Running node count should be "+strconv.Itoa(int(f.MinNodeCount)))

	// Fetching the openebs component artifacts
	artifactsOpenEBS, errs := artifacts.GetArtifactsListUnstructuredFromFile(f.Artifacts)
	Expect(errs).Should(HaveLen(0))

	By("Deploying OpenEBS components in openebs namespace")
	for _, artifact := range artifactsOpenEBS {
		openebsBuilder := installer.BuilderForObject(artifact)

		defaultInstaller, err := openebsBuilder.Build()
		Expect(err).ShouldNot(HaveOccurred())

		err = defaultInstaller.Install()
		Expect(err).ShouldNot(HaveOccurred())

		/// defaultOpenebsComponents is
		defaultOpenebsComponents = append(defaultOpenebsComponents, defaultInstaller)
		Expect(err).ShouldNot(HaveOccurred())
	}

	By("Verifying 'maya-apiserver' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.MayaAPIServerLabelSelector),
		f.MinNodeCount)

	By("Verifying 'openebs-provisioner' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSProvisionerLabelSelector),
		f.MinNodeCount)

	By("Verifying 'snapshot-operator' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSSnapshotOperatorLabelSelector),
		f.MinNodeCount)

	By("Verifying 'admission-server' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSAdmissionServerLabelSelector),
		f.MinNodeCount)

	By("Verifying 'Node-device-manager' pods status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSNDMLabelSelector),
		f.MinNodeCount)

	By("Verifying 'cstor-pool' pods status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSCStorPoolLabelSelector),
		f.MinNodeCount)

	// check validationWebhookConfiguration API enable in cluster
	_, err = validatehook.KubeClient().List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())

	By("OpenEBS control-plane components are in running state")
}

// AfterSuite clean up openebs control plane components created
// by BeforeSuite
func (f *Framework) AfterSuite() {
	for _, component := range defaultOpenebsComponents {
		err := component.UnInstall()
		Expect(err).ShouldNot(HaveOccurred())
	}

}

// checkComponentStatus checks the status of given component
func checkComponentStatus(namespace, lselector string, Count int) (pods *corev1.PodList) {
	// Verify phase of the pod
	var err error
	Eventually(func() int {
		pods, err = pod.
			NewKubeClient().WithNamespace(namespace).
			List(metav1.ListOptions{LabelSelector: lselector})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.
			ListBuilderForAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		DefaultTimeOut, DefaultPollingInterval).
		Should(Equal(Count), "Pod count should be "+string(Count))
	return
}

// GetPodRunningCountEventually gives the number of pods running eventually
func (f *Framework) GetPodRunningCountEventually(namespace, lselector string, expectedPodCount int) int {
	var podCount int
	for i := 0; i < maxRetry; i++ {
		podCount = f.GetPodRunningCount(namespace, lselector)
		if podCount == expectedPodCount {
			return podCount
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}

// GetPodRunningCount gives number of pods running currently
func (f *Framework) GetPodRunningCount(namespace, lselector string) int {
	pods, err := f.PodClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		WithFilter(pod.IsRunning()).
		List().
		Len()
}

// IsPVCBound checks if the pvc is bound or not
func (f *Framework) IsPVCBound(pvcName string) bool {
	volume, err := f.PVCClient.
		Get(pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return pvc.NewForAPIObject(volume).IsBound()
}

// GetSnapshotTypeEventually returns type of snapshot eventually
func (f *Framework) GetSnapshotTypeEventually(snapName string) string {
	var snaptype string
	for i := 0; i < maxRetry; i++ {
		snaptype = f.GetSnapshotType(snapName)
		if snaptype == "Ready" {
			return snaptype
		}
		time.Sleep(5 * time.Second)
	}
	return snaptype
}

// GetSnapshotType returns type of snapshot currently
func (f *Framework) GetSnapshotType(snapName string) string {
	snap, err := f.SnapClient.
		Get(snapName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	if len(snap.Status.Conditions) > 0 {
		return string(snap.Status.Conditions[0].Type)
	}
	return "NotReady"
}

// IsSnapshotDeleted checks if the snapshot is deleted or not
func (f *Framework) IsSnapshotDeleted(snapName string) bool {
	for i := 0; i < maxRetry; i++ {
		_, err := f.SnapClient.
			Get(snapName, metav1.GetOptions{})
		if err != nil {
			return true
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// IsPVCDeleted tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (f *Framework) IsPVCDeleted(pvcName string) bool {
	_, err := f.PVCClient.
		Get(pvcName, metav1.GetOptions{})
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
