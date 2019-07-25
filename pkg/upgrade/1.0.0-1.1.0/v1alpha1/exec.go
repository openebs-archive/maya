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

	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	cv "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pv "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	upgradeVersion = ""
	currentVersion = ""
	urlPrefix      = ""
	imageTag       = ""

	buffer bytes.Buffer

	cvClient      = cv.NewKubeclient()
	cvrClient     = cvr.NewKubeclient()
	deployClient  = deploy.NewKubeClient()
	serviceClient = svc.NewKubeClient()
	pvClient      = pv.NewKubeClient()
	podClient     = pod.NewKubeClient()
	cspClient     = csp.KubeClient()
)

// Exec ...
func Exec(fromVersion, toVersion, kind, name,
	openebsNamespace, urlprefix, imagetag string) error {

	// verify openebs namespace and check maya-apiserver version
	upgradeVersion = toVersion
	currentVersion = fromVersion
	urlPrefix = urlprefix
	imageTag = imagetag

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

	switch kind {
	case "jivaVolume":
		err = jivaUpgrade(name, openebsNamespace)
		if err != nil {
			return err
		}
	case "storagePoolClaim":
		err = spcUpgrade(name, openebsNamespace)
		if err != nil {
			return err
		}
	case "cstorPool":
		err = cspUpgrade(name, openebsNamespace)
		if err != nil {
			return err
		}
	case "cstorVolume":
		err = cstorVolumeUpgrade(name, openebsNamespace)
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("Invalid kind for upgrade")
	}
	return nil
}
