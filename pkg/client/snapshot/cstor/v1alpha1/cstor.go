package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-grpc/v1alpha1"
	"google.golang.org/grpc"
)

// constants
const (
	VolumeGrpcListenPort = 7777
	ProtocolVersion      = 1
)

//CommandStatus is the response from istgt for control commands
type CommandStatus struct {
	Response string `json:"response"`
}

//CreateSnapshot creates snapshots
func CreateSnapshot(ip, volName, snapName string) (*v1alpha1.VolumeSnapCreateResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		glog.Errorf("Unable to dial gRPC server on port %d error : %s", VolumeGrpcListenPort, err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunSnapCommandClient(conn)
	response, err := c.RunVolumeSnapCreateCommand(context.Background(),
		&v1alpha1.VolumeSnapCreateRequest{
			Version:  ProtocolVersion,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		glog.Errorf("Error when calling RunVolumeSnapCreateCommand: %s", err)
		return nil, err
	}

	if response != nil {
		var responseStatus CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return response, fmt.Errorf("Snapshot create failed with error : %v", responseStatus.Response)
		}

	}
	return response, err
}

//DestroySnapshot destroys snapshots
func DestroySnapshot(ip, volName, snapName string) (*v1alpha1.VolumeSnapDeleteResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		glog.Errorf("did not connect: %s", err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunSnapCommandClient(conn)
	response, err := c.RunVolumeSnapDeleteCommand(context.Background(),
		&v1alpha1.VolumeSnapDeleteRequest{
			Version:  ProtocolVersion,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		glog.Errorf("Error when calling RunVolumeSnapDeleteCommand: %s", err)
		return nil, err
	}
	if response != nil {
		var responseStatus CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return response, fmt.Errorf("Snapshot deletion failed with error : %v", responseStatus.Response)
		}

	}
	return response, err
}
