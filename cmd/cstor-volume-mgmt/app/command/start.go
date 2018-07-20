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

	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/start-controller"
	"github.com/spf13/cobra"
)

// CmdStartOptions has flags for starting CStorVolume watcher.
type CmdStartOptions struct {
	kubeconfig string
}

// NewCmdStart starts watching for CStorVolume resource events.
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	getCmd := &cobra.Command{
		Use:   "start",
		Short: "starts CStorVolume watcher",
		Long:  `CStorVolume resources will be watched for added, updated, deleted events`,
		Run: func(cmd *cobra.Command, args []string) {
			startcontroller.StartControllers(options.kubeconfig)
		},
	}
	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset.
	getCmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})
	getCmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "",
		`kubeconfig needs to be specified if out of cluster`)
	return getCmd
}
