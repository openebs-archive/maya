package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cstor_volume_mgmt_client "github.com/openebs/maya/pkg/client/generated/cstor-volume-mgmt/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// constants
const (
	// VolumeGrpcListenPort listen on port 7777
	VolumeGrpcListenPort = 7777
	ProtocolVersion      = 1
)

//CommandStatus is the response from istgt for control commands
type CommandStatus struct {
	Response string `json:"response"`
}

//ResizeVolume resizes cStor volume
func ResizeVolume(ip, volName, capacity string) (*cstor_volume_mgmt_client.VolumeResizeResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Errorf("Unable to dail gRPC server on port %d error: %s", VolumeGrpcListenPort, err)
	}
	defer conn.Close()

	c := cstor_volume_mgmt_client.NewRunCommandClient(conn)
	response, err := c.RunVolumeResizeCommand(context.Background(),
		&cstor_volume_mgmt_client.VolumeResizeRequest{
			Volume: volName,
			Size:   capacity,
		})
	if err != nil {
		return nil, errors.Errorf("error when calling the RunVolumeResizeCommand: '%s'", err)
	}

	if response != nil {
		var responseStatus CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return nil, errors.Errorf("ResizeVolume command failed with error: '%v'", responseStatus.Response)
		}
	}

	return response, nil
}
