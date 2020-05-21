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

package collector

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sync"
	"syscall"
	"testing"
	"time"

	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	SplittedResponse             = "{ \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\", \"WriteIOPS\": \"0\", \"ReadIOPS\": \"0\", \"TotalWriteBytes\": \"0\", \"TotalReadBytes\": \"0\", \"Size\": \"10737418240\", \"UsedLogicalBlocks\":\"19\", \"SectorSize\":\"512\", \"UpTime\":\"20\", \"TotalReadBlockCount\":\"12\", \"TotalWriteBlockCount\":\"15\", \"TotalReadTime\":\"13\", \"TotalWriteTime\":\"132\", \"RevisionCounter\":\"1000\", \"ReplicaCounter\":\"3\", \"Replicas\":[{\"Address\":\"tcp://172.18.0.3:9502\",\"Mode\":\"Degraded\"},{\"Address\":\"tcp://172.18.0.4:9502\",\"Mode\":\"Healthy\"},{\"Address\":\"tcp://172.18.0.5:9502\",\"Mode\":\"Healthy\"}] }"
	NilCstorResponse             = "OK IOSTATS\r\n"
	CstorResponse                = "IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\", \"WriteIOPS\": \"0\", \"ReadIOPS\": \"0\", \"TotalWriteBytes\": \"0\", \"TotalReadBytes\": \"0\", \"Size\": \"10737418240\", \"UsedLogicalBlocks\":\"19\", \"SectorSize\":\"512\", \"UpTime\":\"20\", \"TotalReadBlockCount\":\"12\", \"TotalWriteBlockCount\":\"15\", \"TotalReadTime\":\"13\", \"TotalWriteTime\":\"132\", \"RevisionCounter\":\"1000\", \"ReplicaCounter\":\"3\", \"Replicas\":[{\"Address\":\"tcp://172.18.0.3:9502\",\"Mode\":\"Degraded\"},{\"Address\":\"tcp://172.18.0.4:9502\",\"Mode\":\"Healthy\"},{\"Address\":\"tcp://172.18.0.5:9502\",\"Mode\":\"Healthy\"}] }\r\nOK IOSTATS\r\n"
	JSONFormatedResponse         = "{ \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\", \"WriteIOPS\": \"0\", \"ReadIOPS\": \"0\", \"TotalWriteBytes\": \"0\", \"TotalReadBytes\": \"0\", \"Size\": \"10737418240\", \"UsedLogicalBlocks\":\"19\", \"SectorSize\":\"512\", \"UpTime\":\"20\", \"TotalReadBlockCount\":\"12\", \"TotalWriteBlockCount\":\"15\", \"TotalReadTime\":\"13\", \"TotalWriteTime\":\"132\", \"RevisionCounter\":\"1000\", \"ReplicaCounter\":\"3\", \"Replicas\":[{\"Address\":\"tcp://172.18.0.3:9502\",\"Mode\":\"Degraded\"},{\"Address\":\"tcp://172.18.0.4:9502\",\"Mode\":\"Healthy\"},{\"Address\":\"tcp://172.18.0.5:9502\",\"Mode\":\"Healthy\"}] }"
	ImproperJSONFormatedResponse = `IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\", \"WriteIOPS\": \"0\", \"ReadIOPS\": \"0\", \"TotalWriteBytes\": \"0\", \"TotalReadBytes\": \"0\", \"Size\": \"10737418240\", \"UsedLogicalBlocks\":\"19\", \"SectorSize\":\"512\", \"UpTime\":\"20\", \"TotalReadBlockCount\":\"12\", \"TotalWriteBlockCount\":\"15\", \"TotalReadTime\":\"13\", \"TotalWriteTime\":\"132\", \"RevisionCounter\":\"1000\", \"ReplicaCounter\":\"3\", \"Replicas\":[{\"Address\":\"tcp://172.18.0.3:9502\",\"Mode\":\"Degraded\"},{\"Address\":\"tcp://172.18.0.4:9502\",\"Mode\":\"Healthy\"},{\"Address\":\"tcp://172.18.0.5:9502\",\"Mode\":\"Healthy\"}] }\r\nOK IOSTATS\r\n`
)

func fakeCstor(path string) *cstor {
	return &cstor{
		socketPath: path,
	}
}

func runFakeUnixServer(t *testing.T, wg *sync.WaitGroup, response string) {
	go func() {
		// start the server
		t.Log("fake unix server started")
		listener, err := net.Listen("unix", "/tmp/go.sock")
		if err != nil {
			t.Fatal(err)
		}
		wg.Done()
		for {
			fd, err := listener.Accept()
			if err != nil {
				t.Fatal("Accept error: ", err)
			}
			fd.Write([]byte(HeaderPrefix + EOF))
			go sendFakeResponse(t, fd, response)
		}
	}()
}

func sendFakeResponse(t *testing.T, c net.Conn, resp string) {
	for {
		buf := make([]byte, 1024)
		_, err := c.Read(buf)
		if err != nil {
			return
		}
		data := resp
		_, err = c.Write([]byte(data))
		if err != nil {
			panic("Write: " + err.Error())
		}
		t.Log("Data written:", string(data))
	}
}

// unlink the socket and close the connection
func Unlink(t *testing.T) {
	err := syscall.Unlink("/tmp/go.sock")
	if err != nil {
		t.Log("Unlink()", err)
	}
}

func TestCstorCollector(t *testing.T) {
	// Unlink the existing socket connection /tmp/go.sock if exists
	// else ignore.
	Unlink(t)
	cases := map[string]struct {
		expectedResponse     string
		isFakeServerRequired bool
		match, unmatch       []*regexp.Regexp
	}{
		"[Success] istgt is reachable and giving expected stats": {
			expectedResponse:     CstorResponse,
			isFakeServerRequired: true,
			// match matches the response with the expected input.
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_reads 0`),
				regexp.MustCompile(`openebs_total_read_bytes 0`),
				regexp.MustCompile(`openebs_writes 0`),
				regexp.MustCompile(`openebs_total_write_bytes 0`),
				regexp.MustCompile(`openebs_size_of_volume 10`),
				regexp.MustCompile(`openebs_read_block_count 12`),
				regexp.MustCompile(`openebs_write_block_count 15`),
				regexp.MustCompile(`openebs_read_time 13`),
				regexp.MustCompile(`openebs_write_time 132`),
				regexp.MustCompile(`openebs_target_reject_request_counter 0`),
			},
			// unmatch is used for negative test, but this use case is for
			// positive test, so passing default value.
			unmatch: []*regexp.Regexp{},
		},
		"[Failure] istgt is not reachable and expected stats is null": {
			// match matches the response with the expected input.
			isFakeServerRequired: false,
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_reads 0`),
				regexp.MustCompile(`openebs_total_read_bytes 0`),
				regexp.MustCompile(`openebs_writes 0`),
				regexp.MustCompile(`openebs_total_write_bytes 0`),
				regexp.MustCompile(`openebs_size_of_volume 0`),
				regexp.MustCompile(`openebs_read_block_count 0`),
				regexp.MustCompile(`openebs_write_block_count 0`),
				regexp.MustCompile(`openebs_read_time 0`),
				regexp.MustCompile(`openebs_write_time 0`),
				regexp.MustCompile(`openebs_target_reject_request_counter 0`),
			},
			// unmatch is used for negative test, but this use case is for
			// positive test, so passing default value.
			unmatch: []*regexp.Regexp{},
		},
		"[Failure] istgt is reachable and giving valid stats but compare with incorrect output": {
			expectedResponse:     CstorResponse,
			isFakeServerRequired: true,
			match:                []*regexp.Regexp{},
			unmatch: []*regexp.Regexp{
				// every field is empty for negative testing
				regexp.MustCompile(`openebs_reads `),
				regexp.MustCompile(`openebs_total_read_bytes `),
				regexp.MustCompile(`openebs_writes `),
				regexp.MustCompile(`openebs_total_write_bytes `),
				regexp.MustCompile(`openebs_size_of_volume `),
				regexp.MustCompile(`openebs_read_block_count `),
				regexp.MustCompile(`openebs_write_block_count `),
				regexp.MustCompile(`openebs_read_time `),
				regexp.MustCompile(`openebs_write_time `),
				regexp.MustCompile(`openebs_target_reject_request_counter `),
			},
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			if tt.isFakeServerRequired {
				var wg sync.WaitGroup
				wg.Add(1)
				runFakeUnixServer(t, &wg, tt.expectedResponse)
				wg.Wait()
			}
			// col is an instance of the Volume exporter which gets
			// /v1/stats api along with url.
			c := Cstor("/tmp/go.sock")
			col := New(c)
			if err := prometheus.Register(col); err != nil {
				t.Fatalf("collector failed to register: %s", err)
			}

			server := httptest.NewServer(promhttp.Handler())
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
			// unlink the socketpath
			Unlink(t)
			prometheus.Unregister(col)
			server.Close()
		})
	}
}

func TestUnmarshal(t *testing.T) {
	cases := map[string]struct {
		response string
		output   v1.VolumeStats
	}{
		"[Success]Unmarshal Response into Metrics struct": {
			response: JSONFormatedResponse,
			output: v1.VolumeStats{
				Size:                 "10737418240",
				Iqn:                  "iqn.2017-08.OpenEBS.cstor:vol1",
				Writes:               "0",
				Reads:                "0",
				TotalReadBytes:       "0",
				TotalWriteBytes:      "0",
				UsedLogicalBlocks:    "19",
				SectorSize:           "512",
				TotalReadBlockCount:  "12",
				TotalWriteBlockCount: "15",
				TotalReadTime:        "13",
				TotalWriteTime:       "132",
				UpTime:               "20",
				ReplicaCounter:       "3",
				RevisionCounter:      "1000",
				Replicas: []v1.Replica{
					{
						Address: "tcp://172.18.0.3:9502",
						Mode:    "Degraded",
					},
					{
						Address: "tcp://172.18.0.4:9502",
						Mode:    "Healthy",
					},
					{
						Address: "tcp://172.18.0.5:9502",
						Mode:    "Healthy",
					},
				},
			},
		},
		"[Failure]Unmarshal Response returns empty Metrics": {
			response: ImproperJSONFormatedResponse,
			output:   v1.VolumeStats{},
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			c := Cstor("/tmp/go.sock")
			got, _ := c.unmarshal(tt.response)
			if !reflect.DeepEqual(got, tt.output) {
				t.Fatalf("unmarshal(%v) : expected %v, got %v", tt.response, tt.output, got)
			}
		})
	}
}

func TestInitiateConnection(t *testing.T) {
	c := fakeCstor("/tmp")
	dialFunc = func(path string) (net.Conn, error) {
		return nil, fmt.Errorf("No connection available")
	}

	err := c.initiateConnection()
	if err == nil {
		t.Fatalf("initiateConnection(%s): expected: error, got: nil", "/tmp")
	}
	t.Logf("%v", err)
}

func TestSplitter(t *testing.T) {
	cases := map[string]struct {
		response         string
		splittedResponse string
	}{
		"[Success] If response is as expected": {
			response:         CstorResponse,
			splittedResponse: SplittedResponse,
		},
		"[Failure] If response is not as expected, splitter should return nil string": {
			response:         NilCstorResponse,
			splittedResponse: "",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			c := Cstor("/tmp/go.sock")
			if got := c.splitter(tt.response); !reflect.DeepEqual(got, tt.splittedResponse) {
				t.Fatalf("splitter(%v) => expected %v, got %v", tt.response, tt.splittedResponse, got)
			}
		})
	}

}
