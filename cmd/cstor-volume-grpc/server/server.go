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

package server

import (
	"fmt"
	"log"
	"net"

	"github.com/openebs/maya/cmd/cstor-volume-grpc/api"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/volume"
	"github.com/openebs/maya/pkg/util"
	"google.golang.org/grpc"
)

// StartServer instantiates CStorVolume gRPC server
// and watches for snapshot requests.
func StartServer(kubeconfig string) {

	// Making RunnerVar to use RealRunner
	// volume.RunnerVar = util.RealRunner{}

	// volume.FileOperatorVar = util.RealFileOperator{}

	api.ApiUnixSockVar = util.RealUnixSock{}
	volume.UnixSockVar = util.RealUnixSock{}

	// Blocking call for checking status of istgt running in cstor-volume container.
	volume.CheckForIscsi()

	// Blocking call for running the gRPC server
	RunCStorVolumeGrpcServer()
}

// RunCStorVolumeGrpcServer is Blocking call for listen for grpc requests of CStorVolume.
func RunCStorVolumeGrpcServer() {
	// create a listener on TCP port 7777
	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", api.VolumeGrpcListenPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// create a server instance
	s := api.Server{}
	// create a gRPC server object
	grpcServer := grpc.NewServer()
	// attach the RunCommand service to the server
	api.RegisterRunCommandServer(grpcServer, &s)
	// start the server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
