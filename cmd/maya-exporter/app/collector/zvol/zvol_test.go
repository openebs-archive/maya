package zvol

import (
	"errors"
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

var count int

type testRunner struct {
	stdout      []byte
	isError     bool
	injectDelay bool
}

func (r testRunner) RunCombinedOutput(cmd string, args ...string) ([]byte, error) {
	return nil, nil
}

func (r testRunner) RunStdoutPipe(cmd string, args ...string) ([]byte, error) {
	return nil, nil
}

func (r testRunner) RunCommandWithTimeoutContext(timeout time.Duration, cmd string, args ...string) ([]byte, error) {
	if r.isError {
		return nil, errors.New("some dummy error")
	}
	if r.injectDelay {
		time.Sleep(50 * time.Millisecond)
		return r.stdout, nil
	}
	return r.stdout, nil
}

func TestGetZfsStats(t *testing.T) {
	cases := map[string]struct {
		run   testRunner
		match []*regexp.Regexp
	}{
		"Test0": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 100`),
				regexp.MustCompile(`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 10`),
				regexp.MustCompile(`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 150`),
				regexp.MustCompile(`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 200`),
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`),
				regexp.MustCompile(`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 50`),
				regexp.MustCompile(`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 500`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
			},
		},
		"Test1": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Offline","rebuildStatus": "INIT","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`),
			},
		},
		"Test2": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Degraded","rebuildStatus": "INIT","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`),
			},
		},
		"Test3": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Rebuilding","rebuildStatus": "SNAP REBUILD INPROGRESS","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2`),
			},
		},
		"Test4": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Rebuilding","rebuildStatus": "ACTIVE DATASET REBUILD INPROGRESS","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
			},
		},
		"Test5": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Offline","rebuildStatus": "ERRORED  ","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 4`),
			},
		},

		"Test6": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Offline","rebuildStatus": "FAILED","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 5`),
			},
		},
		"Test7": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Offline","rebuildStatus": "UNKNOWN","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 0`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 6`),
			},
		},
		"Test8": {
			run: testRunner{
				isError: true,
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_zfs_command_error 1`),
			},
		},
		"Test9": {
			run: testRunner{
				stdout: []byte(``),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_zfs_stats_parse_error_counter 1`),
			},
		},
		"Test10": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50},{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 100`),
				regexp.MustCompile(`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 100`),
				regexp.MustCompile(`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 10`),
				regexp.MustCompile(`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 10`),
				regexp.MustCompile(`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 150`),
				regexp.MustCompile(`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 150`),
				regexp.MustCompile(`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 200`),
				regexp.MustCompile(`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 200`),
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`),
				regexp.MustCompile(`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`),
				regexp.MustCompile(`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 50`),
				regexp.MustCompile(`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 50`),
				regexp.MustCompile(`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 500`),
				regexp.MustCompile(`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 500`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
			},
		},
		"Test11": {
			run: testRunner{
				stdout: []byte(`{"stats": [{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50},{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50},{"name": "cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a/pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a","status": "Healthy","rebuildStatus": "DONE","runningIONum": 0,"rebuildBytes": 500,"rebuildCnt": 3,"rebuildDoneCnt": 2,"rebuildFailedCnt": 0,"readCount": 1000,"readLatency": 150,"readByte": 1024,"writeCount": 1000,"writeLatency": 200,"writeByte": 1024,"syncCount": 100,"syncLatency": 10,"inflightIOCnt": 2000,"dispatchedIOCnt": 50}]}`),
			},
			match: []*regexp.Regexp{
				regexp.MustCompile(`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_read_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_write_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1024`),
				regexp.MustCompile(`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_read_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_total_write_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1000`),
				regexp.MustCompile(`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 100`),
				regexp.MustCompile(`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 100`),
				regexp.MustCompile(`openebs_sync_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 100`),
				regexp.MustCompile(`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 10`),
				regexp.MustCompile(`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 10`),
				regexp.MustCompile(`openebs_sync_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 10`),
				regexp.MustCompile(`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 150`),
				regexp.MustCompile(`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 150`),
				regexp.MustCompile(`openebs_read_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 150`),
				regexp.MustCompile(`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 200`),
				regexp.MustCompile(`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 200`),
				regexp.MustCompile(`openebs_write_latency{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 200`),
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_replica_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`),
				regexp.MustCompile(`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`),
				regexp.MustCompile(`openebs_inflight_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 2000`),
				regexp.MustCompile(`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 50`),
				regexp.MustCompile(`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 50`),
				regexp.MustCompile(`openebs_dispatched_io_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 50`),
				regexp.MustCompile(`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_rebuild_count{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 3`),
				regexp.MustCompile(`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 500`),
				regexp.MustCompile(`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 500`),
				regexp.MustCompile(`openebs_rebuild_bytes{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 500`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c1698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c2698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
				regexp.MustCompile(`openebs_rebuild_status{pool="cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a",vol="pvc-1c3698bb-2dc6-11e9-bbe3-42010a80017a"} 1`),
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			runner = tt.run
			vol := New()
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
