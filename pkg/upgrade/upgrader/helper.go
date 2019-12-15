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
	"strings"
	"text/template"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	templates "github.com/openebs/maya/pkg/upgrade/templates/v1"
	utask "github.com/openebs/maya/pkg/upgrade/v1alpha2"
	retry "github.com/openebs/maya/pkg/util/retry"
	errors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
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
	err = deploy.NewForAPIObject(&(deployList.Items[0])).VerifyReplicaStatus()
	if err != nil {
		return nil, err
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
			lastIndex := strings.LastIndex(con.Image, ":")
			baseImage := con.Image[:lastIndex]
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
		klog.Infof("targetservice %s patched", targetServiceName)
	} else {
		klog.Infof("service already in %s version", upgradeVersion)
	}
	return nil
}

// createUtask creates a UpgradeTask CR for the resource
func createUtask(utaskObj *apis.UpgradeTask, openebsNamespace string,
) (*apis.UpgradeTask, error) {
	var err error
	if utaskObj == nil {
		return nil, errors.Errorf("failed to create upgradetask : nil object")
	}
	utaskObj, err = utaskClient.WithNamespace(openebsNamespace).Create(utaskObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create upgradetask")
	}
	return utaskObj, nil
}

func updateUpgradeDetailedStatus(utaskObj *apis.UpgradeTask,
	uStatusObj apis.UpgradeDetailedStatuses, openebsNamespace string,
) (*apis.UpgradeTask, error) {
	var err error
	if !utask.IsValidStatus(uStatusObj) {
		return nil, errors.Errorf(
			"failed to update upgradetask status: invalid status %v",
			uStatusObj,
		)
	}
	uStatusObj.LastUpdatedTime = metav1.Now()
	if uStatusObj.Phase == apis.StepWaiting {
		uStatusObj.StartTime = uStatusObj.LastUpdatedTime
		utaskObj.Status.UpgradeDetailedStatuses = append(
			utaskObj.Status.UpgradeDetailedStatuses,
			uStatusObj,
		)
	} else {
		l := len(utaskObj.Status.UpgradeDetailedStatuses)
		uStatusObj.StartTime = utaskObj.Status.UpgradeDetailedStatuses[l-1].StartTime
		utaskObj.Status.UpgradeDetailedStatuses[l-1] = uStatusObj
	}
	utaskObj, err = utaskClient.WithNamespace(openebsNamespace).Update(utaskObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update upgradetask ")
	}
	return utaskObj, nil
}

// getOrCreateUpgradeTask fetches upgrade task if provided or creates a new upgradetask CR
func getOrCreateUpgradeTask(kind, name, openebsNamespace string) (*apis.UpgradeTask, error) {
	var utaskObj *apis.UpgradeTask
	var err error
	if openebsNamespace == "" {
		return nil, errors.Errorf("missing openebsNamespace")
	}
	if kind == "" {
		return nil, errors.Errorf("missing kind for upgradeTask")
	}
	if name == "" {
		return nil, errors.Errorf("missing name for upgradeTask")
	}
	utaskObj = buildUpgradeTask(kind, name, openebsNamespace)
	// the below logic first tries to fetch the CR if not found
	// then creates a new CR
	utaskObj1, err1 := utaskClient.WithNamespace(openebsNamespace).
		Get(utaskObj.Name, metav1.GetOptions{})
	if err1 != nil {
		if k8serror.IsNotFound(err1) {
			utaskObj, err = createUtask(utaskObj, openebsNamespace)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err1
		}
	} else {
		utaskObj = utaskObj1
	}

	if utaskObj.Status.StartTime.IsZero() {
		utaskObj.Status.Phase = apis.UpgradeStarted
		utaskObj.Status.StartTime = metav1.Now()
	}

	utaskObj.Status.UpgradeDetailedStatuses = []apis.UpgradeDetailedStatuses{}
	utaskObj, err = utaskClient.WithNamespace(openebsNamespace).
		Update(utaskObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update upgradetask")
	}
	return utaskObj, nil
}

func buildUpgradeTask(kind, name, openebsNamespace string) *apis.UpgradeTask {
	// TODO builder
	utaskObj := &apis.UpgradeTask{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: openebsNamespace,
		},
		Spec: apis.UpgradeTaskSpec{
			FromVersion: currentVersion,
			ToVersion:   upgradeVersion,
			ImageTag:    imageTag,
			ImagePrefix: urlPrefix,
		},
		Status: apis.UpgradeTaskStatus{
			Phase:     apis.UpgradeStarted,
			StartTime: metav1.Now(),
		},
	}
	switch kind {
	case "jivaVolume":
		utaskObj.Name = "upgrade-jiva-volume-" + name
		utaskObj.Spec.ResourceSpec = apis.ResourceSpec{
			JivaVolume: &apis.JivaVolume{
				PVName: name,
			},
		}
	case "cstorPool":
		utaskObj.Name = "upgrade-cstor-pool-" + name
		utaskObj.Spec.ResourceSpec = apis.ResourceSpec{
			CStorPool: &apis.CStorPool{
				PoolName: name,
			},
		}
	case "cstorVolume":
		utaskObj.Name = "upgrade-cstor-volume-" + name
		utaskObj.Spec.ResourceSpec = apis.ResourceSpec{
			CStorVolume: &apis.CStorVolume{
				PVName: name,
			},
		}
	}
	return utaskObj
}
