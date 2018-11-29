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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
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
	// VolumeAPIPath is the api path to get volume information
	VolumeAPIPath      = "/latest/volumes/"
	controllerStatusOk = "running"
	volumeStatusOK     = "Running"
	// JivaStorageEngine is constant for jiva engine
	JivaStorageEngine CASType = "jiva"
	// CstorStorageEngine is constant for cstor engine
	CstorStorageEngine CASType = "cstor"
	timeout                    = 5 * time.Second
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
   $ mayactl volume describe --volname <vol>

 # Info of a Volume created in 'test' namespace:
   $ mayactl volume describe --volname <vol> --namespace test

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
			return errors.New("error: --snapname not specified.")
		}
	}
	if sourcenameverify {
		if len(c.sourceVolumeName) == 0 {
			return errors.New("error: --sourcevol not specified.")
		}
	}
	if volnameverify {
		if len(c.volName) == 0 {
			return errors.New("error: --volname not specified.")
		}
	}
	return nil
}

// NewVolumeInfo fetches and fills CASVolume structure from URL given to it
func NewVolumeInfo(URL string, volname string, namespace string) (volInfo *VolumeInfo, err error) {
	url := URL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("namespace", namespace)

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("Can't get a response, error found: %v", err)
		return
	}
	if resp != nil && resp.StatusCode != 200 {
		if resp.StatusCode == 500 {
			fmt.Printf("Sorry something went wrong with service. Please raise an issue on: https://github.com/openebs/openebs/issues")
			err = util.ErrInternalServerError
			return
		} else if resp.StatusCode == 503 {
			fmt.Printf("maya apiservice not reachable at %q\n", mapiserver.GetURL())
			err = util.ErrServerUnavailable
			return
		} else if resp.StatusCode == 404 {
			fmt.Printf("Volume: %s not found at namespace: %q error: %s\n", volname, namespace, http.StatusText(resp.StatusCode))
			err = util.ErrPageNotFound
			return
		}
		fmt.Printf("Received an error from maya apiservice: statuscode: %d", resp.StatusCode)
		err = fmt.Errorf("Received an error from maya apiservice: statuscode: %d", resp.StatusCode)
		return
	}
	defer resp.Body.Close()
	casVol := v1alpha1.CASVolume{}
	err = json.NewDecoder(resp.Body).Decode(&casVol)
	if err != nil {
		fmt.Printf("Response decode failed: error '%+v'", err)
		return
	}
	if casVol.Status.Reason == "pending" {
		fmt.Println("VOLUME status Unknown to maya apiservice")
		err = fmt.Errorf("VOLUME status Unknown to maya apiservice")
		return
	}
	volInfo = &VolumeInfo{
		Volume: casVol,
	}
	return
}

// GetCASType returns the CASType of the volume in lowercase
func (volInfo *VolumeInfo) GetCASType() string {
	if len(volInfo.Volume.Spec.CasType) == 0 {
		return string(JivaStorageEngine)
	}
	return strings.ToLower(volInfo.Volume.Spec.CasType)
}

// GetClusterIP returns the ClusterIP of the cluster
func (volInfo *VolumeInfo) GetClusterIP() string {
	if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/cluster-ips"]; ok {
		return val
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["vsm.openebs.io/cluster-ips"]; ok {
		return val
	}
	return ""
}

// GetControllerStatus returns the status of the volume controller
func (volInfo *VolumeInfo) GetControllerStatus() string {
	if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/controller-status"]; ok {
		return val
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["vsm.openebs.io/controller-status"]; ok {
		return val
	}
	return ""
}

// GetIQN returns the IQN of the volume
func (volInfo *VolumeInfo) GetIQN() string {
	if len(volInfo.Volume.Spec.Iqn) > 0 {
		return volInfo.Volume.Spec.Iqn
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/iqn"]; ok {
		return val
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["vsm.openebs.io/iqn"]; ok {
		return val
	}
	return ""

}

// GetVolumeName returns the volume name
func (volInfo *VolumeInfo) GetVolumeName() string {
	return volInfo.Volume.ObjectMeta.Name
}

// GetTargetPortal returns the TargetPortal of the volume
func (volInfo *VolumeInfo) GetTargetPortal() string {
	if len(volInfo.Volume.Spec.TargetPortal) > 0 {
		return volInfo.Volume.Spec.TargetPortal
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/targetportals"]; ok {
		return val
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["vsm.openebs.io/targetportals"]; ok {
		return val
	}
	return ""
}

// GetVolumeSize returns the capacity of the volume
func (volInfo *VolumeInfo) GetVolumeSize() string {
	if len(volInfo.Volume.Spec.Capacity) > 0 {
		return volInfo.Volume.Spec.Capacity
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/volume-size"]; ok {
		return val
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["vsm.openebs.io/volume-size"]; ok {
		return val
	}
	return ""
}

// GetReplicaCount returns the volume replica count
func (volInfo *VolumeInfo) GetReplicaCount() string {
	if len(volInfo.Volume.Spec.Replicas) > 0 {
		return volInfo.Volume.Spec.Replicas
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/replica-count"]; ok {
		return val
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["vsm.openebs.io/replica-count"]; ok {
		return val
	}
	return ""
}

// GetReplicaStatus returns the replica status of the volume replica
func (volInfo *VolumeInfo) GetReplicaStatus() string {
	if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/replica-status"]; ok {
		return val
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["vsm.openebs.io/replica-status"]; ok {
		return val
	}
	return ""
}

// GetReplicaIP returns the IP of volume replica
func (volInfo *VolumeInfo) GetReplicaIP() string {
	if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/replica-ips"]; ok {
		return val
	} else if val, ok := volInfo.Volume.ObjectMeta.Annotations["vsm.openebs.io/replica-ips"]; ok {
		return val
	}
	return ""
}

// GetStoragePool returns the name of the storage pool
func (volInfo *VolumeInfo) GetStoragePool() string {
	if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/pool-names"]; ok {
		return val
	}
	return ""
}

// GetCVRName returns the name of the CVR
func (volInfo *VolumeInfo) GetCVRName() string {
	if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/cvr-names"]; ok {
		return val
	}
	return ""
}

// GetNodeName returns the name of the node
func (volInfo *VolumeInfo) GetNodeName() string {
	if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/node-names"]; ok {
		return val
	}
	return ""
}

// GetControllerNode returns the node name of the controller
func (volInfo *VolumeInfo) GetControllerNode() string {
	if val, ok := volInfo.Volume.ObjectMeta.Annotations["openebs.io/controller-node-name"]; ok {
		return val
	}
	return ""
}

func print(format string, obj interface{}) error {
	// New Instance of tabwriter
	w := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', 0)
	// New Instance of template
	tmpl, err := template.New("ReplicaStats").Parse(format)
	if err != nil {
		return fmt.Errorf("Error in parsing replica template, found error : %v", err)
	}
	// Parse struct with template
	err = tmpl.Execute(w, obj)
	if err != nil {
		return fmt.Errorf("Error in executing replica template, found error : %v", err)
	}
	return w.Flush()
}
