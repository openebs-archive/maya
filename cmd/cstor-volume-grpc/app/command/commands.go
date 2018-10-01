/*
Copyright 2018 The OpenEBS Authors.

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

package command

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
	grpc_util "github.com/openebs/maya/pkg/grpc"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	cmdName = "cstor-volume-grpc"
	usage   = fmt.Sprintf("%s", cmdName)
)

// NewCmdOptions creates an options Cobra command to return usage.
func NewCmdOptions() *cobra.Command {
	cmd := &cobra.Command{
		Use: "options",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	return cmd
}

// NewCmdServer provides options for volume gRPC server functionality
func NewCmdServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Provides operations related to Volume gRPC server",
		Long:  "Provides operations related to Volume gRPC server",
	}

	cmd.AddCommand(
		NewCmdStart(),
	)
	return cmd
}

// NewCmdClient provides options for volume gRPC client functionality
func NewCmdClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Provides operations related to Volume gRPC client",
		Long:  "Provides operations related to Volume gRPC client",
	}

	cmd.AddCommand(
		NewCmdSnapshot(),
	)
	return cmd
}

// NewCmdSnapshot operates on volume snapshots
func NewCmdSnapshot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Provides operations related to snapshot of a Volume",
		Long:  "Provides operations related to snapshot of a Volume",
	}

	cmd.AddCommand(
		NewCmdSnapshotCreate(),
		NewCmdSnapshotDestroy(),
	)

	return cmd
}

// NewCmdSnapshotCreate creates a volume snapshot
func NewCmdSnapshotCreate() *cobra.Command {
	options := grpc_util.CmdSnaphotOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a new Snapshot",
		//Long:  SnapshotCreateCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotCreate(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.VolName, "volname", "n", options.VolName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")

	cmd.Flags().StringVarP(&options.SnapName, "snapname", "s", options.SnapName,
		"unique snapshot name")
	cmd.MarkPersistentFlagRequired("snapname")

	return cmd
}

// NewCmdSnapshotDestroy destroys a volume snapshot
func NewCmdSnapshotDestroy() *cobra.Command {
	options := grpc_util.CmdSnaphotOptions{}

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroys an existing Snapshot",
		Long:  "Destroys an existing Snapshot",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotDestroy(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.VolName, "volname", "n", options.VolName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")

	cmd.Flags().StringVarP(&options.SnapName, "snapname", "s", options.SnapName,
		"unique snapshot name")
	cmd.MarkPersistentFlagRequired("snapname")

	return cmd
}

// NewCStorVolumeGrpc creates a new CStorVolumeGrpc. This cmd includes logging,
// cmd option parsing from flags.
func NewCStorVolumeGrpc() (*cobra.Command, error) {
	// Create a new command.
	cmd := &cobra.Command{
		Use:   usage,
		Short: "CStor Volume gRPC",
		Long: `interfaces between external grpc and the CStorVolume
		 objects and snapshot creation`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd), util.Fatal)
		},
	}
	cmd.AddCommand(
		NewCmdServer(),
		NewCmdClient(),
		// NewCmdStart(),
	)

	// add the glog flags
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// TODO: switch to a different logging library.
	flag.CommandLine.Parse([]string{})

	return cmd, nil
}

// Run is to CStorVolumeGrpc.
func Run(cmd *cobra.Command) error {
	glog.Infof("cstor-volume-grpc for CStorVolume objects")
	return nil
}
