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
	"fmt"
	"strings"

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	upgrade100to110 "github.com/openebs/maya/pkg/upgrade/1.0.0-1.1.0/v1alpha1"
	utask "github.com/openebs/maya/pkg/upgrade/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	upgradeTaskJobUpgradeCmdHelpText = `
This command upgrades the resource mentioned in upgradeTask env

Usage: upgrade upgrade-task
`
)

// UpgradeTaskOptions stores information required for upgradeTask upgrade
type UpgradeTaskOptions struct {
	resourceName string
}

// NewUpgradeTaskJob upgrade a resource from upgradeTask
func NewUpgradeTaskJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upgrade-task",
		Short:   "Upgrade UpgradeTask Resource",
		Long:    upgradeTaskJobUpgradeCmdHelpText,
		Example: `upgrade upgrade-task`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.InitializeFromUpgradeTask(cmd), util.Fatal)
			util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
			util.CheckErr(options.RunUpgradeTaskUpgradeChecks(cmd), util.Fatal)
			util.CheckErr(options.InitializeDefaults(cmd), util.Fatal)
			util.CheckErr(options.RunUpgradeTaskUpgrade(cmd), util.Fatal)
		},
	}

	return cmd
}

// InitializeFromUpgradeTask will populate the UpgradeOptions from given UpgradeTask
func (u *UpgradeOptions) InitializeFromUpgradeTask(cmd *cobra.Command) error {
	utaskName := getUpgradeTask()
	if len(strings.TrimSpace(u.openebsNamespace)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: namespace is missing")
	}
	utaskObj, _ := utask.NewKubeClient().WithNamespace(u.openebsNamespace).
		Get(utaskName, metav1.GetOptions{})

	if len(strings.TrimSpace(utaskObj.Spec.FromVersion)) != 0 {
		u.fromVersion = utaskObj.Spec.FromVersion
	}

	if len(strings.TrimSpace(utaskObj.Spec.ToVersion)) != 0 {
		u.toVersion = utaskObj.Spec.ToVersion
	}

	switch {
	case utaskObj.Spec.ResourceSpec.JivaVolume != nil:
		u.resourceKind = "jivaVolume"
		u.upgradeTask.resourceName = utaskObj.Spec.ResourceSpec.JivaVolume.PVName

	case utaskObj.Spec.ResourceSpec.CStorPool != nil:
		u.resourceKind = "cstorPool"
		u.upgradeTask.resourceName = utaskObj.Spec.ResourceSpec.CStorPool.PoolName

	case utaskObj.Spec.ResourceSpec.CStorVolume != nil:
		u.resourceKind = "cstorVolume"
		u.upgradeTask.resourceName = utaskObj.Spec.ResourceSpec.CStorVolume.PVName
	}

	return nil
}

// RunUpgradeTaskUpgradeChecks will ensure the sanity of the upgradeTask upgrade options
func (u *UpgradeOptions) RunUpgradeTaskUpgradeChecks(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.upgradeTask.resourceName)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: resource name is missing")
	}

	return nil
}

// RunUpgradeTaskUpgrade upgrades the given upgradeTask
func (u *UpgradeOptions) RunUpgradeTaskUpgrade(cmd *cobra.Command) error {

	from := strings.Split(u.fromVersion, "-")[0]
	to := strings.Split(u.toVersion, "-")[0]

	switch from + "-" + to {
	case "1.0.0-1.1.0":
		fmt.Println("Upgrading to 1.1.0")
		err := upgrade100to110.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.upgradeTask.resourceName,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			fmt.Println(err)
			return errors.Errorf("Failed to upgrade %v %v:", u.resourceKind, u.upgradeTask.resourceName)
		}
	case "1.1.0-1.2.0":
		fmt.Println("Upgrading to 1.2.0")
		err := upgrade100to110.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.upgradeTask.resourceName,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			fmt.Println(err)
			return errors.Errorf("Failed to upgrade %v %v:", u.resourceKind, u.upgradeTask.resourceName)
		}
	default:
		return errors.Errorf("Invalid from version %s or to version %s", u.fromVersion, u.toVersion)
	}
	return nil
}
