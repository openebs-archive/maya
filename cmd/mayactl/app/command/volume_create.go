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
	"github.com/spf13/cobra"
)

var (
	volumeCreateCommandHelpText = `
	Usage: maya volume create -volname <vol> [-size <size>]

	This command creates a new Volume.

	Volume create options:
	-size
	Provisioning size of the volume(default is 5G)

	`
)

type CmdVolumeCreateOptions struct {
	volName string
	size    string
}

// NewCmdVolumeCreate creates a new OpenEBS Volume
func NewCmdVolumeCreate() *cobra.Command {
	options := CmdVolumeCreateOptions{}

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

// Run does tasks related to mayaserver.
func (c *CmdVolumeCreateOptions) Validate(cmd *cobra.Command) error {
	if c.volName == "" {
		return errors.New("--volname is missing. Please specify an unique name")
	}
	return nil
}

// Run does tasks related to mayaserver.
func (c *CmdVolumeCreateOptions) RunVolumeCreate(cmd *cobra.Command) error {
	fmt.Println("Executing volume create...")

	resp := mapiserver.CreateVolume(c.volName, c.size)
	if resp != nil {
		return fmt.Errorf("Error: %v", resp)
	}

	fmt.Printf("Volume Successfully Created:%v\n", c.volName)

	return nil
}
