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

package provisioning

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cspc_v1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	cspcspecs_v1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/cstorpoolspecs"
	cspcrg_v1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/raidgroups"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createCSPCObjectForStripe() {
	createCSPCObject(1, "stripe")
}

func createCSPCObjectForMirror() {
	createCSPCObject(2, "mirror")
}

func createCSPCObjectForRaidz() {
	createCSPCObject(3, "raidz")
}

func createCSPCObjectForRaidz2() {
	createCSPCObject(6, "raidz2")
}

func createCSPCObject(blockDeviceCount int, poolType string) {
	var err error
	cspcObj, err = cspc_v1alpha1.NewBuilder().
		WithGenerateName(cspcName).
		WithNamespace(ops.NameSpace).
		WithPoolSpecBuilder(cspcspecs_v1alpha1.NewBuilder().
			WithNodeSelector(NodeList.Items[0].Labels).
			WithRaidGroupBuilder(
				cspcrg_v1alpha1.NewBuilder().
					// TODO : PAss the entire label -- kubernetes.io/hostname
					WithCSPCBlockDeviceList(ops.GetCSPCBDListForNode(&NodeList.Items[0], blockDeviceCount)).
					WithType(poolType),
			),
		).
		WithPoolSpecBuilder(cspcspecs_v1alpha1.NewBuilder().
			WithNodeSelector(NodeList.Items[1].Labels).
			WithRaidGroupBuilder(
				cspcrg_v1alpha1.NewBuilder().
					WithCSPCBlockDeviceList(ops.GetCSPCBDListForNode(&NodeList.Items[1], blockDeviceCount)).
					WithType(poolType),
			),
		).
		WithPoolSpecBuilder(cspcspecs_v1alpha1.NewBuilder().
			WithNodeSelector(NodeList.Items[2].Labels).
			WithRaidGroupBuilder(
				cspcrg_v1alpha1.NewBuilder().
					WithCSPCBlockDeviceList(ops.GetCSPCBDListForNode(&NodeList.Items[2], blockDeviceCount)).
					WithType(poolType),
			),
		).
		GetObj()
	Expect(err).ShouldNot(HaveOccurred())
	cspcObj, err = ops.CSPCClient.WithNamespace(ops.NameSpace).Create(cspcObj)
	Expect(err).To(BeNil())

	Cspc, err = cspc_v1alpha1.BuilderForAPIObject(cspcObj).Build()
	Expect(err).To(BeNil())
}

func verifyDesiredCSPICount() {
	cspiCount := ops.GetHealthyCSPICount(cspcObj.Name, 3)
	Expect(cspiCount).To(Equal(3))

	// Check are there any extra created csps
	cspiCount = ops.GetCSPICount(getLabelSelector(cspcObj))
	Expect(cspiCount).To(Equal(3), "Mismatch Of CSPI Count")
}

func verifyDesiredCSPICountTo(count int) {
	cspiCount := ops.GetHealthyCSPICount(cspcObj.Name, count)
	Expect(cspiCount).To(Equal(count))

	// Check are there any extra created csps
	cspiCount = ops.GetCSPICount(getLabelSelector(cspcObj))
	Expect(cspiCount).To(Equal(count), "Mismatch Of CSPI Count")
}

func verifyDesiredCSPIResourceCountTo(count int) {
	// Check are there any extra created csps
	cspiCount := ops.GetCSPIResourceCountEventually(cspcObj.Name, count)
	Expect(cspiCount).To(Equal(count), "Mismatch Of CSPI Resource Count")
}

// This function is local to this package
func getLabelSelector(cspc *apis.CStorPoolCluster) string {
	return string(apis.CStorPoolClusterCPK) + "=" + cspc.Name
}

func downScaleCSPCObject() {
	// getting the object to avoid update failure
	cspcObj, err := ops.CSPCClient.WithNamespace(cspcObj.Namespace).
		Get(cspcObj.Name, metav1.GetOptions{})
	Expect(err).To(BeNil())
	// downsclaing cspc by 1
	cspcObj.Spec.Pools = cspcObj.Spec.Pools[:2]
	_, err = ops.CSPCClient.WithNamespace(ops.NameSpace).Update(cspcObj)
	Expect(err).To(BeNil())
	cspiCount := ops.GetCSPIResourceCountEventually(getLabelSelector(cspcObj), 2)
	Expect(cspiCount).To(Equal(2))
}

func cleanCSPCObject() {
	When("Cleaning up cspc", func() {
		It("should delete the cspc", func() {
			SkipTest(skipPositiveCaseIfRequired)
			err := ops.CSPCClient.Delete(cspcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
			cspiCount := ops.GetCSPIResourceCountEventually(getLabelSelector(cspcObj), 0)
			Expect(cspiCount).To(BeZero())
			Expect(ops.IsCSPCNotExists(cspcObj.Name)).To(BeTrue())
		})
	})
}
