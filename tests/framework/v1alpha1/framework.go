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
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	installer "github.com/openebs/maya/tests/artifacts/installer/v1alpha1"
	"github.com/openebs/maya/tests/kubernetes"

	validatehook "github.com/openebs/maya/pkg/kubernetes/webhook/validate/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
)

var (
	defaultOpenebsComponents []*installer.DefaultInstaller
)

// Framework supports common operations used by tests
type Framework struct {
	BaseName string
	// configuration for framework's options
	Options FrameworkOptions
}

// FrameworkOptions provide different options specfic to tests
type FrameworkOptions struct {
	MinNodeCount int
	Artifacts    artifacts.ArtifactSource
}

//func NewFrameworkDefault(baseName string, MinNodeCount int) *Framework {
//	options := FrameworkOptions{
//		MinNodeCount: MinNodeCount,
//	}
//	return NewFramework(baseName, options)
//}

// NewFrameworkDefault makes a framework using default values and sets
// up a BeforeEach/AfterEach
func NewFrameworkDefault(baseName string, options ...FrameworkOptions) *Framework {
	Options := FrameworkOptions{}
	for _, value := range options {
		Options = FrameworkOptions{
			MinNodeCount: value.MinNodeCount,
			Artifacts:    value.Artifacts,
		}
	}
	return NewFramework(baseName, Options)
}

// NewFramework creates a test framework.
func NewFramework(baseName string, options FrameworkOptions) *Framework {
	f := &Framework{
		BaseName: baseName,
		Options:  options,
	}

	BeforeEach(f.BeforeSuite)
	AfterEach(f.AfterSuite)

	return f
}

// BeforeSuite installs openebs control plane
// components
func (f *Framework) BeforeSuite() {
	// Fetching the kube config path
	configPath, err := kubernetes.GetConfigPath()
	Expect(err).ShouldNot(HaveOccurred())

	// Setting the path in environemnt variable
	err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	Expect(err).ShouldNot(HaveOccurred())
	// Getting clientset
	cl, err := kubernetes.GetClientSet()
	Expect(err).ShouldNot(HaveOccurred())

	// Checking appropriate node numbers
	nodes, _ := cl.CoreV1().Nodes().List(v1.ListOptions{})
	Expect(nodes.Items).Should(HaveLen(f.Options.MinNodeCount))

	// Fetching the openebs component artifacts
	artifactsOpenEBS, errs := artifacts.GetArtifactsListUnstructuredFromFile(f.Options.Artifacts)
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
		f.Options.MinNodeCount)

	By("Verifying 'openebs-provisioner' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSProvisionerLabelSelector),
		f.Options.MinNodeCount)

	By("Verifying 'snapshot-operator' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSSnapshotOperatorLabelSelector),
		f.Options.MinNodeCount)

	By("Verifying 'admission-server' pod status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSAdmissionServerLabelSelector),
		f.Options.MinNodeCount)

	By("Verifying 'Node-device-manager' pods status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSNDMLabelSelector),
		f.Options.MinNodeCount)

	By("Verifying 'cstor-pool' pods status as running")
	_ = checkComponentStatus(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSCStorPoolLabelSelector),
		f.Options.MinNodeCount)

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
			NewKubeClient().
			WithNamespace(namespace).
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
