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

func TestZfsStatsCollector(t *testing.T) {
	var runner types.Runner
	cases := map[string]struct {
		zfsStatsOutput string
		isError        bool
		expectedOutput []string
	}{
		"Test0": {
			// expected output if there is one volume with different stats and status is Healthy and rebuild status id DONE.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
			},
		},
		"Test1": {
			// expected output if there is one volume with different stats and replica status is Offline and rebuild status is INIT.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Offline","rebuildStatus": "INIT","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`,
			},
		},
		"Test2": {
			// expected output if there is one volume with different stats and replica status is Degraded and rebuild status is INIT.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Degraded","rebuildStatus": "INIT","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`,
			},
		},
		"Test3": {
			// expected output if there is one volume with different stats and replica status is Rebuilding and rebuild status is SNAP REBUILD INPROGRESS.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Rebuilding","rebuildStatus": "SNAP REBUILD INPROGRESS","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2`,
			},
		},
		"Test4": {
			// expected output if there is one volume with different stats and replica status is Rebuilding and rebuild status is ACTIVE DATASET REBUILD INPROGRESS.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Rebuilding","rebuildStatus": "ACTIVE DATASET REBUILD INPROGRESS","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
			},
		},
		"Test5": {
			// expected output if there is one volume with different stats and replica status is Offline and rebuild status is ERRORED.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Offline","rebuildStatus": "ERRORED  ","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 4`,
			},
		},

		"Test6": {
			// expected output if there is one volume with different stats and replica status is Offline and rebuild status is FAILED.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Offline","rebuildStatus": "FAILED","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 5`,
			},
		},
		"Test7": {
			// expected output if there is one volume with different stats and replica status is Offline and rebuild status is UNKNOWN.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Offline","rebuildStatus": "UNKNOWN","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 6`,
			},
		},
		"Test8": {
			// expected error while running zfs binary
			isError: true,
			expectedOutput: []string{
				`openebs_zfs_stats_command_error 1`,
			},
		},
		"Test9": {
			// expected empty output from zfs
			zfsStatsOutput: ``,
			expectedOutput: []string{
				`openebs_zfs_stats_parse_error_counter 1`,
			},
		},
		"Test10": {
			// expected output if there are two volumes with different stats and status is Healthy and rebuild status id DONE.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50},{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
			},
		},
		"Test11": {
			// expected output if there are three volumes with different stats and status is Healthy and rebuild status id DONE.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50},{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50},{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
			},
		},
		"Test12": {
			// expected output if there are three volumes with different stats and status of one is Healthy and rebuild status is DONE but status of other two volumes is Degraded and rebuild status is init.
			zfsStatsOutput: `{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50},{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a","status": "Degraded","rebuildStatus": "INIT","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50},{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a","status": "Degraded","rebuildStatus": "INIT","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`,
			expectedOutput: []string{
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 100`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 10`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 150`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 200`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 2`,
				`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 2`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 50`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 3`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 500`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 0`,
				`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 0`,
			},
		},
		"Test13": {
			// if failed to initialize libuzfs client
			zfsStatsOutput: zvol.InitializeLibuzfsClientErr.String(),
			expectedOutput: []string{
				`openebs_zfs_stats_failed_to_initialize_libuzfs_client_error_counter 1`,
			},
		},
		"Test14": {
			// if no dataset available
			zfsStatsOutput: zvol.NoDataSetAvailable.String(),
			expectedOutput: []string{
				`openebs_zfs_stats_no_dataset_available_error_counter 1`,
			},
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			if tt.isError {
				runner = mock.StdoutBuilder().Error().Build()
			} else {
				out := tt.zfsStatsOutput
				runner = mock.StdoutBuilder().WithOutput(out).Build()
			}
			// Build prometheus like output using regular expressions
			out := tt.expectedOutput
			regex := mockServer.BuildRegex(out)
			vol := New(runner)
			stop := make(chan struct{})
			buf := mockServer.PrometheusService(vol, stop)
			// expectedOutput the regex after parsing the expected output of zfs
			// stats command into prometheus's format.
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
