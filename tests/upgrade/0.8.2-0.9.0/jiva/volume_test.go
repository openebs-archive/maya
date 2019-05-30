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
	"encoding/json"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	jiva "github.com/openebs/maya/pkg/client/jiva"
	"github.com/openebs/maya/tests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	jobLable      = "job-name=jiva-volume-upgrade"
	urLable       = "upgradejob.openebs.io/name=jiva-volume-upgrade,upgradeitem.openebs.io/name="
	upgradedLabel = "openebs.io/persistent-volume-claim=jiva-volume-claim,openebs.io/version=0.9.0"
	data          map[string]string
	nodes         []string
)

var _ = Describe("[jiva] TEST VOLUME UPGRADE", func() {

	When("jiva pvc is upgraded", func() {
		It("should create new controller and replica pods of version 0.9.0", func() {

			// fetching replica pods before upgrade
			podList, err := ops.PodClient.List(metav1.ListOptions{LabelSelector: replicaLabel})
			Expect(err).ShouldNot(HaveOccurred(), "while listing replica pod")
			for _, pod := range podList.Items {
				nodes = append(nodes, pod.Spec.NodeName)
			}

			// fetching name of pv to update the configmap with the resource to
			// be upgraded
			pvName := ops.GetPVNameFromPVCName(pvcName)
			data = make(map[string]string)
			data["upgrade"] = "casTemplate: jiva-volume-update-0.8.2-0.9.0\nresources:\n- name: " + pvName + "\n  kind: jiva-volume\n  namespace: default\n"
			urLable = urLable + pvName

			By("applying rbac.yaml")
			applyFromURL(rbacURL)

			By("applying cr.yaml")
			applyFromURL(crURL)

			By("applying jiva_upgrade_runtask.yaml")
			applyFromURL(runtaskURL)

			By("applying volume-upgrade-job.yaml")
			applyFromURL(jobURL)

			By("verifying completed pod count as 1")
			completedPodCount := ops.GetPodCompletedCountEventually("default", jobLable, 1)
			Expect(completedPodCount).To(Equal(1), "while checking complete pod count")

			By("verifying upgraderesult")
			status := ops.VerifyUpgradeResultTasksIsNotFail(nsName, urLable)
			Expect(status).To(Equal(true), "while checking upgraderesult")

			By("verifying controller pod count")
			controllerPodCount := ops.GetPodRunningCountEventually("default", ctrlLabel, 1)
			Expect(controllerPodCount).To(Equal(1), "while checking controller pod count")

			By("verifying replica pod count")
			replicaPodCount := ops.GetPodRunningCountEventually("default", replicaLabel, replicaCount)
			Expect(replicaPodCount).To(Equal(replicaCount), "while checking replica pod count")

			By("verifying pod version as 0.9.0")
			podCount := ops.GetPodRunningCountEventually("default", upgradedLabel, replicaCount+1)
			Expect(podCount).To(Equal(replicaCount+1), "while checking pod version")

			By("verifying node stickiness after upgrade")
			// fetching replica pods after upgrade
			podList, err = ops.PodClient.List(metav1.ListOptions{LabelSelector: replicaLabel})
			Expect(err).ShouldNot(HaveOccurred(), "while listing replica pod")
			for _, pod := range podList.Items {
				Expect(nodes).To(
					ContainElement(pod.Spec.NodeName),
					"while verifying node stickness of replicas",
				)
			}

			By("verifying registered replica count and replication factor")
			podList, err = ops.PodClient.List(metav1.ListOptions{LabelSelector: ctrlLabel})
			Expect(err).ShouldNot(HaveOccurred(), "while listing controller pod")

			replicationFactor := ""
			for _, env := range podList.Items[0].Spec.Containers[0].Env {
				if env.Name == "REPLICATION_FACTOR" {
					replicationFactor = env.Value
				}
			}
			Expect(replicationFactor).ToNot(Equal(""), "while fetching replication factor")

			curl := "curl http://localhost:9501/v1/volumes"
			cmd := []string{"/bin/bash", "-c", curl}
			opts := tests.NewOptions().
				WithPodName(podList.Items[0].Name).
				WithNamespace(podList.Items[0].Namespace).
				WithContainer(podList.Items[0].Spec.Containers[0].Name).
				WithCommand(cmd...)

			out, err := ops.ExecPod(opts)
			Expect(err).To(BeNil(), "while executing command in container ", cmd)

			volumes := jiva.VolumeCollection{}
			err = json.Unmarshal(out, &volumes)
			Expect(err).To(BeNil(), "while unmarshalling volumes", string(out))

			registeredReplicaCount := strconv.Itoa(volumes.Data[0].ReplicaCount)
			Expect(registeredReplicaCount).To(
				Equal(replicationFactor),
				"while verifying registered replica count as replication factor",
			)

		})
	})

})
