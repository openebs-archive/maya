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

package zvol

import (
	"fmt"
	"testing"

	types "github.com/openebs/maya/pkg/exec"
	mock "github.com/openebs/maya/pkg/exec/mock/v1alpha1"
	mockServer "github.com/openebs/maya/pkg/prometheus/exporter/mock/v1alpha1"
	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
)

func TestPoolSyncTimeMetricCollector(t *testing.T) {
	cases := map[string]struct {
		zpoolLastSyncTimeOutput string
		isError                 bool
		expectedOutput          []string
		runner                  types.Runner
	}{
		"Test0": {
			// expected output if there is openebs.io:timestamp set
			zpoolLastSyncTimeOutput: `cstor-c6f17743-e5d7-11e9-b673-42010a800112      io.openebs:livenesstimestamp    1570625404      local
			cstor-c6f17743-e5d7-11e9-b673-42010a800112/pvc-053647ca-e5d8-11e9-b673-42010a800112     io.openebs:livenesstimestamp    1570625404      inherited from cstor-c6f17743-e5d7-11e9-b673-42010a800112`,
			expectedOutput: []string{
				`openebs_zpool_last_sync_time{pool="cstor-c6f17743-e5d7-11e9-b673-42010a800112"} 1.570625404e\+09`,
				`openebs_zpool_state_unknown{pool="cstor-c6f17743-e5d7-11e9-b673-42010a800112"} 0`,
				`openebs_zpool_sync_time_command_error{pool="cstor-c6f17743-e5d7-11e9-b673-42010a800112"} 0`,
			},
		},
		"Test1": {
			// if expected output from zfs binary is empty
			zpoolLastSyncTimeOutput: ``,
			expectedOutput: []string{
				`openebs_zpool_last_sync_time{pool=""} 0`,
				`openebs_zpool_state_unknown{pool=""} 1`,
				`openebs_zpool_sync_time_command_error{pool=""} 0`,
			},
		},
		"Test2": {
			// if expected output is No pool Available
			zpoolLastSyncTimeOutput: string(zpool.NoPoolAvailable),
			expectedOutput: []string{
				`openebs_zpool_last_sync_time{pool=""} 0`,
				`openebs_zpool_state_unknown{pool=""} 1`,
				`openebs_zpool_sync_time_command_error{pool=""} 0`,
			},
		},
		"Test3": {
			// if there is error executing the command
			isError: true,
			expectedOutput: []string{
				`openebs_zpool_last_sync_time{pool=""} 0`,
				`openebs_zpool_state_unknown{pool=""} 0`,
				`openebs_zpool_sync_time_command_error{pool=""} 1`,
			},
		},
		"Test4": {
			// if there is unexpected output of last sync time  which cannot be parsed
			zpoolLastSyncTimeOutput: `cstor-c6d62069-e5d7-11e9-b673-42010a800112      io.openebs:livenesstimestamp    -       -
			cstor-c6d62069-e5d7-11e9-b673-42010a800112/pvc-053647ca-e5d8-11e9-b673-42010a800112     io.openebs:livenesstimestamp    -       -`,
			expectedOutput: []string{
				`openebs_zpool_last_sync_time{pool="cstor-c6d62069-e5d7-11e9-b673-42010a800112"} 0`,
				`openebs_zpool_state_unknown{pool="cstor-c6d62069-e5d7-11e9-b673-42010a800112"} 0`,
				`openebs_zpool_sync_time_command_error{pool="cstor-c6d62069-e5d7-11e9-b673-42010a800112"} 0`,
			},
		},
		"Test5": {
			// if expected output from zfs binary is empty
			zpoolLastSyncTimeOutput: string(zvol.NoDataSetAvailable),
			expectedOutput: []string{
				`openebs_zpool_last_sync_time{pool=""} 0`,
				`openebs_zpool_state_unknown{pool=""} 1`,
				`openebs_zpool_sync_time_command_error{pool=""} 0`,
			},
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			if tt.isError {
				tt.runner = mock.StdoutBuilder().Error().Build()
			} else {
				out := tt.zpoolLastSyncTimeOutput
				tt.runner = mock.StdoutBuilder().WithOutput(out).Build()
			}
			// Build prometheus like output using regular expressions
			out := tt.expectedOutput
			regex := mockServer.BuildRegex(out)
			vol := NewPoolSyncMetric(tt.runner)
			stop := make(chan struct{})
			buf := mockServer.PrometheusService(vol, stop)
			// expectedOutput the regex after parsing the expected output of zfs list command into prometheus's format.
			for _, re := range regex {
				if !re.Match(buf) {
					fmt.Println(string(buf))
					t.Errorf("failed expectedOutputing: %q", re)
				}
			}
			mockServer.Unregister(vol)
			stop <- struct{}{}
		})
	}
}
