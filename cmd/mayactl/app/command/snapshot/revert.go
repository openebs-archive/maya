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

/*type CmdSnaphotCreateOptions struct {
	volName  string
	snapName string
}*/

// NewCmdSnapshotRevert reverts a snapshot of OpenEBS Volume
func NewCmdSnapshotRevert() *cobra.Command {
	options := CmdSnaphotCreateOptions{}

	cmd := &cobra.Command{
		Use:   "revert",
		Short: "Reverts to specific snapshot of a Volume",
		Long:  "Reverts to specific snapshot of a Volume",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotRevert(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "n", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")
	cmd.MarkPersistentFlagRequired("snapname")

	cmd.Flags().StringVarP(&options.snapName, "snapname", "s", options.snapName,
		"unique snapshot name")

	return cmd
}

// RunSnapshotRevert does tasks related to mayaserver.
func (c *CmdSnaphotCreateOptions) RunSnapshotRevert(cmd *cobra.Command) error {
	fmt.Println("Executing volume snapshot revert ...")

	resp := mapiserver.RevertSnapshot(c.volName, c.snapName)
	if resp != nil {
		return fmt.Errorf("Error: %v", resp)
	}

	fmt.Printf("Reverting to snapshot [%s] of volume [%s]\n", c.snapName, c.volName)
	return nil
}
