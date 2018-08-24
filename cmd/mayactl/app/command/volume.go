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
	"net/http"
	"strings"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

// VolumeInfo stores the volume information
type VolumeInfo struct {
	Volume v1alpha1.CASVolume
}

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

// CASType is engine type
type CASType string

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
			fmt.Printf("Volume: %s not found at namespace: %q\n", volname, namespace)
			err = util.InternalServerError
		} else if resp.StatusCode == 503 {
			fmt.Println("M_API server not reachable")
			err = util.ServerUnavailable
		} else if resp.StatusCode == 404 {
			fmt.Printf("Volume: %s not found at namespace: %q error: %s\n", volname, namespace, http.StatusText(resp.StatusCode))
			err = util.PageNotFound
		}
		fmt.Printf("Received an error from M_API server: statuscode: %d", resp.StatusCode)
		err = fmt.Errorf("Received an error from M_API server: statuscode: %d", resp.StatusCode)
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
		fmt.Println("VOLUME status Unknown to M_API server")
		err = fmt.Errorf("VOLUME status Unknown to M_API server")
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
	return volInfo.Volume.Spec.Iqn
}

// GetVolumeName returns the volume name
func (volInfo *VolumeInfo) GetVolumeName() string {
	return volInfo.Volume.ObjectMeta.Name
}

// GetTargetPortal returns the TargetPortal of the volume
func (volInfo *VolumeInfo) GetTargetPortal() string {
	return volInfo.Volume.Spec.TargetPortal
}

// GetVolumeSize returns the capacity of the volume
func (volInfo *VolumeInfo) GetVolumeSize() string {
	return volInfo.Volume.Spec.Capacity
}

// GetReplicaCount returns the volume replica count
func (volInfo *VolumeInfo) GetReplicaCount() string {
	return volInfo.Volume.Spec.Replicas
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

var (
	volumeCommandHelpText = `
The following commands helps in operating a Volume such as create, list, and so on.

Usage: mayactl volume <subcommand> [options] [args]

Examples:

 # Create a Volume:
   $ mayactl volume create --volname <vol> --size <size>

 # List Volumes:
   $ mayactl volume list

 # Delete a Volume:
   $ mayactl volume delete --volname <vol>

 # Delete a Volume created in 'test' namespace:
   $ mayactl volume delete --volname <vol> --namespace test

 # Statistics of a Volume:
   $ mayactl volume stats --volname <vol>

 # Statistics of a Volume created in 'test' namespace:
   $ mayactl volume stats --volname <vol> --namespace test

 # Info of a Volume:
   $ mayactl volume info --volname <vol>

 # Info of a Volume created in 'test' namespace:
   $ mayactl volume info --volname <vol> --namespace test
`
	options = &CmdVolumeOptions{
		namespace: "default",
	}
)

// CmdVolumeOptions stores information of volume being operated
type CmdVolumeOptions struct {
	volName          string
	sourceVolumeName string
	snapshotName     string
	size             string
	namespace        string
	json             string
}

// NewCmdVolume provides options for managing OpenEBS Volume
func NewCmdVolume() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Provides operations related to a Volume",
		Long:  volumeCommandHelpText,
	}

	cmd.AddCommand(
		NewCmdVolumeCreate(),
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
