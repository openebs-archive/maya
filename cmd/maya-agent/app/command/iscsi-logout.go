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
	"github.com/openebs/maya/cmd/maya-agent/storage/iscsi"
	"github.com/spf13/cobra"
)

//NewSubCmdIscsiLogout logs out of particular portal or all discovered portals
func NewSubCmdIscsiLogout() *cobra.Command {
	var target string
	getCmd := &cobra.Command{
		Use:   "logout",
		Short: "logout of block devices with iscsi",
		Long:  `Single and multiple logout to set of block devices on the storage area network `,
		Run: func(cmd *cobra.Command, args []string) {

			iscsi.IscsiLogout(target)

		},
	}
	getCmd.Flags().StringVar(&target, "portal", "127.0.0.1",
		"target portal-ip to iscsi logout, 'all' to logout all")

	return getCmd
}
