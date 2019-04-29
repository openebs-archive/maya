// Copyright © 2017-2019 The OpenEBS Authors
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

package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestMayaConfig_Parse(t *testing.T) {
	cases := []struct {
		File   string
		Result *MayaConfig
		Err    bool
	}{
		{
			"dummy_mayaserver_config.hcl",
			&MayaConfig{
				Region:      "BANG-EAST",
				Datacenter:  "dc2",
				NodeName:    "my-vsm",
				DataDir:     "/tmp/mayaserver",
				LogLevel:    "ERR",
				BindAddr:    "192.168.0.1",
				EnableDebug: true,
				Ports: &Ports{
					HTTP: 1234,
				},
				Addresses: &Addresses{
					HTTP: "127.0.0.1",
				},
				AdvertiseAddrs: &AdvertiseAddrs{},
				LeaveOnInt:     true,
				LeaveOnTerm:    true,
				EnableSyslog:   true,
				SyslogFacility: "LOCAL1",
				HTTPAPIResponseHeaders: map[string]string{
					"Access-Control-Allow-Origin": "*",
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Logf("Testing parse: %s", tc.File)

		path, err := filepath.Abs(filepath.Join("../mockit", tc.File))
		if err != nil {
			t.Fatalf("file: %s\n\n%s", tc.File, err)
			continue
		}

		actual, err := ParseMayaConfigFile(path)
		if (err != nil) != tc.Err {
			t.Fatalf("file: %s\n\n%s", tc.File, err)
			continue
		}

		if !reflect.DeepEqual(actual, tc.Result) {
			t.Fatalf("file: %s\nactual:%v\n\nexpected:%v", tc.File, actual, tc.Result)
		}
	}
}
