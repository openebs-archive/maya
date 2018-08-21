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
func createSnapshot(volName, snapName, ip string) (*v1alpha1.VolumeSnapResponse, error) {
	var conn *grpc.ClientConn
	target := fmt.Sprintf("%s:%d", ip, api.VolumeGrpcListenPort)
	glog.V(3).Infof("Dialing server at %s", target)
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunSnapCommandClient(conn)
	response, err := c.RunVolumeSnapCommand(context.Background(),
		&v1alpha1.VolumeSnapRequest{
			Command:  api.CmdSnapCreate,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		glog.Errorf("Error when calling RunVolumeCommand: %s", err)
		return nil, err
	}

	if response != nil {
		var responseStatus api.CommandStatus
		err = json.Unmarshal(response.Status, &responseStatus)
		if err != nil {
			glog.Errorf("Error reading response: %s", err)
			return nil, err
		}
		if strings.Contains(responseStatus.Response, "ERR") {
			return response, fmt.Errorf("Snapshot create failed with error : %v", responseStatus.Response)
		}

	}
	return response, err
}

//DestroySnapshot destroys snapshots
func destroySnapshot(volName, snapName, ip string) (*v1alpha1.VolumeSnapResponse, error) {
	var conn *grpc.ClientConn
	target := fmt.Sprintf("%s:%d", ip, api.VolumeGrpcListenPort)
	glog.V(3).Infof("Dialing server at %s", target)
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunSnapCommandClient(conn)
	response, err := c.RunVolumeSnapCommand(context.Background(),
		&v1alpha1.VolumeSnapRequest{
			Command:  api.CmdSnapDestroy,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		glog.Errorf("Error when calling RunVolumeCommand: %s", err)
		return nil, err
	}

	if response != nil {
		var responseStatus api.CommandStatus
		err = json.Unmarshal(response.Status, &responseStatus)
		if err != nil {
			glog.Errorf("Error reading response: %s", err)
			return nil, err
		}
		if strings.Contains(responseStatus.Response, "ERR") {
			return response, fmt.Errorf("Snapshot create failed with error : %v", responseStatus.Response)
		}

	}
	return response, err
}
