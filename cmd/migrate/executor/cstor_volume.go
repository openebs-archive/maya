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

	migrate "github.com/openebs/maya/pkg/cstor/migrate"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	errors "github.com/pkg/errors"
)

var (
	cstorVolumeMigrateCmdHelpText = `
This command migrates the cStor volume

Usage: migrate cstor-volume --pv-name <pv-name>
`
)

// NewMigrateCStorVolumeJob migrates all the cStor Volumes associated with
// a given Storage Pool Claim
func NewMigrateCStorVolumeJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cstor-volume",
		Short:   "Migrate cStor Volume",
		Long:    cstorVolumeMigrateCmdHelpText,
		Example: `migrate cstor-volume --pv-name <pv-name>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
			util.CheckErr(options.RunCStorVolumeMigrateChecks(cmd), util.Fatal)
			util.CheckErr(options.RunCStorVolumeMigrate(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.pvName,
		"pv-name", "",
		options.pvName,
		"cstor volume name to be migrated. Run \"kubectl get pv\", to get pv-name")

	return cmd
}

// RunCStorVolumeMigrateChecks will ensure the sanity of the cstor SPC migrate options
func (u *MigrateOptions) RunCStorVolumeMigrateChecks(cmd *cobra.Command) error {
	if len(strings.TrimSpace(u.pvName)) == 0 {
		fmt.Println("name", u.pvName)
		return errors.Errorf("Cannot execute migrate job: cstor pv name is missing")
	}
	return nil
}

// RunCStorVolumeMigrate migrates the given spc.
func (u *MigrateOptions) RunCStorVolumeMigrate(cmd *cobra.Command) error {
	klog.Infof("Migrating spc %s to cspc", u.pvName)
	err := migrate.Volume(u.pvName, u.openebsNamespace)
	if err != nil {
		klog.Error(err)
		return errors.Errorf("Failed to migrate cStor volume : %s", u.pvName)
	}
	klog.Infof("Successfully migrated volume %s to csi-volume. Please scale up the application to trigger CSI driver.", u.pvName)
	return nil
}
