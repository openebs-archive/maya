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
	"html/template"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

// VolumeList struct holds the volume's information like status and volume type.
type VolumeList struct {
	Namespace  string
	Name       string
	Status     string
	VolumeType string
}

const (
	// listTemplate is used for formating the list output.
	listTemplate = `
{{ printf "NAMESPACE\t NAME\t STATUS\t TYPE" }}
{{ printf "---------\t ----\t ------\t ----" }} {{range $key,$value := .}}
{{ printf "%v\t" $value.Namespace }} {{ printf "%v\t" $value.Name }} {{ printf "%s\t" $value.Status }} {{ printf "%s" $value.VolumeType }} {{end}}

`
)

var (
	volumesListCommandHelpText = `
This command displays status of available Volumes.
If no volume ID is given, a list of all known volumes will be displayed.

Usage: mayactl volume list [options]
	`
)

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

//RunVolumesList fetches the volumes from maya-apiserver
func (c *CmdVolumeOptions) RunVolumesList(cmd *cobra.Command) error {

	// Calling to m-api service to fetch volume list.
	cvols, err := mapiserver.ListVolumes()
	if err != nil {
		return fmt.Errorf("Volume list error: %s", err)
	}

	// Create a slice of VolumeList struct that can be binded to template.
	volumes := []VolumeList{}

	// Filling the slice with required fields.
	for _, vol := range cvols.Items {
		volume := VolumeInfo{
			Volume: vol,
		}
		volumes = append(volumes, VolumeList{
			Namespace:  volume.GetVolumeNamespace(),
			Name:       volume.GetVolumeName(),
			Status:     volume.GetVolumeStatus(),
			VolumeType: strings.Title(volume.GetCASType()),
		})
	}

	// Check for volumes length.
	if len(volumes) == 0 {
		fmt.Println("No volumes found")
		return nil
	}

	// Creating template instance and parsing structure into it.
	volumeListTemplate, err := template.New("VolumeList").Parse(listTemplate)
	if err != nil {
		fmt.Println("Error displaying output, found error:", err)
		return nil
	}

	// Creating tabwriter instance and executing it with template.
	w := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', 0)
	err = volumeListTemplate.Execute(w, volumes)
	if err != nil {
		fmt.Println("Error displaying output, found error:", err)
	}

	// Flushing the tabwriter instance.
	err = w.Flush()
	if err != nil {
		fmt.Println("Error displaying output, found error:", err)
	}

	return nil
}
