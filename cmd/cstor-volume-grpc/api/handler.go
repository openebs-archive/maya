package api

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"

	"github.com/openebs/maya/pkg/client/generated/cstor-volume-grpc/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"golang.org/x/net/context"
)

// sudo $ISTGTCONTROL snapdestroy vol1 snapname1 0
// sudo $ISTGTCONTROL snapcreate vol1 snapname1

// constants
const (
	VolumeGrpcListenPort = 7777
	CmdSnapCreate        = "SNAPCREATE"
	CmdSnapDestroy       = "SNAPDESTROY"
	//IoWaitTime is the time interval for which the IO has to be stopped before doing snapshot operation
	IoWaitTime = 10
	//TotalWaitTime is the max time duration to wait for doing snapshot operation on all the replicas
	TotalWaitTime = 60
)

//CommandStatus is the response from istgt for control commands
type CommandStatus struct {
	Response string `json:"response"`
}

//APIUnixSockVar is unix socker variable
var APIUnixSockVar util.UnixSock

// Server represents the gRPC server
type Server struct {
}

// RunVolumeSnapCommand generates response to a RunCommand request
func (s *Server) RunVolumeSnapCommand(ctx context.Context, in *v1alpha1.VolumeSnapRequest) (*v1alpha1.VolumeSnapResponse, error) {
	glog.Infof("Received command %s", in.Command)

	switch in.Command {
	case CmdSnapCreate:
		volcmd, err := SendVolumeSnapCommand(ctx, in)
		return volcmd, err

	case CmdSnapDestroy:
		volcmd, err := SendVolumeSnapCommand(ctx, in)
		return volcmd, err
	}

	status := CommandStatus{
		Response: "INVALIDCOMMAND",
	}
	jsonresp, _ := json.Marshal(status)
	return &v1alpha1.VolumeSnapResponse{
		Status: jsonresp,
	}, fmt.Errorf("Invalid VolumeCommand : %s", in.Command)

}

// SendVolumeSnapCommand sends snapcreate or snapdelete command to istgt
func SendVolumeSnapCommand(ctx context.Context, in *v1alpha1.VolumeSnapRequest) (*v1alpha1.VolumeSnapResponse, error) {
	sockresp, err := APIUnixSockVar.SendCommand(fmt.Sprintf("%s %s %s %v %v",
		in.Command, in.Volume, in.Snapname, IoWaitTime, TotalWaitTime))
	respstr := "ERR"
	if len(sockresp) > 1 {
		respstr = sockresp[1]
	}
	status := CommandStatus{
		Response: respstr,
	}
	jsonresp, _ := json.Marshal(status)
	resp := &v1alpha1.VolumeSnapResponse{
		Status: jsonresp,
	}
	return resp, err
}
