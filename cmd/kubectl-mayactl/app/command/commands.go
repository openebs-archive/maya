// Copyright © 2017 The OpenEBS Authors
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

package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/openebs/maya/cmd/kubectl-mayactl/app/command/pool"
	"k8s.io/klog"

	//"github.com/openebs/maya/cmd/mayactl/app/command/snapshot"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/spf13/cobra"
)

// NewMayaCommand creates the `maya` command and its nested children.
func NewMayaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl-mayactl",
		Short: "Maya means 'Magic' a tool for storage orchestration",
		Long:  `Maya means 'Magic' a tool for storage orchestration`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if len(mapiserver.MAPIAddr) == 0 {
				mapiserver.Initialize()
				if mapiserver.GetConnectionStatus() != "running" {
					fmt.Println("Unable to connect to mapi server address")
					// Not exiting here to get the actual standard error in
					// case.The error will contains the exact IP endpoint to
					// its trying to send a http request which is more helpful
					// 1. maya-apiserver not running
					// 2. maya-apiserver not reachable
					//os.Exit(1)
				}
			} else if mapiserver.GetConnectionStatus() != "running" {
				fmt.Println("Invalid m-apiserver address")
				os.Exit(1)
			}
		},
	}

	cmd.AddCommand(
		NewCmdCompletion(cmd),
		NewCmdVersion(),
		NewCmdVolume(),
		//snapshot.NewCmdSnapshot(),
		pool.NewCmdPool(),
	)

	// add the klog flags
	klog.InitFlags(flag.CommandLine)
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// add the api addr flag
	cmd.PersistentFlags().StringVarP(&mapiserver.MAPIAddr, "mapiserver", "m", "", "Maya API Service IP address. You can obtain the IP address using kubectl get svc -n < namespace where openebs is installed >")
	cmd.PersistentFlags().StringVarP(&mapiserver.MAPIAddrPort, "mapiserverport", "p", "5656", "Maya API Service Port.")
	// TODO: switch to a different logging library.
	flag.CommandLine.Parse([]string{})

	return cmd
}
