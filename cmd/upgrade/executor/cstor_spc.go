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

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	upgrader "github.com/openebs/maya/pkg/upgrade/upgrader"
	errors "github.com/pkg/errors"
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
			util.CheckErr(options.RunCStorSPCUpgradeChecks(cmd, args), util.Fatal)
			options.resourceKind = "storagePoolClaim"
			if options.cstorSPC.spcName != "" {
				singleCstorSPCUpgrade(cmd)
			}
			if len(args) != 0 {
				bulkCstorSPCUpgrade(cmd, args)
			}
		},
	}

	cmd.Flags().StringVarP(&options.cstorSPC.spcName,
		"spc-name", "",
		options.cstorSPC.spcName,
		"cstor SPC name to be upgraded. Run \"kubectl get spc\", to get spc-name")

	return cmd
}

func singleCstorSPCUpgrade(cmd *cobra.Command) {
	util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
	util.CheckErr(options.InitializeDefaults(cmd), util.Fatal)
	util.CheckErr(options.RunCStorSPCUpgrade(cmd), util.Fatal)
}

func bulkCstorSPCUpgrade(cmd *cobra.Command, args []string) {
	for _, name := range args {
		options.cstorSPC.spcName = name
		singleCstorSPCUpgrade(cmd)
	}
}

// RunCStorSPCUpgradeChecks will ensure the sanity of the cstor SPC upgrade options
func (u *UpgradeOptions) RunCStorSPCUpgradeChecks(cmd *cobra.Command, args []string) error {
	if len(strings.TrimSpace(u.cstorSPC.spcName)) == 0 && len(args) == 0 {
		return errors.Errorf("Cannot execute upgrade job:" +
			" neither spc-name flag is set nor spc name list is provided")
	}

	return nil
}

// RunCStorSPCUpgrade upgrades the given Jiva Volume.
func (u *UpgradeOptions) RunCStorSPCUpgrade(cmd *cobra.Command) error {
	klog.V(4).Infof("Started upgrading %s{%s} from %s to %s",
		u.resourceKind,
		u.cstorSPC.spcName,
		u.fromVersion,
		u.toVersion)

	if apis.IsCurrentVersionValid(u.fromVersion) && apis.IsDesiredVersionValid(u.toVersion) {
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
	} else {
		return errors.Errorf("Invalid from version %s or to version %s", u.fromVersion, u.toVersion)
	}
	return nil
}
