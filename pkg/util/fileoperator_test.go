// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bytes"
	"testing"
	"text/template"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	testConfFile = `# Global section
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
  DesiredReplicationFactor {{.Spec.ReplicationFactor}}
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
  Replica 6161 6061
  Replica 6162 6062334
`
)

var (
	fakePath     = "/tmp/istgt.conf"
	defaultLines = 69
)

func fakeCreateConfFile() error {
	fakeCStorVolume := &apis.CStorVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-volume",
			UID:  types.UID("1234"),
		},
		Spec: apis.CStorVolumeSpec{
			NodeBase:                 "fake-NodeBase",
			TargetIP:                 "127.0.0.1",
			DesiredReplicationFactor: 3,
			ReplicationFactor:        3,
			ConsistencyFactor:        2,
			Capacity:                 resource.MustParse("5G"),
		},
	}
	var dataBytes []byte
	buffer := &bytes.Buffer{}
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"CapacityStr": func(q resource.Quantity) string { return q.String() },
	}).Parse(testConfFile)
	if err != nil {
		return errors.Wrapf(err, "failed to build istgtconffile from template")
	}
	err = tmpl.Execute(buffer, fakeCStorVolume)
	if err != nil {
		return errors.Wrapf(err, "failed execute istgtconfile template")
	}
	dataBytes = buffer.Bytes()
	buffer.Reset()
	fileOperatorVar := RealFileOperator{}
	err = fileOperatorVar.Write(fakePath, dataBytes, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write istgtconfile template")
	}
	return nil
}

func TestUpdateOrAppendMultipleLines(t *testing.T) {
	tests := map[string]struct {
		expectedErr    bool
		keyUpdateValue map[string]string
		totalLines     int
		searchString   map[int]string
	}{
		"Matched values": {
			expectedErr: false,
			keyUpdateValue: map[string]string{
				"  ReplicationFactor": "  ReplicationFactor 4",
				"  ConsistencyFactor": "  ConsistencyFactor 3",
				"  Replica 6162":      "  Replica 6162 6162",
				"  Replica 6163":      "  Replica 6163 6163",
				"  Replica 6164":      "  Replica 6164 6164",
			},
			searchString: map[int]string{
				1: "  Replica 6163 6163",
				2: "  ReplicationFactor 4",
				3: "  Replica 6162 6162",
				4: "  Replica 6161 6161",
				5: "  Replica 6164 6164",
			},
			totalLines: defaultLines + 1,
		},
	}
	fakeFileOperator := RealFileOperator{}
	err := fakeCreateConfFile()
	if err != nil {
		t.Fatalf("test failed: expected error to be nil during confFile Creation")
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		err = fakeFileOperator.UpdateOrAppendMultipleLines(fakePath, mock.keyUpdateValue, 0644)
		if err != nil {
			t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
		}
		for _, value := range mock.searchString {
			_, _, err = fakeFileOperator.GetLineDetails(fakePath, value)
			if err != nil {
				t.Fatalf("test %q failed :to get %s details from file %s", name, value, fakePath)
			}
		}
	}
}
