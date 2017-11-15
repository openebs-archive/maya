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
	mtypesv1 "github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

var (
	volumesListCommandHelpText = `
	Usage: maya volume list [options]

	This command displays status of available Volumes.
	If no volume ID is given, a list of all known volume will be dumped.
	`
)

// CmdVolumesListOptions captures the CLI flags
type CmdVolumesListOptions struct {
	volName string
}

// NewCmdVolumesList display status of OpenEBS Volume(s)
func NewCmdVolumesList() *cobra.Command {
	options := CmdVolumesListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display status information about Volume(s)",
		Long:  volumesListCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.RunVolumesList(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")

	return cmd
}

//RunVolumesList will fetch the volumes from maya-apiserver
func (c *CmdVolumesListOptions) RunVolumesList(cmd *cobra.Command) error {
	//fmt.Println("Executing volume list...")

	var vsms mtypesv1.VolumeList
	err := mapiserver.ListVolumes(&vsms)
	if err != nil {
		return err
	}

	out := make([]string, len(vsms.Items)+1)
	out[0] = "Name|Status"
	for i, items := range vsms.Items {
		if items.Status.Reason == "" {
			items.Status.Reason = "Running"
		}
		out[i+1] = fmt.Sprintf("%s|%s",
			items.ObjectMeta.Name,
			items.Status.Reason)
	}
	if len(out) == 1 {
		fmt.Println("No Volumes are running")
		return nil
	}
	fmt.Println(util.FormatList(out))
	return nil
}
