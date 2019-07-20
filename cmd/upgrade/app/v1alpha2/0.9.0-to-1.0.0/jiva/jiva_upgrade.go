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
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
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
	kubeConfigPath = "/home/user/.kube/config"
	deployClient   = deploy.NewKubeClient(deploy.WithKubeConfigPath(kubeConfigPath))
	serviceClient  = svc.NewKubeClient(svc.WithKubeConfigPath(kubeConfigPath))
	pvClient       = pv.NewKubeClient(pv.WithKubeConfigPath(kubeConfigPath))
)

type replicaPatchDetails struct {
	Version, PVName, ReplicaContainerName, ReplicaImage string
}

type controllerPatchDetails struct {
	Version, ControllerContainerName, ControllerImage, MExporterImage string
}

func getOpenEBSVersion(d *appsv1.Deployment) (string, error) {
	if d.Labels["openebs.io/version"] == "" {
		return "", errors.Errorf("missing openebs version")
	}
	return d.Labels["openebs.io/version"], nil
}

func getDeployment(labels, namespace string) (*appsv1.Deployment, error) {
	deployList, err := deployClient.WithNamespace(namespace).List(
		&metav1.ListOptions{
			LabelSelector: labels,
		})
	if err != nil {
		return nil, err
	}
	if len(deployList.Items) == 0 {
		return nil, errors.Errorf("no deployments found for %s", labels)
	}
	return &(deployList.Items[0]), nil
}

func getReplicaPatchDetails(d *appsv1.Deployment) (
	*replicaPatchDetails,
	error,
) {
	rd := &replicaPatchDetails{}
	// verify delpoyment name
	if d.Name == "" {
		return nil, errors.New("missing deployment name")
	}
	name, err := getContainerName(d)
	if err != nil {
		return nil, err
	}
	rd.ReplicaContainerName = name
	image, err := getBaseImage(d, rd.ReplicaContainerName)
	if err != nil {
		return nil, err
	}
	rd.ReplicaImage = image
	return rd, nil
}

func getControllerPatchDetails(d *appsv1.Deployment) (
	*controllerPatchDetails,
	error,
) {
	rd := &controllerPatchDetails{}
	// verify delpoyment name
	if d.Name == "" {
		return nil, errors.New("missing deployment name")
	}
	name, err := getContainerName(d)
	if err != nil {
		return nil, err
	}
	rd.ControllerContainerName = name
	image, err := getBaseImage(d, rd.ControllerContainerName)
	if err != nil {
		return nil, err
	}
	rd.ControllerImage = image
	image, err = getBaseImage(d, "maya-volume-exporter")
	if err != nil {
		return nil, err
	}
	rd.MExporterImage = image
	return rd, nil
}

func patchDelpoyment(
	deployName,
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

func getContainerName(d *appsv1.Deployment) (string, error) {
	containerList := d.Spec.Template.Spec.Containers
	// verify length of container list
	if len(containerList) == 0 {
		return "", errors.New("missing container")
	}
	name := containerList[0].Name
	// verify replica container name
	if name == "" {
		return "", errors.New("missing container name")
	}
	return name, nil
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
	pvName := "pvc-8399fbcd-ab15-11e9-afd4-54e1ad5e8320"
	openebsNamespace := "openebs"

	var (
		pvLabel         = "openebs.io/persistent-volume=" + pvName
		replicaLabel    = "openebs.io/replica=jiva-replica," + pvLabel
		controllerLabel = "openebs.io/controller=jiva-controller," + pvLabel
		ns              string
		err             error
	)

	// verifying whether the pvc is deployed with DeployInOpenebsNamespace cas config
	deployInOpenebs, err := deployClient.WithNamespace(openebsNamespace).List(
		&metav1.ListOptions{
			LabelSelector: pvLabel,
		})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
	replicaDeployObj, err := getDeployment(replicaLabel, ns)
	if err != nil {
		fmt.Println("failed to get replica deployment")
		os.Exit(1)
	}
	replicaVersion, err := getOpenEBSVersion(replicaDeployObj)
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
	replicaDetails, err := getReplicaPatchDetails(replicaDeployObj)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	replicaDetails.PVName = pvName
	replicaDetails.Version = upgradeVersion

	// fetching controller deployment details
	controllerDeployObj, err := getDeployment(controllerLabel, ns)
	if err != nil {
		fmt.Println("failed to get controller deployment")
		os.Exit(1)
	}
	controllerVersion, err := getOpenEBSVersion(controllerDeployObj)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if (controllerVersion != currentVersion) && (controllerVersion != upgradeVersion) {
		fmt.Printf(
			"controller version %s is neither %s nor %s",
			controllerVersion,
			currentVersion,
			upgradeVersion,
		)
		os.Exit(1)
	}
	controllerDetails, err := getControllerPatchDetails(controllerDeployObj)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	controllerDetails.Version = upgradeVersion

	// fetching controller service details
	controllerServiceList, err := serviceClient.WithNamespace(ns).List(
		metav1.ListOptions{
			LabelSelector: pvLabel,
		})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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

	var buffer bytes.Buffer

	// replica patch
	if replicaVersion == currentVersion {
		tmpl, err := template.ParseFiles("jiva-replica-patch.tpl.json")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = tmpl.Execute(&buffer, *replicaDetails)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		replicaPatch := buffer.String()
		err = patchDelpoyment(
			replicaDeployObj.Name,
			ns,
			types.StrategicMergePatchType,
			[]byte(replicaPatch),
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(replicaDeployObj.Name, " patched")
	} else {
		fmt.Printf("replica deployment already in %s version\n", upgradeVersion)
	}

	buffer.Reset()

	// controller patch
	if controllerVersion == currentVersion {
		tmpl, err := template.ParseFiles("jiva-target-patch.tpl.json")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = tmpl.Execute(&buffer, *controllerDetails)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		controllerPatch := buffer.String()

		err = patchDelpoyment(
			controllerDeployObj.Name,
			ns,
			types.StrategicMergePatchType,
			[]byte(controllerPatch),
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(controllerDeployObj.Name, " patched")
	} else {
		fmt.Printf("controller deployment already in %s version", upgradeVersion)
	}

	// service patch
	if controllerServiceVersion == currentVersion {
		tmpl, err := template.ParseFiles("jiva-target-svc-patch.tpl.json")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		servicePatch := buffer.String()
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
