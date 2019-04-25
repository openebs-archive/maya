// Copyright © 2019 The OpenEBS Authors
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
	"github.com/openebs/maya/cmd/maya-agent/storage/block"
	"github.com/openebs/maya/cmd/maya-agent/types/v1"
	"github.com/spf13/cobra"
)

// NewSubCmdListBlockDevice is to list block device is created
func NewSubCmdListBlockDevice() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "list",
		Short: "List block devices",
		Long: `the set of block devices on the local machine/minion
		can be listed`,
		Run: func(cmd *cobra.Command, args []string) {
			//resJsonDecoded is the decoded value of block disk
			var resJsonDecoded v1.BlockDeviceInfo
			err := block.ListBlockExec(&resJsonDecoded)
			if err != nil {
				panic(err)
			}
			//to print after formatting to end user
			block.FormatOutputForUser(&resJsonDecoded)

		},
	}

	return getCmd
}
