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

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/openebs/maya/cmd/maya-apiserver/app/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ugorji/go/codec"
)

var (
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "latest_openebs_volume_request_duration_seconds",
			Help:    "Request response time of the /latest/volumes.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, .5, 1, 2.5, 5, 10},
		},
		// code is http code and method is http method returned by
		// endpoint "/latest/volumes"
		[]string{"code", "method"},
	)
	// latestOpenEBSVolumeRequestCounter Count the no of request Since a
	// request has been made on /latest/volumes
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "latest_openebs_volume_requests_total",
			Help: "Total number of /latest/volumes requests.",
		},
		[]string{"code", "method"},
	)
)

type TestServer struct {
	T      testing.TB
	Dir    string
	Maya   *MayaApiServer
	Server *HTTPServer
}

func (s *TestServer) Cleanup() {
	s.Server.Shutdown()
	s.Maya.Shutdown()
	os.RemoveAll(s.Dir)
}

// makeHTTPTestServer returns a test server with full logging.
func makeHTTPTestServer(t testing.TB, fnmc func(mc *config.MayaConfig)) *TestServer {
	return makeHTTPTestServerWithWriter(t, nil, fnmc)
}

// makeHTTPTestServerNoLogs returns a test server which only prints maya logs and
// no http server logs
func makeHTTPTestServerNoLogs(t testing.TB, fnmc func(mc *config.MayaConfig)) *TestServer {
	return makeHTTPTestServerWithWriter(t, ioutil.Discard, fnmc)
}

// makeHTTPTestServerWithWriter returns a test server whose logs will be written to
// the passed writer. If the writer is nil, the logs are written to stderr.
func makeHTTPTestServerWithWriter(t testing.TB, w io.Writer, fnmc func(mc *config.MayaConfig)) *TestServer {
	dir, maya := makeMayaServer(t, fnmc)
	if w == nil {
		w = maya.logOutput
	}
	srv, err := NewHTTPServer(maya, maya.config, w)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	s := &TestServer{
		T:      t,
		Dir:    dir,
		Maya:   maya,
		Server: srv,
	}
	return s
}

func BenchmarkHTTPRequests(b *testing.B) {
	s := makeHTTPTestServerNoLogs(b, func(mc *config.MayaConfig) {

	})

	defer s.Cleanup()

	handler := func(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
		// TODO we are returning a num;
		// instead return some big payload i.e. big array of any structure
		return 1000, nil
	}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/v1/kv/key", nil)
			s.Server.wrap(RequestCounter, RequestDuration, handler)(resp, req)
		}
	})
}

func TestSetIndex(t *testing.T) {
	resp := httptest.NewRecorder()
	setIndex(resp, 1000)
	header := resp.Header().Get("X-Maya-Index")
	if header != "1000" {
		t.Fatalf("Bad: %v", header)
	}
	setIndex(resp, 2000)
	if v := resp.Header()["X-Maya-Index"]; len(v) != 1 {
		t.Fatalf("bad: %#v", v)
	}
}

func TestSetLastContact(t *testing.T) {
	resp := httptest.NewRecorder()
	setLastContact(resp, 123456*time.Microsecond)
	header := resp.Header().Get("X-Maya-LastContact")
	if header != "123" {
		t.Fatalf("Bad: %v", header)
	}
}

func TestSetHeaders(t *testing.T) {
	s := makeHTTPTestServer(t, nil)
	s.Maya.config.HTTPAPIResponseHeaders = map[string]string{"foo": "bar"}
	defer s.Cleanup()

	resp := httptest.NewRecorder()
	handler := func(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
		return "noop", nil
	}

	req, _ := http.NewRequest("GET", "/v1/kv/key", nil)
	s.Server.wrap(RequestCounter, RequestDuration, handler)(resp, req)
	header := resp.Header().Get("foo")

	if header != "bar" {
		t.Fatalf("expected header: %v, actual: %v", "bar", header)
	}

}

func TestContentTypeIsJSON(t *testing.T) {
	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()

	handler := func(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
		return "noop", nil
	}

	req, _ := http.NewRequest("GET", "/v1/kv/key", nil)
	s.Server.wrap(RequestCounter, RequestDuration, handler)(resp, req)

	contentType := resp.Header().Get("Content-Type")

	if contentType != "application/json" {
		t.Fatalf("Content-Type header was not 'application/json'")
	}
}

func TestPrettyPrint(t *testing.T) {
	testPrettyPrint("pretty=1", true, t)
}

func TestPrettyPrintOff(t *testing.T) {
	testPrettyPrint("pretty=0", false, t)
}

func TestPrettyPrintBare(t *testing.T) {
	testPrettyPrint("pretty", true, t)
}

func testPrettyPrint(pretty string, prettyFmt bool, t *testing.T) {
	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	r := struct {
		Name string
		Role string
		Org  string
	}{
		"das",
		"hacker",
		"openebs",
	}

	resp := httptest.NewRecorder()
	handler := func(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
		return r, nil
	}

	urlStr := "/v1/kv/key?" + pretty
	req, _ := http.NewRequest("GET", urlStr, nil)
	s.Server.wrap(RequestCounter, RequestDuration, handler)(resp, req)

	var expected bytes.Buffer
	if prettyFmt {
		enc := codec.NewEncoder(&expected, jsonHandlePretty)
		err := enc.Encode(r)
		if err == nil {
			expected.Write([]byte("\n"))
		} else {
			t.Fatalf("err while pretty encoding: %v", err)
		}
	} else {
		enc := codec.NewEncoder(&expected, jsonHandle)
		err := enc.Encode(r)

		if err != nil {
			t.Fatalf("err while encoding: %v", err)
		}
	}

	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !bytes.Equal(expected.Bytes(), actual) {
		t.Fatalf("bad:\nexpected:\t%q\n\nactual:\t\t%q", expected.String(), string(actual))
	}
}

func TestParseRegion(t *testing.T) {

	var region string

	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	req, err := http.NewRequest("GET",
		"/v1/kv/key?region=foo", nil)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	s.Server.parseRegion(req, &region)

	if region != "foo" {
		t.Fatalf("bad %s", region)
	}

	// reset the region
	region = ""
	req, err = http.NewRequest("GET", "/v1/kv/key", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	s.Server.parseRegion(req, &region)
	if region != "global" {
		t.Fatalf("bad %s", region)
	}
}

// assertIndex tests that X-Maya-Index is set and non-zero
func assertIndex(t *testing.T, resp *httptest.ResponseRecorder) {
	header := resp.Header().Get("X-Maya-Index")
	if header == "" || header == "0" {
		t.Fatalf("Bad: %v", header)
	}
}

// checkIndex is like assertIndex but returns an error
func checkIndex(resp *httptest.ResponseRecorder) error {
	header := resp.Header().Get("X-Maya-Index")
	if header == "" || header == "0" {
		return fmt.Errorf("Bad: %v", header)
	}
	return nil
}

// getIndex parses X-Maya-Index
func getIndex(t *testing.T, resp *httptest.ResponseRecorder) uint64 {
	header := resp.Header().Get("X-Maya-Index")
	if header == "" {
		t.Fatalf("Bad: %v", header)
	}
	val, err := strconv.Atoi(header)
	if err != nil {
		t.Fatalf("Bad: %v", header)
	}
	return uint64(val)
}

func httpTest(t testing.TB, fnmc func(mc *config.MayaConfig), f func(srv *TestServer)) {
	s := makeHTTPTestServer(t, fnmc)
	defer s.Cleanup()
	f(s)
}

func encodeReq(obj interface{}) io.ReadCloser {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.Encode(obj)
	return ioutil.NopCloser(buf)
}

func TestCodedError(t *testing.T) {
	tests := map[string]struct {
		code            int
		message         string
		expectedCode    int
		expectedMessage string
	}{
		"401 error": {401, "unauthorized", 401, "unauthorized"},
		"401 none":  {401, "", 401, ""},
		"500 error": {500, "internal server error", 500, "internal server error"},
		"500 none":  {500, "", 500, ""},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			httpErr := CodedError(mock.code, mock.message)
			if httpErr.Error() != mock.expectedMessage {
				t.Errorf("test '%s' failed: expected '%s' got '%s'", name, mock.expectedMessage, httpErr.Error())
			}
			if httpErr.Code() != mock.expectedCode {
				t.Errorf("test '%s' failed: expected '%d' got '%d'", name, mock.expectedCode, httpErr.Code())
			}
		})
	}
}

func TestCodedErrorf(t *testing.T) {
	tests := map[string]struct {
		code            int
		message         string
		args            []interface{}
		expectedCode    int
		expectedMessage string
	}{
		"401 error": {401, "%s %d", []interface{}{"unauthorized", 401}, 401, "{unauthorized 401}"},
		"500 error": {500, "%s %d", []interface{}{"internal error", 500}, 500, "{internal error 500}"},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			httpErr := CodedErrorf(mock.code, mock.message, mock.args...)
			if httpErr.Error() != mock.expectedMessage {
				t.Errorf("test '%s' failed: expected '%s' got '%s'", name, mock.expectedMessage, httpErr.Error())
			}
			if httpErr.Code() != mock.expectedCode {
				t.Errorf("test '%s' failed: expected '%d' got '%d'", name, mock.expectedCode, httpErr.Code())
			}
		})
	}
}

func TestCodedErrorWrap(t *testing.T) {
	tests := map[string]struct {
		code            int
		err             error
		expectedCode    int
		expectedMessage string
	}{
		"401 error": {
			401, fmt.Errorf("unauthorized"), 401, "{unauthorized}",
		},
		"500 error": {
			500, fmt.Errorf("internal error"), 500, "{internal error}",
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			httpErr := CodedErrorWrap(mock.code, mock.err)
			if httpErr.Error() != mock.expectedMessage {
				t.Errorf("test '%s' failed: expected '%s' got '%s'", name, mock.expectedMessage, httpErr.Error())
			}
			if httpErr.Code() != mock.expectedCode {
				t.Errorf("test '%s' failed: expected '%d' got '%d'", name, mock.expectedCode, httpErr.Code())
			}
		})
	}
}

func TestCodedErrorWrapf(t *testing.T) {
	tests := map[string]struct {
		code            int
		err             error
		message         string
		args            []interface{}
		expectedCode    int
		expectedMessage string
	}{
		"401 error": {
			401, fmt.Errorf("unauthorized"), "%s", []interface{}{"who is this?"},
			401, "error: {unauthorized}, msg: {who is this?}",
		},
		"500 error": {
			500, fmt.Errorf("internal error"), "%s", []interface{}{"what is this?"},
			500, "error: {internal error}, msg: {what is this?}",
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			httpErr := CodedErrorWrapf(mock.code, mock.err, mock.message, mock.args...)
			if httpErr.Error() != mock.expectedMessage {
				t.Errorf("test '%s' failed: expected '%s' got '%s'", name, mock.expectedMessage, httpErr.Error())
			}
			if httpErr.Code() != mock.expectedCode {
				t.Errorf("test '%s' failed: expected '%d' got '%d'", name, mock.expectedCode, httpErr.Code())
			}
		})
	}
}
