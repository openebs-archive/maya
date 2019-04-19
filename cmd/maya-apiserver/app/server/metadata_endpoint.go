package server

import (
	"net/http"
	"strings"
)

const (
	// AnyInstance stands for OpenEBS
	// being able to act as a persistence
	// mechanism for any type of compute
	// instance
	AnyInstance = "any-compute"

	// AnyZone specifies OpenEBS' availability zone
	AnyZone = "any-zone"
)

// MetaSpecificRequest is a handler responsible to
// perform validation and meta variable substitution
// into request paths
func (s *HTTPServer) MetaSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	path := strings.TrimPrefix(req.URL.Path, "/latest/meta-data")

	// Is req valid ?
	if path == req.URL.Path {
		return nil, CodedError(421, ErrInvalidPath)
	}

	// We do an exact suffix comparison
	switch {
	case strings.Compare(path, "/instance-id") == 0:
		return s.metaInstanceID(resp, req)
	case strings.Compare(path, "/placement/availability-zone") == 0:
		return s.metaAvailabilityZone(resp, req)
	default:
		return nil, CodedError(421, ErrInvalidPath)
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
