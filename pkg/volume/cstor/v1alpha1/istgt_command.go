// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/openebs/maya/pkg/client/generated/cstor-volume-mgmt/v1alpha1"
	file "github.com/openebs/maya/pkg/file"
	v1_strings "github.com/openebs/maya/pkg/string/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"golang.org/x/net/context"
)

// sudo $ISTGTCONTROL snapdestroy vol1 snapname1 0
// sudo $ISTGTCONTROL snapcreate vol1 snapname1
// sudo $ISTGTCONTROL resize vol1 30G

// TODO: Need to modify the volume resize command based on istgt
// constants
const (
	VolumeGrpcListenPort = 7777
	CmdSnapCreate        = "SNAPCREATE"
	CmdSnapDestroy       = "SNAPDESTROY"
	CmdVolResize         = "RESIZE"
	//IoWaitTime is the time interval for which the IO has to be stopped before doing snapshot operation
	IoWaitTime = 10
	//TotalWaitTime is the max time duration to wait for doing snapshot operation on all the replicas
	TotalWaitTime   = 60
	ProtocolVersion = 1
)

//CommandStatus is the response from istgt for control commands
type CommandStatus struct {
	Response string `json:"response"`
}

//APIFileOperatorVar is used for doing File Operations
var APIFileOperatorVar file.FileOperator

//APIUnixSockVar is unix socker variable
var APIUnixSockVar util.UnixSock

// Server represents the gRPC server
type Server struct {
}

func init() {
	APIUnixSockVar = util.RealUnixSock{}
	APIFileOperatorVar = file.RealFileOperator{}
}

// RunVolumeSnapCreateCommand performs snapshot create operation and sends back the response
func (s *Server) RunVolumeSnapCreateCommand(ctx context.Context, in *v1alpha1.VolumeSnapCreateRequest) (*v1alpha1.VolumeSnapCreateResponse, error) {
	glog.Infof("Received snapshot create request. volname = %s, snapname = %s, version = %d", in.Volume, in.Snapname, in.Version)
	volcmd, err := CreateSnapshot(ctx, in)
	return volcmd, err

}

// RunVolumeSnapDeleteCommand performs snapshot create operation and sends back the response
func (s *Server) RunVolumeSnapDeleteCommand(ctx context.Context, in *v1alpha1.VolumeSnapDeleteRequest) (*v1alpha1.VolumeSnapDeleteResponse, error) {
	glog.Infof("Received snapshot delete request. volname = %s, snapname = %s, version = %d", in.Volume, in.Snapname, in.Version)
	volcmd, err := DeleteSnapshot(ctx, in)
	return volcmd, err
}

// RunVolumeResizeCommand perform volume resize operation and sends back the response
func (s *Server) RunVolumeResizeCommand(ctx context.Context, in *v1alpha1.VolumeResizeRequest) (*v1alpha1.VolumeResizeResponse, error) {
	glog.Infof("Received volume resize request. volname: '%s', capacity: '%s'", in.Volume, in.Size)
	volcmd, err := ResizeVolume(ctx, in)
	return volcmd, err
}

// CreateSnapshot sends snapcreate command to istgt and returns the response
func CreateSnapshot(ctx context.Context, in *v1alpha1.VolumeSnapCreateRequest) (*v1alpha1.VolumeSnapCreateResponse, error) {
	sockresp, err := APIUnixSockVar.SendCommand(fmt.Sprintf("%s %s %s %v %v",
		CmdSnapCreate, in.Volume, in.Snapname, IoWaitTime, TotalWaitTime))
	respstr := "ERR"
	if nil != sockresp && len(sockresp) > 1 {
		respstr = sockresp[1]
	}
	status := CommandStatus{
		Response: respstr,
	}
	jsonresp, _ := json.Marshal(status)
	resp := &v1alpha1.VolumeSnapCreateResponse{
		Status: jsonresp,
	}
	return resp, err
}

// DeleteSnapshot sends snapdelete command to istgt and returns the response
func DeleteSnapshot(ctx context.Context, in *v1alpha1.VolumeSnapDeleteRequest) (*v1alpha1.VolumeSnapDeleteResponse, error) {
	sockresp, err := APIUnixSockVar.SendCommand(fmt.Sprintf("%s %s %s %v %v",
		CmdSnapDestroy, in.Volume, in.Snapname, IoWaitTime, TotalWaitTime))
	respstr := "ERR"
	if nil != sockresp && len(sockresp) > 1 {
		respstr = sockresp[1]
	}
	status := CommandStatus{
		Response: respstr,
	}
	jsonresp, _ := json.Marshal(status)
	resp := &v1alpha1.VolumeSnapDeleteResponse{
		Status: jsonresp,
	}
	return resp, err
}

// ResizeVolume sends resize volume command to istgt and returns the response
func ResizeVolume(ctx context.Context, in *v1alpha1.VolumeResizeRequest) (*v1alpha1.VolumeResizeResponse, error) {
	updateStorageVal := fmt.Sprintf("  LUN0 Storage %s 32k", in.Size)
	index, oldStorageVal, err := APIFileOperatorVar.GetLineDetails(util.IstgtConfPath, "LUN0 Storage")
	if err != nil {
		return nil, err
	} else if index == -1 {
		return nil, errors.Wrapf(err, "failed to get the Storage details from '%v'", util.IstgtConfPath)
	}
	err = APIFileOperatorVar.Updatefile(util.IstgtConfPath, updateStorageVal, index, 0644)
	if err != nil {
		return nil, err
	}
	glog.Infof("Updated the '%s' file with capacity '%s'", util.IstgtConfPath, in.Size)
	sockresp, err := APIUnixSockVar.SendCommand(fmt.Sprintf("%s %v %v",
		CmdVolResize, IoWaitTime, TotalWaitTime))
	list := v1_strings.List(sockresp...)
	if err != nil || list.Contains("ERR") {
		glog.Infof("Reverting the changes to file '%s'", util.IstgtConfPath)
		errOp := APIFileOperatorVar.Updatefile(util.IstgtConfPath, oldStorageVal, index, 0644)
		if errOp != nil {
			glog.Errorf("Failed to revert the changes on file '%v'", errOp)
		}
	}

	respstr := "ERR"
	if nil != sockresp && len(sockresp) > 1 {
		respstr = sockresp[1]
	}
	status := CommandStatus{
		Response: respstr,
	}
	jsonresp, _ := json.Marshal(status)
	resp := &v1alpha1.VolumeResizeResponse{
		Status: jsonresp,
	}
	return resp, err
}
