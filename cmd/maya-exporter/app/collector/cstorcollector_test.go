package collector

import (
	"errors"
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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	CstorResponse                = "IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\", \"writes\": \"0\", \"reads\": \"0\", \"totalwritebytes\": \"0\", \"totalreadbytes\": \"0\", \"size\": \"10737418240\" }\r\nOK IOSTATS\r\n"
	JSONFormatedResponse         = "{ \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\", \"writes\": \"0\", \"reads\": \"0\", \"totalwritebytes\": \"0\", \"totalreadbytes\": \"0\", \"size\": \"10737418240\" }"
	ImproperJSONFormatedResponse = `IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\", \"writes\": \"0\", \"reads\": \"0\", \"totalwritebytes\": \"0\", \"totalreadbytes\": \"0\", \"size\": \"10737418240\" }\r\nOK IOSTATS\r\n`
)

func TestUnmarshaller(t *testing.T) {
	cases := map[string]struct {
		response string
		output   Response
	}{
		"[Success]Unmarshal Response into Metrics struct": {
			response: JSONFormatedResponse,
			output: Response{
				Size:            "10737418240",
				Iqn:             "iqn.2017-08.OpenEBS.cstor:vol1",
				Writes:          "0",
				Reads:           "0",
				TotalReadBytes:  "0",
				TotalWriteBytes: "0",
			},
		},
		"[Failure]Unmarshal Response returns empty Metrics": {
			response: ImproperJSONFormatedResponse,
			output:   Response{},
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := unmarshaller(tt.response)
			if !reflect.DeepEqual(got, tt.output) {
				t.Fatalf("unmarshal(%v) : expected %v, got %v", tt.response, tt.output, got)
			}
		})
	}
}

func runFakeUnixServer(t *testing.T, wg *sync.WaitGroup) {
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
			go sendFakeResponse(t, fd)
		}
	}()
}

func sendFakeResponse(t *testing.T, c net.Conn) {
	for {
		buf := make([]byte, 512)
		_, err := c.Read(buf)
		if err != nil {
			return
		}

		data := CstorResponse
		_, err = c.Write([]byte(data))
		if err != nil {
			panic("Write: " + err.Error())
		}
		t.Log("Data written:", string(data))
	}
}

func TestReadHeader(t *testing.T) {
	cases := map[string]struct {
		exporter       *CstorStatsExporter
		err            error
		fakeUnixServer bool
		header         string
	}{
		// Only success case of this can be performed, because you
		// can't control the server's status at the run time in unit test.
		// To test the failure case, we need to down the server at
		// the run time and determining that situation (time) is not
		// possible in unit test.
		"[Success] Read header if server is available": {
			exporter: &CstorStatsExporter{
				// conn value will be set at the run time after
				// making the connection.
				Conn: nil,
			},
			err:            nil,
			fakeUnixServer: true,
			header:         "iSCSI Target Controller version istgt:0.5.20121028:15:21:49:Jun  8 2018 on  from \r\n",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			if tt.fakeUnixServer {
				var wg sync.WaitGroup
				wg.Add(1)
				runFakeUnixServer(t, &wg)
				wg.Wait()
				conn, err := net.Dial("unix", "/tmp/go.sock")
				if err != nil {
					t.Fatal("err in dial :", err)
				}
				// overwrite the value of conn from nil to
				// expected value.
				tt.exporter.Conn = conn
				CstorResponse = tt.header
				tt.exporter.writer()
			}
			got := ReadHeader(tt.exporter.Conn)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("ReadHeader(%v) : expected %v, got %v", tt.exporter.Conn, tt.err, got)
			}
		})
		// unlink the socketpath
		Unlink(t)
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
	for _, tt := range []struct {
		input          string
		response       string
		match, unmatch []*regexp.Regexp
	}{
		// this is the input we are passing for positive testing
		// match will expect similar output from response.
		{
			input: `
{
	"stats": {
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
		"size_of_volume": 10,
	}
}`,
			// match matches the response with the expected input.
			match: []*regexp.Regexp{
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
				regexp.MustCompile(`OpenEBS_size_of_volume 10`),
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
			var wg sync.WaitGroup
			cstorResponse := "IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\", \"writes\": \"0\", \"reads\": \"0\", \"totalwritebytes\": \"0\", \"totalreadbytes\": \"0\", \"size\": \"10737418240\" }\r\nOK IOSTATS\r\n"
			CstorResponse = cstorResponse
			wg.Add(1)
			runFakeUnixServer(t, &wg)
			wg.Wait()
			conn, err := net.Dial("unix", "/tmp/go.sock")
			if err != nil {
				t.Fatal("err in dial :", err)
			}
			// col is an instance of the Volume exporter which gets
			// /v1/stats api along with url.
			col := NewCstorStatsExporter(conn)
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
			// unlink the socketpath
			Unlink(t)
		}()
	}
}
func TestCstorStatsCollector(t *testing.T) {
	// Unlink the existing socket connection /tmp/go.sock if exists
	// else ignore.
	Unlink(t)
	cases := map[string]struct {
		exporter       *CstorStatsExporter
		err            error
		fakeUnixServer bool
	}{
		"[Success] If controller is cstor and its running": {
			exporter: &CstorStatsExporter{
				// Value of Conn will be overwritten at run time
				Conn:    nil,
				Metrics: metrics,
			},
			fakeUnixServer: true,
			err:            nil,
		},
		"[failure] If controller is cstor and its not running": {
			exporter: &CstorStatsExporter{
				Conn:    nil,
				Metrics: metrics,
			},
			err: errors.New("error in initiating connection with socket"),
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			if tt.fakeUnixServer {
				var wg sync.WaitGroup
				wg.Add(1)
				runFakeUnixServer(t, &wg)
				wg.Wait()
				conn, err := net.Dial("unix", "/tmp/go.sock")
				if err != nil {
					t.Fatal("err in dial :", err)
				}
				// overwriting the value of Conn from nil to some valid
				// value at run time.
				tt.exporter.Conn = conn
			}
			got := tt.exporter.collector()
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("collector() : expected %v, got %v", tt.err, got)
			}
		})
		Unlink(t)
	}
}
