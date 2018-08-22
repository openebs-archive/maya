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
	"fmt"
	"strconv"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// VolumeOperator is the name of the tool that makes volume-related operations.
const (
	VolumeOperator   = "iscsi"
	IstgtConfPath    = "/usr/local/etc/istgt/istgt.conf"
	IstgtStatusCmd   = "STATUS"
	IstgtRefreshCmd  = "REFRESH"
	WaitTimeForIscsi = 3 * time.Second
)

//FileOperatorVar is used for doing File Operations
var FileOperatorVar util.FileOperator

//UnixSockVar is used for communication through Unix Socket
var UnixSockVar util.UnixSock

// CreateVolume creates a new cStor volume istgt config.
func CreateVolume(cStorVolume *apis.CStorVolume) error {
	// create conf file
	text := CreateIstgtConf(cStorVolume)
	err := FileOperatorVar.Write(IstgtConfPath, text, 0644)
	if err != nil {
		glog.Errorf("Failed to write istgt.conf")
	}
	glog.Info("Done writing istgt.conf")

	// send refresh command to istgt and read the response
	_, err = UnixSockVar.SendCommand(IstgtRefreshCmd)
	if err != nil {
		glog.Info("Failed to refresh iscsi service with new configuration.")
	}
	glog.Info("Creating Iscsi Volume Successful")
	return nil

}

// CreateIstgtConf creates istgt.conf file
func CreateIstgtConf(cStorVolume *apis.CStorVolume) []byte {

	var buffer bytes.Buffer
	buffer.WriteString(`# Global section
[Global]
`)
	buffer.WriteString("  NodeBase \"" + cStorVolume.Spec.NodeBase + "\"")
	buffer.WriteString(`
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
`)
	buffer.WriteString("  Portal UC1 " + cStorVolume.Spec.TargetIP + ":3261\n")
	buffer.WriteString("  Netmask " + cStorVolume.Spec.TargetIP + "/8\n")
	buffer.WriteString(`
# PortalGroup section
[PortalGroup1]
`)
	buffer.WriteString("  Portal DA1 " + cStorVolume.Spec.TargetIP + ":3260\n")
	buffer.WriteString(`
# InitiatorGroup section
[InitiatorGroup1]
  InitiatorName "ALL"
  Netmask "ALL"

[InitiatorGroup2]
  InitiatorName "None"
  Netmask "None"

# LogicalUnit section
[LogicalUnit1]
`)
	buffer.WriteString("  TargetName " + cStorVolume.Name + "\n")
	buffer.WriteString("  TargetAlias nicknamefor-" + cStorVolume.Name)
	buffer.WriteString(`
  Mapping PortalGroup1 InitiatorGroup1
  AuthMethod None
  AuthGroup None
  UseDigest Auto
  ReadOnly No
`)
	buffer.WriteString("  ReplicationFactor " + strconv.Itoa(cStorVolume.Spec.ReplicationFactor) + "\n")
	buffer.WriteString("  ConsistencyFactor " + strconv.Itoa(cStorVolume.Spec.ConsistencyFactor))
	buffer.WriteString(`
  UnitType Disk
  UnitOnline Yes
  BlockLength 512
  QueueDepth 32
  Luworkers 1
`)
	buffer.WriteString("  UnitInquiry \"OpenEBS\" \"iscsi\" \"0\" \"" + string(cStorVolume.UID) + "\"")
	buffer.WriteString(`
  PhysRecordLength 4096
`)
	buffer.WriteString("  LUN0 Storage " + cStorVolume.Spec.Capacity + " 32k")
	buffer.WriteString(`
  LUN0 Option Unmap Disable
  LUN0 Option WZero Disable
  LUN0 Option ATS Disable
  LUN0 Option XCOPY Disable
`)

	return buffer.Bytes()
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
	if len(string(cStorVolume.Spec.Capacity)) == 0 {
		return fmt.Errorf("capacity cannot be empty")
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

// CheckForIscsi is blocking call for checking status of istgt in cstor-istgt container.
func CheckForIscsi() {
	for {
		_, err := UnixSockVar.SendCommand(IstgtStatusCmd)
		if err != nil {
			time.Sleep(WaitTimeForIscsi)
			glog.Warningf("Waiting for istgt... err : %v", err)
			continue
		}
		break
	}
}
