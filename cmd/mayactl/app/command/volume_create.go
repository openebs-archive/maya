/*
Copyright 2017 The OpenEBS Authors.

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
	"errors"
	"fmt"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	mayav1 "github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

var (
	volumeCreateCommandHelpText = `
This command creates a new Volume.

Usage: mayactl volume create --volname <vol> [-size <size>]
`
)

// NewCmdVolumeCreate creates a new OpenEBS Volume
func NewCmdVolumeCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a new Volume",
		Long:  volumeCreateCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunVolumeCreate(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")

	cmd.Flags().StringVarP(&options.size, "size", "", options.size,
		"volume capacity in GB (example: 10G) (default: 5G")
	return cmd
}

// Validate verifies whether a volume name is provided or not followed by
// stats command. It returns nil and proceeds to execute the command if there is
// no error and returns an error if it is missing.
func (c *CmdVolumeOptions) Validate(cmd *cobra.Command) error {
	if len(c.volName) == 0 {
		return errors.New("--volname is missing. Please specify a unique name")
	}
	return nil
}

// RunVolumeCreate makes create volume request to maya-apiserver after verifying whether the volume already exists or not. In case if the volume already exists it returns the error and come out of execution.
func (c *CmdVolumeOptions) RunVolumeCreate(cmd *cobra.Command) error {
	fmt.Println("Executing volume create...")
	err := IsVolumeExist(c.volName)
	if err != nil {
		return err
	}
	resp := mapiserver.CreateVolume(c.volName, c.size)
	if resp != nil {
		return fmt.Errorf("Volume creation failed: %v", resp)
	}
	fmt.Printf("Volume Successfully Created:%v\n", c.volName)
	return nil
}

// IsVolumeExist checks whether the volume already exists or not
func IsVolumeExist(volname string) error {
	var vols mayav1.VolumeList
	err := mapiserver.ListVolumes(&vols)
	if err != nil {
		return err
	}
	for _, items := range vols.Items {
		if volname == items.ObjectMeta.Name {
			return fmt.Errorf("Error: Volume %v already exist ", volname)
		}
	}
	return nil
}
