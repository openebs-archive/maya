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

package v1alpha1

import (
	"fmt"
	"net"
	"strconv"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-mgmt/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"google.golang.org/grpc"
)

// StartServer instantiates CStorVolume gRPC server
// and watches for snapshot requests.
func StartServer(unixSockVar util.UnixSock, port string) error {
	// Blocking call for checking status of istgt running in cstor-volume container.
	util.CheckForIscsi(unixSockVar)

	if len(port) == 0 {
		i, err := strconv.Atoi(port)
		if err == nil && i != 0 {
			// Blocking call for running the gRPC server
			return RunCStorVolumeGrpcServer(i)
		}
		glog.Warningf("Invalid listen port. Using default port %d ", VolumeGrpcListenPort)
	}
	return RunCStorVolumeGrpcServer(VolumeGrpcListenPort)
}

// RunCStorVolumeGrpcServer is Blocking call for listen for grpc requests of CStorVolume.
func RunCStorVolumeGrpcServer(port int) error {
	glog.Infof("Starting gRPC server on port : %d", port)
	// create a listener on TCP port 7777
	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}
	// create a server instance
	s := Server{}
	// create a gRPC server object
	grpcServer := grpc.NewServer()
	// attach the RunCommand service to the server
	v1alpha1.RegisterRunSnapCommandServer(grpcServer, &s)
	// start the server
	if err := grpcServer.Serve(lis); err != nil {
		glog.Fatalf("failed to serve: %s", err)
	}
	return nil
}
