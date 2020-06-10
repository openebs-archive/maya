/*
Copyright 2018 The OpenEBS Authors.

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

package pool

import (
	"github.com/spf13/cobra"
)

var (
	poolCommandHelpText = `
Command provides operations related to a storage pools.

Usage: mayactl pool <subcommand> [options] [args]

Examples:
  # Lists pool:
    $ mayactl pool list 
`

	options = &CmdPoolOptions{}
)

// CmdPoolOptions holds information of pool being operated
type CmdPoolOptions struct {
	poolName string
}

// NewCmdPool adds command for operating on snapshot
func NewCmdPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool",
		Short: "Provides operations related to a storage pool",
		Long:  poolCommandHelpText,
	}

	cmd.AddCommand(
		NewCmdPoolList(),
		NewCmdPoolDescribe(),
	)
	return cmd
}
