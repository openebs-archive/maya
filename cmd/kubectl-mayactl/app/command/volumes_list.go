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

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
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
	//fmt.Println("Executing volume list...")

	var cvols v1alpha1.CASVolumeList
	err := mapiserver.ListVolumes(&cvols)
	if err != nil {
		return fmt.Errorf("Volume list error: %s", err)
	}

	out := make([]string, len(cvols.Items)+2)
	out[0] = "Namespace|Name|Status|Type|Capacity|StorageClass|Access Mode"
	out[1] = "---------|----|------|----|--------|-------------|-----------"
	for i, item := range cvols.Items {
		if len(item.Status.Reason) == 0 {
			item.Status.Reason = volumeStatusOK
		}
		out[i+2] = fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", item.ObjectMeta.Namespace,
			item.ObjectMeta.Name,
			item.Status.Reason, item.Spec.CasType, item.Spec.Capacity, item.ObjectMeta.Annotations["openebs.io/storage-class"], item.Spec.AccessMode)
	}
	if len(out) == 2 {
		fmt.Println("No Volumes are running")
		return nil
	}
	fmt.Println(util.FormatList(out))
	return nil
}
