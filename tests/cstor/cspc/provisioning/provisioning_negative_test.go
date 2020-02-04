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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	"github.com/openebs/maya/pkg/debug"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const HostIpPort = "146.148.72.182:30036"

var skipNegativeCaseIfRequired = !skipPositiveCaseIfRequired

var _ = Describe("[CSPC-NEGATIVE] CSTOR STRIPE POOL PROVISIONING AND RECONCILIATION ", func() {

	provisioningAndReconciliationNegativeTest(createCSPCObjectForStripe)
})

func SkipTest(skip bool) {
	if skip {
		Skip("Skipping negative cases")
	}
}

func provisioningAndReconciliationNegativeTest(createCSPCObject func()) {
	When("A CSPC Is Created with CSPC Update error", func() {
		It("No pool resource will be created  ", func() {
			SkipTest(skipNegativeCaseIfRequired)
			// Inject CSPC Error
			injectClient := debug.NewClient(HostIpPort)
			err := injectClient.PostInject(debug.NewErrorInjection().WithCSPCUpdateError(debug.Inject))
			Expect(err).To(BeNil())

			By("Preparing A CSPC Object, No Error Should Occur", createCSPCObject)

			By("Creating A CSPC Object, No CSPIs Should Be Created")

			verifyDesiredCSPICountTo(0)
		})
	})

	When("CSPC update error is removed", func() {
		It("cStor Pools Should be Provisioned ", func() {
			SkipTest(skipNegativeCaseIfRequired)
			// Eject CSPI Error
			injectClient := debug.NewClient(HostIpPort)
			err := injectClient.PostInject(debug.NewErrorInjection().WithCSPCUpdateError(debug.Eject))

			Expect(err).To(BeNil())
			// Wait for few seconds
			time.Sleep(3 * time.Second)
			By("Creating A CSPC Object, Desired Number of CSPIs Should Be Created", verifyDesiredCSPICount)
		})
	})

	Cleanup()

	When("A CSPC Is Created with CSPI create error", func() {
		It("No cStor Pools Should be Provisioned ", func() {
			SkipTest(skipNegativeCaseIfRequired)
			// Inject CSPI Error
			injectClient := debug.NewClient(HostIpPort)
			err := injectClient.PostInject(debug.NewErrorInjection().WithCSPICreateError(debug.Inject))
			Expect(err).To(BeNil())

			By("Preparing A CSPC Object, No Error Should Occur", createCSPCObject)

			By("Creating A CSPC Object, Desired Number of CSPIs Should Be Created")

			verifyDesiredCSPICountTo(0)
		})
	})

	When("CSPI create error is removed", func() {
		It("cStor Pools Should be Provisioned ", func() {
			SkipTest(skipNegativeCaseIfRequired)
			// Eject CSPI Error
			injectClient := debug.NewClient(HostIpPort)
			err := injectClient.PostInject(debug.NewErrorInjection().WithCSPICreateError(debug.Eject))

			Expect(err).To(BeNil())
			// Wait for few seconds
			time.Sleep(3 * time.Second)
			By("Creating A CSPC Object, Desired Number of CSPIs Should Be Created", verifyDesiredCSPICount)
		})
	})

	Cleanup()

	When("A CSPC Is Created with deployment create error", func() {
		It("CSPI should get created but not the corresponding pool deployments", func() {
			SkipTest(skipNegativeCaseIfRequired)
			// Inject CSPI Error
			injectClient := debug.NewClient(HostIpPort)
			err := injectClient.PostInject(debug.NewErrorInjection().WithDeploymentCreateError(debug.Inject))
			Expect(err).To(BeNil())
			time.Sleep(3 * time.Second)
			By("Preparing A CSPC Object, No Error Should Occur", createCSPCObject)

			By("Creating A CSPC Object, Desired Number of CSPI(non-healthy) Resource Should Be Created")

			verifyDesiredCSPIResourceCountTo(3)
		})
	})

	When("Deployment create error is removed", func() {
		It("Pool deployments should come up and corresponding CSPIs should become healthy ", func() {
			SkipTest(skipNegativeCaseIfRequired)
			// Eject CSPI Error
			injectClient := debug.NewClient(HostIpPort)
			err := injectClient.PostInject(debug.NewErrorInjection().WithDeploymentCreateError(debug.Eject))

			Expect(err).To(BeNil())
			By("Creating A CSPC Object, Desired Number of CSPIs Should Be Created", verifyDesiredCSPICount)
		})
	})

	Cleanup()

	When("A CSPC Is Created with 30 % random error injection threshold ", func() {
		It("Pools should get provisioned eventually after all the errors has been ejected", func() {
			SkipTest(skipNegativeCaseIfRequired)
			// Inject CSPI Error
			injectClient := debug.NewClient(HostIpPort)
			err := injectClient.PostInject(debug.NewErrorInjection().
				WithCSPIThreshold(30).
				WithCSPCThreshold(30).
				WithDeploymentThreshold(30))
			Expect(err).To(BeNil())
			time.Sleep(3 * time.Second)
			By("Preparing A CSPC Object, No Error Should Occur", createCSPCObject)
			healthyCSPICount := ops.GetCSPICountWithCSPCName(cspcObj.Name, 3,
				[]cspi.Predicate{cspi.IsStatus("ONLINE")})
			if healthyCSPICount == 3 {
				fmt.Fprint(GinkgoWriter, "Pools got provisioned successfully without requiring errors to be ejected")
			} else {
				err := injectClient.PostInject(debug.NewErrorInjection())
				Expect(err).To(BeNil())
				verifyDesiredCSPICount()
			}

		})
	})

	Cleanup()
}

func Cleanup() {
	When("Cleaning up cspc", func() {
		It("should delete the cspc", func() {
			SkipTest(skipNegativeCaseIfRequired)
			err := ops.CSPCClient.Delete(cspcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
			bdcCount := ops.GetBDCCountEventually(
				metav1.ListOptions{
					LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspcObj.Name},
				0, string(ops.NameSpace))
			Expect(bdcCount).To(BeZero())
			Expect(ops.IsCSPCNotExists(cspcObj.Name)).To(BeTrue())
		})
	})
}
