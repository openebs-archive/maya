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

	"github.com/openebs/maya/cmd/mayactl/app/command/snapshot"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/spf13/cobra"
)

// Variable to capture value from --apiserver flag
var apiserver *string

// NewCommand creates the `maya` command and its nested children.
func NewMayaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mayactl",
		Short: "Maya means 'Magic'a tool for storage orchestration",
		Long:  `Maya means 'Magic' a tool for storage orchestration`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		// PersistentPreRun is used so that it is inherited in all child commands
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			flagValue := cmd.Flag("apiserver").Value.String()
			if !(flagValue == "") {
				mapiserver.SetFlag(flagValue)
			}
		},
	}
	cmd.AddCommand(
		NewCmdVersion(),
		NewCmdVolume(),
		snapshot.NewCmdSnapshot(),
	)
	// Register --apiserver flag to mayactl command with persistence to all child commands
	apiserver = cmd.PersistentFlags().StringP("apiserver", "a", "", "IP to connect to maya server[format:scheme://apiserverIP:port]")
	// add the glog flags
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// TODO: switch to a different logging library.
	return cmd
}
