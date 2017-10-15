package server

import (
	"bytes"
	"github.com/ugorji/go/codec"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

/*var (
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
*/
func TestInvalidReqMetaData(t *testing.T) {
	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()
	// valid uri is meta-data & not metadata
	req, _ := http.NewRequest("GET", "/metadata/instance-id", nil)

	out, err := s.Server.MetaSpecificRequest(resp, req)

	if err == nil {
		t.Fatalf("ERR: expected: %v, got: %v", ErrInvalidMethod, err)
	}

	if err.Error() != ErrInvalidMethod {
		t.Fatalf("ERR: expected: %v, got: %v", ErrInvalidMethod, err.Error())
	}

	if out != nil {
		t.Fatalf("Service must not return any value, for invalid request")
	}
}

func TestMetaInstanceID(t *testing.T) {
	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/latest/meta-data/instance-id", nil)

	out, err := s.Server.MetaSpecificRequest(resp, req)

	if err != nil {
		t.Fatalf("ERR: %v", err)
	}

	if out == "" || out == nil {
		t.Fatalf("Service must return a non empty instance")
	}
}

func TestInvalidReqMetaInstanceID(t *testing.T) {
	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/latest/meta-data/instance-id", nil)

	out, err := s.Server.MetaSpecificRequest(resp, req)

	if err == nil {
		t.Fatalf("ERR: expected: %v, got: %v", CodedError(405, ErrInvalidMethod), err)
	}

	if err.Error() != ErrInvalidMethod {
		t.Fatalf("ERR: expected: %v, got: %v", ErrInvalidMethod, err.Error())
	}

	if out != nil {
		t.Fatalf("Service must not return any value, for invalid request")
	}
}

func TestMetaAvailZone(t *testing.T) {
	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/latest/meta-data/placement/availability-zone", nil)

	out, err := s.Server.MetaSpecificRequest(resp, req)

	if err != nil {
		t.Fatalf("ERR: %v", err)
	}

	if out == "" || out == nil {
		t.Fatalf("Service must return a non empty instance")
	}
}

func TestInvalidReqMetaAvailZone(t *testing.T) {
	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/latest/meta-data/placement/availability-zone", nil)

	out, err := s.Server.MetaSpecificRequest(resp, req)

	if err == nil {
		t.Fatalf("ERR: expected: %v, got: %v", CodedError(405, ErrInvalidMethod), err)
	}

	if err.Error() != ErrInvalidMethod {
		t.Fatalf("ERR: expected: %v, got: %v", ErrInvalidMethod, err.Error())
	}

	if out != nil {
		t.Fatalf("Service must not return any value, for invalid request")
	}
}

func TestInvalidReqPathMetaViaWrap1(t *testing.T) {

	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()

	// An invalid URL path
	req, err := http.NewRequest("GET",
		"/oddy/meta-data/placement/availability-zone", nil)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// The handler i.e. `s.Server.MetaSpecificRequest` is curried via `wrap`
	// function & returned. The returned func is immediately invoked by
	// passing the respective arguments i.e. `resp` & `req`.
	// Learn more by understanding -
	// `Immediately Invoked Function Expression (IIFE)`.
	s.Server.wrap(RequestCounter, RequestDuration, s.Server.MetaSpecificRequest)(resp, req)

	contentType := resp.Header().Get("Content-Type")

	if len(contentType) != 0 {
		t.Fatalf("err content type, expected: nil, got: %s", contentType)
	}

	// This should be an invalid path/method error
	if resp.Code != 405 {
		t.Fatalf("err http resp code, expected: 405, got: %v", resp.Code)
	}

	// actuals
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("err reading response: %v", err)
	}

	// compare expectations with actuals
	if !bytes.Equal([]byte(ErrInvalidMethod), actual) {
		t.Fatalf("bad:\nexpected:\t%q\n\nactual:\t\t%q", ErrInvalidMethod, string(actual))
	}
}

func TestInvalidReqPathMetaViaWrap2(t *testing.T) {

	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()

	// An invalid URL path
	// NOTE: `v1` is invalid here
	req, err := http.NewRequest("GET",
		"/latest/meta-data/v1/placement/availability-zone", nil)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// The handler i.e. `s.Server.MetaSpecificRequest` is curried via `wrap`
	// function & returned. The returned func is immediately invoked by
	// passing the respective arguments i.e. `resp` & `req`.

	// Learn more by understanding -
	// `Immediately Invoked Function Expression (IIFE)`.
	s.Server.wrap(RequestCounter, RequestDuration, s.Server.MetaSpecificRequest)(resp, req)

	contentType := resp.Header().Get("Content-Type")

	if len(contentType) != 0 {
		t.Fatalf("err content type, expected: nil, got: %s", contentType)
	}

	// This should be an invalid path/method error
	if resp.Code != 405 {
		t.Fatalf("err http resp code, expected: 405, got: %v", resp.Code)
	}

	// actuals
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("err reading response: %v", err)
	}

	// compare expectations with actuals
	if !bytes.Equal([]byte(ErrInvalidMethod), actual) {
		t.Fatalf("bad:\nexpected:\t%q\n\nactual:\t\t%q", ErrInvalidMethod, string(actual))
	}
}

func TestMetaAvailZoneViaWrap(t *testing.T) {

	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()

	req, err := http.NewRequest("GET",
		"/latest/meta-data/placement/availability-zone", nil)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// The handler i.e. `s.Server.MetaSpecificRequest` is curried via `wrap`
	// function & returned. The returned func is immediately invoked by
	// passing the respective arguments i.e. `resp` & `req`.
	// Learn more by understanding -
	// `Immediately Invoked Function Expression (IIFE)`.
	s.Server.wrap(RequestCounter, RequestDuration, s.Server.MetaSpecificRequest)(resp, req)

	contentType := resp.Header().Get("Content-Type")

	if contentType != "application/json" {
		t.Fatalf("Content-Type header was not 'application/json'")
	}

	// expectations
	var expected bytes.Buffer
	enc := codec.NewEncoder(&expected, jsonHandle)
	err = enc.Encode(AnyZone)

	if err != nil {
		t.Fatalf("err while encoding: %v", err)
	}

	// actuals
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("err reading response: %v", err)
	}

	// compare
	if !bytes.Equal(expected.Bytes(), actual) {
		t.Fatalf("bad:\nexpected:\t%q\n\nactual:\t\t%q", string(expected.Bytes()), string(actual))
	}
}

func TestInvalidReqMetaAvailZoneViaWrap(t *testing.T) {

	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()

	// NOTE: `POST` req is invalid
	req, err := http.NewRequest("POST",
		"/latest/meta-data/placement/availability-zone", nil)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// The handler i.e. `s.Server.MetaSpecificRequest` is curried via `wrap`
	// function & returned. The returned func is immediately invoked by
	// passing the respective arguments i.e. `resp` & `req`.
	// Learn more by understanding -
	// `Immediately Invoked Function Expression (IIFE)`.
	s.Server.wrap(RequestCounter, RequestDuration, s.Server.MetaSpecificRequest)(resp, req)

	contentType := resp.Header().Get("Content-Type")

	if len(contentType) != 0 {
		t.Fatalf("err content type, expected: nil, got: %s", contentType)
	}

	// This should be an invalid path/method error
	if resp.Code != 405 {
		t.Fatalf("err http resp code, expected: 405, got: %v", resp.Code)
	}

	// actuals
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("err reading response: %v", err)
	}

	// compare expectations with actuals
	if !bytes.Equal([]byte(ErrInvalidMethod), actual) {
		t.Fatalf("bad:\nexpected:\t%q\n\nactual:\t\t%q", ErrInvalidMethod, string(actual))
	}
}

func TestMetaInstanceIDViaWrap(t *testing.T) {

	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()

	req, err := http.NewRequest("GET",
		"/latest/meta-data/instance-id", nil)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// The handler i.e. `s.Server.MetaSpecificRequest` is curried via `wrap`
	// function & returned. The returned func is immediately invoked by
	// passing the respective arguments i.e. `resp` & `req`.
	// Learn more by understanding -
	// `Immediately Invoked Function Expression (IIFE)`.
	s.Server.wrap(RequestCounter, RequestDuration, s.Server.MetaSpecificRequest)(resp, req)

	contentType := resp.Header().Get("Content-Type")

	if contentType != "application/json" {
		t.Fatalf("Content-Type header was not 'application/json'")
	}

	// expectations
	var expected bytes.Buffer
	enc := codec.NewEncoder(&expected, jsonHandle)
	err = enc.Encode(AnyInstance)

	if err != nil {
		t.Fatalf("err while encoding: %v", err)
	}

	// actuals
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("err reading response: %v", err)
	}

	// compare
	if !bytes.Equal(expected.Bytes(), actual) {
		t.Fatalf("bad:\nexpected:\t%q\n\nactual:\t\t%q", string(expected.Bytes()), string(actual))
	}
}

func TestInvalidReqMetaInstanceIDViaWrap(t *testing.T) {

	s := makeHTTPTestServer(t, nil)
	defer s.Cleanup()

	resp := httptest.NewRecorder()

	// NOTE: `POST` req is invalid
	req, err := http.NewRequest("POST",
		"/latest/meta-data/instance-id", nil)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// The handler i.e. `s.Server.MetaSpecificRequest` is curried via `wrap`
	// function & returned. The returned func is immediately invoked by
	// passing the respective arguments i.e. `resp` & `req`.
	// Learn more by understanding -
	// `Immediately Invoked Function Expression (IIFE)`.
	s.Server.wrap(RequestCounter, RequestDuration, s.Server.MetaSpecificRequest)(resp, req)

	contentType := resp.Header().Get("Content-Type")

	if len(contentType) != 0 {
		t.Fatalf("err content type, expected: nil, got: %s", contentType)
	}

	// This should be an invalid path/method error
	if resp.Code != 405 {
		t.Fatalf("err http resp code, expected: 405, got: %v", resp.Code)
	}

	// actuals
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("err reading response: %v", err)
	}

	// compare expectations with actuals
	if !bytes.Equal([]byte(ErrInvalidMethod), actual) {
		t.Fatalf("bad:\nexpected:\t%q\n\nactual:\t\t%q", ErrInvalidMethod, string(actual))
	}
}
