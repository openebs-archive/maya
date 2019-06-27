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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha2"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// stsYaml holds the yaml spec
	// for statefulset application
	stsYaml artifacts.Artifact = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: busybox1
  namespace: default
  labels:
    app: busybox1
spec:
  serviceName: busybox1
  replicas: 3
  selector:
    matchLabels:
      app: busybox1
      openebs.io/replica-anti-affinity: busybox1
  template:
    metadata:
      labels:
        app: busybox1
        openebs.io/replica-anti-affinity: busybox1
    spec:
      containers:
      - name: busybox1
        image: ubuntu
        imagePullPolicy: IfNotPresent
        command:
          - sleep
          - infinity
        volumeMounts:
        - name: busybox1
          mountPath: /busybox1
  volumeClaimTemplates:
  - metadata:
      name: busybox1
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: cstor-sts
      resources:
        requests:
          storage: 1Gi`

	// stsSCYaml holds the yaml spe for
	// storageclass required by the statefulset application
	stsSCYaml artifacts.Artifact = `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: cstor-sts
  annotations:
    openebs.io/cas-type: cstor
    cas.openebs.io/config: |
      - name: ReplicaCount
        value: "1"
      - name: StoragePoolClaim
        value: "cstor-sparse-pool"
provisioner: openebs.io/provisioner-iscsi
`
)

var _ = Describe("StatefulSet", func() {
	BeforeEach(func() {
		// Extracting storageclass artifacts unstructured
		SCUnstructured, err := artifacts.GetArtifactUnstructured(
			artifacts.Artifact(stsSCYaml),
		)
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting statefulset artifacts unstructured
		STSUnstructured, err := artifacts.GetArtifactUnstructured(stsYaml)
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting statefulset application namespace
		stsNamespace := STSUnstructured.GetNamespace()

		// Generating label selector for stsResources
		stsApplicationLabel := "app=" + STSUnstructured.GetName()
		replicaAntiAffinityLabel := "openebs.io/replica-anti-affinity=" + STSUnstructured.GetName()

		// Apply sts storageclass
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(SCUnstructured),
			SCUnstructured.GetNamespace(),
		)
		_, err = cu.Apply(SCUnstructured)
		Expect(err).ShouldNot(HaveOccurred())

		// Apply the sts
		cu = k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(STSUnstructured),
			stsNamespace,
		)
		_, err = cu.Apply(STSUnstructured)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify creation of sts instances

		// Check for pvc to get created and bound
		Eventually(func() int {
			pvcs, err := pvc.
				NewKubeClient().
				WithNamespace(stsNamespace).
				List(metav1.ListOptions{LabelSelector: stsApplicationLabel})
			Expect(err).ShouldNot(HaveOccurred())
			pvcCount, err := pvc.
				ListBuilderForAPIObjects(pvcs).
				WithFilter(pvc.IsBound()).
				Len()
			Expect(err).ShouldNot(HaveOccurred())
			return pvcCount
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(3), "PVC count should be "+string(3))

		// Check for CVR to get healthy
		Eventually(func() int {
			cvrs, err := cvr.
				NewKubeclient(cvr.WithNamespace("")).
				List(metav1.ListOptions{LabelSelector: replicaAntiAffinityLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return cvr.
				NewListBuilder().
				WithAPIList(cvrs).
				WithFilter(cvr.IsHealthy()).
				List().
				Len()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(3), "CVR count should be "+string(3))

		// Check for statefulset pods to get created and running
		Eventually(func() int {
			pods, err := pod.
				NewKubeClient().
				WithNamespace(stsNamespace).
				List(metav1.ListOptions{LabelSelector: stsApplicationLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return pod.
				ListBuilderForAPIList(pods).
				WithFilter(pod.IsRunning()).
				List().
				Len()
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(3), "Pod count should be "+string(3))
	})

	AfterEach(func() {
		// Extracting storageclass artifacts unstructured
		SCUnstructured, err := artifacts.GetArtifactUnstructured(artifacts.Artifact(stsSCYaml))
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting statefulset artifacts unstructured
		STSUnstructured, err := artifacts.GetArtifactUnstructured(stsYaml)
		Expect(err).ShouldNot(HaveOccurred())

		// Extracting statefulset application namespace
		stsNamespace := STSUnstructured.GetNamespace()

		// Generating label selector for stsResources
		stsApplicationLabel := "app=" + STSUnstructured.GetName()

		// Fetch PVCs to be deleted
		pvcs, err := pvc.NewKubeClient().
			WithNamespace(stsNamespace).
			List(metav1.ListOptions{LabelSelector: stsApplicationLabel})
		Expect(err).ShouldNot(HaveOccurred())
		// Delete PVCs
		for _, p := range pvcs.Items {
			err = pvc.NewKubeClient().
				WithNamespace(stsNamespace).
				Delete(p.GetName(), &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred())
		}

		// Delete the sts artifacts
		cu := k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(STSUnstructured),
			stsNamespace,
		)
		err = cu.Delete(STSUnstructured)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify deletion of sts instances
		Eventually(func() int {
			pods, err := pod.
				NewKubeClient().
				WithNamespace(stsNamespace).
				List(metav1.ListOptions{LabelSelector: stsApplicationLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return len(pods.Items)
		}, defaultTimeOut, defaultPollingInterval).
			Should(Equal(0), "pod count should be 0")

		// Verify deletion of pvc instances
		Eventually(func() int {
			pvcs, err := pvc.
				NewKubeClient().
				WithNamespace(stsNamespace).
				List(metav1.ListOptions{LabelSelector: stsApplicationLabel})
			Expect(err).ShouldNot(HaveOccurred())
			return len(pvcs.Items)
		},
			defaultTimeOut, defaultPollingInterval).
			Should(Equal(0), "pvc count should be 0")

		// Delete storageclass
		cu = k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(SCUnstructured),
			SCUnstructured.GetNamespace(),
		)
		err = cu.Delete(SCUnstructured)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("test statefulset application on cstor", func() {
		It("should distribute the cstor volume replicas across pools", func() {
			// Extracting statefulset artifacts unstructured

			STSUnstructured, err := artifacts.GetArtifactUnstructured(stsYaml)
			Expect(err).ShouldNot(HaveOccurred())

			// Extracting statefulset application namespace
			stsNamespace := STSUnstructured.GetNamespace()

			// Generating label selector for stsResources
			stsApplicationLabel := "app=" + STSUnstructured.GetName()
			replicaAntiAffinityLabel := "openebs.io/replica-anti-affinity=" + STSUnstructured.GetName()

			pvcs, err := pvc.
				NewKubeClient().
				WithNamespace(stsNamespace).
				List(metav1.ListOptions{LabelSelector: stsApplicationLabel})
			Expect(err).ShouldNot(HaveOccurred())
			pvcList, err := pvc.
				ListBuilderForAPIObjects(pvcs).
				WithFilter(pvc.ContainsName(STSUnstructured.GetName())).
				List()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pvcList.Len()).Should(Equal(3), "pvc count should be "+string(3))

			cvrs, err := cvr.
				NewKubeclient(cvr.WithNamespace("")).
				List(metav1.ListOptions{LabelSelector: replicaAntiAffinityLabel})
			Expect(cvrs.Items).Should(HaveLen(3), "cvr count should be "+string(3))

			poolNames := cvr.
				NewListBuilder().
				WithAPIList(cvrs).
				List()
			Expect(poolNames.GetUniquePoolNames()).
				Should(HaveLen(3), "pool names count should be "+string(3))

			pools, err := csp.NewKubeClient().List(metav1.ListOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			nodeNames := csp.ListBuilder().WithAPIList(pools).List()
			Expect(nodeNames.GetPoolUIDs()).
				Should(HaveLen(3), "node names count should be "+string(3))
		})

		PIt("should co-locate the cstor volume targets with application instances", func() {
			// future
		})
	})
})
