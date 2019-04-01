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

package command

import (
	"errors"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	utiltesting "k8s.io/client-go/util/testing"
)

func TestNewCmdVolumeStats(t *testing.T) {
	tests := map[string]*struct {
		expectedCmd *cobra.Command
	}{
		"NewCmdVolumeStats": {
			expectedCmd: &cobra.Command{
				Use:     "stats",
				Short:   "Displays the runtime statisics of Volume",
				Long:    volumeStatsCommandHelpText,
				Example: ` mayactl volume stats --volname=vol`,
				Run: func(cmd *cobra.Command, args []string) {
					util.CheckErr(options.Validate(cmd, false, false, true), util.Fatal)
					util.CheckErr(options.runVolumeStats(cmd), util.Fatal)
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewCmdVolumeStats()
			if (got.Use != tt.expectedCmd.Use) || (got.Short != tt.expectedCmd.Short) || (got.Long != tt.expectedCmd.Long) || (got.Example != tt.expectedCmd.Example) {
				t.Fatalf("TestName: %v | processStats() => Got: %v | Want: %v \n", name, got, tt.expectedCmd)
			}
		})
	}
}

func TestRunVolumeStats(t *testing.T) {
	options := CmdVolumeOptions{}
	cmd := &cobra.Command{
		Use:     "stats",
		Short:   "Displays the runtime statisics of Volume",
		Long:    volumeStatsCommandHelpText,
		Example: ` mayactl volume stats --volname=vol`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd, false, false, true), util.Fatal)
			util.CheckErr(options.runVolumeStats(cmd), util.Fatal)
		},
	}

	tests := map[string]*struct {
		cmdVolumeOptions *CmdVolumeOptions
		cmd              *cobra.Command
		fakeHandler      utiltesting.FakeHandler
		err              error
		addr             string
	}{
		"Status OK": {
			cmd:              cmd,
			cmdVolumeOptions: &CmdVolumeOptions{volName: "vol1"},
			fakeHandler:      utiltesting.FakeHandler{StatusCode: 200, ResponseBody: `[{"name":"openebs_actual_used","metric":[{"gauge":{"value":0}}]},{"name":"openebs_logical_size","metric":[{"gauge":{"value":0.0000152587890625}}]},{"name":"openebs_read_block_count","metric":[{"gauge":{"value":0}}]},{"name":"openebs_read_time","metric":[{"gauge":{"value":0}}]},{"name":"openebs_reads","metric":[{"gauge":{"value":0}}]},{"name":"openebs_sector_size","metric":[{"gauge":{"value":4096}}]},{"name":"openebs_size_of_volume","metric":[{"gauge":{"value":5}}]},{"name":"openebs_total_read_bytes","metric":[{"gauge":{"value":0}}]},{"name":"openebs_total_write_bytes","metric":[{"gauge":{"value":0}}]},{"name":"openebs_volume_uptime","metric":[{"label":[{"name":"castype","value":"jiva"},{"name":"iqn","value":"iqn.2016-09.com.openebs.jiva:pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"},{"name":"portal","value":"127.0.0.1"},{"name":"volName","value":"pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"}],"counter":{"value":104.436802}}]},{"name":"openebs_write_block_count","metric":[{"gauge":{"value":0}}]},{"name":"openebs_write_time","metric":[{"gauge":{"value":0}}]},{"name":"openebs_writes","metric":[{"gauge":{"value":0}}]}]`, T: t},
			err:              nil,
			addr:             "MAPI_ADDR",
		},
		"Metric is empty": {
			cmd:              cmd,
			cmdVolumeOptions: &CmdVolumeOptions{volName: "vol1"},
			fakeHandler:      utiltesting.FakeHandler{StatusCode: 200, ResponseBody: `[{"name":"openebs_actual_used","metric":[]},{"name":"openebs_logical_size","metric":[]},{"name":"openebs_read_block_count","metric":[]},{"name":"openebs_read_time","metric":[]},{"name":"openebs_reads","metric":[]},{"name":"openebs_sector_size","metric":[]},{"name":"openebs_size_of_volume","metric":[]},{"name":"openebs_total_read_bytes","metric":[]},{"name":"openebs_total_write_bytes","metric":[]},{"name":"openebs_volume_uptime","metric":[]},{"name":"openebs_write_block_count","metric":[]},{"name":"openebs_write_time","metric":[]},{"name":"openebs_writes","metric":[]}]`, T: t},
			err:              nil,
			addr:             "MAPI_ADDR",
		},
		"Invalid Response": {
			cmd:              cmd,
			cmdVolumeOptions: &CmdVolumeOptions{volName: "vol1"},
			fakeHandler:      utiltesting.FakeHandler{StatusCode: 200, ResponseBody: `[{"name":"openebs_actual_used","metric":[]},{"name":"openebs_logical_size","metric":[]},{"name":"openebs_read_block_count","metric":[]},{"name":"openebs_read_time","metric":[]},{"name":"openebs_reads","metric":[]},{"name":"openebs_sector_size","metric":[]},{"name":"openebs_size_of_volume","metric":[]},{"name":"openebs_total_read_bytes","metric":[]},{"name":"openebs_total_write_bytes","metric":[]},{"name":"openebs_volume_uptime","metric":[]},{"name":"openebs_write_block_count","metric":[]},{"name":"openebs_write_time","metric":[]},{"name":"openebs_writes","metric":[`, T: t},
			err:              errors.New("Volume not found"),
			addr:             "MAPI_ADDR",
		},
		"BadRequest": {
			cmd:              cmd,
			cmdVolumeOptions: &CmdVolumeOptions{volName: "vol1"},
			fakeHandler:      utiltesting.FakeHandler{StatusCode: 404, ResponseBody: "", T: t},
			err:              errors.New("Volume not found"),
			addr:             "MAPI_ADDR",
		},
		"Response code 500": {
			cmd:              cmd,
			cmdVolumeOptions: &CmdVolumeOptions{volName: "vol1"},
			fakeHandler:      utiltesting.FakeHandler{StatusCode: 500, ResponseBody: "", T: t},
			err:              errors.New("Volume not found"),
			addr:             "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := tt.cmdVolumeOptions.runVolumeStats(tt.cmd)
			if !checkErr(got, tt.err) {
				t.Fatalf("TestName: %v | runVolumeStats() => Got: %v | Want: %v \n", name, got, tt.err)
			}
		})
	}
}

func TestProcessStats(t *testing.T) {
	tests := map[string]struct {
		stats1, stats2 map[string]v1alpha1.MetricsFamily
		Output         v1alpha1.StatsJSON
	}{
		"When length of stats1 is not equal to stat2": {
			stats1: map[string]v1alpha1.MetricsFamily{},
			stats2: map[string]v1alpha1.MetricsFamily{
				"openebs_reads":             v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_writes":            v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_read_time":         v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 0}},
				"openebs_write_time":        v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 0}},
				"openebs_read_block_count":  v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 0}},
				"openebs_write_block_count": v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 0}},
				"openebs_sector_size":       v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_logical_size":      v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 1.52587890625e-5}},
				"openebs_actual_used":       v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_size_of_volume":    v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 5}},
				"openebs_volume_uptime":     v1alpha1.MetricsFamily{Label: []v1alpha1.LabelItem{v1alpha1.LabelItem{Name: "castype", Value: "jiva"}, v1alpha1.LabelItem{Name: "iqn", Value: "iqn.2016-09.com.openebs.jiva:pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"}, v1alpha1.LabelItem{Name: "portal", Value: "127.0.0.1"}, v1alpha1.LabelItem{Name: "volName", Value: "pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"}}},
			},
			Output: v1alpha1.StatsJSON{Volume: "pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff", Size: "5.000000", CASType: "jiva", ReadIOPS: 4096, WriteIOPS: 4096, ReadThroughput: 0, WriteThroughput: 0, ReadLatency: 0, WriteLatency: 0, AvgReadBlockSize: 0, AvgWriteBlockSize: 0, SectorSize: 4096, ActualUsed: 4096, LogicalSize: 1.52587890625e-05},
		},
		"When length of stats1 is  equal to stat2": {
			stats1: map[string]v1alpha1.MetricsFamily{
				"openebs_reads":             v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2048}},
				"openebs_writes":            v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2048}},
				"openebs_read_time":         v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2}},
				"openebs_write_time":        v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2}},
				"openebs_read_block_count":  v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2}},
				"openebs_write_block_count": v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2}},
				"openebs_sector_size":       v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_logical_size":      v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 1.52587890625e-5}},
				"openebs_actual_used":       v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2048}},
				"openebs_size_of_volume":    v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 5}},
				"openebs_volume_uptime":     v1alpha1.MetricsFamily{Label: []v1alpha1.LabelItem{v1alpha1.LabelItem{Name: "castype", Value: "jiva"}, v1alpha1.LabelItem{Name: "iqn", Value: "iqn.2016-09.com.openebs.jiva:pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"}, v1alpha1.LabelItem{Name: "portal", Value: "127.0.0.1"}, v1alpha1.LabelItem{Name: "volName", Value: "pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"}}},
			},
			stats2: map[string]v1alpha1.MetricsFamily{
				"openebs_reads":             v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_writes":            v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_read_time":         v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4}},
				"openebs_write_time":        v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4}},
				"openebs_read_block_count":  v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4}},
				"openebs_write_block_count": v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4}},
				"openebs_sector_size":       v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_logical_size":      v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 1.52687890625e-5}},
				"openebs_actual_used":       v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 4096}},
				"openebs_size_of_volume":    v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 5}},
				"openebs_volume_uptime":     v1alpha1.MetricsFamily{Label: []v1alpha1.LabelItem{v1alpha1.LabelItem{Name: "castype", Value: "jiva"}, v1alpha1.LabelItem{Name: "iqn", Value: "iqn.2016-09.com.openebs.jiva:pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"}, v1alpha1.LabelItem{Name: "portal", Value: "127.0.0.1"}, v1alpha1.LabelItem{Name: "volName", Value: "pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff"}}},
			},
			Output: v1alpha1.StatsJSON{Volume: "pvc-66a9a0b4-e7ef-11e8-a279-b4b686bd0cff", Size: "5.000000", CASType: "jiva", ReadIOPS: 2048, WriteIOPS: 2048, ReadThroughput: 1.9073650038576458e-06, WriteThroughput: 1.9073650038576458e-06, ReadLatency: 9.765625e-10, WriteLatency: 9.765625e-10, AvgReadBlockSize: 0, AvgWriteBlockSize: 0, SectorSize: 4096, ActualUsed: 4096, LogicalSize: 1.52687890625e-05},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			Output := processStats(tt.stats1, tt.stats2)
			if !reflect.DeepEqual(Output, tt.Output) {
				t.Fatalf("TestName: %v | processStats() => Got: %+v | Want: %+v \n", name, Output, tt.Output)
			}
		})
	}
}

func TestConvertMappedResponse(t *testing.T) {
	tests := map[string]struct {
		rawResponse v1alpha1.VolumeMetricsList
		output      map[string]v1alpha1.MetricsFamily
	}{
		"MetricsList containing MetricsFamily": {
			rawResponse: v1alpha1.VolumeMetricsList{
				{Name: "openebs_reads", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 2048}}}},
				{Name: "openebs_writes", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 2048}}}},
				{Name: "openebs_read_time", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 204}}}},
				{Name: "openebs_write_time", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 204}}}},
			},
			output: map[string]v1alpha1.MetricsFamily{
				"openebs_reads":      v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2048}},
				"openebs_writes":     v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2048}},
				"openebs_read_time":  v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 204}},
				"openebs_write_time": v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 204}},
			},
		},
		"When metricsFamily is not present in some metric item": {
			rawResponse: v1alpha1.VolumeMetricsList{
				{Name: "openebs_reads", Metric: []v1alpha1.MetricsFamily{}},
				{Name: "openebs_writes", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 2048}}}},
				{Name: "openebs_read_time", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 204}}}},
				{Name: "openebs_write_time", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 204}}}},
			},
			output: map[string]v1alpha1.MetricsFamily{
				"openebs_reads":      v1alpha1.MetricsFamily{},
				"openebs_writes":     v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2048}},
				"openebs_read_time":  v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 204}},
				"openebs_write_time": v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 204}},
			},
		},
		"When more than one metricFamily is present in some metric item": {
			rawResponse: v1alpha1.VolumeMetricsList{
				{Name: "openebs_reads", Metric: []v1alpha1.MetricsFamily{}},
				{Name: "openebs_writes", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 2048}}, {Gauge: v1alpha1.Gauge{Value: 2048}}}},
				{Name: "openebs_read_time", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 204}}}},
				{Name: "openebs_write_time", Metric: []v1alpha1.MetricsFamily{{Gauge: v1alpha1.Gauge{Value: 204}}}},
			},
			output: map[string]v1alpha1.MetricsFamily{
				"openebs_reads":      v1alpha1.MetricsFamily{},
				"openebs_writes":     v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 2048}},
				"openebs_read_time":  v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 204}},
				"openebs_write_time": v1alpha1.MetricsFamily{Gauge: v1alpha1.Gauge{Value: 204}},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if output := convertMappedResponse(tt.rawResponse); !reflect.DeepEqual(tt.output, output) {
				t.Fatalf("Test Name: %q | convertMappedResponse() => Got: %+v | Want: %+v \n", name, output, tt.output)
			}
		})
	}
}
