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
	"github.com/spf13/cobra"
)

// NewCmdIscsi creates NewCmdBlockDevice and its nested children
func NewCmdIscsi() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iscsi",
		Short: "Operations using iscsi",
		Long: `The block devices on the local machine/minion can be 
		operated using maya-agent`,
	}
	//New sub command to list block device is added
	cmd.AddCommand(
		NewSubCmdIscsiDiscover(),
		NewSubCmdIscsiLogin(),
		NewSubCmdIscsiLogout(),
	)

	return cmd
}
