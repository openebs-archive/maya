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
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-volume-grpc/api"
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-grpc/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	cmdName = "cstor-volume-grpc"
	usage   = fmt.Sprintf("%s", cmdName)
)

// CmdSnaphotOptions holds the options for snapshot
// create command
type CmdSnaphotOptions struct {
	volName  string
	snapName string
}

// Validate validates the flag values
func (c *CmdSnaphotOptions) Validate(cmd *cobra.Command) error {
	if c.volName == "" {
		return errors.New("--volname is missing. Please specify a unique name")
	}
	if c.snapName == "" {
		return errors.New("--snapname is missing. Please specify a unique name")
	}

	return nil
}

//CreateSnapshot creates snapshots
func CreateSnapshot(volName, snapName string) (*v1alpha1.VolumeSnapCreateResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", api.VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("Unable to dial gRPC server on port %d error : %s", api.VolumeGrpcListenPort, err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunSnapCommandClient(conn)
	response, err := c.RunVolumeSnapCreateCommand(context.Background(),
		&v1alpha1.VolumeSnapCreateRequest{
			Version:  api.ProtocolVersion,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		glog.Fatalf("Error when calling RunVolumeSnapCreateCommand: %s", err)
	}

	if response != nil {
		var responseStatus api.CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return response, fmt.Errorf("Snapshot create failed with error : %v", responseStatus.Response)
		}

	}
	return response, err
}

//DestroySnapshot destroys snapshots
func DestroySnapshot(volName, snapName string) (*v1alpha1.VolumeSnapDeleteResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", api.VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunSnapCommandClient(conn)
	response, err := c.RunVolumeSnapDeleteCommand(context.Background(),
		&v1alpha1.VolumeSnapDeleteRequest{
			Version:  api.ProtocolVersion,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		glog.Fatalf("Error when calling RunVolumeSnapDeleteCommand: %s", err)
	}
	if response != nil {
		var responseStatus api.CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return response, fmt.Errorf("Snapshot deletion failed with error : %v", responseStatus.Response)
		}

	}
	return response, err
}

// RunSnapshotCreate does tasks related to grpc snapshot create.
func (c *CmdSnaphotOptions) RunSnapshotCreate(cmd *cobra.Command) error {
	glog.Info("Executing volume snapshot create...")
	response, err := CreateSnapshot(c.volName, c.snapName)
	if response != nil {
		glog.Infof("Response from server: %s", response.Status)
		if err == nil {
			glog.Infof("Volume Snapshot Successfully Created:%v@%v\n", c.volName, c.snapName)
		}

	}

	return err
}

//RunSnapshotDestroy will initiate deletion of snapshot
func (c *CmdSnaphotOptions) RunSnapshotDestroy(cmd *cobra.Command) error {
	glog.Info("Executing snapshot destroy...")
	response, err := DestroySnapshot(c.volName, c.snapName)
	if response != nil {
		glog.Infof("Response from server: %s", response.Status)
		if err == nil {
			glog.Infof("Snapshot deletion initiated:%v@%v\n", c.volName, c.snapName)
		}
	}

	return err
}

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
	options := CmdSnaphotOptions{}

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

	cmd.Flags().StringVarP(&options.snapName, "snapname", "s", options.snapName,
		"unique snapshot name")
	cmd.MarkPersistentFlagRequired("snapname")

	return cmd
}

// NewCmdSnapshotDestroy destroys a volume snapshot
func NewCmdSnapshotDestroy() *cobra.Command {
	options := CmdSnaphotOptions{}

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroys an existing Snapshot",
		Long:  "Destroys an existing Snapshot",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotDestroy(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "n", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")

	cmd.Flags().StringVarP(&options.snapName, "snapname", "s", options.snapName,
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
