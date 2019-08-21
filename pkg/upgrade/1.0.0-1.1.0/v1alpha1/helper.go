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

package v1alpha1

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	templates "github.com/openebs/maya/pkg/upgrade/templates/v1"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	retry "github.com/openebs/maya/pkg/util/retry"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

func getDeployment(labels, namespace string) (*appsv1.Deployment, error) {
	deployList, err := deployClient.WithNamespace(namespace).List(
		&metav1.ListOptions{
			LabelSelector: labels,
		})
	if err != nil {
		return nil, err
	}
	if len(deployList.Items) == 0 {
		return nil, errors.Errorf("no deployments found for %s in %s", labels, namespace)
	}
	return &(deployList.Items[0]), nil
}

func getOpenEBSVersion(d *appsv1.Deployment) (string, error) {
	if d.Labels["openebs.io/version"] == "" {
		return "", errors.Errorf("missing openebs version for %s", d.Name)
	}
	return d.Labels["openebs.io/version"], nil
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
			rolloutStatus, err1 := deployClient.WithNamespace(namespace).
				RolloutStatus(deployName)
			if err1 != nil {
				return err1
			}
			if !rolloutStatus.IsRolledout {
				return errors.Errorf("failed to rollout because %s", rolloutStatus.Message)
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
			baseImage := strings.Split(con.Image, ":")[0]
			if urlPrefix != "" {
				// urlPrefix is the url to the directory where the images are present
				// the below logic takes the image name from current baseImage and
				// appends it to the given urlPrefix
				// For example baseImage is abc/quay.io/openebs/jiva
				// and urlPrefix is xyz/aws-56546546/openebsdirectory/
				// it will take jiva from current url and append it to urlPrefix
				// and return xyz/aws-56546546/openebsdirectory/jiva
				urlSubstr := strings.Split(baseImage, "/")
				baseImage = urlPrefix + urlSubstr[len(urlSubstr)-1]
			}
			return baseImage, nil
		}
	}
	return "", errors.Errorf("image not found for container %s", name)
}

func patchService(targetServiceLabel, namespace string) error {
	targetServiceObj, err := serviceClient.WithNamespace(namespace).List(
		metav1.ListOptions{
			LabelSelector: targetServiceLabel,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to get service for %s", targetServiceLabel)
	}
	if len(targetServiceObj.Items) == 0 {
		return errors.Errorf("no service found for %s in %s", targetServiceLabel, namespace)
	}
	targetServiceName := targetServiceObj.Items[0].Name
	if targetServiceName == "" {
		return errors.Errorf("missing service name")
	}
	version := targetServiceObj.Items[0].
		Labels["openebs.io/version"]
	if version != currentVersion && version != upgradeVersion {
		return errors.Errorf(
			"service version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	if version == currentVersion {
		tmpl, err := template.New("servicePatch").
			Parse(templates.OpenebsVersionPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for service patch")
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for service patch")
		}
		servicePatch := buffer.String()
		buffer.Reset()
		_, err = serviceClient.WithNamespace(namespace).Patch(
			targetServiceName,
			types.StrategicMergePatchType,
			[]byte(servicePatch),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to patch service %s", targetServiceName)
		}
		fmt.Printf("targetservice %s patched\n", targetServiceName)
	} else {
		fmt.Printf("service already in %s version\n", upgradeVersion)
	}
	return nil
}
