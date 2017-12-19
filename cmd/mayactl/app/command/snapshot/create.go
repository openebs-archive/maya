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

/*func init() {
	host := os.Getenv("MAPI_ADDR")
	port := os.Getenv("MAPI_PORT")
	defaultEndpoint := fmt.Sprintf("%s:%s", host, port)
	if host == "" || port == "" {
		fmt.Println("$MAPI_ADDR or $MAPI_ADDR are not set. Check if the maya-apiserver is running.")
		defaultEndpoint = ""
	}

	cmd.PersistentFlags().StringVar(&APIServerEndpoint, "api-server-endpoint", defaultEndpoint, "IP endpoint of API server instance (required)")
	cmd.PersistentFlags().StringVar(&logLevelRaw, "log-level", "WARNING", "logging level for logging/tracing output (valid values: CRITICAL,ERROR,WARNING,NOTICE,INFO,DEBUG,TRACE)")

	cmd.MarkFlagRequired("api-server-endpoint")

	// load the environment variables
	//flags.SetFlagsFromEnv(cmd.PersistentFlags(), "MAYA")
}
*/

// CmdSnaphotCreateOptions holds the options for snapshot
// create command
type CmdSnaphotCreateOptions struct {
	volName  string
	snapName string
}

// NewCmdSnapshotCreate creates a snapshot of OpenEBS Volume
func NewCmdSnapshotCreate() *cobra.Command {
	options := CmdSnaphotCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a new Snapshot",
		//Long:  SnapshotCreateCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotCreate(cmd), util.Fatal)
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

// Validate validates the flag values
func (c *CmdSnaphotCreateOptions) Validate(cmd *cobra.Command) error {
	if c.volName == "" {
		return errors.New("--volname is missing. Please specify an unique name")
	}
	if c.snapName == "" {
		return errors.New("--snapname is missing. Please specify an unique name")
	}

	return nil
}

// RunSnapshotCreate does tasks related to mayaserver.
func (c *CmdSnaphotCreateOptions) RunSnapshotCreate(cmd *cobra.Command) error {
	fmt.Println("Executing volume snapshot create...")

	resp := mapiserver.CreateSnapshot(c.volName, c.snapName)
	if resp != nil {
		return fmt.Errorf("Snapshot create failed: %v", resp)
	}

	fmt.Printf("Volume snapshot Successfully Created:%v\n", c.volName)

	return nil
}
