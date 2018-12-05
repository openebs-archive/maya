// Copyright © 2017 The OpenEBS Authors
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
	"errors"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	utiltesting "k8s.io/client-go/util/testing"
)

func TestNewCmdPoolList(t *testing.T) {
	tests := map[string]*struct {
		expectedCmd *cobra.Command
	}{
		"NewCmdVolumeStats": {
			expectedCmd: &cobra.Command{
				Use:   "list",
				Short: "Lists all the pools",
				Long:  poolListCommandHelpText,
				Run: func(cmd *cobra.Command, args []string) {
					util.CheckErr(options.runPoolList(cmd), util.Fatal)
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewCmdPoolList()
			if (got.Use != tt.expectedCmd.Use) || (got.Short != tt.expectedCmd.Short) || (got.Long != tt.expectedCmd.Long) || (got.Example != tt.expectedCmd.Example) {
				t.Fatalf("TestName: %v | NewCmdPoolList() => Got: %v | Want: %v \n", name, got, tt.expectedCmd)
			}
		})
	}
}

func TestRunPoolList(t *testing.T) {
	options := CmdPoolOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all the pools",
		Long:  poolListCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.runPoolList(cmd), util.Fatal)
		},
	}
	tests := map[string]*struct {
		cmdPoolOptions *CmdPoolOptions
		cmd            *cobra.Command
		fakeHandler    utiltesting.FakeHandler
		err            error
		addr           string
	}{
		"StatusOK": {
			cmd:            cmd,
			cmdPoolOptions: &CmdPoolOptions{},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: `{"items":[{"apiVersion":"openebs.io/v1alpha1","kind":"StoragePool","metadata":{"creationTimestamp":"2018-11-15T07:53:56Z","generation":1,"labels":{"openebs.io/cas-type":"cstor","openebs.io/cstor-pool":"cstor-sparse-pool-g5pi","openebs.io/storage-pool-claim":"cstor-sparse-pool","openebs.io/version":"0.7.0","kubernetes.io/hostname":"127.0.0.1","openebs.io/cas-template-name":"cstor-pool-create-default-0.7.0"},"name":"cstor-sparse-pool-g5pi","resourceVersion":"580","selfLink":"/apis/openebs.io/v1alpha1/storagepools/cstor-sparse-pool-g5pi","uid":"9a6a4b68-e8ab-11e8-b96a-b4b686bd0cff"},"spec":{"disks":{"diskList":["sparse-5a92ced3e2ee21eac7b930f670b5eab5"]},"format":"","message":"","mountpoint":"","name":"","nodename":"","path":"","poolSpec":{"cacheFile":"/tmp/cstor-sparse-pool.cache","overProvisioning":false,"poolType":"striped"}}}],"metadata":{"resourceVersion":"658","selfLink":"/apis/openebs.io/v1alpha1/storagepools"}}`,
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"No pools present": {
			cmd:            cmd,
			cmdPoolOptions: &CmdPoolOptions{},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: `{"items":[],"metadata":{"resourceVersion":"650","selfLink":"/apis/openebs.io/v1alpha1/storagepools"}}`,
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"Invalid Response": {
			cmd:            cmd,
			cmdPoolOptions: &CmdPoolOptions{},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: `"name":"go_gc_duration_seconds","help":"A summary of the GC invocation durations.","type":2,"metric":[{"summary":{"sample_count":0,"sample_sum":0,"quantile":[{"quantile":0,"value":0},{"quantile":0.25,"value":0},{"quantile":0.5,"value":0},{"quantile":0.75,"value":0},{"quantile":1,"value":0}]}}]},{"name":"go_goroutines","help":"Number of goroutines that currently exist.","type":1,"metric":[{"gauge":{"value":12}}]},{"name":"go_memstats_alloc_bytes","help":"Number of bytes allocated and still in use.","type":1,"metric":[{"gauge":{"value":1221368}}]},{"name":"go_memstats_alloc_bytes_total","help":"Total number of bytes allocated, even if freed.","type":0,"metric":[{"counter":{"value":1221368}}]},{"name":"go_memstats_buck_hash_sys_bytes","help":"Number of bytes used by the profiling bucket hash table.","type":1,"metric":[{"gauge":{"value":2792}}]},{"name":"go_memstats_frees_total","help":"Total number of frees.","type":0,"metric":[{"counter":{"value":431}}]},{"name":"go_memstats_gc_sys_bytes","help":"Number of bytes used for garbage collection system metadata.","type":1,"metric":[{"gauge":{"value":169984}}]},{"name":"go_memstats_heap_alloc_bytes","help":"Number of heap bytes allocated and still in use.","type":1,"metric":[{"gauge":{"value":1221368}}]},{"name":"go_memstats_heap_idle_bytes","help":"Number of heap bytes waiting to be used.","type":1,"metric":[{"gauge":{"value":237568}}]},{"name":"go_memstats_heap_inuse_bytes","help":"Number of heap bytes that are in use.","type":1,"metric":[{"gauge":{"value":2318336}}]},{"name":"go_memstats_heap_objects","help":"Number of allocated objects.","type":1,"metric":[{"gauge":{"value":7375}}]},{"name":"go_memstats_heap_released_bytes_total","help":"Total number of heap bytes released to OS.","type":0,"metric":[{"counter":{"value":0}}]},{"name":"go_memstats_heap_sys_bytes","help":"Number of heap bytes obtained from system.","type":1,"metric":[{"gauge":{"value":2555904}}]},{"name":"go_memstats_last_gc_time_seconds","help":"Number of seconds since 1970 of last garbage collection.","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"go_memstats_lookups_total","help":"Total number of pointer lookups.","type":0,"metric":[{"counter":{"value":24}}]},{"name":"go_memstats_mallocs_total","help":"Total number of mallocs.","type":0,"metric":[{"counter":{"value":7806}}]},{"name":"go_memstats_mcache_inuse_bytes","help":"Number of bytes in use by mcache structures.","type":1,"metric":[{"gauge":{"value":13888}}]},{"name":"go_memstats_mcache_sys_bytes","help":"Number of bytes used for mcache structures obtained from system.","type":1,"metric":[{"gauge":{"value":16384}}]},{"name":"go_memstats_mspan_inuse_bytes","help":"Number of bytes in use by mspan structures.","type":1,"metric":[{"gauge":{"value":32832}}]},{"name":"go_memstats_mspan_sys_bytes","help":"Number of bytes used for mspan structures obtained from system.","type":1,"metric":[{"gauge":{"value":49152}}]},{"name":"go_memstats_next_gc_bytes","help":"Number of heap bytes when next garbage collection will take place.","type":1,"metric":[{"gauge":{"value":4473924}}]},{"name":"go_memstats_other_sys_bytes","help":"Number of bytes used for other system allocations.","type":1,"metric":[{"gauge":{"value":1305880}}]},{"name":"go_memstats_stack_inuse_bytes","help":"Number of bytes in use by the stack allocator.","type":1,"metric":[{"gauge":{"value":589824}}]},{"name":"go_memstats_stack_sys_bytes","help":"Number of bytes obtained from system for stack allocator.","type":1,"metric":[{"gauge":{"value":589824}}]},{"name":"go_memstats_sys_bytes","help":"Number of bytes obtained by system. Sum of all system allocations.","type":1,"metric":[{"gauge":{"value":4689920}}]},{"name":"http_request_duration_microseconds","help":"The HTTP request latencies in microseconds.","type":2,"metric":[{"label":[{"name":"handler","value":"prometheus"}],"summary":{"sample_count":1,"sample_sum":16749.686,"quantile":[{"quantile":0.5,"value":16749.686},{"quantile":0.9,"value":16749.686},{"quantile":0.99,"value":16749.686}]}}]},{"name":"http_request_size_bytes","help":"The HTTP request sizes in bytes.","type":2,"metric":[{"label":[{"name":"handler","value":"prometheus"}],"summary":{"sample_count":1,"sample_sum":65,"quantile":[{"quantile":0.5,"value":65},{"quantile":0.9,"value":65},{"quantile":0.99,"value":65}]}}]},{"name":"http_requests_total","help":"Total number of HTTP requests made.","type":0,"metric":[{"label":[{"name":"code","value":"200"},{"name":"handler","value":"prometheus"},{"name":"method","value":"get"}],"counter":{"value":1}}]},{"name":"http_response_size_bytes","help":"The HTTP response sizes in bytes.","type":2,"metric":[{"label":[{"name":"handler","value":"prometheus"}],"summary":{"sample_count":1,"sample_sum":7948,"quantile":[{"quantile":0.5,"value":7948},{"quantile":0.9,"value":7948},{"quantile":0.99,"value":7948}]}}]},{"name":"openebs_actual_used","help":"Actual volume size used","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"openebs_logical_size","help":"Logical size of volume","type":1,"metric":[{"gauge":{"value":0.0000152587890625}}]},{"name":"openebs_read_block_count","help":"Read Block count of volume","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"openebs_read_time","help":"Read time on volume","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"openebs_reads","help":"Read Input/Outputs on Volume","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"openebs_sector_size","help":"sector size of volume","type":1,"metric":[{"gauge":{"value":4096}}]},{"name":"openebs_size_of_volume","help":"Size of the volume requested","type":1,"metric":[{"gauge":{"value":5}}]},{"name":"openebs_total_read_bytes","help":"Total read bytes","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"openebs_total_write_bytes","help":"Total write bytes","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"openebs_volume_uptime","help":"Time since volume has registered","type":0,"metric":[{"label":[{"name":"castype","value":"jiva"},{"name":"iqn","value":"iqn.2016-09.com.openebs.jiva:pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"},{"name":"portal","value":"127.0.0.1"},{"name":"volName","value":"pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"}],"counter":{"value":104.436802}}]},{"name":"openebs_write_block_count","help":"Write Block count of volume","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"openebs_write_time","help":"Write time on volume","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"openebs_writes","help":"Write Input/Outputs on Volume","type":1,"metric":[{"gauge":{"value":0}}]},{"name":"process_cpu_seconds_total","help":"Total user and system CPU time spent in seconds.","type":0,"metric":[{"counter":{"value":0.05}}]},{"name":"process_max_fds","help":"Maximum number of open file descriptors.","type":1,"metric":[{"gauge":{"value":1048576}}]},{"name":"process_open_fds","help":"Number of open file descriptors.","type":1,"metric":[{"gauge":{"value":8}}]},{"name":"process_resident_memory_bytes","help":"Resident memory size in bytes.","type":1,"metric":[{"gauge":{"value":6705152}}]},{"name":"process_start_time_seconds","help":"Start time of the process since unix epoch in seconds.","type":1,"metric":[{"gauge":{"value":1542187608.82}}]},{"name":"process_virtual_memory_bytes","help":"Virtual memory size in bytes.","type":1,"metric":[{"gauge":{"value":13996032}}]}]`,
				T:            t,
			},
			err:  errors.New("Error listing pools: invalid character ':' after top-level value"),
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			cmd:            cmd,
			cmdPoolOptions: &CmdPoolOptions{},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: "",
				T:            t,
			},
			err:  errors.New("Error listing pools: Server status error: Not Found"),
			addr: "MAPI_ADDR",
		},
		"Response code 500": {
			cmd:            cmd,
			cmdPoolOptions: &CmdPoolOptions{},
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   500,
				ResponseBody: "",
				T:            t,
			},
			err:  errors.New("Error listing pools: Server status error: Internal Server Error"),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := tt.cmdPoolOptions.runPoolList(tt.cmd)
			if !checkErr(got, tt.err) {
				t.Fatalf("TestName: %v | runPoolList() => Got: %v | Want: %v \n", name, got, tt.err)
			}
		})
	}
}
