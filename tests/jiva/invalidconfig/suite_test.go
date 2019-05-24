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

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	tests "github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/artifacts"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	kubeConfigPath string
	replicaCount   string
	repCountInt    int
	ops            *tests.Operations
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test openebs by applying invalid configuration in sc and pvc")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
	flag.StringVar(&replicaCount, "replicas", "", "No.of storage replicas need to be created")
}

var _ = BeforeSuite(func() {
	var err error

	ops = tests.NewOperations(tests.WithKubeConfigPath(kubeConfigPath))

	repCountInt, err = strconv.Atoi(replicaCount)
	Expect(err).ShouldNot(HaveOccurred(), "while converting replicaCount to integer{%s}", replicaCount)

	By("Waiting for maya-apiserver pod to come into running state")
	podCount := ops.GetPodRunningCountEventually(string(artifacts.OpenebsNamespace), string(artifacts.MayaAPIServerLabelSelector), 1)
	Expect(podCount).To(Equal(1))

	By("Waiting for openebs-provisioner pod to come into running state")
	podCount = ops.GetPodRunningCountEventually(string(artifacts.OpenebsNamespace), string(artifacts.OpenEBSProvisionerLabelSelector), 1)
	Expect(podCount).To(Equal(1))
})

var _ = AfterSuite(func() {
})
