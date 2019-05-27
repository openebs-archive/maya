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

package jiva

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	jobLable      = "job-name=jiva-volume-upgrade"
	urLable       = "upgradejob.openebs.io/name=jiva-volume-upgrade,upgradeitem.openebs.io/name="
	upgradedLabel = "openebs.io/persistent-volume-claim=jiva-volume-claim,openebs.io/version=0.9.0"
	data          map[string]string
)

var _ = Describe("[jiva] TEST VOLUME UPGRADE", func() {

	When("jiva pvc is upgraded", func() {
		It("should create new controller and replica pods of version 0.9.0", func() {

			// fetching name of pv to update the configmap with the resource to
			// be upgraded
			pvName := ops.GetPVNameFromPVCName(pvcName)
			data = make(map[string]string)
			data["upgrade"] = "casTemplate: jiva-volume-update-0.8.2-0.9.0\nresources:\n- name: " + pvName + "\n  kind: jiva-volume\n  namespace: default\n"
			urLable = urLable + pvName

			By("applying rbac.yaml")
			applyYAML(rbacYAML, "")

			By("applying cr.yaml")
			applyYAML(crYAML, "")

			By("applying jiva_upgrade_runtask.yaml")
			applyYAML(runtaskYAML, "")

			By("applying volume-upgrade-job.yaml")
			applyYAML(jobYAML, "job")

			By("verifying completed pod count as 1")
			completedPodCount := ops.GetPodCompletedCountEventually("default", jobLable, 1)
			Expect(completedPodCount).To(Equal(1), "while checking complete pod count")

			By("verifying upgraderesult")
			status := ops.VerifyUpgradeResultTasksIsSuccess(nsName, urLable)
			Expect(status).To(Equal(true), "while checking upgraderesult")

			By("verifying controller pod count")
			controllerPodCount := ops.GetPodRunningCountEventually("default", ctrlLabel, replicaCount)
			Expect(controllerPodCount).To(Equal(replicaCount), "while checking controller pod count")

			By("verifying replica pod count")
			replicaPodCount := ops.GetPodRunningCountEventually("default", replicaLabel, replicaCount)
			Expect(replicaPodCount).To(Equal(replicaCount), "while checking replica pod count")

			By("verifying pod version as 0.9.0")
			podCount := ops.GetPodRunningCountEventually("default", upgradedLabel, replicaCount+1)
			Expect(podCount).To(Equal(replicaCount+1), "while checking pod version")

		})
	})

})
