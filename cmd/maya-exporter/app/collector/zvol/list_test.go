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
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
)

func TestZfsListCollector(t *testing.T) {
	cases := map[string]struct {
		zfsListOutput  string
		isError        bool
		expectedOutput []string
		runner         types.Runner
	}{
		"Test0": {
			// expected output if there is one volume with different used size
			zfsListOutput: `cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`,
			expectedOutput: []string{
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 238592`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`,
			},
		},
		"Test1": {
			// expected output if there is one volume with different used size
			zfsListOutput: `cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	3055	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`,
			expectedOutput: []string{
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 238592`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3055`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`,
			},
		},
		"Test2": {
			// expected output if there are two volumes
			zfsListOutput: `cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`,
			expectedOutput: []string{
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 238592`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`,
			},
		},
		"Test3": {
			// Expected output when there are three volumes
			zfsListOutput: `cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`,
			expectedOutput: []string{
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`,
				`openebs_volume_replica_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 238592`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`,
				`openebs_volume_replica_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`,
			},
		},
		"Test4": {
			// if there is unexpected output from zpool
			zfsListOutput: `cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	liaub	kzjsfvn	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`,
			expectedOutput: []string{
				`openebs_zfs_list_parse_error 2`,
			},
		},
		"Test5": {
			// if there is any error while running zfs list command
			isError: true,
			expectedOutput: []string{
				`openebs_zfs_list_command_error 1`,
			},
		},
		"Test6": {
			// if expected output from zfs binary is empty
			zfsListOutput: ``,
			expectedOutput: []string{
				`openebs_zfs_list_parse_error 1`,
			},
		},
		"Test7": {
			// if expected output from is "failed to initialize libuzfs client"
			zfsListOutput: string(zvol.InitializeLibuzfsClientErr),
			expectedOutput: []string{
				`zfs_list_failed_to_initialize_libuzfs_client_error_counter 1`,
			},
		},
		"Test8": {
			// if expected output from zfs binary is empty
			zfsListOutput: string(zvol.NoDataSetAvailable),
			expectedOutput: []string{
				`zfs_list_no_dataset_available_error_counter 1`,
			},
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			if tt.isError {
				tt.runner = mock.StdoutBuilder().Error().Build()
			} else {
				out := tt.zfsListOutput
				tt.runner = mock.StdoutBuilder().WithOutput(out).Build()
			}
			// Build prometheus like output using regular expressions
			out := tt.expectedOutput
			regex := mockServer.BuildRegex(out)
			vol := NewVolumeList(tt.runner)
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
