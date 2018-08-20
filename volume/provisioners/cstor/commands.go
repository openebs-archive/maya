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

package cstor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-volume-grpc/api"
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-grpc/v1alpha1"
	"google.golang.org/grpc"
)

//createSnapshot creates snapshots
func createSnapshot(volName, snapName, ip string) (*v1alpha1.VolumeCommand, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, api.VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunCommandClient(conn)
	response, err := c.RunVolumeCommand(context.Background(),
		&v1alpha1.VolumeCommand{
			Command:  api.CmdSnapCreate,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		glog.Fatalf("Error when calling RunVolumeCommand: %s", err)
	}

	if response != nil {
		var responseStatus api.CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response[0], "ERR") {
			return response, fmt.Errorf("Snapshot create failed with error : %v", responseStatus.Response[0])
		}

	}
	return response, err
}

//destroySnapshot destroys snapshots
func destroySnapshot(volName, snapName, ip string) (*v1alpha1.VolumeCommand, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, api.VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunCommandClient(conn)
	response, err := c.RunVolumeCommand(context.Background(),
		&v1alpha1.VolumeCommand{
			Command:  api.CmdSnapDestroy,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		glog.Fatalf("Error when calling RunVolumeCommand: %s", err)
	}
	if response != nil {
		var responseStatus api.CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response[0], "ERR") {
			return response, fmt.Errorf("Snapshot deletion failed with error : %v", responseStatus.Response[0])
		}

	}
	return response, err
}
