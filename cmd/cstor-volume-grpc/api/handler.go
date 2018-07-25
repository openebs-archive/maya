package api

import (
	"fmt"
	"log"

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
	IoWaitTime           = 10
	TotalWaitTime        = 60
)

//ApiUnixSockVar is unix socker variable
var ApiUnixSockVar util.UnixSock

// Server represents the gRPC server
type Server struct {
}

// RunVolumeCommand generates response to a RunCommand request
func (s *Server) RunVolumeCommand(ctx context.Context, in *VolumeCommand) (*VolumeCommand, error) {
	log.Printf("Received command %s", in.Command)

	switch in.Command {
	case CmdSnapCreate:
		volcmd, err := SendVolumeSnapCommand(ctx, in)
		return volcmd, err

	case CmdSnapDestroy:
		volcmd, err := SendVolumeSnapCommand(ctx, in)
		return volcmd, err
	}

	return &VolumeCommand{
		Command:  in.Command,
		Volume:   in.Volume,
		Snapname: in.Snapname,
		Status:   "INVALIDCOMMAND",
	}, fmt.Errorf("Invalid VolumeCommand : %s", in.Command)

}

// SendVolumeSnapCommand sends snapcreate or snapdelete command to istgt
func SendVolumeSnapCommand(ctx context.Context, in *VolumeCommand) (*VolumeCommand, error) {
	sockresp, err := ApiUnixSockVar.SendCommand(fmt.Sprintf("%s %s %s %v %v\r\n",
		in.Command, in.Volume, in.Snapname, IoWaitTime, TotalWaitTime))
	resp := &VolumeCommand{
		Command:  in.Command,
		Volume:   in.Volume,
		Snapname: in.Snapname,
		Status:   string(sockresp),
	}
	return resp, err
}
