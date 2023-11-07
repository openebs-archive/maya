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

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	upgrader "github.com/openebs/maya/pkg/upgrade/upgrader"
	errors "github.com/pkg/errors"
)

// CStorVolumeOptions stores information required for cstor volume upgrade
type CStorVolumeOptions struct {
	pvName string
}

var (
	cstorVolumeUpgradeCmdHelpText = `
This command upgrades one or many CStor Persistent Volume
`
	cstorVolumeUpgradeCmdExampleText = `  # Upgrade one volume at a time
  upgrade cstor-volume --pv-name <pv-name> --options...

  # Upgrade multiple volumes at a time
  upgrade cstor-volume <pv-name>... --options...`
)

// NewUpgradeCStorVolumeJob upgrade a CStor Volume
func NewUpgradeCStorVolumeJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cstor-volume",
		Short:   "Upgrade CStor Volume",
		Long:    cstorVolumeUpgradeCmdHelpText,
		Example: cstorVolumeUpgradeCmdExampleText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.RunCStorVolumeUpgradeChecks(args), util.Fatal)
			options.resourceKind = "cstorVolume"
			if options.cstorVolume.pvName != "" {
				singleCStorVolumeUpgrade(cmd)
			}
			if len(args) != 0 {
				bulkCStorVolumeUpgrade(cmd, args)
			}
		},
	}

	cmd.Flags().StringVarP(&options.cstorVolume.pvName,
		"pv-name", "",
		options.cstorVolume.pvName,
		"cstor persistent volume name. Run \"kubectl get pv\" to get pv-name.")

	return cmd
}

func singleCStorVolumeUpgrade(cmd *cobra.Command) {
	util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
	util.CheckErr(options.InitializeDefaults(cmd), util.Fatal)
	util.CheckErr(options.RunCStorVolumeUpgrade(cmd), util.Fatal)
}

func bulkCStorVolumeUpgrade(cmd *cobra.Command, args []string) {
	for _, name := range args {
		options.cstorVolume.pvName = name
		singleCStorVolumeUpgrade(cmd)
	}
}

// RunCStorVolumeUpgradeChecks will ensure the sanity of the cstor upgrade options
func (u *UpgradeOptions) RunCStorVolumeUpgradeChecks(args []string) error {
	if len(strings.TrimSpace(u.cstorVolume.pvName)) == 0 && len(args) == 0 {
		return errors.Errorf("Cannot execute upgrade job:" +
			" neither pv-name flag is set nor pv name list is provided")
	}

	return nil
}

// RunCStorVolumeUpgrade upgrades the given CStor Volume.
func (u *UpgradeOptions) RunCStorVolumeUpgrade(cmd *cobra.Command) error {
	klog.V(4).Infof("Started upgrading %s{%s} from %s to %s",
		u.resourceKind,
		u.cstorVolume.pvName,
		u.fromVersion,
		u.toVersion)

	if apis.IsCurrentVersionValid(u.fromVersion) && apis.IsDesiredVersionValid(u.toVersion) {
		err := upgrader.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.cstorVolume.pvName,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			klog.Error(err)
			return errors.Errorf("Failed to upgrade CStor Volume %v:", u.cstorVolume.pvName)
		}
	} else {
		return errors.Errorf("Invalid from version %s or to version %s", u.fromVersion, u.toVersion)
	}
	klog.V(4).Infof("Successfully upgraded %s{%s} from %s to %s",
		u.resourceKind,
		u.cstorVolume.pvName,
		u.fromVersion,
		u.toVersion)
	return nil
}
