package collector

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestCollector tests collector.go
func TestCollector(t *testing.T) {

	for _, tt := range []struct {
		input          string
		match, unmatch []*regexp.Regexp
	}{
		// this is the input we are passing for positive testing
		// match will expect similar output from response.
		{
			input: `
{
	"stats": {
		"actual_used": 0,
		"logical_size": 0,
		"sector_size": 4096,
		"read_iops": 0,
		"read_time_per_second": 0,
		"read_block_count_per_second": ,
		"write_iops": 0,
		"write_time_per_second": 0,
		"write_block_count_per_second": 0,
		"read_latency": 0,
		"write_latency": 0,
		"avg_read_block_count_per_second": 0,
		"avg_write_block_count_per_second": ,
		"size_of_volume": 0,
	}
}`,
			// match matches the response with the expected input.
			match: []*regexp.Regexp{
				regexp.MustCompile(`OpenEBS_actual_used 4`),
				regexp.MustCompile(`OpenEBS_logical_size 4`),
				regexp.MustCompile(`OpenEBS_sector_size 4096`),
				regexp.MustCompile(`OpenEBS_read_iops 0`),
				regexp.MustCompile(`OpenEBS_read_time_per_second 0`),
				regexp.MustCompile(`OpenEBS_read_block_count_per_second 0`),
				regexp.MustCompile(`OpenEBS_write_iops 0`),
				regexp.MustCompile(`OpenEBS_write_time_per_second 0`),
				regexp.MustCompile(`OpenEBS_write_block_count_per_second 0`),
				regexp.MustCompile(`OpenEBS_read_latency 0`),
				regexp.MustCompile(`OpenEBS_write_latency 0`),
				regexp.MustCompile(`OpenEBS_avg_read_block_count_per_second 0`),
				regexp.MustCompile(`OpenEBS_avg_write_block_count_per_second 0`),
				regexp.MustCompile(`OpenEBS_size_of_volume 0`),
			},
			// unmatch is used for negative test, but this use case is for
			// positive test, so passing default value.
			unmatch: []*regexp.Regexp{},
		},
		{
			// this is the input we are passing for negative test.
			// unmatch will match the response with this input.
			input: `
{
	"stats": {
		"actual_used": 0,
		"logical_size": 0,
		"sector_size": 4096,
		"read_iops": 0,
		"read_time_per_second": 0,
		"read_block_count_per_second": ,
		"write_iops": 0,
		"write_time_per_second": 0,
		"write_block_count_per_second": 0,
		"read_latency": 0,
		"write_latency": 0,
		"avg_read_block_count_per_second": 0,
		"avg_write_block_count_per_second": ,
		"size_of_volume": 0,
	}
}`,
			match: []*regexp.Regexp{},
			unmatch: []*regexp.Regexp{
				// every field is empty for negative testing
				regexp.MustCompile(`OpenEBS_actual_used`),
				regexp.MustCompile(`OpenEBS_logical_size`),
				regexp.MustCompile(`OpenEBS_sector_size`),
				regexp.MustCompile(`OpenEBS_read_iops`),
				regexp.MustCompile(`OpenEBS_read_time_per_second`),
				regexp.MustCompile(`OpenEBS_read_block_count_per_second`),
				regexp.MustCompile(`OpenEBS_write_iops`),
				regexp.MustCompile(`OpenEBS_write_time_per_second`),
				regexp.MustCompile(`OpenEBS_write_block_count_per_second`),
				regexp.MustCompile(`OpenEBS_read_latency`),
				regexp.MustCompile(`OpenEBS_write_latency`),
				regexp.MustCompile(`OpenEBS_avg_read_block_count_per_second`),
				regexp.MustCompile(`OpenEBS_avg_write_block_count_per_second`),
				regexp.MustCompile(`OpenEBS_size_of_volume`),
			},
		},
	} {
		func() {
			// response is the response expected from the test server.
			var response = `{"Name":"vol","ReadIOPS":"1","ReplicaCounter":6,"RevisionCounter":100,"SCSIIOCount":null,"SectorSize":"4096","Size":"5G","TotalReadBlockCount":"10","TotalReadTime":"10","TotalWriteTime":"15","TotatWriteBlockCount":"10","UpTime":10,"UsedBlocks":"1048576","UsedLogicalBlocks":"1048576","WriteIOPS":"15","actions":{},"links":{"self":"http://localhost:9501/v1/stats"},"type":"stats"}`
			// This is dummy server which gives response in json format and it
			// is used to map the response with the fields of struct VolumeMetrics.
			controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, response)
			}))

			defer controller.Close()
			control, err := url.Parse(controller.URL)
			// col is an instance of the Volume exporter which gets
			// /v1/stats api along with url.
			col := NewExporter(control)
			if err := prometheus.Register(col); err != nil {
				t.Fatalf("collector failed to register: %s", err)
			}
			defer prometheus.Unregister(col)

			server := httptest.NewServer(promhttp.Handler())
			defer server.Close()

			client := http.DefaultClient
			client.Timeout = 5 * time.Second
			resp, err := client.Get(server.URL)
			if err != nil {
				t.Fatalf("unexpected failed response from prometheus: %s", err)
			}
			defer resp.Body.Close()

			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed reading server response: %s", err)
			}

			for _, re := range tt.match {
				if !re.Match(buf) {
					t.Errorf("failed matching: %q", re)
				}
			}

			for _, re := range tt.unmatch {
				if !re.Match(buf) {
					t.Errorf("failed unmatching: %q", re)
				}
			}

		}()
	}
}
