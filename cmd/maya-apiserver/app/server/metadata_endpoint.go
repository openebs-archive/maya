package server

import (
	"net/http"
	"strings"
)

const (
	// OpenEBS can be used as a persistence mechanism for
	// any type of compute instance
	AnyInstance = "any-compute"

	// TODO We shall see how to construct an Availability Zone
	AnyZone = "any-zone"
)

func (s *HTTPServer) MetaSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	path := strings.TrimPrefix(req.URL.Path, "/latest/meta-data")

	// Is req valid ?
	if path == req.URL.Path {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	// We do an exact suffix comparision
	switch {

	case strings.Compare(path, "/instance-id") == 0:
		return s.metaInstanceID(resp, req)

	case strings.Compare(path, "/placement/availability-zone") == 0:
		return s.metaAvailabilityZone(resp, req)

	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// EBS demands a particular instance id to be returned during
// aws session creation.
func (s *HTTPServer) metaInstanceID(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	if req.Method != "GET" {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	return AnyInstance, nil
}

func (s *HTTPServer) metaAvailabilityZone(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	if req.Method != "GET" {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	return AnyZone, nil
}
