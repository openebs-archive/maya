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

package exporter

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstorpool/v1alpha3"
	cv "github.com/openebs/maya/pkg/cstorvolume/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests"
	framework "github.com/openebs/maya/tests/framework/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	online, healthy float64 = 1, 1
)

const (
	poolStatus                = "openebs_pool_status"
	replicaStatus             = "openebs_replica_status"
	noPoolAvailableErrorCount = "openebs_zpool_list_no_pool_available_error"
)

var _ = Describe("Test maya-exporter [single-pool-pod]", func() {
	var (
		err error
		pod *corev1.PodList
	)
	BeforeEach(func() {
		When("we are creating pool deployment", func() {
			By("building object of storageclass")
			scObj, err = sc.NewBuilder().
				WithName(scName).
				WithAnnotations(annotations).
				WithProvisioner(openebsProvisioner).Build()
			Expect(err).ShouldNot(HaveOccurred(), "while building storageclass obj for storageclass {%s}", scName)

			By("creating storageclass")
			_, err = ops.SCClient.Create(scObj)
			Expect(err).To(BeNil(), "while creating storageclass", scName)

			By("building spc object")
			spcObj = spc.NewBuilder().
				WithName(spcName).
				WithDiskType(string(apis.TypeBlockDeviceCPV)).
				WithMaxPool(1).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build().Object

			By("creating storagepoolclaim")
			_, err = ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil(), "while creating spc", spcName)

			By("verifying healthy csp count")
			Eventually(func() int {
				cspAPIList, err = ops.CSPClient.List(metav1.ListOptions{})
				Expect(err).To(BeNil())
				count := csp.
					ListBuilderForAPIObject(cspAPIList).
					List().
					Filter(csp.HasLabel(string(apis.StoragePoolClaimCPK), spcName), csp.IsStatus("Healthy")).Len()
				return count
			},
				framework.DefaultTimeOut, framework.DefaultPollingInterval).
				Should(Equal(1), "while getting healthy csp count")

			By("listing cstor pool pods")
			selector := map[string]string{
				string(apis.StoragePoolClaimCPK): spcName,
				"app":                            "cstor-pool",
			}

			ls := labels.Set(selector).
				AsSelector().
				String()
			pod, err = ops.PodClient.List(metav1.ListOptions{
				LabelSelector: ls,
			})

			Expect(err).To(BeNil(), "while listing pool pods with selector ", ls)

			By("verifying pod items")
			Expect(len(pod.Items)).To(Equal(1), "while getting pod items length", pod)

			By("verifying no of containers in pool pod")
			Expect(len(pod.Items[0].Spec.Containers)).To(Equal(3), "while getting no of containers", pod)

			By("verifying whether maya-exporter container exists")
			Expect(pod.Items[0].Spec.Containers[2].Name).To(Equal("maya-exporter"), "while verifying container name", pod)

		})
	})

	AfterEach(func() {
		When("we are deleting resources created for testing maya-exporter", func() {
			By("getting the pvclaim name")
			pvcObj, err = pvc.
				NewKubeClient(pvc.WithKubeConfigPath(kubeConfigPath)).
				WithNamespace(nsName).
				Get(pvcName, metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pvcObj.Spec.VolumeName).ShouldNot(BeEmpty())

			By("deleting pvc")
			err = pvc.
				NewKubeClient(pvc.WithKubeConfigPath(kubeConfigPath)).
				WithNamespace(nsName).
				Delete(pvcName, new(metav1.DeleteOptions))
			Expect(err).ShouldNot(HaveOccurred())

			By("listing pvc to verify if it is deleted")
			Eventually(func() int {
				pvcs, err = pvc.
					NewKubeClient(pvc.WithKubeConfigPath(kubeConfigPath)).
					WithNamespace(nsName).
					List(metav1.ListOptions{LabelSelector: "name=exporter-volume"})
				Expect(err).ShouldNot(HaveOccurred())
				return len(pvcs.Items)
			},
				framework.DefaultTimeOut, framework.DefaultPollingInterval).
				Should(Equal(0), "while listing pvc")

			CstorVolumeLabel := "openebs.io/persistent-volume=" + pvcObj.Spec.VolumeName

			By("verifying if cv is deleted")
			// verify deletion of cstorvolume
			Eventually(func() int {
				cvs, err = cv.
					NewKubeclient(cv.WithNamespace("openebs"), cv.WithKubeConfigPath(kubeConfigPath)).
					List(metav1.ListOptions{LabelSelector: CstorVolumeLabel})
				Expect(err).ShouldNot(HaveOccurred())
				return len(cvs.Items)
			},
				framework.DefaultTimeOut, framework.DefaultPollingInterval).
				Should(Equal(0), "while listing cvs")

			By("deleting storageclass")
			err = ops.SCClient.Delete(scName, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting storageclass", scName)

			By("listing spc")
			spcList, err := ops.SPCClient.List(metav1.ListOptions{})
			Expect(err).To(BeNil(), "while listing spc clients", spcList)

			By("deleting spc")
			for _, spc := range spcList.Items {
				_, err = ops.SPCClient.Delete(spc.Name, &metav1.DeleteOptions{})
				Expect(err).To(BeNil(), "while deleting the spc's", spc)
			}
		})
	})

	Context("Test maya-exporter's response", func() {
		It("should show pool status and volume as online (1) and no errors", func() {
			By("sending get request to maya-exporter without pvc")
			curl := "curl localhost:9500/metrics/?format=json"
			cmd := []string{"/bin/bash", "-c", curl}
			opts := tests.NewOptions().
				WithPodName(pod.Items[0].Name).
				WithNamespace(pod.Items[0].Namespace).
				WithContainer(pod.Items[0].Spec.Containers[2].Name).
				WithCommand(cmd...)

			out, err := ops.ExecPod(opts)

			Expect(err).To(BeNil(), "while executing command in container ", cmd)

			By("unmarshalling pool metrics")
			stats := apis.PoolMetricsList{}
			err = json.Unmarshal(out, &stats)
			Expect(err).To(BeNil(), "while unmarshalling metrics", string(out))

			mapResp := stats.ToMap()

			By("verifying whether pool status is online")
			Expect(apis.GetValue(poolStatus, mapResp)).To(Equal(online), "while getting pool status of", pod.Items[0].Name)

			By("verifying whether there is no error")
			Expect(apis.GetValue(noPoolAvailableErrorCount, mapResp)).To(Equal(float64(0)), "while getting total no of no pool available errors", pod.Items[0].Name, mapResp)

			By("building pvc object")
			pvcObj, err = pvc.NewBuilder().
				WithName(pvcName).
				WithNamespace(nsName).
				WithStorageClass(scName).
				WithAccessModes([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}).
				WithCapacity("1G").
				Build()
			Expect(err).To(BeNil(), "while creating pvc", pvcName)

			By("creating pvc")
			_, err = ops.PVCClient.WithNamespace(nsName).Create(pvcObj)
			Expect(err).To(BeNil(), "while creating pvc", pvcName)

			By("verifying pvc to be created and bound with pv")
			Eventually(
				func() bool {
					return ops.IsPVCBound(pvcName)
				},
				framework.DefaultTimeOut, framework.DefaultPollingInterval).
				Should(BeTrue())

			Eventually(
				func() bool {
					By("getting the pvclaim name")
					pvcObj, err = pvc.
						NewKubeClient(pvc.WithKubeConfigPath(kubeConfigPath)).
						WithNamespace(nsName).
						Get(pvcName, metav1.GetOptions{})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(pvcObj.Spec.VolumeName).ShouldNot(BeEmpty())

					By("verifying whether cvr is created and healthy")
					csv, err = cv.
						NewKubeclient(cv.WithNamespace("openebs"), cv.WithKubeConfigPath(kubeConfigPath)).
						Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
					Expect(err).ShouldNot(HaveOccurred())
					return cv.
						NewForAPIObject(csv).IsHealthy()
				},
				framework.DefaultTimeOut, framework.DefaultPollingInterval).
				Should(BeTrue())

			By("sending get request to maya-exporter")
			out, err = ops.ExecPod(opts)
			Expect(err).To(BeNil(), "while executing command in container ", cmd)
			err = json.Unmarshal(out, &stats)

			By("unmarshalling the metrics")
			Expect(err).To(BeNil(), "while unmarshalling the stats", string(out))
			mapResp = stats.ToMap()
			Expect(apis.GetValue(replicaStatus, mapResp)).To(Equal(healthy), "while getting pool status of", pod.Items[0].Name, mapResp)
		})
	})
})
