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
	volumesListCommandHelpText = `
This command displays status of available Volumes.
If no volume ID is given, a list of all known volumes will be displayed.

Usage: mayactl volume list [options]
	`
)

const (
	volumeListTemplate = `
{{ printf "NAMESPACE\t NAME\t STATUS\t TYPE" }}
{{ printf "---------\t ----\t ------\t ----" }} {{range $key,$value := .}}
{{ printf "%v\t" $value.Namespace }} {{ printf "%v\t" $value.Name }} {{ printf "%s\t" $value.Status }} {{ printf "%s" $value.VolumeType }} {{end}}

`
)

// VolumeList struct holds the volume's information like status and volume type.
type VolumeList struct {
	Namespace  string
	Name       string
	Status     string
	VolumeType string
}

// NewCmdVolumesList displays status of OpenEBS Volume(s)
func NewCmdVolumesList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Displays status information about Volume(s)",
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

//RunVolumesList fetchs the volumes from maya-apiserver
func (c *CmdVolumeOptions) RunVolumesList(cmd *cobra.Command) error {
	// Call to m-api service to fetch volume list.
	cvols, err := mapiserver.ListVolumes()
	if err != nil {
		CheckError(err)
	}

	// Create and Fill the slice with required fields after process
	volumes := []VolumeList{}
	for index, volume := range cvols.Items {
		cvols.Items[index], err = processCASVolume(volume, false)
		if err != nil {
			CheckError(err)
		}
		volumes = append(volumes, VolumeList{
			Namespace:  cvols.Items[index].ObjectMeta.Namespace,
			Name:       cvols.Items[index].ObjectMeta.Name,
			Status:     cvols.Items[index].Status.Reason,
			VolumeType: cvols.Items[index].Spec.CasType,
		})
	}

	// Check for volumes length.
	if len(volumes) == 0 {
		fmt.Println("No volumes found")
		return nil
	}

	renderTemplate("VolumeList", volumeListTemplate, volumes)
	return nil
}
