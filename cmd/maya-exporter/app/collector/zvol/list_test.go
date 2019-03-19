package zvol

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestGetZfsList(t *testing.T) {
	cases := map[string]struct {
		run   testRunner
		match []*regexp.Regexp
	}{
		"Test0": {
			run: testRunner{
				stdout: []byte(`cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 238592`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`),
			},
		},
		"Test1": {
			run: testRunner{
				stdout: []byte(`cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	3055	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 238592`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3055`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`),
			},
		},
		"Test2": {
			run: testRunner{
				stdout: []byte(`cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 238592`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`),
			},
		},
		"Test3": {
			run: testRunner{
				stdout: []byte(`cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 3000`),
				regexp.MustCompile(`openebs_available_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 3000`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5"} 238592`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c3a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c1a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5"} 6144`),
				regexp.MustCompile(`openebs_used_size{name="cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone"} 0`),
			},
		},
		"Test4": {
			run: testRunner{
				stdout: []byte(`cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	liaub	kzjsfvn	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
						cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
						cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_zfs_list_parse_error 2`),
			},
		},
		"Test5": {
			run: testRunner{
				isError: true,
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_zfs_list_command_error 1`),
			},
		},
		"Test6": {
			run: testRunner{
				stdout: []byte(``),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_zfs_list_parse_error 1`),
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			runner = tt.run
			vol := NewVolumeList()
			if err := prometheus.Register(vol); err != nil {
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
					fmt.Println(string(buf))
					t.Errorf("failed matching: %q", re)
				}
			}

			prometheus.Unregister(vol)
			server.Close()
		})
	}
}

//func TestRejectRequestCounter(t *testing.T) {
//	cases := map[string]struct {
//		run      testRunner
//		reqCount int
//		col      prometheus.Collector
//		output   *regexp.Regexp
//	}{
//		"Test0": {
//			run: testRunner{
//				stdout: []byte(`cstor-f1ea249b-417d-11e9-9c76-42010a8001a5	238592	3000	512	/cstor-f1ea249b-417d-11e9-9c76-42010a8001a5
//				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5	6144	3000	6144	-
//				cstor-f1ea249b-417d-11e9-9c76-42010a8001a5/pvc-c4a68fa3-4183-11e9-9c76-42010a8001a5_rebuild_clone	0	3000	6144	-`),
//			},
//			reqCount: 200,
//			col:      NewVolumeList(),
//			output:   regexp.MustCompile(`openebs_zfs_list_request_reject_count\s\d+`),
//		},
//		"Test1": {
//			run: testRunner{
//				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Rebuilding","rebuildStatus": "SNAP REBUILD INPROGRESS","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
//			},
//			reqCount: 200,
//			col:      New(),
//			output:   regexp.MustCompile(`openebs_zfs_stats_reject_request_count\s\d+`),
//		},
//	}
//	for name, tt := range cases {
//		t.Run(name, func(t *testing.T) {
//			runner = tt.run
//			if err := prometheus.Register(tt.col); err != nil {
//				t.Fatalf("collector failed to register: %s", err)
//			}
//
//			server := httptest.NewServer(promhttp.Handler())
//			var body io.ReadCloser
//
//			wg := sync.WaitGroup{}
//			wg.Add(tt.reqCount)
//			for i := 0; i < tt.reqCount; i++ {
//				go func(server *httptest.Server) {
//					defer wg.Done()
//					client := http.DefaultClient
//					client.Timeout = 5 * time.Second
//					resp, err := client.Get(server.URL)
//					body = resp.Body
//					if err != nil {
//						t.Fatalf("unexpected failed response from prometheus: %s", err)
//					}
//				}(server)
//			}
//
//			wg.Wait()
//			defer body.Close()
//
//			buf, err := ioutil.ReadAll(body)
//			if err != nil {
//				t.Fatalf("failed reading server response: %s", err)
//			}
//
//			str := tt.output.FindStringSubmatch(string(buf))
//			newStr := strings.Fields(str[0])
//			reqCount, _ := strconv.Atoi(newStr[1])
//			if reqCount <= 0 {
//				t.Fatalf("Failed to test reject request count, fount str: %s, buf: %s", str, string(buf))
//			}
//			prometheus.Unregister(tt.col)
//			server.Close()
//		})
//	}
//}
