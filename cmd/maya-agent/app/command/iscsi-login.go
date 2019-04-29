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

//NewSubCmdIscsiLogin logs in to particular portal or all discovered portals
func NewSubCmdIscsiLogin() *cobra.Command {
	var target string
	getCmd := &cobra.Command{
		Use:   "login",
		Short: "iscsi login to block devices",
		Long:  `Single and multiple login to set of block devices on the storage area network `,
		Run: func(cmd *cobra.Command, args []string) {

			iscsi.IscsiLogin(target)

		},
	}

	getCmd.Flags().StringVar(&target, "portal", "10.107.180.120",
		"target portal-ip to iscsi login, 'all' to login all")
	return getCmd
}
