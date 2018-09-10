/*
Copyright 2017 The OpenEBS Authors.

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
	"errors"
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	client "github.com/openebs/maya/pkg/client/jiva"
	"github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

// VolumeInfo stores the volume information
type VolumeInfo struct {
	Volume v1alpha1.CASVolume
}

// CmdVolumeOptions stores information of volume being operated
type CmdVolumeOptions struct {
	volName          string
	sourceVolumeName string
	snapshotName     string
	size             string
	namespace        string
	json             string
}

// CASType is engine type
type CASType string

const (
	// JivaStorageEngine is constant for jiva engine
	JivaStorageEngine CASType = "jiva"
	// CstorStorageEngine is constant for cstor engine
	CstorStorageEngine CASType = "cstor"

	infoNotAvailable = "N/A"
	statusRunning    = "Running"

	// Keys of annotations
	replicaStatus     = "openebs.io/replica-status"
	replicaIP         = "openebs.io/replica-ips"
	controllerStatus  = "openebs.io/controller-status"
	clusterIP         = "openebs.io/cluster-ips"
	replicaAccessMode = "openebs.io/replica-access-mode"
	replicaNodeName   = "openebs.io/replica-node-names"
	storagePool       = "openebs.io/pool-names"
	cvrName           = "openebs.io/cvr-names"
	nodeName          = "openebs.io/node-names"
	replicaPodName    = "openebs.io/replica-pod-names"
)

// # Create a Volume:
// $ mayactl volume create --volname <vol> --size <size>

var (
	volumeCommandHelpText = `
The following commands helps in operating a Volume such as create, list, and so on.

Usage: mayactl volume <subcommand> [options] [args]

Examples:
 # List Volumes:
   $ mayactl volume list

 # Statistics of a Volume:
   $ mayactl volume stats --volname <vol>

 # Statistics of a Volume created in 'test' namespace:
   $ mayactl volume stats --volname <vol> --namespace test

 # Info of a Volume:
   $ mayactl volume info --volname <vol>

 # Info of a Volume created in 'test' namespace:
   $ mayactl volume info --volname <vol> --namespace test

 # Delete a Volume:
   $ mayactl volume delete --volname <vol>

 # Delete a Volume created in 'test' namespace:
   $ mayactl volume delete --volname <vol> --namespace test
`
	options = &CmdVolumeOptions{
		namespace: "default",
	}
)

// NewCmdVolume provides options for managing OpenEBS Volume
func NewCmdVolume() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Provides operations related to a Volume",
		Long:  volumeCommandHelpText,
	}

	cmd.AddCommand(
		// NewCmdVolumeCreate(),
		NewCmdVolumesList(),
		NewCmdVolumeDelete(),
		NewCmdVolumeStats(),
		NewCmdVolumeInfo(),
	)
	cmd.PersistentFlags().StringVarP(&options.namespace, "namespace", "n", options.namespace,
		"namespace name, required if volume is not in the default namespace")

	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse([]string{})
	return cmd
}

// Validate verifies whether a volume name,source name or snapshot name is provided or not followed by
// stats command. It returns nil and proceeds to execute the command if there is
// no error and returns an error if it is missing.
func (c *CmdVolumeOptions) Validate(cmd *cobra.Command, snapshotnameverify, sourcenameverify, volnameverify bool) error {
	if snapshotnameverify {
		if len(c.snapshotName) == 0 {
			return errors.New("--snapname is missing. Please provide a snapshotname")
		}
	}
	if sourcenameverify {
		if len(c.sourceVolumeName) == 0 {
			return errors.New("--sourcevol is missing. Please specify a sourcevolumename")
		}
	}
	if volnameverify {
		if len(c.volName) == 0 {
			return errors.New("--volname is missing. Please specify a unique volumename")
		}
	}
	return nil
}

// processCASVolume process the response with values to be displayed
func processCASVolume(volume v1alpha1.CASVolume, fetchAccessMode bool) (v1alpha1.CASVolume, error) {

	// For support to 0.6 volumes
	if len(volume.Spec.CasType) == 0 {
		volume.Spec.CasType = string(JivaStorageEngine)
	}

	// Assigning N/A if info not available
	if len(volume.ObjectMeta.Namespace) == 0 {
		volume.ObjectMeta.Namespace = infoNotAvailable
	}

	// Assigning reason as Running when no error found
	if len(volume.Status.Reason) == 0 {
		volume.Status.Reason = statusRunning
	}

	if fetchAccessMode && volume.Spec.CasType == string(JivaStorageEngine) {
		controllerClient := client.ControllerClient{}
		collection := client.ReplicaCollection{}
		controllerStatuses := strings.Split(volume.ObjectMeta.Annotations[controllerStatus], ",")
		// Iterating over controllerStatus
		for _, controllerStatus := range controllerStatuses {
			if controllerStatus != statusRunning {
				fmt.Printf("Unable to fetch volume details, Volume controller's status is '%s'.\n", controllerStatus)
				return volume, errors.New("Unable to fetch volume details")
			}
		}
		// controllerIP:9501/v1/replicas is to be parsed into this structure via GetVolumeStats.
		// An API needs to be passed as argument.
		_, err := controllerClient.GetVolumeStats(volume.ObjectMeta.Annotations[clusterIP]+v1.ControllerPort, v1.InfoAPI, &collection)
		if err != nil {
			return volume, fmt.Errorf("Cannot get volume stats %v", err)
		}

		replica := make(map[string]string)
		accessMode := []string{}
		for _, repl := range collection.Data {
			replica[repl.Address] = repl.Mode
		}
		for _, ip := range strings.Split(volume.ObjectMeta.Annotations[replicaIP], ",") {
			if val, ok := replica[ip]; ok {
				accessMode = append(accessMode, val)
			} else {
				accessMode = append(accessMode, "N/A")
			}
		}
		volume.ObjectMeta.Annotations[replicaAccessMode] = strings.Join(accessMode, ",")
	}
	return volume, nil
}

func renderTemplate(templateName, templatePreview string, templateStructure interface{}) (err error) {
	// Create template instance and pass template structure into it.
	templ, err := template.New(templateName).Parse(templatePreview)
	if err != nil {
		fmt.Println("Error displaying output, found error:", err)
		return
	}

	// Create tabwriter instance and execute it with template.
	w := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', 0)
	err = templ.Execute(w, templateStructure)
	if err != nil {
		fmt.Println("Error displaying output, found error:", err)
		return
	}

	// Flush the tabwriter instance.
	err = w.Flush()
	if err != nil {
		fmt.Println("Error displaying output, found error:", err)
	}
	return err
}
