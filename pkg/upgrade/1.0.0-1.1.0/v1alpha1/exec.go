/*
Copyright 2019 The OpenEBS Authors.

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
	"bytes"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	cv "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	job "github.com/openebs/maya/pkg/kubernetes/job"
	pv "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	utask "github.com/openebs/maya/pkg/upgrade/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	upgradeVersion = ""
	currentVersion = ""
	urlPrefix      = ""
	imageTag       = ""

	buffer   bytes.Buffer
	utaskObj *apis.UpgradeTask

	isENVPresent bool

	cvClient      = cv.NewKubeclient()
	cvrClient     = cvr.NewKubeclient()
	deployClient  = deploy.NewKubeClient()
	serviceClient = svc.NewKubeClient()
	pvClient      = pv.NewKubeClient()
	podClient     = pod.NewKubeClient()
	jobClient     = job.NewKubeClient()
	cspClient     = csp.KubeClient()
	utaskClient   = utask.NewKubeClient()
)

// Exec ...
func Exec(fromVersion, toVersion, kind, name,
	openebsNamespace, urlprefix, imagetag string) error {

	if menv.Get("UPGRADE_TASK_LABEL") != "" {
		isENVPresent = true
	}

	// verify openebs namespace and check maya-apiserver version
	upgradeVersion = toVersion
	currentVersion = fromVersion
	urlPrefix = urlprefix
	imageTag = imagetag

	err := verifyMayaApiserver(openebsNamespace)
	if err != nil {
		return err
	}

	switch kind {
	case "jivaVolume":
		utaskObj, err = jivaUpgrade(name, openebsNamespace)

	case "storagePoolClaim":
		utaskObj, err = spcUpgrade(name, openebsNamespace)

	case "cstorPool":
		utaskObj, err = cspUpgrade(name, openebsNamespace)

	case "cstorVolume":
		utaskObj, err = cstorVolumeUpgrade(name, openebsNamespace)

	default:
		err = errors.Errorf("Invalid kind for upgrade")
	}

	if err != nil {
		if utaskObj != nil {
			backoffLimit, uerr := getBackoffLimit(openebsNamespace)
			if uerr != nil {
				return uerr
			}
			if utaskObj.Status.Retries == backoffLimit {
				utaskObj.Status.Phase = apis.UpgradeError
				utaskObj.Status.CompletedTime = metav1.Now()
			}
			utaskObj.Status.Retries = utaskObj.Status.Retries + 1
			_, uerr = utaskClient.WithNamespace(openebsNamespace).
				Update(utaskObj)
			if uerr != nil && isENVPresent {
				return uerr
			}
		}
		return err
	}
	if utaskObj != nil {
		utaskObj.Status.Phase = apis.UpgradeSuccess
		utaskObj.Status.CompletedTime = metav1.Now()
		_, uerr := utaskClient.WithNamespace(openebsNamespace).
			Update(utaskObj)
		if uerr != nil && isENVPresent {
			return uerr
		}
	}
	return nil
}

func getBackoffLimit(openebsNamespace string) (int, error) {
	podName := menv.Get("POD_NAME")
	podObj, err := podClient.WithNamespace(openebsNamespace).
		Get(podName, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}
	jobObj, err := jobClient.WithNamespace(openebsNamespace).
		Get(podObj.OwnerReferences[0].Name, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}
	// if backoffLimit not present it returns the default as 6
	if jobObj.Spec.BackoffLimit == nil {
		return 6, nil
	}
	backoffLimit := int(*jobObj.Spec.BackoffLimit)
	return backoffLimit, nil
}

func verifyMayaApiserver(openebsNamespace string) error {
	mayaLabels := "name=maya-apiserver"
	mayaPods, err := podClient.WithNamespace(openebsNamespace).
		List(
			metav1.ListOptions{
				LabelSelector: mayaLabels,
			},
		)
	if err != nil {
		return errors.Wrapf(err, "failed to get maya-apiserver deployment")
	}
	if len(mayaPods.Items) == 0 {
		return errors.Errorf(
			"failed to get maya-apiserver deployment in %s",
			openebsNamespace,
		)
	}
	if len(mayaPods.Items) > 1 {
		return errors.Errorf("control plane upgrade is not complete try after some time")
	}
	if mayaPods.Items[0].Labels["openebs.io/version"] != upgradeVersion {
		return errors.Errorf(
			"maya-apiserver deployment is in %s but required version is %s",
			mayaPods.Items[0].Labels["openebs.io/version"],
			upgradeVersion,
		)
	}
	return nil
}
