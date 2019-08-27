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
	upgrade090to100 "github.com/openebs/maya/pkg/upgrade/0.9.0-1.0.0/v1alpha1"
	upgrade100to110 "github.com/openebs/maya/pkg/upgrade/1.0.0-1.1.0/v1alpha1"
)

// CStorVolumeOptions stores information required for cstor volume upgrade
type CStorVolumeOptions struct {
	pvName string
}

var (
	cstorVolumeUpgradeCmdHelpText = `
This command upgrades the CStor Persistent Volume

Usage: upgrade cstor-volume --volname <pv-name> --options...
`
)

// NewUpgradeCStorVolumeJob upgrade a CStor Volume
func NewUpgradeCStorVolumeJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cstor-volume",
		Short:   "Upgrade CStor Volume",
		Long:    cstorVolumeUpgradeCmdHelpText,
		Example: `upgrade cstor-volume --pv-name <pv-name>`,
		Run: func(cmd *cobra.Command, args []string) {
			options.resourceKind = "cstorVolume"
			util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
			util.CheckErr(options.RunCStorVolumeUpgradeChecks(cmd), util.Fatal)
			util.CheckErr(options.InitializeDefaults(cmd), util.Fatal)
			util.CheckErr(options.RunCStorVolumeUpgrade(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.cstorVolume.pvName,
		"pv-name", "",
		options.cstorVolume.pvName,
		"cstor persistent volume name. Run \"kubectl get pv\" to get pv-name.")

	return cmd
}

// RunCStorVolumeUpgradeChecks will ensure the sanity of the cstor upgrade options
func (u *UpgradeOptions) RunCStorVolumeUpgradeChecks(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.cstorVolume.pvName)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: cstor pv name is missing")
	}

	return nil
}

// RunCStorVolumeUpgrade upgrades the given CStor Volume.
func (u *UpgradeOptions) RunCStorVolumeUpgrade(cmd *cobra.Command) error {

	from := strings.Split(u.fromVersion, "-")[0]
	to := strings.Split(u.toVersion, "-")[0]

	switch from + "-" + to {
	case "0.9.0-1.0.0":
		fmt.Println("Upgrading to 1.0.0")
		err := upgrade090to100.Exec(u.resourceKind,
			u.cstorVolume.pvName,
			u.openebsNamespace)
		if err != nil {
			fmt.Println(err)
			return errors.Errorf("Failed to upgrade CStor Volume %v:", u.cstorVolume.pvName)
		}
	case "1.0.0-1.1.0":
		fmt.Println("Upgrading to 1.1.0")
		err := upgrade100to110.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.cstorVolume.pvName,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			fmt.Println(err)
			return errors.Errorf("Failed to upgrade CStor Volume %v:", u.cstorVolume.pvName)
		}
	case "1.1.0-1.2.0":
		fmt.Println("Upgrading to 1.2.0")
		err := upgrade100to110.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.cstorVolume.pvName,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			fmt.Println(err)
			return errors.Errorf("Failed to upgrade CStor Volume %v:", u.cstorVolume.pvName)
		}
	default:
		return errors.Errorf("Invalid from version %s or to version %s", u.fromVersion, u.toVersion)
	}
	return nil
}
