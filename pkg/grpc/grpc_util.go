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
package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-grpc/v1alpha1"
	"github.com/openebs/maya/pkg/grpc/api"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// CmdSnaphotOptions holds the options for snapshot
// create command
type CmdSnaphotOptions struct {
	VolName  string
	SnapName string
}

// Validate validates the flag values
func (c *CmdSnaphotOptions) Validate(cmd *cobra.Command) error {
	if c.VolName == "" {
		return errors.New("--volname is missing. Please specify a unique name")
	}
	if c.SnapName == "" {
		return errors.New("--snapname is missing. Please specify a unique name")
	}

	return nil
}

//CreateSnapshot creates snapshots
func CreateSnapshot(volName, snapName, ip string) (*v1alpha1.VolumeSnapCreateResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, api.VolumeGrpcListenPort), grpc.WithInsecure())
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
		return nil, fmt.Errorf("error when calling RunVolumeSnapCreateCommand: %s", err)
	}

	if response != nil {
		var responseStatus api.CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return response, fmt.Errorf("snapshot create failed with error : %v", responseStatus.Response)
		}

	}
	return response, nil
}

//DestroySnapshot destroys snapshots
func DestroySnapshot(volName, snapName, ip string) (*v1alpha1.VolumeSnapDeleteResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, api.VolumeGrpcListenPort), grpc.WithInsecure())
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
		return nil, fmt.Errorf("error when calling RunVolumeSnapDeleteCommand: %s", err)

	}
	if response != nil {
		var responseStatus api.CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return response, fmt.Errorf("snapshot deletion failed with error : %v", responseStatus.Response)
		}

	}
	return response, nil
}

// RunSnapshotCreate does tasks related to grpc snapshot create.
func (c *CmdSnaphotOptions) RunSnapshotCreate(cmd *cobra.Command) error {
	glog.Info("Executing volume snapshot create...")
	response, err := CreateSnapshot(c.VolName, c.SnapName, "")
	if response != nil {
		glog.Infof("Response from server: %s", response.Status)
		if err == nil {
			glog.Infof("Volume Snapshot Successfully Created:%v@%v\n", c.VolName, c.SnapName)
		}

	}

	return err
}

//RunSnapshotDestroy will initiate deletion of snapshot
func (c *CmdSnaphotOptions) RunSnapshotDestroy(cmd *cobra.Command) error {
	glog.Info("Executing snapshot destroy...")
	response, err := DestroySnapshot(c.VolName, c.SnapName, "")
	if response != nil {
		glog.Infof("Response from server: %s", response.Status)
		if err == nil {
			glog.Infof("Snapshot deletion initiated:%v@%v\n", c.VolName, c.SnapName)
		}
	}

	return err
}
