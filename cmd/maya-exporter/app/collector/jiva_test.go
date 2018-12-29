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

	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	utiltesting "k8s.io/client-go/util/testing"
)

var (
	fakeResponse          = `{"Name":"vol","ReadIOPS":"1","ReplicaCounter":3,"RevisionCounter":100,"SCSIIOCount":null,"SectorSize":"4096","Size":"1073741824","TotalReadBlockCount":"10","TotalReadTime":"10","TotalWriteTime":"15","TotalWriteBlockCount":"10","UpTime":10,"UsedBlocks":"1048576","UsedLogicalBlocks":"1048576","WriteIOPS":"15","actions":{},"links":{"self":"http://localhost:9501/v1/stats"},"type":"stats", "Replicas":[{"Address":"tcp://172.18.0.3:9502","Mode":"RW"},{"Address":"tcp://172.18.0.4:9502","Mode":"RW"},{"Address":"tcp://172.18.0.5:9502","Mode":"RW"}] }`
	controllerResponse    = `{"Name":"vol1","ReadIOPS":"0","ReplicaCounter":0,"RevisionCounter":0,"SCSIIOCount":{},"SectorSize":"4096","Size":"1073741824","TotalReadBlockCount":"0","TotalReadTime":"0","TotalWriteTime":"0","TotalWriteBlockCount":"0","UpTime":158.667823193,"UsedBlocks":"5","UsedLogicalBlocks":"0","WriteIOPS":"0","actions":{},"links":{"self":"http://10.42.0.1:9501/v1/stats"},"type":"stats", "Replicas":[{"Address":"tcp://172.18.0.3:9502","Mode":"RW"},{"Address":"tcp://172.18.0.4:9502","Mode":"RW"},{"Address":"tcp://172.18.0.5:9502","Mode":"RW"}] }`
	validControllerResp   = `{"Name":"vol1","ReadIOPS":"5","ReplicaCounter":2,"RevisionCounter":10,"SCSIIOCount":{},"SectorSize":"4096","Size":"1073741824","TotalReadBlockCount":"25","TotalReadTime":"45","TotalWriteTime":"30","TotalWriteBlockCount":"6","UpTime":158.667823193,"UsedBlocks":"5","UsedLogicalBlocks":"23","WriteIOPS":"11","actions":{},"links":{"self":"http://10.42.0.1:9501/v1/stats"},"type":"stats", "Replicas":[{"Address":"tcp://172.18.0.3:9502","Mode":"RW"},{"Address":"tcp://172.18.0.4:9502","Mode":"RW"},{"Address":"tcp://172.18.0.5:9502","Mode":"RW"}] }`
	invalidControllerResp = `404 Page not found`
)

// TestCollector tests collector.go
func TestJivaCollector(t *testing.T) {

	cases := map[string]struct {
		input          string
		match, unmatch []*regexp.Regexp
	}{
		// this is the input we are passing for positive testing
		// match will expect similar output from response.
		"[Success] controller is reachable and giving expected stats": {
			input: fakeResponse,
			// match matches the response with the expected input.
			match: []*regexp.Regexp{
				// these regex are the actual expected output from exporter
				// based on the fakeResponse
				regexp.MustCompile(`openebs_actual_used 4`),
				regexp.MustCompile(`openebs_logical_size 4`),
				regexp.MustCompile(`openebs_sector_size 4096`),
				regexp.MustCompile(`openebs_reads 1`),
				regexp.MustCompile(`openebs_read_time 10`),
				regexp.MustCompile(`openebs_read_block_count 10`),
				regexp.MustCompile(`openebs_writes 15`),
				regexp.MustCompile(`openebs_write_time 15`),
				regexp.MustCompile(`openebs_write_block_count 10`),
				regexp.MustCompile(`openebs_size_of_volume 1`),
			},
			// unmatch is used for negative test, but this use case is for
			// positive test, so passing default value.
			unmatch: []*regexp.Regexp{},
		},
		"[Failure] controller is not reachable and expected stats is null": {
			input: invalidControllerResp,
			// match matches the response with the expected input.
			match: []*regexp.Regexp{
				// these regex are the actual expected output from exporter
				// based on the fakeResponse
				regexp.MustCompile(`openebs_actual_used 0`),
				regexp.MustCompile(`openebs_logical_size 0`),
				regexp.MustCompile(`openebs_sector_size 0`),
				regexp.MustCompile(`openebs_reads 0`),
				regexp.MustCompile(`openebs_read_time 0`),
				regexp.MustCompile(`openebs_read_block_count 0`),
				regexp.MustCompile(`openebs_writes 0`),
				regexp.MustCompile(`openebs_write_time 0`),
				regexp.MustCompile(`openebs_write_block_count 0`),
				regexp.MustCompile(`openebs_size_of_volume 0`),
			},
			// unmatch is used for negative test, but this use case is for
			// positive test, so passing default value.
			unmatch: []*regexp.Regexp{},
		},
		"[Failure] controller is reachable and giving valid stats but compare with incorrect output": {
			// this is the input we are passing for negative test.
			// unmatch will match the response with this input.
			input: fakeResponse,
			match: []*regexp.Regexp{},
			unmatch: []*regexp.Regexp{
				// every field is empty for negative testing
				regexp.MustCompile(`openebs_actual_used`),
				regexp.MustCompile(`openebs_logical_size`),
				regexp.MustCompile(`openebs_sector_size`),
				regexp.MustCompile(`openebs_reads`),
				regexp.MustCompile(`openebs_read_time`),
				regexp.MustCompile(`openebs_read_block_count`),
				regexp.MustCompile(`openebs_writes`),
				regexp.MustCompile(`openebs_write_time`),
				regexp.MustCompile(`openebs_write_block_count`),
				regexp.MustCompile(`openebs_size_of_volume`),
			},
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			// This is dummy server which gives response in json format and it
			// is used to map the response with the fields of struct VolumeMetrics.
			controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, tt.input)
			}))

			defer controller.Close()
			control, err := url.Parse(controller.URL)
			if err != nil {
				t.Fatalf("Couldn't parse the controller URL, found error %v", err)
			}
			jiva := Jiva(control)
			// exporter is an instance of the Volume exporter which gets
			// /v1/stats api along with url.
			col := New(jiva)
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
					fmt.Println(string(buf))
					t.Errorf("failed matching: %q", re)
				}
			}

			for _, re := range tt.unmatch {
				if !re.Match(buf) {
					t.Errorf("failed unmatching: %q", re)
				}
			}
		})
	}
}

func TestGetVolumeStats(t *testing.T) {

	cases := map[string]struct {
		jiva        jiva
		obj         v1.VolumeStats
		fakeHandler utiltesting.FakeHandler
		isErr       bool
	}{
		"Valid Response from jiva controller": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(validControllerResp),
				T:            t,
			},
			isErr: false,
		},
		"Invalid Response from jiva controller": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(invalidControllerResp),
				T:            t,
			},
			isErr: true,
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			defer server.Close()
			tt.jiva.url = server.URL
			err := tt.jiva.getVolumeStats(&tt.obj)
			if err != nil && !tt.isErr {
				t.Fatalf("getVolumeStats(%v) => got %v, want isErr to be %v", server.URL, err, tt.isErr)
			}
		})
	}
}
