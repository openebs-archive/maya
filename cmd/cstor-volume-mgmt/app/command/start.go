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
	goflag "flag"
	"sync"

	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/start-controller"
	"github.com/openebs/maya/cmd/cstor-volume-mgmt/volume"
	serverclient "github.com/openebs/maya/pkg/cstor/volume/serverclient/v1alpha1"
	"github.com/spf13/cobra"
)

// CmdStartOptions has flags for starting CStorVolume watcher.
type CmdStartOptions struct {
	kubeconfig string
	port       string
}

// NewCmdStart starts gRPC server and watcher for CStorVolume.
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	cmd := &cobra.Command{
		Use:   "start",
		Short: "starts CStorVolume gRPC and watcher",
		Long: `The grpc server would be serving snapshot requests whereas
		the watcher would be watching for add, updat, delete events`,
		Run: func(cmd *cobra.Command, args []string) {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				serverclient.StartServer(volume.UnixSockVar, options.port)
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				startcontroller.StartControllers(options.kubeconfig)
				wg.Done()
			}()
			wg.Wait()
		},
	}
	goflag.CommandLine.Parse([]string{})
	cmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "",
		`kubeconfig needs to be specified if out of cluster`)
	cmd.Flags().StringVarP(&options.port, "port", "p", options.port,
		"port on which the server should listen on")

	return cmd
}
