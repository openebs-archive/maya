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
	snapshotCreateCommandHelpText = `
This command creates a new snapshot.

Usage: mayactl snapshot create [options]

$ mayactl snapshot create --volname <vol> --snapname <snap>
`
)

// NewCmdSnapshotCreate creates a snapshot of OpenEBS Volume
func NewCmdSnapshotCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a new Snapshot",
		Long:  snapshotCreateCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotCreate(cmd), util.Fatal)
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

// Validate validates the flag values
func (c *CmdSnaphotOptions) Validate(cmd *cobra.Command) error {
	if len(c.volName) == 0 {
		return errors.New("--volname is missing. Please specify an unique name")
	}
	if len(c.snapName) == 0 {
		return errors.New("--snapname is missing. Please specify an unique name")
	}

	return nil
}

// RunSnapshotCreate does tasks related to mayaserver.
func (c *CmdSnaphotOptions) RunSnapshotCreate(cmd *cobra.Command) error {
	fmt.Println("Executing volume snapshot create...")

	resp := mapiserver.CreateSnapshot(c.volName, c.snapName, c.namespace)
	if resp != nil {
		return fmt.Errorf("Snapshot creation failed: %v", resp)
	}

	fmt.Printf("Volume snapshot created for volume %s : '%s'\n", c.volName, c.snapName)
	return nil
}
