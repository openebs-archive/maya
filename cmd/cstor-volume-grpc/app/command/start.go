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

package command

import (
	"github.com/openebs/maya/cmd/cstor-volume-grpc/server"
	"github.com/spf13/cobra"
)

// CmdStartOptions has flags for starting CStorVolume watcher.
type CmdStartOptions struct {
	port string
}

// NewCmdStart starts gRPC server for CStorVolume.
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	cmd := &cobra.Command{
		Use:   "start",
		Short: "starts CStorVolume gRPC",
		Long:  `CStorVolume gRPC will be serving snapshot requests`,
		Run: func(cmd *cobra.Command, args []string) {
			server.StartServer(options.port)
		},
	}

	cmd.Flags().StringVarP(&options.port, "port", "p", options.port,
		"port on which the server should listen on")

	return cmd
}
