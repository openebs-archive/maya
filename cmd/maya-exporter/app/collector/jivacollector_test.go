package collector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/openebs/maya/types/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	utiltesting "k8s.io/client-go/util/testing"
)

var (
	controllerResponse    = `{"Name":"vol1","ReadIOPS":"0","ReplicaCounter":0,"RevisionCounter":0,"SCSIIOCount":{},"SectorSize":"4096","Size":"1073741824","TotalReadBlockCount":"0","TotalReadTime":"0","TotalWriteTime":"0","TotatWriteBlockCount":"0","UpTime":158.667823193,"UsedBlocks":"5","UsedLogicalBlocks":"0","WriteIOPS":"0","actions":{},"links":{"self":"http://10.42.0.1:9501/v1/stats"},"type":"stats"}`
	validControllerResp   = `{"Name":"vol1","ReadIOPS":"5","ReplicaCounter":2,"RevisionCounter":10,"SCSIIOCount":{},"SectorSize":"4096","Size":"1073741824","TotalReadBlockCount":"25","TotalReadTime":"45","TotalWriteTime":"30","TotatWriteBlockCount":"6","UpTime":158.667823193,"UsedBlocks":"5","UsedLogicalBlocks":"23","WriteIOPS":"11","actions":{},"links":{"self":"http://10.42.0.1:9501/v1/stats"},"type":"stats"}`
	invalidControllerResp = `404 Page not found`
)

// TestCollector tests collector.go
func TestJivaCollector(t *testing.T) {

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
		"actual_used": 4,
		"logical_size": 4,
		"sector_size": 4096,
		"reads": 1,
		"read_time": 10,
		"read_block_count": 10,
		"writes": 15,
		"write_time": 15,
		"write_block_count": 10,
		"size_of_volume": 1073741824,
	}
}`,
			// match matches the response with the expected input.
			match: []*regexp.Regexp{
				regexp.MustCompile(`OpenEBS_jiva_actual_used 4`),
				regexp.MustCompile(`OpenEBS_jiva_logical_size 4`),
				regexp.MustCompile(`OpenEBS_jiva_sector_size 4096`),
				regexp.MustCompile(`OpenEBS_jiva_reads 1`),
				regexp.MustCompile(`OpenEBS_jiva_read_time 10`),
				regexp.MustCompile(`OpenEBS_jiva_read_block_count 10`),
				regexp.MustCompile(`OpenEBS_jiva_writes 15`),
				regexp.MustCompile(`OpenEBS_jiva_write_time 15`),
				regexp.MustCompile(`OpenEBS_jiva_write_block_count 10`),
				regexp.MustCompile(`OpenEBS_jiva_size_of_volume 1`),
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
						"actual_used": 4,
						"logical_size": 4,
						"sector_size": 4096,
						"reads": 1,
						"total_read_time": 10,
						"total_read_block_count": 10,
						"writes": 15,
						"total_write_time": 15,
						"total_write_block_count": 10,
						"size_of_volume": 1073741824,
					}
				}`,
			match: []*regexp.Regexp{},
			unmatch: []*regexp.Regexp{
				// every field is empty for negative testing
				regexp.MustCompile(`OpenEBS_jiva_actual_used`),
				regexp.MustCompile(`OpenEBS_jiva_logical_size`),
				regexp.MustCompile(`OpenEBS_jiva_sector_size`),
				regexp.MustCompile(`OpenEBS_jiva_reads`),
				regexp.MustCompile(`OpenEBS_jiva_read_time`),
				regexp.MustCompile(`OpenEBS_jiva_read_block_count`),
				regexp.MustCompile(`OpenEBS_jiva_writes`),
				regexp.MustCompile(`OpenEBS_jiva_write_time`),
				regexp.MustCompile(`OpenEBS_jiva_write_block_count`),
				regexp.MustCompile(`OpenEBS_jiva_size_of_volume`),
			},
		},
	} {
		func() {
			// response is the response expected from the test server.
			var response = `{"Name":"vol","ReadIOPS":"1","ReplicaCounter":6,"RevisionCounter":100,"SCSIIOCount":null,"SectorSize":"4096","Size":"1073741824","TotalReadBlockCount":"10","TotalReadTime":"10","TotalWriteTime":"15","TotatWriteBlockCount":"10","UpTime":10,"UsedBlocks":"1048576","UsedLogicalBlocks":"1048576","WriteIOPS":"15","actions":{},"links":{"self":"http://localhost:9501/v1/stats"},"type":"stats"}`
			// This is dummy server which gives response in json format and it
			// is used to map the response with the fields of struct VolumeMetrics.
			controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, response)
			}))

			defer controller.Close()
			control, err := url.Parse(controller.URL)
			if err != nil {
				t.Fatalf("Couldn't parse the controller URL, found error %v", err)
			}
			// col is an instance of the Volume exporter which gets
			// /v1/stats api along with url.
			col := NewJivaStatsExporter(control, "jiva")
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
			//fmt.Println(string(buf))
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

func TestJivaStatsCollector(t *testing.T) {
	cases := map[string]struct {
		exporter    *VolumeStatsExporter
		err         error
		fakehandler utiltesting.FakeHandler
		testServer  bool
	}{
		"[Success] If controller is Jiva and its running": {
			exporter: &VolumeStatsExporter{
				CASType: "jiva",
				Jiva: Jiva{
					VolumeControllerURL: "localhost:9500",
				},
				Metrics: *MetricsInitializer("jiva"),
			},
			testServer: true,
			fakehandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(controllerResponse),
				T:            t,
			},

			err: nil,
		},
		"[Failure] If controller is Jiva and it is not reachable": {
			exporter: &VolumeStatsExporter{
				CASType: "jiva",
				Jiva: Jiva{
					VolumeControllerURL: "localhost:9500",
				},
				Metrics: *MetricsInitializer("jiva"),
			},
			err: errors.New("error in collecting metrics"),
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			if tt.testServer {
				server := httptest.NewServer(&tt.fakehandler)
				tt.exporter.VolumeControllerURL = server.URL
			}
			got := tt.exporter.Jiva.collector(&tt.exporter.Metrics)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("collector() : expected %v, got %v", tt.err, got)
			}
		})
	}
}

func TestGetVolumeStats(t *testing.T) {

	cases := map[string]struct {
		jiva        Jiva
		obj         v1.VolumeStats
		fakeHandler utiltesting.FakeHandler
		err         error
	}{
		"Valid Response from jiva controller": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(validControllerResp),
				T:            t,
			},
			err: nil,
		},
		"Invalid Response from jiva controller": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(invalidControllerResp),
				T:            t,
			},
			err: errors.New("Error in unmarshalling the json response"),
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			defer server.Close()
			tt.jiva.VolumeControllerURL = server.URL
			got := tt.jiva.getVolumeStats(&tt.obj)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("getVolumeStats(%v) => got %v, want %v", server.URL, got, tt.err)
			}
		})
	}
}
