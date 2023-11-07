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

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	cmdName = "cstor-volume-mgmt"
	usage   = fmt.Sprintf("%s", cmdName)
)

// CmdSnaphotOptions holds the options for snapshot
// create command
type CmdSnaphotOptions struct {
	volName  string
	snapName string
}

// NewCStorVolumeMgmt creates a new CStorVolume CRD watcher and grpc command.
func NewCStorVolumeMgmt() (*cobra.Command, error) {
	// Create a new command.
	cmd := &cobra.Command{
		Use:   usage,
		Short: "CStor Volume Watcher and GRPC server",
		Long: `interfaces between observing the CStorVolume
		 objects and volume controller creation and GRPC server`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd), util.Fatal)
		},
	}
	cmd.AddCommand(
		NewCmdStart(),
	)
	return cmd, nil
}

// Run is to run cstor-volume-mgmt command without any arguments
func Run(cmd *cobra.Command) error {
	klog.Infof("cstor-volume-mgmt for CStorVolume objects and grpc server")
	return nil
}
