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

package executor

import (
	"strings"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

// UpgradeOptions stores information required for upgrade
type UpgradeOptions struct {
	fromVersion       string
	toVersion         string
	openebsNamespace  string
	imageURLPrefix    string
	toVersionImageTag string
	resourceKind      string
	jivaVolume        JivaVolumeOptions
	cstorSPC          CStorSPCOptions
	cstorVolume       CStorVolumeOptions
	resource          ResourceOptions
}

var (
	options = &UpgradeOptions{
		openebsNamespace: "openebs",
		imageURLPrefix:   "quay.io/openebs/",
	}
)

// RunPreFlightChecks will ensure the sanity of the common upgrade options
func (u *UpgradeOptions) RunPreFlightChecks(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.openebsNamespace)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: namespace is missing")
	}

	if len(strings.TrimSpace(u.fromVersion)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: from-version is missing")
	}

	if len(strings.TrimSpace(u.toVersion)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: to-version is missing")
	}

	if len(strings.TrimSpace(u.resourceKind)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: resource details are missing")
	}

	return nil
}

// InitializeDefaults will ensure the default values for optional options are
// set.
func (u *UpgradeOptions) InitializeDefaults(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.toVersionImageTag)) == 0 {
		u.toVersionImageTag = u.toVersion
	}

	return nil
}

// getUpgradePath gives the path for the upgrade
func (u *UpgradeOptions) getUpgradePath() (string, error) {
	podClient := pod.NewKubeClient()
	from := strings.Split(u.fromVersion, "-")[0]
	to := strings.Split(u.toVersion, "-")[0]
	mayaLabels := "name=maya-apiserver"
	mayaPods, err := podClient.WithNamespace(u.openebsNamespace).
		List(
			metav1.ListOptions{
				LabelSelector: mayaLabels,
			},
		)
	if err != nil {
		return "", err
	}
	if len(mayaPods.Items) != 1 {
		return "", errors.Errorf("Expecting 1 maya pod got %d", len(mayaPods.Items))
	}
	// mayaVersion is the version of the control plane
	mayaVersion := strings.Split(mayaPods.Items[0].Labels["openebs.io/version"], "-")[0]

	// if the from version and to version have equal prefix to control plane
	// version for example 1.3.0-RC1, 1.3.0-RC2, 1.3.0, then this will require
	// a RC upgrade path like RC1-RC2, RC2-, RC1-.
	if from == mayaVersion && to == mayaVersion {
		from = ""
		if len(strings.Split(u.fromVersion, "-")) == 2 {
			from = strings.Split(u.fromVersion, "-")[1]
		}
		to = ""
		if len(strings.Split(u.toVersion, "-")) == 2 {
			to = strings.Split(u.toVersion, "-")[1]
		}
	}
	// if from and to version don't have same prefix, return the normal path
	return from + "-" + to, nil
}
