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
	"github.com/spf13/cobra"
)

var (
	options = &CmdSnaphotOptions{
		namespace: "default",
	}
)

// CmdSnaphotOptions holds information of snapshots being operated
type CmdSnaphotOptions struct {
	volName   string
	snapName  string
	namespace string
}

var (
	snapshotCommandHelpText = `
Command provides operations related to a volume snapshot.

If you have specified an OpenEBS volume in a namespace other than the 'default',
you must use --namespace flag with a command. If not, you will see the results only
in the 'default' namespace.

Usage: mayactl snapshot <subcommand> [options] [args]

Examples:
  # Create a snapshot:
    $ mayactl snapshot create --volname <vol> --snapname <snap>

  # Create a snapshot for a volume created in 'test' namespace
    $ mayactl snapshot create --volname <vol> --snapname <snap> --namespace test

  # Lists snapshot:
    $ mayactl snapshot list --volname <vol>

  # Lists snapshots for a volume created in 'test' namespace
    $ mayactl snapshot list --volname <vol> --namespace test

  # Reverts a snapshot:
    $ mayactl snapshot revert --volname <vol> --snapname <snap>

  # Revert a snapshot for a volume created in 'test' namespace
    $ mayactl snapshot revert --volname <vol> --snapname <snap> --namespace test
`
)

// NewCmdSnapshot adds command for operating on snapshot
func NewCmdSnapshot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Provides operations related to a Volume snapshot",
		Long:  snapshotCommandHelpText,
	}

	cmd.AddCommand(
		NewCmdSnapshotCreate(),
		NewCmdSnapshotList(),
		NewCmdSnapshotRevert(),
	)
	cmd.PersistentFlags().StringVarP(&options.namespace, "namespace", "n", options.namespace,
		"namespace name, required if volume is not in the default namespace")

	return cmd
}
