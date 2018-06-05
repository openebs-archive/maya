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

package iscsi

import (
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-iscsi-mgmt/cstorops/unixsock"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// IscsiOperator is the name of the tool that makes iscsi-related operations.
const (
	IscsiOperator = "iscsi"
	IstgtConfPath = "/usr/local/etc/istgt/istgt.conf"
)

// CreateIscsi creates a new cStor iscsi.
func CreateIscsi(cStorIscsiUpdated *apis.CStorVolume) error {

	generateSparseFile(cStorIscsiUpdated)

	// create conf file
	CreateIstgtConf(cStorIscsiUpdated)

	// send refresh command to istgt
	err := unixsock.SendCommand("REFRESH\n")
	if err != nil {
		glog.Info("refresh failed")
	}
	glog.Info("Creating Iscsi Successful")
	return nil
}

func generateSparseFile(cStorIscsiUpdated *apis.CStorVolume) {

	touchcmd := exec.Command("/usr/bin/touch", "/tmp/cstor/"+cStorIscsiUpdated.Spec.VolumeName)
	_, toucherr := touchcmd.CombinedOutput()
	if toucherr != nil {
		glog.Infof("failed to touch file /tmp/cstor/" + cStorIscsiUpdated.Spec.VolumeName)
		return
	}

	trunccmd := exec.Command("/usr/bin/truncate", "-s", cStorIscsiUpdated.Spec.Capacity,
		"/tmp/cstor/"+cStorIscsiUpdated.Spec.VolumeName)
	_, truncerr := trunccmd.CombinedOutput()
	if truncerr != nil {
		glog.Infof("failed to truncate file /tmp/cstor/" + cStorIscsiUpdated.Spec.VolumeName +
			" with capacity " + cStorIscsiUpdated.Spec.Capacity)
	}

}

// CreateIstgtConf creates istgt.conf file
func CreateIstgtConf(cStorIscsiUpdated *apis.CStorVolume) {
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

	portaluc1 := []byte("  Portal UC1 " + cStorIscsiUpdated.Spec.CStorControllerIP + ":3261\n")
	text = append(text, portaluc1...)

	netmask := []byte("  Netmask " + cStorIscsiUpdated.Spec.CStorControllerIP + "/8\n")
	text = append(text, netmask...)

	text1 := []byte(`
# PortalGroup section
[PortalGroup1]
`)
	text = append(text, text1...)

	portalda1 := []byte("  Portal DA1 " + cStorIscsiUpdated.Spec.CStorControllerIP + ":3260\n")
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

	targetName := []byte("  TargetName " + cStorIscsiUpdated.Spec.VolumeName + "\n")
	text = append(text, targetName...)
	targetAlias := []byte("  TargetAlias nicknamefor-" + cStorIscsiUpdated.Spec.VolumeName)
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
		cStorIscsiUpdated.Spec.VolumeName + " " + cStorIscsiUpdated.Spec.Capacity + " 32k")
	text = append(text, lun0storage...)

	text4 := []byte(`
  LUN0 Option Unmap Disable
  LUN0 Option WZero Disable
  LUN0 Option ATS Disable
  LUN0 Option XCOPY Disable
`)

	text = append(text, text4...)

	err := ioutil.WriteFile(IstgtConfPath, text, 0644)
	if err != nil {
		glog.Errorf("Failed to write istgt.conf...")
	}
	glog.Info("Done writing istgt.conf")
}

// CheckValidIscsi checks for validity of CStorVolume resource.
func CheckValidIscsi(cStorIscsiUpdated *apis.CStorVolume) error {

	return nil
}

// CheckForIscsi is blocking call for checking status of istgt in cstor-iscsi container.
func CheckForIscsi() {
	for {
		err := unixsock.SendCommand("STATUS\n")
		if err != nil {
			time.Sleep(3 * time.Second)
			glog.Infof("Waiting for iscsi...")
			continue
		}
		break
	}
}
