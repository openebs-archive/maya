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
	"fmt"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	volumeDeleteCommandHelpText = `
This command initiates a deletion process for an OpenEBS Volume.

Usage: kubectl-mayactl volume delete --volname <vol>
`
)

// NewCmdVolumeDelete creates a new OpenEBS Volume
func NewCmdVolumeDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes a Volume",
		Long:  volumeDeleteCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd, false, false, true), util.Fatal)
			util.CheckErr(options.RunVolumeDelete(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")
	return cmd
}

//RunVolumeDelete will initiate the process of deleting a volume from maya-apiserver
func (c *CmdVolumeOptions) RunVolumeDelete(cmd *cobra.Command) error {
	fmt.Println("Executing volume delete...")

	resp := mapiserver.DeleteVolume(c.volName, c.namespace)
	if resp != nil {
		return fmt.Errorf("Volume deletion failed: %v", resp)
	}

	fmt.Printf("Volume deletion initiated:%v\n", c.volName)

	return nil
}
