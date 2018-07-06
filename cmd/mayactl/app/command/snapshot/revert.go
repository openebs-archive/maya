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
	"fmt"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	snapshotrevertHelpText = `
This command rolls back volume data to the specified snapshot. Once the roll back
to snapshot is successful, all data changes made after the snapshot was taken will
be posted. This command should be used cautiously and only if there is an issue with
the current state of data.

Usage: mayactl snapshot revert [options]

$ mayactl snapshot revert --volname <vol> --snapname <snap>

`
)

// NewCmdSnapshotRevert reverts a snapshot of OpenEBS Volume
func NewCmdSnapshotRevert() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revert",
		Short: "Reverts to specific snapshot of a Volume",
		Long:  snapshotrevertHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotRevert(cmd), util.Fatal)
		},
	}
	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")
	cmd.Flags().StringVarP(&options.snapName, "snapname", "s", options.snapName,
		"unique snapshot name")
	cmd.MarkPersistentFlagRequired("snapname")
	return cmd
}

// RunSnapshotRevert does tasks related to mayaserver.
func (c *CmdSnaphotOptions) RunSnapshotRevert(cmd *cobra.Command) error {
	fmt.Println("Executing volume snapshot revert ...")

	resp := mapiserver.RevertSnapshot(c.volName, c.snapName, c.namespace)
	if resp != nil {
		return fmt.Errorf("Snapshot revert failed: %v", resp)
	}
	fmt.Printf("Reverting to snapshot [%s] of volume [%s]\n", c.snapName, c.volName)
	return nil
}
