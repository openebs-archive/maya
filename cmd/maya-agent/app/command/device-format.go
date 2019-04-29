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
	"github.com/spf13/cobra"
)

// NewSubCmdFormat formats the specified disk
func NewSubCmdFormat() *cobra.Command {
	var disk string
	var ftype string
	getCmd := &cobra.Command{
		Use:   "format",
		Short: "format disk",
		Long:  `the block devices on the storage area network can be formatted`,
		Run: func(cmd *cobra.Command, args []string) {

			block.Format(disk, ftype)

		},
	}
	getCmd.Flags().StringVar(&disk, "disk", "sdc",
		"disk name")
	getCmd.Flags().StringVar(&ftype, "type", "ext4",
		"formatting type")
	return getCmd
}
