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

	"github.com/golang/glog"

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	upgrade100to120 "github.com/openebs/maya/pkg/upgrade/1.0.0-1.1.0/v1alpha1"
	utask "github.com/openebs/maya/pkg/upgrade/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	resourceUpgradeCmdHelpText = `
This command upgrades the resource mentioned in the UpgradeTask CR.
The name of the UpgradeTask CR is extracted from the ENV UPGRADE_TASK

Usage: upgrade resource
`
)

// ResourceOptions stores information required for upgradeTask upgrade
type ResourceOptions struct {
	name string
}

// NewUpgradeResourceJob upgrade a resource from upgradeTask
func NewUpgradeResourceJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Short:   "Upgrade a resource using the details specified in the UpgradeTask CR.",
		Long:    resourceUpgradeCmdHelpText,
		Example: `upgrade resource`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.InitializeFromUpgradeTaskResource(cmd), util.Fatal)
			util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
			util.CheckErr(options.RunResourcekUpgradeChecks(cmd), util.Fatal)
			util.CheckErr(options.InitializeDefaults(cmd), util.Fatal)
			util.CheckErr(options.RunResourceUpgrade(cmd), util.Fatal)
		},
	}

	return cmd
}

// InitializeFromUpgradeTaskResource will populate the UpgradeOptions from given UpgradeTask
func (u *UpgradeOptions) InitializeFromUpgradeTaskResource(cmd *cobra.Command) error {
	upgradeTaskCRName := getUpgradeTaskCRName()
	if len(strings.TrimSpace(u.openebsNamespace)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: namespace is missing")
	}
	upgradeTaskCRObj, err := utask.NewKubeClient().WithNamespace(u.openebsNamespace).
		Get(upgradeTaskCRName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(upgradeTaskCRObj.Spec.FromVersion)) != 0 {
		u.fromVersion = upgradeTaskCRObj.Spec.FromVersion
	}

	if len(strings.TrimSpace(upgradeTaskCRObj.Spec.ToVersion)) != 0 {
		u.toVersion = upgradeTaskCRObj.Spec.ToVersion
	}

	switch {
	case upgradeTaskCRObj.Spec.ResourceSpec.JivaVolume != nil:
		u.resourceKind = "jivaVolume"
		u.resource.name = upgradeTaskCRObj.Spec.ResourceSpec.JivaVolume.PVName

	case upgradeTaskCRObj.Spec.ResourceSpec.CStorPool != nil:
		u.resourceKind = "cstorPool"
		u.resource.name = upgradeTaskCRObj.Spec.ResourceSpec.CStorPool.PoolName

	case upgradeTaskCRObj.Spec.ResourceSpec.CStorVolume != nil:
		u.resourceKind = "cstorVolume"
		u.resource.name = upgradeTaskCRObj.Spec.ResourceSpec.CStorVolume.PVName
	}

	return nil
}

// RunResourcekUpgradeChecks will ensure the sanity of the upgradeTask upgrade options
func (u *UpgradeOptions) RunResourcekUpgradeChecks(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.resource.name)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: resource name is missing")
	}

	return nil
}

// RunResourceUpgrade upgrades the given upgradeTask
func (u *UpgradeOptions) RunResourceUpgrade(cmd *cobra.Command) error {

	from := strings.Split(u.fromVersion, "-")[0]
	to := strings.Split(u.toVersion, "-")[0]

	switch from + "-" + to {
	case "1.0.0-1.1.0", "1.0.0-1.2.0", "1.1.0-1.2.0":
		glog.Infof("Upgrading to %s", u.toVersion)
		err := upgrade100to120.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.resource.name,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			return errors.Errorf("Failed to upgrade %v %v:", u.resourceKind, u.resource.name)
		}
	default:
		return errors.Errorf("Invalid from version %s or to version %s", u.fromVersion, u.toVersion)
	}
	return nil
}
