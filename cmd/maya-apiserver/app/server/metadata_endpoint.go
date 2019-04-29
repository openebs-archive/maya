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
