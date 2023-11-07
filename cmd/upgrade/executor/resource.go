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

	"k8s.io/klog/v2"

	utaskapis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"

	upgrader "github.com/openebs/maya/pkg/upgrade/upgrader"
	utask "github.com/openebs/maya/pkg/upgrade/v1alpha2"
	errors "github.com/pkg/errors"
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
			upgradeTaskLabel := getUpgradeTaskLabel()
			openebsNamespace := getOpenEBSNamespace()
			upgradeTaskList, err := utask.NewKubeClient().
				WithNamespace(openebsNamespace).
				List(metav1.ListOptions{
					LabelSelector: upgradeTaskLabel,
				})
			util.CheckErr(err, util.Fatal)
			if len(upgradeTaskList.Items) == 0 {
				util.Fatal("No resource found for given label")
			}
			for _, cr := range upgradeTaskList.Items {
				util.CheckErr(options.InitializeFromUpgradeTaskResource(
					cr, cmd), util.Fatal)
				util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
				util.CheckErr(options.RunResourceUpgradeChecks(cmd), util.Fatal)
				util.CheckErr(options.InitializeDefaults(cmd), util.Fatal)
				util.CheckErr(options.RunResourceUpgrade(cmd), util.Fatal)
			}
		},
	}
	// TODO
	// cmd.Flags().StringVarP(&options.resource.label,
	// 	"label", "",
	// 	options.resource.label,
	// 	"label of the upgradetasks to be upgraded. Run \"kubectl get utask \", to get label")

	return cmd
}

// InitializeFromUpgradeTaskResource will populate the UpgradeOptions from given UpgradeTask
func (u *UpgradeOptions) InitializeFromUpgradeTaskResource(
	upgradeTaskCRObj utaskapis.UpgradeTask, cmd *cobra.Command) error {

	if len(strings.TrimSpace(u.openebsNamespace)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: namespace is missing")
	}
	if len(strings.TrimSpace(upgradeTaskCRObj.Spec.FromVersion)) != 0 {
		u.fromVersion = upgradeTaskCRObj.Spec.FromVersion
	}

	if len(strings.TrimSpace(upgradeTaskCRObj.Spec.ToVersion)) != 0 {
		u.toVersion = upgradeTaskCRObj.Spec.ToVersion
	}

	if len(strings.TrimSpace(upgradeTaskCRObj.Spec.ImagePrefix)) != 0 {
		u.imageURLPrefix = upgradeTaskCRObj.Spec.ImagePrefix
	}

	if len(strings.TrimSpace(upgradeTaskCRObj.Spec.ImageTag)) != 0 {
		u.toVersionImageTag = upgradeTaskCRObj.Spec.ImageTag
	}

	switch {
	case upgradeTaskCRObj.Spec.ResourceSpec.JivaVolume != nil:
		u.resourceKind = "jivaVolume"
		u.resource.name = upgradeTaskCRObj.Spec.ResourceSpec.JivaVolume.PVName

	case upgradeTaskCRObj.Spec.ResourceSpec.CStorPool != nil:
		u.resourceKind = "cstorPool"
		u.resource.name = upgradeTaskCRObj.Spec.ResourceSpec.CStorPool.PoolName

	case upgradeTaskCRObj.Spec.ResourceSpec.StoragePoolClaim != nil:
		u.resourceKind = "storagePoolClaim"
		u.resource.name = upgradeTaskCRObj.Spec.ResourceSpec.StoragePoolClaim.SPCName

	case upgradeTaskCRObj.Spec.ResourceSpec.CStorVolume != nil:
		u.resourceKind = "cstorVolume"
		u.resource.name = upgradeTaskCRObj.Spec.ResourceSpec.CStorVolume.PVName
	}

	return nil
}

// RunResourceUpgradeChecks will ensure the sanity of the upgradeTask upgrade options
func (u *UpgradeOptions) RunResourceUpgradeChecks(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.resource.name)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: resource name is missing")
	}

	return nil
}

// RunResourceUpgrade upgrades the given upgradeTask
func (u *UpgradeOptions) RunResourceUpgrade(cmd *cobra.Command) error {

	if apis.IsCurrentVersionValid(u.fromVersion) && apis.IsDesiredVersionValid(u.toVersion) {
		klog.Infof("Upgrading to %s", u.toVersion)
		err := upgrader.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.resource.name,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			return errors.Wrapf(err, "Failed to upgrade %v %v", u.resourceKind, u.resource.name)
		}
	} else {
		return errors.Errorf("Invalid from version %s or to version %s", u.fromVersion, u.toVersion)
	}
	return nil
}
