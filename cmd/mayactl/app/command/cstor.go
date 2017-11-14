// Copyright © 2017 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"flag"

	"github.com/openebs/maya/cmd/mayactl/app/command/cstor"
	"github.com/spf13/cobra"
)

// NewCommand creates the `maya` command and its nested children.
func NewCstorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cstor",
		Short: "cstor cli command",
		Long:  ``,
	}

	cmd.AddCommand(
		cstor.NewPoolCmd(),
		cstor.NewVolumeCmd(),
		cstor.NewSnapshotCmd(),
	)

	// add the glog flags
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// TODO: switch to a different logging library.
	flag.CommandLine.Parse([]string{})

	return cmd
}
