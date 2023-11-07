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

package volume

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/openebs/maya/pkg/alertlog"

	"strings"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cvapis "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

// VolumeOperator is the name of the tool that makes volume-related operations.
const (
	VolumeOperator = "iscsi"
)

var (
	istgtConfFile = `# Global section
[Global]
  NodeBase {{.Spec.NodeBase}}
  PidFile "/var/run/istgt.pid"
  AuthFile "/usr/local/etc/istgt/auth.conf"
  LogFile "/usr/local/etc/istgt/logfile"
  Luworkers 6
  MediaDirectory "/mnt"
  Timeout 60
  NopInInterval 20
  MaxR2T 16
  DiscoveryAuthMethod None
  DiscoveryAuthGroup None
  MaxSessions 32
  MaxConnections 4
  FirstBurstLength 262144
  MaxBurstLength 1048576
  MaxRecvDataSegmentLength 262144
  MaxOutstandingR2T 16
  DefaultTime2Wait 2
  DefaultTime2Retain 20
  OperationalMode 0

# UnitControl section
[UnitControl]
  AuthMethod None
  AuthGroup None
  Portal UC1 {{.Spec.TargetIP}}:3261
  Netmask {{.Spec.TargetIP}}/8

# PortalGroup section
[PortalGroup1]
  Portal DA1 {{.Spec.TargetIP}}:3260

# InitiatorGroup section
[InitiatorGroup1]
  InitiatorName "ALL"
  Netmask "ALL"

[InitiatorGroup2]
  InitiatorName "None"
  Netmask "None"

# LogicalUnit section
[LogicalUnit1]
  TargetName {{.Name}}
  TargetAlias nicknamefor-{{.Name}}
  Mapping PortalGroup1 InitiatorGroup1
  AuthMethod None
  AuthGroup None
  UseDigest Auto
  ReadOnly No
  DesiredReplicationFactor {{.Spec.DesiredReplicationFactor}}
  ReplicationFactor {{.Spec.ReplicationFactor}}
  ConsistencyFactor {{.Spec.ConsistencyFactor}}
  UnitType Disk
  UnitOnline Yes
  BlockLength 512
  QueueDepth 32
  Luworkers 6
  UnitInquiry "OpenEBS" "iscsi" "0" "{{.UID}}"
  PhysRecordLength 4096
  LUN0 Storage {{CapacityStr .Spec.Capacity}} 32k
  LUN0 Option Unmap Disable
  LUN0 Option WZero Disable
  LUN0 Option ATS Disable
  LUN0 Option XCOPY Disable
  {{- range $k, $v := .Spec.ReplicaDetails.KnownReplicas }}
  Replica {{$k}} {{$v}}
  {{- end }}
`
)

// FileOperatorVar is used for doing File Operations
var FileOperatorVar util.FileOperator

// UnixSockVar is used for communication through Unix Socket
var UnixSockVar util.UnixSock

func init() {
	UnixSockVar = util.RealUnixSock{}
	FileOperatorVar = util.RealFileOperator{}
}

// CreateVolumeTarget creates a new cStor volume istgt config.
func CreateVolumeTarget(cStorVolume *apis.CStorVolume) error {
	// create conf file
	data, err := CreateIstgtConf(cStorVolume)
	if err != nil {
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.volume.target.create.failure",
			"msg", "Failed to create CStor volume target",
			"rname", cStorVolume.Name,
		)
		return errors.Wrapf(err, "failed to create istgtconf file data")
	}
	err = FileOperatorVar.Write(util.IstgtConfPath, data, 0644)
	if err != nil {
		klog.Errorf("Failed to write istgt.conf")
	}
	klog.Info("Done writing istgt.conf")

	// send refresh command to istgt and read the response
	_, err = UnixSockVar.SendCommand(util.IstgtRefreshCmd)
	if err != nil {
		klog.Info("Failed to refresh iscsi service with new configuration.")
	}
	klog.Info("Creating Iscsi Volume Successful")
	alertlog.Logger.Infow("",
		"eventcode", "cstor.volume.target.create.success",
		"msg", "Successfully created CStor volume target",
		"rname", cStorVolume.Name,
	)
	return nil
}

// GetVolumeStatus retrieves an array of replica statuses.
func GetVolumeStatus(cStorVolume *apis.CStorVolume) (*apis.CVStatus, error) {
	// send replica command to istgt and read the response
	statuses, err := UnixSockVar.SendCommand(util.IstgtReplicaCmd)
	if err != nil {
		klog.Errorf("Failed to list replicas.")
		return nil, err
	}
	stringResp := fmt.Sprintf("%s", statuses)
	// Here it is assumed that the arrays statuses contains only one json and
	// the chars '}' and '{' are present only in the json string.
	// Therefore, the json string begins with '{' and ends with '}'
	//
	// TODO: Find a better approach
	jsonBeginIndex := strings.Index(stringResp, "{")
	jsonEndIndex := strings.LastIndex(stringResp, "}")
	if jsonBeginIndex >= jsonEndIndex {
		return nil, errors.Errorf("invalid data from %v command", util.IstgtReplicaCmd)
	}
	return extractReplicaStatusFromJSON(stringResp[jsonBeginIndex : jsonEndIndex+1])
}

// extractReplicaStatusFromJSON recieves a volume name and a json string.
// It then extracts and returns an array of replica statuses.
func extractReplicaStatusFromJSON(str string) (*apis.CVStatus, error) {
	// Unmarshal json into CVStatusResponse
	cvResponse := apis.CVStatusResponse{}
	err := json.Unmarshal([]byte(str), &cvResponse)
	if err != nil {
		return nil, err
	}
	if len(cvResponse.CVStatuses) == 0 {
		return nil, errors.Errorf("empty volume status from istgt")
	}
	return &cvResponse.CVStatuses[0], nil
}

// CreateIstgtConf creates istgt.conf file
func CreateIstgtConf(cStorVolume *apis.CStorVolume) ([]byte, error) {
	var dataBytes []byte
	buffer := &bytes.Buffer{}
	if cStorVolume == nil {
		return dataBytes, errors.Errorf("nil cstorvolume object")
	}
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"CapacityStr": func(q resource.Quantity) string { return q.String() },
	}).Parse(istgtConfFile)
	if err != nil {
		return dataBytes, errors.Wrapf(err, "failed to build istgtconffile from template")
	}
	cvObj := cStorVolume.DeepCopy()
	if cvObj.Spec.DesiredReplicationFactor == 0 {
		cvObj.Spec.DesiredReplicationFactor = cvObj.Spec.ReplicationFactor
	}
	err = tmpl.Execute(buffer, cvObj)
	if err != nil {
		return dataBytes, errors.Wrapf(err, "failed execute istgtconfile template")
	}
	dataBytes = buffer.Bytes()
	buffer.Reset()
	return dataBytes, nil
}

// ResizeTargetVolume sends resize volume command to istgt and get the response
func ResizeTargetVolume(cStorVolume *apis.CStorVolume) error {
	// send resize command to istgt and read the response
	resizeCmd := getResizeCommand(cStorVolume)
	sockResp, err := UnixSockVar.SendCommand(resizeCmd)
	if err != nil {
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.volume.target.resize.failure",
			"msg", "Failed to resize CStor volume target",
			"rname", cStorVolume.Name,
			"capacity", cStorVolume.Spec.Capacity,
		)

		return errors.Wrapf(
			err,
			"failed to execute istgt %s command on volume %s",
			util.IstgtResizeCmd,
			cStorVolume.Name)
	}
	for _, resp := range sockResp {
		if strings.Contains(resp, "ERR") {
			alertlog.Logger.Errorw("",
				"eventcode", "cstor.volume.target.resize.failure",
				"msg", "Failed to resize CStor volume target",
				"rname", cStorVolume.Name,
				"capacity", cStorVolume.Spec.Capacity,
			)
			return errors.Errorf(
				"failed to execute istgt %s command on volume %s resp: %s",
				util.IstgtResizeCmd,
				cStorVolume.Name,
				resp,
			)
		}
	}
	updateStorageVal := fmt.Sprintf("  LUN0 Storage %s 32K", cStorVolume.Spec.Capacity.String())
	cvapis.ConfFileMutex.Lock()
	err = FileOperatorVar.Updatefile(util.IstgtConfPath, updateStorageVal, "LUN0 Storage", 0644)
	if err != nil {
		cvapis.ConfFileMutex.Unlock()
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.volume.target.resize.failure",
			"msg", "Failed to resize CStor volume target",
			"rname", cStorVolume.Name,
			"capacity", cStorVolume.Spec.Capacity,
		)
		return errors.Wrapf(err,
			"failed to update %s file with %s details",
			util.IstgtConfPath,
			updateStorageVal)
	}
	cvapis.ConfFileMutex.Unlock()
	klog.Infof("Updated '%s' file with capacity '%s'", util.IstgtConfPath, updateStorageVal)
	alertlog.Logger.Infow("",
		"eventcode", "cstor.volume.target.resize.success",
		"msg", "Successfully resized CStor volume target",
		"rname", cStorVolume.Name,
		"capacity", cStorVolume.Spec.Capacity,
	)
	return nil
}

// ExecuteDesiredReplicationFactorCommand executes istgtcontrol command to update
// desired replication factor
func ExecuteDesiredReplicationFactorCommand(
	cStorVolume *apis.CStorVolume,
	getDRFCmd func(*apis.CStorVolume) string) error {
	// send desiredReplicationFactor command to istgt and read the response
	drfCmd := getDRFCmd(cStorVolume)
	sockResp, err := UnixSockVar.SendCommand(drfCmd)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to execute istgtcontrol %s command on volume %s",
			util.IstgtDRFCmd,
			cStorVolume.Name)
	}
	for _, resp := range sockResp {
		if strings.Contains(resp, "ERR") {
			return errors.Errorf(
				"failed to execute istgtcontrol %s command on volume %s resp: %s",
				util.IstgtDRFCmd,
				cStorVolume.Name,
				resp,
			)
		}
	}
	return nil
}

// GetScaleUpCommand will return data required to execute istgtcontrol drf
// command
// Ex command: drf <vol_name> <value>
func GetScaleUpCommand(cstorVolume *apis.CStorVolume) string {
	return fmt.Sprintf("%s %s %d", util.IstgtDRFCmd,
		cstorVolume.Name,
		cstorVolume.Spec.DesiredReplicationFactor,
	)
}

// GetScaleDownCommand return replica scale down command
// Ex command: drf <vol_name> <value> <known replica list>
func GetScaleDownCommand(cStorVolume *apis.CStorVolume) string {
	cmd := fmt.Sprintf("%s %s %d ", util.IstgtDRFCmd,
		cStorVolume.Name,
		cStorVolume.Spec.DesiredReplicationFactor,
	)
	for repID := range cStorVolume.Spec.ReplicaDetails.KnownReplicas {
		cmd = cmd + fmt.Sprintf("%s ", repID)
	}
	return cmd
}

// getResizeCommand returns resize used to resize volumes
// Ex command for resize: Resize volname 10G 10 30
func getResizeCommand(cstorVolume *apis.CStorVolume) string {
	return fmt.Sprintf("%s %s %s %v %v", util.IstgtResizeCmd,
		cstorVolume.Name,
		cstorVolume.Spec.Capacity.String(),
		cvapis.IoWaitTime,
		cvapis.TotalWaitTime,
	)
}

// CheckValidVolume checks for validity of CStorVolume resource.
func CheckValidVolume(cStorVolume *apis.CStorVolume) error {
	if len(string(cStorVolume.ObjectMeta.UID)) == 0 {
		return fmt.Errorf("Invalid volume resource")
	}
	if len(string(cStorVolume.Spec.TargetIP)) == 0 {
		return fmt.Errorf("targetIP cannot be empty")
	}
	if len(string(cStorVolume.Name)) == 0 {
		return fmt.Errorf("volumeName cannot be empty")
	}
	if len(string(cStorVolume.UID)) == 0 {
		return fmt.Errorf("volumeID cannot be empty")
	}
	if cStorVolume.Spec.Capacity.IsZero() {
		return fmt.Errorf("capacity cannot be zero")
	}
	if cStorVolume.VersionDetails.Status.Current >= "1.3.0" {
		if cStorVolume.Spec.DesiredReplicationFactor == 0 {
			return fmt.Errorf("DesiredReplicationFactor cannot be zero")
		}
	}
	if cStorVolume.Spec.ReplicationFactor == 0 {
		return fmt.Errorf("replicationFactor cannot be zero")
	}
	if cStorVolume.Spec.ConsistencyFactor == 0 {
		return fmt.Errorf("consistencyFactor cannot be zero")
	}

	if cStorVolume.Spec.ReplicationFactor < cStorVolume.Spec.ConsistencyFactor {
		return fmt.Errorf("replicationFactor cannot be less than consistencyFactor")
	}
	return nil
}
