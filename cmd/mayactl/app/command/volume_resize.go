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

package command

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/pkg/validation"
	"github.com/spf13/cobra"
)

var (
	volumeResizeCommandHelpText = `
This command resizes the volume.

Usage: mayactl volume resize [options]

$ mayactl volume resize --volume <vol> --size <size> --namespace <namespace>
  Supported units: M, Mi, G, Gi, T, Ti, P, Pi, E, Ei

`
)

// NewCmdVolumeResize resizes a existing OpenEBS Volume
func NewCmdVolumeResize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resizes the volume",
		Long:  volumeResizeCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.ValidateResize(cmd), util.Fatal)
			util.CheckErr(options.RunVolumeResize(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name")
	cmd.MarkPersistentFlagRequired("volname")
	cmd.Flags().StringVarP(&options.size, "size", "", options.size,
		"expanding size")
	cmd.MarkPersistentFlagRequired("size")
	return cmd
}

// ValidateResize validates the flag values
func (c *CmdVolumeOptions) ValidateResize(cmd *cobra.Command) error {
	if len(c.volName) == 0 {
		return errors.New("--volname is missing. Please specify the name of the volume to resize")
	}
	if len(c.size) == 0 {
		return errors.New("--size is missing. Please specify value")
	}

	// Validate capacity
	isValid, err := validation.ValidateString(c.size, "^[0-9]+[MGTPE][i]{0,1}$")
	if err != nil {
		return err
	}
	if !isValid {
		return errors.New("invalid size. please specify valid size and size must match to regular expression '^[0-9]+[MGTPE][i]{0,1}$'")
	}
	return nil
}

// RunVolumeResize will initiate the process of resizing a volume from maya-apiserver
func (c *CmdVolumeOptions) RunVolumeResize(cmd *cobra.Command) error {
	fmt.Println("Executing the volume resize...")

	resp := mapiserver.ResizeVolume(c.volName, c.size, c.namespace)
	if resp != nil {
		return fmt.Errorf("Volume resize failed: '%v'", resp)
	}

	fmt.Printf("Volume resize successfull for volume: '%s' with size: '%s' in '%s' namespace\n", c.volName, c.size, c.namespace)
	return nil
}
