// Copyright © 2018-2019 The OpenEBS Authors
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

package replicadeletion

import (
	"flag"
	"strconv"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/integration-tests/artifacts"
	installer "github.com/openebs/maya/integration-tests/artifacts/installer/v1alpha1"
	node "github.com/openebs/maya/pkg/kubernetes/node/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// defaultTimeOut is the default time in seconds
	// for Eventually block
	defaultTimeOut int = 500
	// defaultPollingInterval is the default polling
	// time in seconds for the Eventually block
	defaultPollingInterval int = 10
	// minNodeCount is the minimum number of nodes
	// need to run this test
	minNodeCount int = 3
	// parentDir is the OpenEBS artifacts source directory
	parentDir artifacts.ArtifactSource = "../../"
)

var (
	// defaultInstallerList holds the list of DefaultInstaller instances
	defaultInstallerList []*installer.DefaultInstaller
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node-Stickiness via pod deleteion")
}

var kubeConfigPath string

func init() {
	flag.StringVar(&kubeConfigPath, "kubeConfigPath", "", "Based on arguments test will be triggered on corresponding cluster")
}

// TODO: Refactor below code based on the framework changes
// getPodList returns the list of running pod object
func getPodList(podKubeClient *pod.KubeClient, namespace, lselector string, podCount int) (pods *corev1.PodList) {
	// Verify phase of the pod
	var err error

	if podKubeClient == nil {
		podKubeClient = pod.NewKubeClient(pod.WithKubeConfigPath(kubeConfigPath)).WithNamespace(namespace)
	}

	Eventually(func() int {
		pods, err = podKubeClient.
			List(metav1.ListOptions{LabelSelector: lselector})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.ListBuilderForAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(podCount), "Pod count should be "+string(podCount))
	return
}

var _ = BeforeSuite(func() {
	// Fetching the kube config path
	//configPath, err := kubernetes.GetConfigPath()
	//Expect(err).ShouldNot(HaveOccurred())

	//// Setting the path in environemnt variable
	//err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	//Expect(err).ShouldNot(HaveOccurred())

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
	openebsartifacts, errs := artifacts.GetArtifactsListUnstructuredFromFile(parentDir + artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	By("Installing OpenEBS components")
	// Installing the artifacts to kubernetes cluster
	for _, artifact := range openebsartifacts {
		defaultInstaller, err := installer.
			BuilderForObject(artifact).
			WithKubeClient(unstruct.WithKubeConfigPath(kubeConfigPath)).
			Build()
		Expect(err).ShouldNot(HaveOccurred())
		//		installerClient := defaultInstaller.NewKubeClient()
		//Expect(err).ShouldNot(HaveOccurred())
		err = defaultInstaller.Install()
		Expect(err).ShouldNot(HaveOccurred())
		defaultInstallerList = append(defaultInstallerList, defaultInstaller)
	}

	podKubeClient := pod.NewKubeClient(pod.WithKubeConfigPath(kubeConfigPath)).WithNamespace(string(artifacts.OpenebsNamespace))
	// Check for maya-apiserver pod to get created and running
	_ = getPodList(podKubeClient, string(artifacts.OpenebsNamespace), string(artifacts.MayaAPIServerLabelSelector), 1)

	// Check for provisioner pod to get created and running
	_ = getPodList(podKubeClient, string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSProvisionerLabelSelector), 1)

	// Check for snapshot operator to get created and running
	_ = getPodList(podKubeClient, string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSSnapshotOperatorLabelSelector), 1)

	// Check for admission server to get created and running
	_ = getPodList(podKubeClient, string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSAdmissionServerLabelSelector), 1)

	// Check for NDM pods to get created and running
	_ = getPodList(podKubeClient, string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSNDMLabelSelector), minNodeCount)

	// Check for cstor storage pool pods to get created and running
	_ = getPodList(podKubeClient, string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSCStorPoolLabelSelector), minNodeCount)

	By("OpenEBS components are in running state")
})

var _ = AfterSuite(func() {
	By("Uinstalling OpenEBS Components and test namespace")
	for _, componentInstaller := range defaultInstallerList {
		err := componentInstaller.UnInstall()
		Expect(err).ShouldNot(HaveOccurred())
	}
})
