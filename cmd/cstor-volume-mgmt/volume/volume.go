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
	VolumeOperator = "iscsi"
	IstgtConfPath  = "/usr/local/etc/istgt/istgt.conf"
)

//RunnerVar is
var RunnerVar util.Runner

//FileOperatorVar is
var FileOperatorVar util.FileOperator

//UnixSockVar is
var UnixSockVar util.UnixSock

// CreateVolume creates a new cStor volume istgt config.
func CreateVolume(cStorVolume *apis.CStorVolume) error {

	//generate sparse file
	// generateSparseFile(cStorVolume)
	touchArgs := sparseFileCommandBuilder(cStorVolume, "touch")

	stdoutStderr, err := RunnerVar.RunCombinedOutput("/usr/bin/touch", touchArgs...)
	if err != nil {
		glog.Error("failed to touch file /tmp/cstor/"+cStorVolume.Spec.VolumeName,
			err.Error(), string(stdoutStderr))
		return err
	}

	truncArgs := sparseFileCommandBuilder(cStorVolume, "truncate")

	stdoutStderr, err = RunnerVar.RunCombinedOutput("/usr/bin/truncate", truncArgs...)
	if err != nil {
		glog.Error("failed to truncate file /tmp/cstor/"+cStorVolume.Spec.VolumeName+
			" with capacity "+cStorVolume.Spec.Capacity,
			err.Error(), string(stdoutStderr))
		return err
	}

	// create conf file
	text := CreateIstgtConf(cStorVolume)
	err = FileOperatorVar.Write(IstgtConfPath, text, 0644)
	if err != nil {
		glog.Errorf("Failed to write istgt.conf...")
	}
	glog.Info("Done writing istgt.conf")

	// send refresh command to istgt
	err = UnixSockVar.SendCommand("REFRESH\n")
	if err != nil {
		glog.Info("refresh failed")
	}
	glog.Info("Creating Iscsi Volume Successful")
	return nil

}

// sparseFileArgumentsBuilder is to build sparse file command.
func sparseFileCommandBuilder(cStorVolume *apis.CStorVolume, op string) []string {
	var cmdArgs []string
	if op == "touch" {
		cmdArgs = append(cmdArgs, "/tmp/cstor/"+cStorVolume.Spec.VolumeName)
	} else if op == "truncate" {
		cmdArgs = append(cmdArgs, "-s")
		cmdArgs = append(cmdArgs, cStorVolume.Spec.Capacity)
		cmdArgs = append(cmdArgs, "/tmp/cstor/"+cStorVolume.Spec.VolumeName)
	}

	return cmdArgs
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
  UnitInquiry "OpenEBS" "iscsi" "0" "4059aab98f093c5d95207f7af09d1413"
  PhysRecordLength 4096
`)
	text = append(text, text3...)

	lun0storage := []byte("  LUN0 Storage /tmp/cstor/" +
		cStorVolume.Spec.VolumeName + " " + cStorVolume.Spec.Capacity + " 32k")
	text = append(text, lun0storage...)

	text4 := []byte(`
  LUN0 Option Unmap Disable
  LUN0 Option WZero Disable
  LUN0 Option ATS Disable
  LUN0 Option XCOPY Disable
`)

	text = append(text, text4...)

	return text
}

// CheckValidVolume checks for validity of CStorVolume resource.
func CheckValidVolume(cStorVolume *apis.CStorVolume) error {
	if string(cStorVolume.ObjectMeta.UID) == "" {
		return fmt.Errorf("Volumename cannot be empty")
	}
	return nil
}

// CheckForIscsi is blocking call for checking status of istgt in cstor-iscsi container.
func CheckForIscsi() {
	for {
		err := UnixSockVar.SendCommand("STATUS\n")
		if err != nil {
			time.Sleep(3 * time.Second)
			glog.Infof("Waiting for iscsi...")
			continue
		}
		break
	}
}
