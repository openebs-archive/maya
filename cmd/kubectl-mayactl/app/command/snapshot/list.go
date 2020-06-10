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

package snapshot

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	snapshotListCommandHelpText = `
This command displays available snapshots on a volume.

Usage: mayactl snapshot list [options]

$ mayactl snapshot list --volname <vol>
`
)

// NewCmdSnapshotList displays list of volumes
func NewCmdSnapshotList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all the snapshots of a Volume",
		Long:  snapshotListCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.ValidateList(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotList(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")
	return cmd
}

// ValidateList validates the flag values
func (c *CmdSnaphotOptions) ValidateList(cmd *cobra.Command) error {
	if len(c.volName) == 0 {
		return errors.New("--volname is missing. Please specify an unique name")
	}
	return nil
}

// RunSnapshotList makes snapshot-list API request to maya-apiserver
func (c *CmdSnaphotOptions) RunSnapshotList(cmd *cobra.Command) error {

	resp := mapiserver.ListSnapshot(c.volName, c.namespace)
	if resp != nil {
		return fmt.Errorf("Error list available snapshot: %v", resp)
	}
	return nil
}
