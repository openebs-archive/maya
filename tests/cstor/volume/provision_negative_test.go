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

package volume

import (
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/debug"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("[Cstor Volume Provisioning Negative] Volume Provisioning", func() {
	When("SPC is created", func() {
		It("cStor Pools Should be Provisioned ", func() {
			By("Build And Create StoragePoolClaim object")
			// Populate configurations and create
			spcConfig := &tests.SPCConfig{
				Name:               spcName,
				DiskType:           "sparse",
				PoolCount:          cstor.PoolCount,
				IsOverProvisioning: false,
				PoolType:           "striped",
			}
			ops.Config = spcConfig
			spcObj = ops.BuildAndCreateSPC()
			By("Creating SPC, Desired Number of CSP Should Be Created", func() {
				ops.VerifyDesiredCSPCount(spcObj, cstor.PoolCount)
			})
			By("Build And Create StorageClass", buildAndCreateSC)
		})
	})
	When("Service is Applied", func() {
		It("Should Create Service", func() {
			var nodeIP string
			var err error
			By("Creating Service To Inject Errors During Volume Provisioning")
			spcLabel := SPCLabel + "=" + spcObj.Name
			poolPodList, err = ops.PodClient.
				WithNamespace(openebsNamespace).
				List(metav1.ListOptions{LabelSelector: spcLabel})
			Expect(err).To(BeNil())

			Expect(len(poolPodList.Items)).Should(BeNumerically(">=", 1),
				"Pool Pods are not yet created")
			servicePort := []corev1.ServicePort{
				corev1.ServicePort{
					Name:     "http",
					Port:     int32(8080),
					Protocol: "TCP",
					NodePort: int32(30031),
				},
			}
			nodeName := poolPodList.Items[0].Spec.NodeSelector[hostLabel]
			nodeObj, err := ops.NodeClient.Get(nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())

			//GetNode Ip
			for _, address := range nodeObj.Status.Addresses {
				if address.Type == corev1.NodeExternalIP {
					nodeIP = address.Address
					break
				}
			}
			Expect(nodeIP).NotTo(BeEmpty())
			hostIPPort = nodeIP + ":30031"
			serviceConfig := &tests.ServiceConfig{
				Name:        svcName,
				Namespace:   openebsNamespace,
				Selectors:   poolPodList.Items[0].Labels,
				ServicePort: servicePort,
			}
			ops.Config = serviceConfig
			serviceObj = ops.BuildAndCreateService()
		})
	})

	When("PersistentVolumeClaim Is Created", func() {
		It("Volume Should be Created and Provisioned", func() {
			// Populate PVC configurations
			pvcConfig := &tests.PVCConfig{
				Name:        pvcName,
				Namespace:   nsObj.Name,
				SCName:      scObj.Name,
				Capacity:    "5G",
				AccessModes: accessModes,
			}
			ops.Config = pvcConfig

			//Injecting Errors During CVRUpdate and ZFS Creation Time
			injectError := debug.NewClient(hostIPPort)
			err := injectError.PostInject(
				debug.NewErrorInjection().
					WithZFSCreateError(debug.Inject).
					WithCVRUpdateError(debug.Inject))
			Expect(err).To(BeNil())

			pvcObj = ops.BuildAndCreatePVC()
			// n-1 Replicas should be Healthy
			By("Creating PVC, Desired Number of CVR Should Be Created", func() {
				ops.VerifyVolumeStatus(pvcObj, cstor.ReplicaCount-1)
			})

			// GetLatest PVC object
			pvcObj, err = ops.PVCClient.
				WithNamespace(nsObj.Name).
				Get(pvcObj.Name, metav1.GetOptions{})
			Expect(err).To(BeNil())
		})
	})

	When("PVC Created During Error Injection", func() {
		It("All Volume Replicas Should Not Be Healthy", func() {
			By("Verify Volume Status after Injecting Errors", func() {
				command := "zfs get guid | grep " + pvcObj.Spec.VolumeName + " | awk '{print $3}'"
				_ = ops.ExecuteCMDEventually(
					&poolPodList.Items[0],
					"cstor-pool-mgmt", command, false)

				// Eject the ZFSCreate error
				injectError := debug.NewClient(hostIPPort)
				err := injectError.PostInject(
					debug.NewErrorInjection().
						WithZFSCreateError(debug.Eject).
						WithCVRUpdateError(debug.Inject))
				Expect(err).To(BeNil())

				//After ejecting the ZFS Create volume dataset should be created
				zvolGUID := ops.ExecuteCMDEventually(
					&poolPodList.Items[0],
					"cstor-pool-mgmt", command, true)
				Expect(zvolGUID).NotTo(BeEmpty())
				// CVR Update will error due to error injection. So replica count
				// should be n-1
				ops.VerifyVolumeStatus(pvcObj, cstor.ReplicaCount-1)

				//Eject CVRUpdate error then CVR should become Healthy
				err = injectError.PostInject(
					debug.NewErrorInjection().
						WithCVRUpdateError(debug.Eject))
				Expect(err).To(BeNil())

				ops.VerifyVolumeStatus(pvcObj, cstor.ReplicaCount)
			})
		})
	})

	When("CleanUp Negative Volume Provisioned Resources", func() {
		It("Should Delete All The Resources Related To Test", func() {
			err := ops.SVCClient.Delete(serviceObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
			By("Delete persistentVolumeClaim then volume resources should be deleted", func() {
				ops.DeleteVolumeResources(pvcObj, scObj)
			})
			By("Delete StoragePoolClaim then pool related resources should be deleted", func() {
				ops.DeleteStoragePoolClaim(spcObj.Name)
			})
		})
	})
})

func buildAndCreateSC() {
	var err error
	casConfig := strings.Replace(
		openebsCASConfigValue, "$spcName", spcObj.Name, 1)
	casConfig = strings.Replace(
		casConfig, "$count", strconv.Itoa(cstor.ReplicaCount), 1)
	annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
	annotations[string(apis.CASConfigKey)] = casConfig
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithAnnotations(annotations).
		WithProvisioner(openebsProvisioner).Build()
	Expect(err).ShouldNot(HaveOccurred(), "while building storageclass obj for storageclass {%s}", scName)

	By("creating storageclass")
	scObj, err = ops.SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating storageclass", scName)
}
