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
	"fmt"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// VolumeOperator is the name of the tool that makes volume-related operations.
const (
	VolumeOperator  = "iscsi"
	IstgtConfPath   = "/usr/local/etc/istgt/istgt.conf"
	IstgtStatusCmd  = "STATUS"
	IstgtRefreshCmd = "REFRESH"
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
		glog.Errorf("Failed to write istgt.conf...")
	}
	glog.Info("Done writing istgt.conf")

	// send refresh command to istgt and read the response
	_, err = UnixSockVar.SendCommand(IstgtRefreshCmd)
	if err != nil {
		glog.Info("refresh failed")
	}
	glog.Info("Creating Iscsi Volume Successful")
	return nil

}

// CreateIstgtConf creates istgt.conf file
func CreateIstgtConf(cStorVolume *apis.CStorVolume) []byte {
	text := []byte(`
# Global section
[Global]
  NodeBase "iqn.2017-08.OpenEBS.cstor"
  PidFile "/var/run/istgt.pid"
  AuthFile "/usr/local/etc/istgt/auth.conf"
  LogFile "/usr/local/etc/istgt/logfile"
  Luworkers 1 
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

	portaluc1 := []byte("  Portal UC1 " + cStorVolume.Spec.CStorControllerIP + ":3261\n")
	text = append(text, portaluc1...)

	netmask := []byte("  Netmask " + cStorVolume.Spec.CStorControllerIP + "/8\n")
	text = append(text, netmask...)

	text1 := []byte(`
# PortalGroup section
[PortalGroup1]
`)
	text = append(text, text1...)

	portalda1 := []byte("  Portal DA1 " + cStorVolume.Spec.CStorControllerIP + ":3260\n")
	text = append(text, portalda1...)

	text2 := []byte(`
# InitiatorGroup section
[InitiatorGroup1]
  InitiatorName "ALL"
  Netmask "ALL"

[InitiatorGroup2]
  InitiatorName "None"
  Netmask "None"

# LogicalUnit section
[LogicalUnit2]
`)

	text = append(text, text2...)

	targetName := []byte("  TargetName " + cStorVolume.Spec.VolumeName + "\n")
	text = append(text, targetName...)
	targetAlias := []byte("  TargetAlias nicknamefor-" + cStorVolume.Spec.VolumeName)
	text = append(text, targetAlias...)

	text3 := []byte(`
  Mapping PortalGroup1 InitiatorGroup1
  AuthMethod None
  AuthGroup None
  UseDigest Auto
  ReadOnly No
  ReplicationFactor 3
  ConsistencyFactor 2
  UnitType Disk
  UnitOnline Yes
  BlockLength 512
  QueueDepth 32
  Luworkers 1
`)
	text = append(text, text3...)

	unitinquiry := []byte("  UnitInquiry \"OpenEBS\" \"iscsi\" \"0\" \"" + cStorVolume.Spec.VolumeID + "\"")
	text = append(text, unitinquiry...)

	text4 := []byte(`
  PhysRecordLength 4096
`)
	text = append(text, text4...)

	lun0storage := []byte("  LUN0 Storage " + cStorVolume.Spec.Capacity + " 32k")
	text = append(text, lun0storage...)

	text5 := []byte(`
  LUN0 Option Unmap Disable
  LUN0 Option WZero Disable
  LUN0 Option ATS Disable
  LUN0 Option XCOPY Disable
`)

	text = append(text, text5...)

	return text
}

// CheckValidVolume checks for validity of CStorVolume resource.
func CheckValidVolume(cStorVolume *apis.CStorVolume) error {
	if len(string(cStorVolume.ObjectMeta.UID)) == 0 {
		return fmt.Errorf("Invalid volume resource")
	}
	if len(string(cStorVolume.Spec.CStorControllerIP)) == 0 {
		return fmt.Errorf("cstorControllerIP cannot be empty")
	}
	if len(string(cStorVolume.Spec.VolumeName)) == 0 {
		return fmt.Errorf("volumeName cannot be empty")
	}
	if len(string(cStorVolume.Spec.VolumeID)) == 0 {
		return fmt.Errorf("volumeID cannot be empty")
	}
	if len(string(cStorVolume.Spec.Capacity)) == 0 {
		return fmt.Errorf("capacity cannot be empty")
	}

	return nil
}

// CheckForIscsi is blocking call for checking status of istgt in cstor-istgt container.
func CheckForIscsi() {
	for {
		_, err := UnixSockVar.SendCommand(IstgtStatusCmd)
		if err != nil {
			time.Sleep(3 * time.Second)
			glog.Warningf("Waiting for istgt... err : %v", err)
			continue
		}
		break
	}
}
