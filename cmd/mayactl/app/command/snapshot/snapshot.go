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

type CmdSnaphotOptions struct {
	volName   string
	snapName  string
	namespace string
}

func NewCmdSnapshot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Provides operations related to snapshot of a Volume",
		Long:  "Provides operations related to snapshot of a Volume",
	}

	cmd.AddCommand(
		NewCmdSnapshotCreate(),
		NewCmdSnapshotList(),
		NewCmdSnapshotRevert(),
	)
	cmd.PersistentFlags().StringVarP(&options.namespace, "namespace", "n", options.namespace,
		"namespace name, required if volume is in other then dafault namespace")

	return cmd
}
