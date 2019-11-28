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

	migrate "github.com/openebs/maya/pkg/cstor/migrate"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	errors "github.com/pkg/errors"
)

var (
	cstorSPCMigrateCmdHelpText = `
This command migrates the cStor SPC

Usage: migrate pool --spc-name <spc-name>
`
)

// NewMigratePoolJob migrates all the cStor Pools associated with
// a given Storage Pool Claim
func NewMigratePoolJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pool",
		Short:   "Migrate cStor SPC",
		Long:    cstorSPCMigrateCmdHelpText,
		Example: `migrate cstor-spc --spc-name <spc-name>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
			util.CheckErr(options.RunCStorSPCMigrateChecks(cmd), util.Fatal)
			util.CheckErr(options.RunCStorSPCMigrate(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.spcName,
		"spc-name", "",
		options.spcName,
		"cstor SPC name to be migrated. Run \"kubectl get spc\", to get spc-name")

	return cmd
}

// RunCStorSPCMigrateChecks will ensure the sanity of the cstor SPC migrate options
func (u *MigrateOptions) RunCStorSPCMigrateChecks(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.spcName)) == 0 {
		return errors.Errorf("Cannot execute migrate job: cstor spc name is missing")
	}

	return nil
}

// RunCStorSPCMigrate migrates the given spc.
func (u *MigrateOptions) RunCStorSPCMigrate(cmd *cobra.Command) error {

	err := migrate.Pool(u.spcName, u.openebsNamespace)
	if err != nil {
		klog.Error(err)
		return errors.Errorf("Failed to migrate cStor SPC : %s", u.spcName)
	}

	return nil
}
