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
	retry "github.com/openebs/maya/pkg/util/retry"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	upgradeVersion = "1.0.0"
	currentVersion = "0.9.0"

	replicaPatchTemplate = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "{{.UpgradeVersion}}",
			  "openebs.io/persistent-volume": "{{.PVName}}",
			  "openebs.io/replica": "jiva-replica"
		   }
		},
		"spec": {
			"selector": {
				"matchLabels":{
					"openebs.io/persistent-volume": "{{.PVName}}",
					"openebs.io/replica": "jiva-replica"
				}
			},
		   "template": {
			   "metadata": {
				   "labels": {
					   "openebs.io/version": "{{.UpgradeVersion}}",
					   "openebs.io/persistent-volume": "{{.PVName}}",
					   "openebs.io/replica": "jiva-replica"
				   }
			   },
			  "spec": {
				 "containers": [
					{
					   "name": "{{.ReplicaContainerName}}",
					   "image": "{{.ReplicaImage}}:{{.UpgradeVersion}}"
					}
				 ],
				 "affinity": {
					 "podAntiAffinity": {
						 "requiredDuringSchedulingIgnoredDuringExecution": [
							 {
								 "labelSelector": {
									 "matchLabels": {
										 "openebs.io/persistent-volume": "{{.PVName}}",
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

	targetPatchTemplate = `{
		"metadata": {
		   "labels": {
			 "openebs.io/version": "{{.UpgradeVersion}}"
		   }
		},
		"spec": {
		   "template": {
			  "metadata": {
				 "labels":{
					"openebs.io/version": "{{.UpgradeVersion}}"
				 }
			  },
			 "spec": {
			   "containers": [
				 {
					"name": "{{.ControllerContainerName}}",
					"image": "{{.ControllerImage}}:{{.UpgradeVersion}}"
				 },
				 {
					"name": "maya-volume-exporter",
					"image": "{{.MExporterImage}}:{{.UpgradeVersion}}"
				 }
			   ]
			 }
		   }
		}
	  }`

	servicePatchTemplate = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "{{.}}"
		   }
		}
	 }`

	buffer bytes.Buffer

	deployClient  = deploy.NewKubeClient()
	serviceClient = svc.NewKubeClient()
	pvClient      = pv.NewKubeClient()
)

type replicaPatchDetails struct {
	UpgradeVersion, PVName, ReplicaContainerName, ReplicaImage string
}

type controllerPatchDetails struct {
	UpgradeVersion, ControllerContainerName, ControllerImage, MExporterImage string
}

type replicaDetails struct {
	patchDetails  *replicaPatchDetails
	version, name string
}

type controllerDetails struct {
	patchDetails  *controllerPatchDetails
	version, name string
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
	_, err := deployClient.WithNamespace(namespace).Patch(
		deployName,
		pt,
		data,
	)
	if err != nil {
		return err
	}

	err = retry.
		Times(60).
		Wait(5 * time.Second).
		Try(func(attempt uint) error {
			rolloutStatus, err := deployClient.WithNamespace(namespace).
				RolloutStatus(deployName)
			if err != nil {
				return err
			}
			if !rolloutStatus.IsRolledout {
				return errors.Errorf("failed to rollout %s", rolloutStatus.Message)
			}
			return nil
		})
	if err != nil {
		return err
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

func getReplica(replicaLabel, namespace string) (*replicaDetails, error) {
	replicaObj := &replicaDetails{}
	deployObj, err := getDeployment(replicaLabel, namespace)
	if err != nil {
		return nil, errors.Errorf("failed to get replica deployment")
	}
	if deployObj.Name == "" {
		return nil, errors.Errorf("missing deployment name for replica")
	}
	replicaObj.name = deployObj.Name
	version, err := getOpenEBSVersion(deployObj)
	if err != nil {
		return nil, err
	}
	if (version != currentVersion) && (version != upgradeVersion) {
		return nil, errors.Errorf(
			"replica version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	replicaObj.version = version
	patchDetails, err := getReplicaPatchDetails(deployObj)
	if err != nil {
		return nil, err
	}
	replicaObj.patchDetails = patchDetails
	replicaObj.patchDetails.UpgradeVersion = upgradeVersion
	return replicaObj, nil
}

func getController(controllerLabel, namespace string) (*controllerDetails, error) {
	controllerObj := &controllerDetails{}
	deployObj, err := getDeployment(controllerLabel, namespace)
	if err != nil {
		return nil, err
	}
	if deployObj.Name == "" {
		return nil, errors.Errorf("missing deployment name for controller")
	}
	controllerObj.name = deployObj.Name
	version, err := getOpenEBSVersion(deployObj)
	if err != nil {
		return nil, err
	}
	if (version != currentVersion) && (version != upgradeVersion) {
		return nil, errors.Errorf(
			"controller version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	controllerObj.version = version
	patchDetails, err := getControllerPatchDetails(deployObj)
	if err != nil {
		return nil, err
	}
	controllerObj.patchDetails = patchDetails
	controllerObj.patchDetails.UpgradeVersion = upgradeVersion
	return controllerObj, nil
}

func patchReplica(replicaObj *replicaDetails, namespace string) error {
	if replicaObj.version == currentVersion {
		tmpl, err := template.New("replicaPatch").Parse(replicaPatchTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(&buffer, replicaObj.patchDetails)
		if err != nil {
			return err
		}
		replicaPatch := buffer.String()
		err = patchDelpoyment(
			replicaObj.name,
			namespace,
			types.StrategicMergePatchType,
			[]byte(replicaPatch),
		)
		if err != nil {
			return err
		}
		fmt.Println(replicaObj.name, " patched")
	} else {
		fmt.Printf("replica deployment already in %s version\n", upgradeVersion)
	}
	return nil
}

func patchController(controllerObj *controllerDetails, namespace string) error {
	if controllerObj.version == currentVersion {
		tmpl, err := template.New("controllerPatch").Parse(targetPatchTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(&buffer, controllerObj.patchDetails)
		if err != nil {
			return err
		}
		controllerPatch := buffer.String()

		err = patchDelpoyment(
			controllerObj.name,
			namespace,
			types.StrategicMergePatchType,
			[]byte(controllerPatch),
		)
		if err != nil {
			return err
		}
		fmt.Println(controllerObj.name, " patched")
	} else {
		fmt.Printf("controller deployment already in %s version\n", upgradeVersion)
	}
	return nil
}

func main() {
	// inputs required for the upgrade
	pvName := os.Args[1]
	openebsNamespace := os.Args[2]

	var (
		pvLabel         = "openebs.io/persistent-volume=" + pvName
		replicaLabel    = "openebs.io/replica=jiva-replica," + pvLabel
		controllerLabel = "openebs.io/controller=jiva-controller," + pvLabel
		ns              string
		err             error
	)

	pvObj, err := pvClient.Get(pvName, metav1.GetOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// verifying whether the pvc is deployed with DeployInOpenebsNamespace cas config
	deployInOpenebs, err := deployClient.WithNamespace(openebsNamespace).List(
		&metav1.ListOptions{
			LabelSelector: pvLabel,
		})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// check whether pvc pods are openebs namespace or not
	if len(deployInOpenebs.Items) > 0 {
		ns = openebsNamespace
	} else {
		// if pvc pods are not in openebs namespace take the namespace of pvc
		if err != nil {
			fmt.Println("namespace missing", pvObj)
			os.Exit(1)
		}
		ns = pvObj.Spec.ClaimRef.Namespace
	}

	// fetching replica deployment details
	replicaObj, err := getReplica(replicaLabel, ns)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	replicaObj.patchDetails.PVName = pvName

	// fetching controller deployment details
	controllerObj, err := getController(controllerLabel, ns)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
	controllerServiceVersion := controllerServiceList.Items[0].
		Labels["openebs.io/version"]
	if controllerServiceVersion != currentVersion &&
		controllerServiceVersion != upgradeVersion {
		fmt.Printf(
			"controller service version %s is neither %s nor %s\n",
			controllerServiceVersion,
			currentVersion,
			upgradeVersion,
		)
		os.Exit(1)
	}

	// replica patch
	err = patchReplica(replicaObj, ns)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	buffer.Reset()

	// controller patch
	err = patchController(controllerObj, ns)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	buffer.Reset()

	// service patch
	if controllerServiceVersion == currentVersion {
		tmpl, err := template.New("servicePatch").Parse(servicePatchTemplate)
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
		fmt.Printf("controller service already in %s version\n", upgradeVersion)
	}

	fmt.Println("Upgrade Successful for", pvName)
}
