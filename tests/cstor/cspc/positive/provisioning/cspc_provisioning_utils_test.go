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
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cspc_v1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	cspcspecs_v1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/cstorpoolspecs"
	cspcrg_v1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/raidgroups"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

func createCSPCObjectWithResources() {
	createCSPCWithResourceLimits(1, "stripe")
}

func getResourceLimits(cpu, memory string) corev1.ResourceList {
	res := corev1.ResourceList{}
	res[corev1.ResourceCPU] = resource.MustParse(cpu)
	res[corev1.ResourceMemory] = resource.MustParse(memory)

	return res
}

func getDefaultResources(cpu, memory string) *corev1.ResourceRequirements {
	Resources := &corev1.ResourceRequirements{
		Limits: getResourceLimits(cpu, memory),
	}
	return Resources
}

func getAuxResources(cpu, memory string) corev1.ResourceRequirements {
	Resources := corev1.ResourceRequirements{
		Limits: getResourceLimits(cpu, memory),
	}
	return Resources
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

func createCSPCWithResourceLimits(blockDeviceCount int, poolType string) {
	var err error
	cspcObj, err = cspc_v1alpha1.NewBuilder().
		WithGenerateName(cspcName).
		WithNamespace(ops.NameSpace).
		WithDefaultResourceRequirement(getDefaultResources("250m", "64Mi")).
		WithAuxResourceRequirement(getAuxResources("500m", "128Mi")).
		WithPoolSpecBuilder(cspcspecs_v1alpha1.NewBuilder().
			WithNodeSelector(NodeList.Items[0].Labels).
			WithRaidGroupBuilder(
				cspcrg_v1alpha1.NewBuilder().
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

// This function is local to this package
func getLabelSelector(cspc *apis.CStorPoolCluster) string {
	return string(apis.CStorPoolClusterCPK) + "=" + cspc.Name
}
