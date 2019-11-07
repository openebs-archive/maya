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

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	upgrader "github.com/openebs/maya/pkg/upgrade/upgrader"
)

// CStorSPCOptions stores information required for cstor SPC upgrade
type CStorSPCOptions struct {
	spcName string
}

var (
	cstorSPCUpgradeCmdHelpText = `
This command upgrades the cStor SPC

Usage: upgrade cstor-spc --spc-name <spc-name> --options...
`
)

// NewUpgradeCStorSPCJob upgrades all the cStor Pools associated with
// a given Storage Pool Claim
func NewUpgradeCStorSPCJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cstor-spc",
		Short:   "Upgrade cStor SPC",
		Long:    cstorSPCUpgradeCmdHelpText,
		Example: `upgrade cstor-spc --spc-name <spc-name>`,
		Run: func(cmd *cobra.Command, args []string) {
			options.resourceKind = "storagePoolClaim"
			util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
			util.CheckErr(options.RunCStorSPCUpgradeChecks(cmd), util.Fatal)
			util.CheckErr(options.InitializeDefaults(cmd), util.Fatal)
			util.CheckErr(options.RunCStorSPCUpgrade(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.cstorSPC.spcName,
		"spc-name", "",
		options.cstorSPC.spcName,
		"cstor SPC name to be upgraded. Run \"kubectl get spc\", to get spc-name")

	return cmd
}

// RunCStorSPCUpgradeChecks will ensure the sanity of the cstor SPC upgrade options
func (u *UpgradeOptions) RunCStorSPCUpgradeChecks(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.cstorSPC.spcName)) == 0 {
		return errors.Errorf("Cannot execute upgrade job: cstor spc name is missing")
	}

	return nil
}

// RunCStorSPCUpgrade upgrades the given Jiva Volume.
func (u *UpgradeOptions) RunCStorSPCUpgrade(cmd *cobra.Command) error {

	path, err := u.getUpgradePath()
	if err != nil {
		return err
	}
	switch path {
	case "1.0.0-1.3.0", "1.1.0-1.3.0", "1.2.0-1.3.0":
		// RC1-RC2 for RC1 to RC2, RC1- for RC1 to GA, RC2- for RC2 to GA
		klog.Infof("Upgrading to %s", u.toVersion)
		err := upgrader.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.cstorSPC.spcName,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			klog.Error(err)
			return errors.Errorf("Failed to upgrade cStor SPC %v:", u.cstorSPC.spcName)
		}
	default:
		return errors.Errorf("Invalid from version %s or to version %s", u.fromVersion, u.toVersion)
	}
	return nil
}
