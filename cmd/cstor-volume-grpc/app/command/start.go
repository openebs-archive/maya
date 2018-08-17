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
	"fmt"
	"strconv"

	"github.com/openebs/maya/cmd/cstor-volume-grpc/server"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

//MaxPortNumber that can be used for ipv4
const MaxPortNumber = 65535

// CmdStartOptions has flags for starting CStorVolume watcher.
type CmdStartOptions struct {
	port string
}

// Validate validates the flag values
func (c *CmdStartOptions) Validate(cmd *cobra.Command) error {
	if len(c.port) != 0 {
		i, err := strconv.Atoi(c.port)
		if err != nil || i > MaxPortNumber {
			return fmt.Errorf("--port should be a valid integer and less than %d ", MaxPortNumber+1)
		}
	}

	return nil
}

// NewCmdStart starts gRPC server for CStorVolume.
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	cmd := &cobra.Command{
		Use:   "start",
		Short: "starts CStorVolume gRPC",
		Long:  `CStorVolume gRPC will be serving snapshot requests`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			server.StartServer(options.port)
		},
	}

	cmd.Flags().StringVarP(&options.port, "port", "p", options.port,
		"port on which the server should listen on")

	return cmd
}
