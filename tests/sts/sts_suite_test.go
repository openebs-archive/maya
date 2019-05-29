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

package sts

import (
	"flag"
	"os"

	"testing"

	"github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	"github.com/openebs/maya/tests/kubernetes"
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
	// neede to run this test
	minNodeCount int = 3
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StatefulSet")
}

func init() {
	flag.Parse()
}

var _ = BeforeSuite(func() {
	// Fetching the kube config path
	configPath, err := kubernetes.GetConfigPath()
	Expect(err).ShouldNot(HaveOccurred())

	// Setting the path in environemnt variable
	err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	Expect(err).ShouldNot(HaveOccurred())
	// Getting clientset
	cl, err := kubernetes.GetClientSet()
	Expect(err).ShouldNot(HaveOccurred())

	// Checking appropriate node numbers. This test is designed to run on a 3 node cluster
	nodes, err := cl.CoreV1().Nodes().List(v1.ListOptions{})
	Expect(nodes.Items).Should(HaveLen(minNodeCount))

	// Fetching the openebs component artifacts
	artifactsOpenEBS, errs := artifacts.GetArtifactsListUnstructuredFromFile(artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	// Installing the artifacts to kubernetes cluster
	for _, artifact := range artifactsOpenEBS {
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(artifact),
			artifact.GetNamespace(),
		)
		_, err := cu.Apply(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}

	// Check for maya-apiserver pod to get created and running
	Eventually(func() int {
		pods, err := pod.
			NewKubeClient().
			WithNamespace(string(artifacts.OpenebsNamespace)).
			List(metav1.ListOptions{LabelSelector: string(artifacts.MayaAPIServerLabelSelector)})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.ListBuilderForAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(1), "Maya-APIServer pod count should be 1")

	// Check for provisioner pod to get created and running
	Eventually(func() int {
		pods, err := pod.
			NewKubeClient().
			WithNamespace(string(artifacts.OpenebsNamespace)).
			List(metav1.ListOptions{LabelSelector: string(artifacts.OpenEBSProvisionerLabelSelector)})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.ListBuilderForAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(1), "OpenEBS provisioner pod count should be 1")

	// Check for snapshot operator to get created and running
	Eventually(func() int {
		pods, err := pod.
			NewKubeClient().
			WithNamespace(string(artifacts.OpenebsNamespace)).
			List(metav1.ListOptions{LabelSelector: string(artifacts.OpenEBSSnapshotOperatorLabelSelector)})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.ListBuilderForAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(1), "OpenEBS snapshot pod count should be 1")

	// Check for admission server to get created and running
	Eventually(func() int {
		pods, err := pod.
			NewKubeClient().
			WithNamespace(string(artifacts.OpenebsNamespace)).
			List(metav1.ListOptions{LabelSelector: string(artifacts.OpenEBSAdmissionServerLabelSelector)})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.ListBuilderForAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(1), "OpenEBS admission server pod count should be 1")

	// Check for NDM pods to get created and running
	Eventually(func() int {
		pods, err := pod.
			NewKubeClient().
			WithNamespace(string(artifacts.OpenebsNamespace)).
			List(metav1.ListOptions{LabelSelector: string(artifacts.OpenEBSNDMLabelSelector)})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.ListBuilderForAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(minNodeCount), "NDM pod count should be "+string(minNodeCount))

	// Check for cstor storage pool pods to get created and running
	Eventually(func() int {
		pods, err := pod.
			NewKubeClient().
			WithNamespace(string(artifacts.OpenebsNamespace)).
			List(metav1.ListOptions{LabelSelector: string(artifacts.OpenEBSCStorPoolLabelSelector)})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.ListBuilderForAPIList(pods).
			WithFilter(pod.IsRunning()).
			List().
			Len()
	},
		defaultTimeOut, defaultPollingInterval).
		Should(Equal(minNodeCount), "CStor pool pod count should be "+string(minNodeCount))
})

var _ = AfterSuite(func() {
	// Fetching the openebs component artifacts
	artifacts, errs := artifacts.GetArtifactsListUnstructuredFromFile(artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	// Deleting the artifacts to kubernetes cluster
	for _, artifact := range artifacts {
		cu := k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(artifact),
			artifact.GetNamespace(),
		)
		err := cu.Delete(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}

	// Unsetting the environment variable
	err := os.Unsetenv(string(v1alpha1.KubeConfigEnvironmentKey))
	Expect(err).ShouldNot(HaveOccurred())
})
