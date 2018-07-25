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
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-volume-grpc/api"
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
		return errors.New("--volname is missing. Please specify an unique name")
	}
	if c.snapName == "" {
		return errors.New("--snapname is missing. Please specify an unique name")
	}

	return nil
}

//CreateSnapshot creates snapshots
func CreateSnapshot(volName, snapName string) (*api.VolumeCommand, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", api.VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := api.NewRunCommandClient(conn)
	response, err := c.RunVolumeCommand(context.Background(),
		&api.VolumeCommand{
			Command:  api.CmdSnapCreate,
			Volume:   volName,
			Snapname: snapName,
			Status:   "requesting",
		})

	if err != nil {
		log.Fatalf("Error when calling RunVolumeCommand: %s", err)
	}
	log.Printf("Response from server: %s, %s, %s, %s",
		response.Command, response.Volume, response.Snapname, response.Status)
	return response, err
}

//DestroySnapshot destroys snapshots
func DestroySnapshot(volName, snapName string) (*api.VolumeCommand, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", api.VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := api.NewRunCommandClient(conn)
	response, err := c.RunVolumeCommand(context.Background(),
		&api.VolumeCommand{
			Command:  api.CmdSnapDestroy,
			Volume:   volName,
			Snapname: snapName,
			Status:   "requesting",
		})

	if err != nil {
		log.Fatalf("Error when calling RunVolumeCommand: %s", err)
	}
	log.Printf("Response from server: %s, %s, %s, %s",
		response.Command, response.Volume, response.Snapname, response.Status)
	return response, err
}

// RunSnapshotCreate does tasks related to grpc snapshot create.
func (c *CmdSnaphotOptions) RunSnapshotCreate(cmd *cobra.Command) error {
	fmt.Println("Executing volume snapshot create...")
	resp, err := CreateSnapshot(c.volName, c.snapName)
	if err != nil {
		return fmt.Errorf("Snapshot create failed: %v", err)
	}
	if resp != nil && strings.Contains(resp.Status, "ERR") {
		return fmt.Errorf("Snapshot create failed with error status: %v", resp.Status)
	}

	fmt.Printf("Volume Snapshot Successfully Created:%v@%v\n", c.volName, c.snapName)
	return nil
}

//RunSnapshotDestroy will initiate deletion of snapshot
func (c *CmdSnaphotOptions) RunSnapshotDestroy(cmd *cobra.Command) error {
	fmt.Println("Executing snapshot destroy...")
	resp, err := DestroySnapshot(c.volName, c.snapName)
	if err != nil {
		return fmt.Errorf("Error: %v", resp)
	}

	if resp != nil && strings.Contains(resp.Status, "ERR") {
		return fmt.Errorf("Snapshot deletion failed with error status: %v", resp.Status)
	}

	fmt.Printf("Snapshot deletion initiated:%v@%v\n", c.volName, c.snapName)

	return nil
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
	cmd.MarkPersistentFlagRequired("snapname")

	cmd.Flags().StringVarP(&options.snapName, "snapname", "s", options.snapName,
		"unique snapshot name")

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
	cmd.MarkPersistentFlagRequired("snapname")

	cmd.Flags().StringVarP(&options.snapName, "snapname", "s", options.snapName,
		"unique snapshot name")

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
