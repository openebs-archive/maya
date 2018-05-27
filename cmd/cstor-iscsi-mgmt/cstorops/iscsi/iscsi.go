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

	// create conf file
	CreateIstgtConf(cStorIscsiUpdated)

	// send refresh command to istgt
	err := unixsock.SendRefresh()
	if err != nil {
		glog.Info("refresh failed")
	}
	glog.Info("Creating Iscsi Successful")
	return nil
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
  Portal UC1 localhost:3261
  Netmask localhost/8
# PortalGroup section
[PortalGroup1]
  Portal DA1 localhost:3260

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
	targetName := []byte("  TargetName " + cStorIscsiUpdated.Spec.VolumeName + "\n")
	text = append(text, targetName...)
	targetAlias := []byte("  TargetAlias nicknamefor-" + cStorIscsiUpdated.Spec.VolumeName)
	text = append(text, targetAlias...)

	text2 := []byte(`
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
  UnitInquiry "CloudByte" "iscsi" "0" "4059aab98f093c5d95207f7af09d1413"
  PhysRecordLength 4096
  LUN0 Storage /home/payes/vol1 1G 32k
  LUN0 Option Unmap Disable
  LUN0 Option WZero Disable
  LUN0 Option ATS Disable
  LUN0 Option XCOPY Disable
  `)

	text = append(text, text2...)

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
		err := unixsock.Status()
		if err != nil {
			time.Sleep(3 * time.Second)
			glog.Infof("Waiting for iscsi...")
			continue
		}
		break
	}
}
