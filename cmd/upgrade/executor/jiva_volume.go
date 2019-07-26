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

var (
	jivaVolumeUpgradeCmdHelpText = `
This command upgrades the Jiva Persistent Volume

Usage: upgrade jiva-volume --volname <pv-name> --options...
`
)

// NewUpgradeJivaVolumeJob upgrade a Jiva Volume
func NewUpgradeJivaVolumeJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jiva-volume",
		Short:   "Upgrade Jiva Volume",
		Long:    jivaVolumeUpgradeCmdHelpText,
		Example: `upgrade jiva-volume --volname <pv-name>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
			util.CheckErr(options.RunJivaVolumeUpgradeChecks(cmd), util.Fatal)
			util.CheckErr(options.RunJivaVolumeUpgrade(cmd), util.Fatal)
		},
	}

	options.resourceKind = "jivaVolume"

	cmd.Flags().StringVarP(&options.resourceName,
		"volname", "",
		options.resourceName,
		"jiva persistent volume name to be upgraded, as obtained using: kubectl get pv")

	return cmd
}

// RunJivaVolumeUpgradeChecks will ensure the sanity of the common upgrade options
func (u *UpgradeOptions) RunJivaVolumeUpgradeChecks( cmd *cobra.Command ) error {
	if len(strings.TrimSpace(u.resourceName)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: jiva pv name is missing")
	}

	if len(strings.TrimSpace(u.toVersionImageTag)) == 0 {
		u.toVersionImageTag = u.toVersion
	}

	return nil
}

// RunJivaVolumeUpgrade upgrades the given Jiva Volume.
func (u *UpgradeOptions) RunJivaVolumeUpgrade(cmd *cobra.Command) error {
	from := u.fromVersion
	to := u.toVersion
	kind := u.resourceKind
	name := u.resourceName
	openebsNamespace := u.namespace
	urlPrefix := u.imageUrlPrefix
	imageTag := u.toVersionImageTag

	fromVersion := strings.Split(from, "-")[0]
	toVersion := strings.Split(to, "-")[0]

	switch fromVersion + "-" + toVersion {
	case "0.9.0-1.0.0":
		fmt.Println("Upgrading to 1.0.0")
		err := upgrade090to100.Exec(kind, name, openebsNamespace)
		if err != nil {
			return errors.Errorf("Failed to upgrade Jiva Volume %v:", name)
		}
	case "1.0.0-1.1.0":
		fmt.Println("Upgrading to 1.1.0")
		err := upgrade100to110.Exec(from, to, kind, name, openebsNamespace, urlPrefix, imageTag)
		if err != nil {
			fmt.Println( err )
			return errors.Errorf("Failed to upgrade Jiva Volume %v:", name)
		}
	default:
		return errors.Errorf("Invalid from version %s or to version %s", from, to)
	}
	return nil
}
