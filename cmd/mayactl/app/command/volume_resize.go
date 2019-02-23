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
	"regexp"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	volumeResizeCommandHelpText = `
This command resizes the volume.

Usage: mayactl volume resize [options]

$ mayactl volume resize --volume <vol> --size <size> --namespace <namespace>
  Supported units: M, Mi, G, Gi, T, Ti, P, Pi, E, Ei, Z, Zi

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
		return errors.New("--volname is missing. Please specify the name of the volume to be resized")
	}
	if len(c.size) == 0 {
		return errors.New("--size is missing. Please specify value")
	}

	// TODO: Below validation is need to be moved into maya-apiserver
	// Regex to say only positive integers and valid size is accepted
	reg, err := regexp.Compile("^[0-9]+[MGTPEZ][i]{0,1}$")
	if err != nil {
		return errors.New("failed to process regular expresion")
	}

	if !reg.MatchString(c.size) {
		return errors.New("Please provide valid size and unit")
	}
	return nil
}

// RunVolumeResize will initiate the process of resizing a volume from maya-apiserver
func (c *CmdVolumeOptions) RunVolumeResize(cmd *cobra.Command) error {
	fmt.Println("Executing the volume resize...")

	resp := mapiserver.ResizeVolume(c.volName, c.size, c.namespace)
	if resp != nil {
		return fmt.Errorf("Volume resize failed: %v", resp)
	}

	fmt.Printf("Volume resize is successfull on volume: %s with size: %s in %s namespace\n", c.volName, c.size, c.namespace)
	return nil
}
