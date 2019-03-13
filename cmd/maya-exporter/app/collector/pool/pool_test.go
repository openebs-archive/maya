package pool

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var count int

type testRunner struct {
	stdout  []byte
	isError bool
}

func (r testRunner) RunCombinedOutput(cmd string, args ...string) ([]byte, error) {
	return nil, nil
}

func (r testRunner) RunStdoutPipe(cmd string, args ...string) ([]byte, error) {
	return nil, nil
}

func (r testRunner) RunCommandWithTimeoutContext(timeout time.Duration, cmd string, args ...string) ([]byte, error) {
	if r.isError {
		switch count {
		case 1:
			count++
			return []byte("no pools available"), nil
		case 2:
			return []byte("ONLINE"), nil
		case 3:
			count = 1
			return nil, errors.New("some dummy error")
		default:
			return nil, errors.New("some dummy error")
		}
	}
	return r.stdout, nil
}

func TestGetZpoolStats(t *testing.T) {
	cases := map[string]struct {
		run            testRunner
		match, unmatch []*regexp.Regexp
	}{
		"Test0": {
			run: testRunner{
				stdout: []byte("cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 ONLINE	-"),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_pool_size 1024`),
				regexp.MustCompile(`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_used_pool_capacity 24`),
				regexp.MustCompile(`openebs_free_pool_capacity 1000`),
				regexp.MustCompile(`openebs_used_pool_capacity_percent 0`),
			},
		},
		"Test1": {
			run: testRunner{
				stdout: []byte("cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 OFFLINE	-"),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_pool_size 1024`),
				regexp.MustCompile(`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 0`),
				regexp.MustCompile(`openebs_used_pool_capacity 24`),
				regexp.MustCompile(`openebs_free_pool_capacity 1000`),
				regexp.MustCompile(`openebs_used_pool_capacity_percent 0`),
			},
		},
		"Test2": {
			run: testRunner{
				stdout: []byte("cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 UNAVAIL	-"),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_pool_size 1024`),
				regexp.MustCompile(`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 5`),
				regexp.MustCompile(`openebs_used_pool_capacity 24`),
				regexp.MustCompile(`openebs_free_pool_capacity 1000`),
				regexp.MustCompile(`openebs_used_pool_capacity_percent 0`),
			},
		},
		"Test3": {
			run: testRunner{
				stdout: []byte("cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 FAULTED	-"),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_pool_size 1024`),
				regexp.MustCompile(`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_used_pool_capacity 24`),
				regexp.MustCompile(`openebs_free_pool_capacity 1000`),
				regexp.MustCompile(`openebs_used_pool_capacity_percent 0`),
			},
		},
		"Test4": {
			run: testRunner{
				stdout: []byte("cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	1024	24	1000	-	0	0	1.00 REMOVED	-"),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_pool_size 1024`),
				regexp.MustCompile(`openebs_pool_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a"} 4`),
				regexp.MustCompile(`openebs_used_pool_capacity 24`),
				regexp.MustCompile(`openebs_free_pool_capacity 1000`),
				regexp.MustCompile(`openebs_used_pool_capacity_percent 0`),
			},
		},
		"Test5": {
			run: testRunner{
				stdout: []byte("no pools available"),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_no_pool_available_error 1`),
			},
		},
		"Test6": {
			run: testRunner{
				stdout: []byte("cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a  1024    24  1000    -"),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_zpool_list_incomplete_stdout_error 1`),
			},
		},

		"Test7": {
			run: testRunner{
				isError: true,
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_zpool_command_error 1`),
			},
		},
		"Test8": {
			run: testRunner{
				stdout: []byte("cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	iaueb7	aiwub	aliubv	-	0	iauwb	1.00 REMOVED	-"),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_zpool_list_parse_error_count 4`),
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			runner = tt.run
			pool := New()
			if err := prometheus.Register(pool); err != nil {
				t.Fatalf("collector failed to register: %s", err)
			}
			defer prometheus.Unregister(pool)

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

func TestGetInitStatus(t *testing.T) {
	cases := map[string]struct {
		run   testRunner
		count int
	}{
		"Test0": {
			run: testRunner{
				stdout: []byte(`
				pool: mypool
				state: ONLINE
				 scan: none requested
			   config:

				   NAME                                  STATE     READ WRITE CKSUM
				   mypool                                ONLINE       0     0     0
					 raidz1-0                            ONLINE       0     0     0
					   /home/infinity/experiments/pool1  ONLINE       0     0     0
					   /home/infinity/experiments/pool2  ONLINE       0     0     0
					 raidz1-1                            ONLINE       0     0     0
					   /home/infinity/experiments/pool3  ONLINE       0     0     0
					   /home/infinity/experiments/pool4  ONLINE       0     0     0`),
			},
		},
		"Test1": {
			run: testRunner{
				isError: true,
			},
			count: 3,
		},
		"Test2": {
			run: testRunner{
				isError: true,
			},
			count: 1,
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			runner = tt.run
			pool := New()
			count = tt.count
			pool.GetInitStatus(1 * time.Second)
		})
	}
}

func TestRejectRequestCounter(t *testing.T) {
	reqCount := 100
	output := regexp.MustCompile(`openebs_zpool_reject_request_count\s\d+`)
	runner = testRunner{
		stdout: []byte("cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	3000	23423 1341	-	0	iauwb	1.00 REMOVED	-"),
	}
	col := New()
	if err := prometheus.Register(col); err != nil {
		t.Fatalf("collector failed to register: %s", err)
	}

	server := httptest.NewServer(promhttp.Handler())
	var body io.ReadCloser

	wg := sync.WaitGroup{}
	wg.Add(reqCount)
	for i := 0; i < reqCount; i++ {
		go func(server *httptest.Server) {
			defer wg.Done()
			client := http.DefaultClient
			client.Timeout = 5 * time.Second
			resp, err := client.Get(server.URL)
			body = resp.Body
			if err != nil {
				t.Fatalf("unexpected failed response from prometheus: %s", err)
			}
		}(server)
	}

	wg.Wait()
	defer body.Close()

	buf, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fatalf("failed reading server response: %s", err)
	}

	str := output.FindStringSubmatch(string(buf))
	newStr := strings.Fields(str[0])
	reqCount, _ = strconv.Atoi(newStr[1])
	if reqCount <= 0 {
		t.Fatalf("Failed to test reject request count, fount str: %s, buf: %s", str, string(buf))
	}
	prometheus.Unregister(col)
	server.Close()
}
