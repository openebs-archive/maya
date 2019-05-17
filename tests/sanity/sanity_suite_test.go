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

package sanity

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/openebs/maya/pkg/client/k8s/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	"github.com/openebs/maya/tests/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	waitTime time.Duration = 10
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sanity")
}

var namespace string

func init() {
	flag.StringVar(&namespace, "namespace", "openebs", "namespace for performing the test")
	flag.Parse()
}

var _ = BeforeSuite(func() {

	// Fetching the kube config path
	configPath, err := kubernetes.GetConfigPath()
	Expect(err).ShouldNot(HaveOccurred())

	// Setting the path in environemnt variable
	err = os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), configPath)
	Expect(err).ShouldNot(HaveOccurred())

	// Fetching the openebs component artifacts
	artifacts, errs := artifacts.GetArtifactsListUnstructuredFromFile(artifacts.OpenEBSArtifacts)
	Expect(errs).Should(HaveLen(0))

	// Installing the artifacts to kubernetes cluster
	for _, artifact := range artifacts {
		cu := k8s.CreateOrUpdate(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
		_, err := cu.Apply(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}

	// Waiting for pods to be ready
	clientset, err := kubernetes.GetClientSet()
	Expect(err).NotTo(HaveOccurred())

	status := false
	for i := 0; i < 300; i++ {
		pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(pods).NotTo(BeNil())
		expectedStoragePoolPods, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		if kubernetes.CheckPodsRunning(*pods, 4+len(expectedStoragePoolPods.Items)) {
			status = true
			break
		}
		time.Sleep(waitTime * time.Second)
	}
	if !status {
		Fail("Pods were not ready in expected time")
	}
})

var _ = AfterSuite(func() {
	// Fetching the openebs component artifacts
	artifacts, err := artifacts.GetArtifactsListUnstructuredFromFile(artifacts.OpenEBSArtifacts)
	Expect(err).ShouldNot(HaveOccurred())

	// Deleting artifacts
	for _, artifact := range artifacts {
		d := k8s.DeleteResource(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
		err := d.Delete(artifact)
		Expect(err).NotTo(HaveOccurred())
	}

	// Waiting for openebs namespace to get terminated
	clientset, errs := kubernetes.GetClientSet()
	Expect(errs).Should(HaveLen(0))

	status := false
	for i := 0; i < 100; i++ {
		namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(namespaces).NotTo(BeNil())

		if kubernetes.CheckForNamespace(*namespaces, namespace) {
			status = true
			break
		}
		time.Sleep(waitTime * time.Second)
	}
	if !status {
		Fail("Pods were not ready in expected time")
	}
})
