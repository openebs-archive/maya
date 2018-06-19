/*
Copyright 2018 The OpenEBS Authors.

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

package command

import (
	"fmt"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	volumeCloneCommandHelpText = `
This command clones a Volume.

Usage: mayactl volume clone --volname <volume name> --snapname <clone name> --sourcename <source volume name>[-size <size>]
`
)

// NewCmdVolumeClone clones a new OpenEBS Volume
func NewCmdVolumeClone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clones a Volume",
		Long:  volumeCreateCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd, true, true, true), util.Fatal)
			util.CheckErr(options.RunVolumeClone(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName, "unique volume name.")
	cmd.Flags().StringVarP(&options.sourceVolumeName, "sourcevol", "", options.sourceVolumeName, "Source name")
	cmd.Flags().StringVarP(&options.snapshotName, "snapname", "", options.snapshotName, "SnapShot name")
	cmd.MarkPersistentFlagRequired("volname")
	cmd.MarkPersistentFlagRequired("sourcevol")
	cmd.MarkPersistentFlagRequired("snapname")
	cmd.Flags().StringVarP(&options.size, "size", "", options.size,
		"volume capacity in GB (example: 10G) (default: 5G")
	return cmd
}

// RunVolumeClone makes create clone volume request to maya-apiserver .
func (c *CmdVolumeOptions) RunVolumeClone(cmd *cobra.Command) error {
	fmt.Println("Executing volume clone...")
	err := IsVolumeExist(c.volName)
	if err != nil {
		return err
	}
	resp := mapiserver.CreateCloneVolume(c.volName, c.size, c.snapshotName, c.sourceVolumeName)
	if resp != nil {
		return fmt.Errorf("Volume creation failed: %v", resp)
	}
	fmt.Printf("Volume successfully cloned:%v\n", c.volName)
	return nil
}
