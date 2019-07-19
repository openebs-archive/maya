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
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pv "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

var (
	replicaPatch = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "@upgrade_version@",
			  "openebs.io/persistent-volume": "@pv_name@",
			  "openebs.io/replica": "jiva-replica"
		   }
		},
		"spec": {
			"selector": {
				"matchLabels":{
					"openebs.io/persistent-volume": "@pv_name@",
					"openebs.io/replica": "jiva-replica"
				}
			},
		   "template": {
			   "metadata": {
				   "labels": {
					   "openebs.io/version": "@upgrade_version@",
					   "openebs.io/persistent-volume": "@pv_name@",
					   "openebs.io/replica": "jiva-replica"
				   }
			   },
			  "spec": {
				 "containers": [
					{
					   "name": "@r_name@",
					   "image": "@replica_image@:@upgrade_version@"
					}
				 ],
				 "affinity": {
					 "podAntiAffinity": {
						 "requiredDuringSchedulingIgnoredDuringExecution": [
							 {
								 "labelSelector": {
									 "matchLabels": {
										 "openebs.io/persistent-volume": "@pv_name@",
										 "openebs.io/replica": "jiva-replica"
									 }
								 },
					 "topologyKey": "kubernetes.io/hostname"
							 }
						 ]
					 }
				 }
			  }
		   }
		}
	 }`

	controllerPatch = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "@upgrade_version@"
		   }
		},
		"spec": {
		   "template": {
			   "metadata": {
				   "labels":{
					   "openebs.io/version": "@upgrade_version@"
				   }
			   },
			  "spec": {
				 "containers": [
					{
					   "name": "@c_name@",
					   "image": "@controller_image@:@upgrade_version@"
					},
					{
					   "name": "maya-volume-exporter",
					   "image": "@m_exporter_image@:@upgrade_version@"
					}
				 ]
			  }
		   }
		}
	 }`

	servicePatch = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "@upgrade_version@"
		   }
		}
	 }`

	kubeConfigPath = "/home/user/.kube/config"

	deployClient  = deploy.NewKubeClient(deploy.WithKubeConfigPath(kubeConfigPath))
	serviceClient = svc.NewKubeClient(svc.WithKubeConfigPath(kubeConfigPath))
	pvClient      = pv.NewKubeClient(pv.WithKubeConfigPath(kubeConfigPath))
)

func getDeploymentDetails(labels, namespace string) (
	deployName string,
	containerName string,
	version string,
	err error,
) {
	deployList, err := deployClient.WithNamespace(namespace).List(
		&metav1.ListOptions{
			LabelSelector: labels,
		})
	if err != nil {
		return "", "", "", err
	}
	if len(deployList.Items) == 1 {
		deployName = deployList.Items[0].Name
		version = deployList.Items[0].Labels["openebs.io/version"]
		if deployName == "" {
			return "", "", "", errors.New("empty deployment name")
		}
		if len(deployList.Items[0].Spec.Template.Spec.Containers) == 0 {
			return "", "", "", errors.New("empty container list")
		}
		containerName = deployList.Items[0].Spec.Template.Spec.Containers[0].Name
		if containerName == "" {
			return "", "", "", errors.New("empty container name")
		}
	} else {
		return "", "", "", errors.New("deployment missing")
	}
	return deployName, containerName, version, nil
}

func patchDelpoyment(
	deployName string,
	namespace string,
	pt types.PatchType,
	data []byte,
) error {
	var (
		retries = 60
	)
	_, err := deployClient.WithNamespace(namespace).Patch(
		deployName,
		pt,
		data,
	)
	if err != nil {
		return err
	}
	for {
		retries = retries - 1
		rolloutStatus, err := deployClient.WithNamespace(namespace).
			RolloutStatus(deployName)
		if err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if retries == 0 || rolloutStatus.IsRolledout {
			if !rolloutStatus.IsRolledout {
				return errors.Errorf("failed to rollout %s", rolloutStatus.Message)
			}
			break
		}
	}
	return nil
}

func getBaseImage(deployObj *appsv1.Deployment, name string) (string, error) {
	for _, con := range deployObj.Spec.Template.Spec.Containers {
		if con.Name == name {
			return strings.Split(con.Image, ":")[0], nil
		}
	}
	return "", errors.Errorf("image not found for %s", name)
}

func main() {
	// inputs required for the upgrade
	upgradeVersion := "1.0.0"
	currentVersion := "0.9.0"
	pvName := "pvc-e9c1b919-aa19-11e9-bea9-54e1ad5e8320"
	openebsNamespace := "openebs"

	var (
		pvLabel         = "openebs.io/persistent-volume=" + pvName
		replicaLabel    = "openebs.io/replica=jiva-replica," + pvLabel
		controllerLabel = "openebs.io/controller=jiva-controller," + pvLabel
		ns              = ""
	)

	// verifying whether the pvc is deployed with DeployInOpenebsNamespace cas config
	deployInOpenebs, err := deployClient.WithNamespace(openebsNamespace).List(
		&metav1.ListOptions{
			LabelSelector: pvLabel,
		})
	if len(deployInOpenebs.Items) > 0 {
		ns = openebsNamespace
	} else {
		pvObj, err := pvClient.Get(pvName, metav1.GetOptions{})
		if err != nil {
			fmt.Println("namespace missing")
			os.Exit(1)
		}
		ns = pvObj.Spec.ClaimRef.Namespace
	}

	// fetching replica deployment details
	replicaDeployName, replicaContainer, replicaVersion, err := getDeploymentDetails(
		replicaLabel,
		ns,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if (replicaVersion != currentVersion) && (replicaVersion != upgradeVersion) {
		fmt.Printf(
			"replica version %s is neither %s nor %s",
			replicaVersion,
			currentVersion,
			upgradeVersion,
		)
		os.Exit(1)
	}
	replicaDeployObj, err := deployClient.WithNamespace(ns).Get(replicaDeployName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// fetching controller deployment details
	controllerDeployName, controllerContainer, controllerVersion, err := getDeploymentDetails(
		controllerLabel,
		ns,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if controllerVersion != currentVersion && controllerVersion != upgradeVersion {
		fmt.Printf(
			"controller version %s is neither %s nor %s",
			controllerVersion,
			currentVersion,
			upgradeVersion,
		)
		os.Exit(1)
	}
	controllerDeployObj, err := deployClient.WithNamespace(ns).Get(controllerDeployName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// fetching controller service details
	controllerServiceList, err := serviceClient.WithNamespace(ns).List(
		metav1.ListOptions{
			LabelSelector: pvLabel,
		})
	// controllerServiceObj := controllerServiceList.Items[0]
	controllerServiceName := controllerServiceList.Items[0].Name
	controllerServiceVersion := controllerServiceList.Items[0].Labels["openebs.io/version"]
	if controllerServiceVersion != currentVersion && controllerServiceVersion != upgradeVersion {
		fmt.Printf(
			"controller service version %s is neither %s nor %s",
			controllerServiceVersion,
			currentVersion,
			upgradeVersion,
		)
		os.Exit(1)
	}

	// replica patch
	if replicaVersion == currentVersion {
		replicaBaseImage, err := getBaseImage(replicaDeployObj, replicaContainer)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		replicaPatch = strings.ReplaceAll(replicaPatch, "@upgrade_version@", upgradeVersion)
		replicaPatch = strings.ReplaceAll(replicaPatch, "@pv_name@", pvName)
		replicaPatch = strings.ReplaceAll(replicaPatch, "@r_name@", replicaContainer)
		replicaPatch = strings.ReplaceAll(replicaPatch, "@replica_image@", replicaBaseImage)

		err = patchDelpoyment(
			replicaDeployName,
			ns,
			types.StrategicMergePatchType,
			[]byte(replicaPatch),
		)
		fmt.Println(replicaDeployName, " patched")
	} else {
		fmt.Printf("replica deployment already in %s version\n", upgradeVersion)
	}

	// controller patch
	if controllerVersion == currentVersion {
		controllerBaseImage, err := getBaseImage(controllerDeployObj, controllerContainer)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		mExpoerterBaseImage, err := getBaseImage(controllerDeployObj, "maya-volume-exporter")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		controllerPatch = strings.ReplaceAll(controllerPatch, "@upgrade_version@", upgradeVersion)
		controllerPatch = strings.ReplaceAll(controllerPatch, "@c_name@", controllerContainer)
		controllerPatch = strings.ReplaceAll(controllerPatch, "@controller_image@", controllerBaseImage)
		controllerPatch = strings.ReplaceAll(controllerPatch, "@m_exporter_image@", mExpoerterBaseImage)

		err = patchDelpoyment(
			controllerDeployName,
			ns,
			types.StrategicMergePatchType,
			[]byte(controllerPatch),
		)
		fmt.Println(controllerDeployName, " patched")
	} else {
		fmt.Printf("controller deployment already in %s version", upgradeVersion)
	}

	// service patch
	if controllerServiceVersion == currentVersion {
		servicePatch = strings.ReplaceAll(servicePatch, "@upgrade_version@", upgradeVersion)

		_, err = serviceClient.WithNamespace(ns).Patch(
			controllerServiceName,
			types.StrategicMergePatchType,
			[]byte(servicePatch),
		)
		if err != nil {
			fmt.Println("Patch failed")
			fmt.Println(err)
		}
		fmt.Println(controllerServiceName, "patched")
	} else {
		fmt.Printf("controller service already in %s version", upgradeVersion)
	}

	fmt.Println("Upgrade Complete")
}
