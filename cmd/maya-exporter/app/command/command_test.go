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

package command

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Short: "Collect metrics from OpenEBS volumes",
		Long: `maya-exporter can be used to monitor openebs volumes and pools.
It can be deployed alongside the openebs volume or pool containers as sidecars.`,
		Example: `maya-exporter -a=http://localhost:8001 -c=:9500 -m=/metrics`,
	}
)

func TestRegisterJiva(t *testing.T) {
	cases := map[string]struct {
		option *VolumeExporterOptions
		output error
	}{
		"ValidURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "http://localhost:9501",
			},
			output: nil,
		},
		"InvalidURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "localhost",
			},
			output: errors.New("Error in parsing the URI"),
		},
		"EmptyURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "",
			},
			output: errors.New("Error in parsing the URI"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := tt.option.RegisterJiva()
			if !reflect.DeepEqual(got, tt.output) {
				t.Fatalf("RegisterJivaStatsExporter() => [%v], want [%v]", got, tt.output)
			}
		})
	}
}
