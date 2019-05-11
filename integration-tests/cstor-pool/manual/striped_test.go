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

package cstorpoolitmanul

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstorpool/v1alpha3"
	cspc "github.com/openebs/maya/pkg/cstorpoolcluster/v1alpha1"
	"time"

	//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"
)

const (
	KubeConfigPath = "/home/ashutosh/.kube/config"
	MaxRetry       = 30
)

// operations encapsulates various kubernetes object client set for operations.
type operations struct {
	cspClient  *csp.Kubeclient
	cspcClient *cspc.Kubeclient
}

// TestIntegrationCstorPool function instantiate the cstor pool test suite.
func TestIntegrationCstorPool(t *testing.T) {
	// RegisterFailHandler is used to register failed test cases and produce readable output.
	RegisterFailHandler(Fail)
	// RunSpecs runs all the test cases in the suite.
	RunSpecs(t, "Cstor pool integration test suite")
}

var _ = Describe("STRIPED SPARSE CSPC With 1 Disk attached to a single node", func() {
	var (
		cspcObj *apis.CStorPoolCluster
	)
	BeforeEach(func() {
		nodeName:="gke-cstor-it-default-pool-569eb31d-18cd"
		cspcObj = cspc.NewBuilder().
			WithName("sparse-striped-manual").
			WithDiskType("sparse").
			WithNode(nodeName).
			WithNodeStripedType(nodeName).
			WithNodeGroup(nodeName,apis.CStorPoolClusterDiskGroups{
				Name:"group-0",
				Disks:[]apis.CStorPoolClusterDisk{
					{"sparse-0ed3e3be315d5d33bbb7721bd56f9ec2",""},
				},
		}).
			Build().Object
	})

	When("We create a sparse-striped-manual cspc  with 1 disk attached to a single node ", func() {
		It("pool resource count should be 1 with no error and healthy status", func() {
			// Create a cstor pool cluster
			_, err := ops.cspcClient.Create(cspcObj)
			Expect(err).To(BeNil())
			cspCount := ops.isHealthyCspCount(cspcObj.Name, 1)
			Expect(cspCount).To(Equal(1))

			onDiffNode:=ops.isCSPOnDifferentNodes(cspcObj.Name)
			Expect(onDiffNode).To(Equal(true))
		})
	})

	AfterEach(func() {
		cspcClient := cspc.NewKubeClient(cspc.WithKubeConfigPath(KubeConfigPath))
		cspcList, err := cspcClient.List(metav1.ListOptions{})
		Expect(err).To(BeNil())
		for _, cspc := range cspcList.Items {
			_, err := cspcClient.Delete(cspc.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		}
	})

})

var _ = Describe("STRIPED SPARSE CSPC With 2 Disk attached to a single node", func() {
	var (
		cspcObj *apis.CStorPoolCluster
	)
	BeforeEach(func() {
		nodeName:="gke-cstor-it-default-pool-569eb31d-18cd"
		cspcObj = cspc.NewBuilder().
			WithName("sparse-striped-manual").
			WithDiskType("sparse").
			WithNode(nodeName).
			WithNodeStripedType(nodeName).
			WithNodeGroup(nodeName,apis.CStorPoolClusterDiskGroups{
				Name:"group-0",
				Disks:[]apis.CStorPoolClusterDisk{
					{"sparse-0ed3e3be315d5d33bbb7721bd56f9ec2",""},
					{"sparse-51558cb2272cf9283285256302f68dc8",""},
				},
			}).
			Build().Object
	})

	When("We create a sparse-striped-manual cspc  with 2 disk attached to a single node ", func() {
		It("pool resource count should be 2 with no error and healthy status", func() {
			// Create a cstor pool cluster
			_, err := ops.cspcClient.Create(cspcObj)
			Expect(err).To(BeNil())
			cspCount := ops.isHealthyCspCount(cspcObj.Name, 1)
			Expect(cspCount).To(Equal(1))

			onDiffNode:=ops.isCSPOnDifferentNodes(cspcObj.Name)
			Expect(onDiffNode).To(Equal(true))
		})
	})

	AfterEach(func() {
		cspcClient := cspc.NewKubeClient(cspc.WithKubeConfigPath(KubeConfigPath))
		cspcList, err := cspcClient.List(metav1.ListOptions{})
		Expect(err).To(BeNil())
		for _, cspc := range cspcList.Items {
			_, err := cspcClient.Delete(cspc.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		}
	})

})


var _ = Describe("STRIPED SPARSE CSPC With 3 Disk attached to a single node", func() {
	var (
		cspcObj *apis.CStorPoolCluster
	)
	BeforeEach(func() {
		nodeName:="gke-cstor-it-default-pool-569eb31d-18cd"
		cspcObj = cspc.NewBuilder().
			WithName("sparse-striped-manual").
			WithDiskType("sparse").
			WithNode(nodeName).
			WithNodeStripedType(nodeName).
			WithNodeGroup(nodeName,apis.CStorPoolClusterDiskGroups{
				Name:"group-0",
				Disks:[]apis.CStorPoolClusterDisk{
					{"sparse-0ed3e3be315d5d33bbb7721bd56f9ec2",""},
					{"sparse-51558cb2272cf9283285256302f68dc8",""},
					{"sparse-589b2914a4d8f0daab0a82b235869c7e",""},
				},
			}).
			Build().Object
	})

	When("We create a sparse-striped-manual cspc  with 3 disk attached to a single node ", func() {
		It("pool resource count should be 3 with no error and healthy status", func() {
			// Create a cstor pool cluster
			_, err := ops.cspcClient.Create(cspcObj)
			Expect(err).To(BeNil())
			cspCount := ops.isHealthyCspCount(cspcObj.Name, 1)
			Expect(cspCount).To(Equal(1))

			onDiffNode:=ops.isCSPOnDifferentNodes(cspcObj.Name)
			Expect(onDiffNode).To(Equal(true))
		})
	})

	AfterEach(func() {
		cspcClient := cspc.NewKubeClient(cspc.WithKubeConfigPath(KubeConfigPath))
		cspcList, err := cspcClient.List(metav1.ListOptions{})
		Expect(err).To(BeNil())
		for _, cspc := range cspcList.Items {
			_, err := cspcClient.Delete(cspc.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		}
	})

})



var _ = AfterSuite(func() {
	cspcClient := cspc.NewKubeClient(cspc.WithKubeConfigPath(KubeConfigPath))
	cspcList, err := cspcClient.List(metav1.ListOptions{})
	Expect(err).To(BeNil())
	for _, cspc := range cspcList.Items {
		_, err := cspcClient.Delete(cspc.Name, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())
	}
})

var _ = BeforeSuite(func() {
	for _, f := range clientBuilderFuncList {
		f()
	}
})

// ops is the initialized empty instance of operations type.
var ops = &operations{}

type clientBuilderFunc func() *operations

var clientBuilderFuncList = []clientBuilderFunc{
	ops.newSpcClient,
	ops.newCspClient,
}

func (ops *operations) newCspClient() *operations {
	newCspClient, err := csp.KubeClient().WithFlag(KubeConfigPath)
	Expect(err).To(BeNil())
	ops.cspClient = newCspClient
	return ops
}

func (ops *operations) newSpcClient() *operations {
	newSpcClient := cspc.NewKubeClient(cspc.WithKubeConfigPath(KubeConfigPath))
	ops.cspcClient = newSpcClient
	return ops
}

func (ops *operations) getHealthyCSPCount(cspcName string) int {
	cspAPIList, err := ops.cspClient.List(metav1.ListOptions{})
	Expect(err).To(BeNil())
	count := csp.
		ListBuilderForAPIObject(cspAPIList).
		List().
		Filter(csp.HasLabel(string(apis.CStorPoolClusterCPK), cspcName), csp.IsStatus("Healthy")).Len()
	return count
}

func (ops *operations) deleteCSP(cspcName string, deleteCount int) {
	cspAPIList, err := ops.cspClient.List(metav1.ListOptions{})
	Expect(err).To(BeNil())
	cspList := csp.
		ListBuilderForAPIObject(cspAPIList).
		List().
		Filter(csp.HasLabel(string(apis.CStorPoolClusterCPK), cspcName), csp.IsStatus("Healthy"))
	cspCount := cspList.Len()
	Expect(deleteCount).Should(BeNumerically("<=", cspCount))

	for i := 0; i < deleteCount; i++ {
		_, err := ops.cspClient.Delete(cspList.ObjectList.Items[i].Name, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())

	}
}

func (ops *operations) isCSPOnDifferentNodes(cspcName string) bool {
	cspAPIList, err := ops.cspClient.List(metav1.ListOptions{})
	Expect(err).To(BeNil())
	cspList := csp.
		ListBuilderForAPIObject(cspAPIList).
		List().
		Filter(csp.HasLabel(string(apis.CStorPoolClusterCPK), cspcName))
	cspCountOnNode:= make(map[string]int)
	for _,val:=range cspList.ObjectList.Items{
		val:=val
		cspObj:=csp.BuilderForAPIObject(&val).Csp
		cspCountOnNode[cspObj.GetNodeName()]++
		if cspCountOnNode[cspObj.GetNodeName()]>1{
			return false
		}
	}
	return true
}

func (ops *operations) isHealthyCspCount(cspcName string, expectedCspCount int) int {
	var maxRetry int
	var cspCount int
	maxRetry = MaxRetry
	for i := 0; i < maxRetry; i++ {
		cspCount = ops.getHealthyCSPCount(cspcName)
		if cspCount == expectedCspCount {
			return expectedCspCount
		}
		if maxRetry == 0 {
			break
		}
		maxRetry--
		time.Sleep(5 * time.Second)
	}
	return cspCount
}


