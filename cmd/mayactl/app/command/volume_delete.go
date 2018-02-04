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
	volumeDeleteCommandHelpText = `
	Usage: maya volume delete -volname <vol>

	This command initiate a delete for OpenEBS Volume.

	`
)

// CmdVolumeDeleteOptions stores the input parameters
type CmdVolumeDeleteOptions struct {
	volName string
}

// NewCmdVolumeDelete creates a new OpenEBS Volume
func NewCmdVolumeDelete() *cobra.Command {
	options := CmdVolumeDeleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes a Volume",
		Long:  volumeDeleteCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.ValidateVolumeDelete(cmd), util.Fatal)
			util.CheckErr(options.RunVolumeDelete(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")

	return cmd
}

//ValidateVolumeDelete validates the arguments passed
func (c *CmdVolumeDeleteOptions) ValidateVolumeDelete(cmd *cobra.Command) error {
	if c.volName == "" {
		return errors.New("--volname is missing. Please specify an unique name")
	}
	return nil
}

//RunVolumeDelete will initiate deletion of volume from maya-apiserver
func (c *CmdVolumeDeleteOptions) RunVolumeDelete(cmd *cobra.Command) error {
	fmt.Println("Executing volume delete...")

	resp := mapiserver.DeleteVolume(c.volName)
	if resp != nil {
		return fmt.Errorf("Error: %v", resp)
	}

	fmt.Printf("Volume deletion initiated:%v\n", c.volName)

	return nil
}
