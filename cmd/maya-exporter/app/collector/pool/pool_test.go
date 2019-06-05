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

package pool

import (
	"fmt"
	"testing"

	types "github.com/openebs/maya/pkg/exec"
	mock "github.com/openebs/maya/pkg/exec/mock/v1alpha1"
	mockServer "github.com/openebs/maya/pkg/prometheus/exporter/mock/v1alpha1"
	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
)

func TestPoolCollector(t *testing.T) {
	var runner types.Runner
	cases := map[string]struct {
		zpoolOutput    string
		isError        bool
		expectedOutput []string
	}{
		// pool status is online
		"Test0": {
			zpoolOutput: "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 ONLINE	-",
			expectedOutput: []string{
				`openebs_pool_size 1024`,
				`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 1`,
				`openebs_used_pool_capacity 24`,
				`openebs_free_pool_capacity 1000`,
				`openebs_used_pool_capacity_percent 0`,
			},
		},
		// pool status is offline
		"Test1": {
			zpoolOutput: "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 OFFLINE	-",
			expectedOutput: []string{
				`openebs_pool_size 1024`,
				`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 0`,
				`openebs_used_pool_capacity 24`,
				`openebs_free_pool_capacity 1000`,
				`openebs_used_pool_capacity_percent 0`,
			},
		},
		// pool status is unavailable
		"Test2": {
			zpoolOutput: "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 UNAVAIL	-",
			expectedOutput: []string{
				`openebs_pool_size 1024`,
				`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 5`,
				`openebs_used_pool_capacity 24`,
				`openebs_free_pool_capacity 1000`,
				`openebs_used_pool_capacity_percent 0`,
			},
		},
		// pool status is faulted
		"Test3": {
			zpoolOutput: "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 FAULTED	-",
			expectedOutput: []string{
				`openebs_pool_size 1024`,
				`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 3`,
				`openebs_used_pool_capacity 24`,
				`openebs_free_pool_capacity 1000`,
				`openebs_used_pool_capacity_percent 0`,
			},
		},
		// pool status is removed
		"Test4": {
			zpoolOutput: "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 REMOVED	-",
			expectedOutput: []string{
				`openebs_pool_size 1024`,
				`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 4`,
				`openebs_used_pool_capacity 24`,
				`openebs_free_pool_capacity 1000`,
				`openebs_used_pool_capacity_percent 0`,
			},
		},
		// no pools available
		"Test5": {
			zpoolOutput: zpool.NoPoolAvailable.String(),
			expectedOutput: []string{
				`openebs_zpool_list_no_pool_available_error 1`,
			},
		},
		// incomplete stdout of zpool list command
		"Test6": {
			zpoolOutput: "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a  1024    24  1000    -",
			expectedOutput: []string{
				`openebs_zpool_list_incomplete_stdout_error 1`,
			},
		},
		// if there is an error while running zpool list command
		"Test7": {
			isError: true,
			expectedOutput: []string{
				`openebs_zpool_list_command_error 1`,
			},
		},
		// if there is unexpected response
		"Test8": {
			zpoolOutput: "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	iaueb7	aiwub	aliubv	-	0	iauwb	1.00 REMOVED	-",
			expectedOutput: []string{
				`openebs_zpool_list_parse_error_count 4`,
			},
		},
		// if failed to initialize libuzfs client err
		"Test9": {
			zpoolOutput: zpool.InitializeLibuzfsClientErr.String(),
			expectedOutput: []string{
				`openebs_zpool_list_failed_to_initialize_libuzfs_client_error_counter 1`,
			},
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			if tt.isError {
				runner = mock.StdoutBuilder().Error().Build()
			} else {
				out := tt.zpoolOutput
				runner = mock.StdoutBuilder().WithOutput(out).Build()
			}
			// Build prometheus like output using regular expressions
			out := tt.expectedOutput
			regex := mockServer.BuildRegex(out)
			pool := New(runner)
			stop := make(chan struct{})
			buf := mockServer.PrometheusService(pool, stop)
			// expectedOutput the regex after parsing the expected output of zfs
			// list command into prometheus's format.
			for _, re := range regex {
				if !re.Match(buf) {
					fmt.Println(string(buf))
					t.Errorf("failed expectedOutputing: %q", re)
				}
			}
			mockServer.Unregister(pool)
			stop <- struct{}{}
		})
	}
}
