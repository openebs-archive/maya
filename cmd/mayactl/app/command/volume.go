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
	"flag"
	"github.com/spf13/cobra"
)

var (
	volumeCommandHelpText = `
The following commands helps in operating a Volume such as create, list, and so on.

Usage: mayactl volume <subcommand> [options] [args]

Examples:

 # Create a Volume:
   $ mayactl volume create --volname <vol> --size <size>

 # List Volumes:
   $ mayactl volume list

 # Delete a Volume:
   $ mayactl volume delete --volname <vol>

 # Delete a Volume created in 'test' namespace:
   $ mayactl volume delete --volname <vol> --namespace test

 # Statistics of a Volume:
   $ mayactl volume stats --volname <vol>

 # Statistics of a Volume created in 'test' namespace:
   $ mayactl volume stats --volname <vol> --namespace test

 # Info of a Volume:
   $ mayactl volume info --volname <vol>

 # Info of a Volume created in 'test' namespace:
   $ mayactl volume info --volname <vol> --namespace test
`
	options = &CmdVolumeOptions{
		namespace: "default",
	}
)

// CmdVolumeOptions stores information of volume being operated
type CmdVolumeOptions struct {
	volName          string
	sourceVolumeName string
	snapshotName     string
	size             string
	namespace        string
	json             string
}

// NewCmdVolume provides options for managing OpenEBS Volume
func NewCmdVolume() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Provides operations related to a Volume",
		Long:  volumeCommandHelpText,
	}

	cmd.AddCommand(
		NewCmdVolumeCreate(),
		NewCmdVolumesList(),
		NewCmdVolumeDelete(),
		NewCmdVolumeStats(),
		NewCmdVolumeInfo(),
	)
	cmd.PersistentFlags().StringVarP(&options.namespace, "namespace", "n", options.namespace,
		"namespace name, required if volume is not in the default namespace")

	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse([]string{})
	return cmd
}

// Validate verifies whether a volume name,source name or snapshot name is provided or not followed by
// stats command. It returns nil and proceeds to execute the command if there is
// no error and returns an error if it is missing.
func (c *CmdVolumeOptions) Validate(cmd *cobra.Command, snapshotnameverify, sourcenameverify, volnameverify bool) error {
	if snapshotnameverify {
		if len(c.snapshotName) == 0 {
			return errors.New("--snapname is missing. Please provide a snapshotname")
		}
	}
	if sourcenameverify {
		if len(c.sourceVolumeName) == 0 {
			return errors.New("--sourcevol is missing. Please specify a sourcevolumename")
		}
	}
	if volnameverify {
		if len(c.volName) == 0 {
			return errors.New("--volname is missing. Please specify a unique volumename")
		}
	}
	return nil
}
