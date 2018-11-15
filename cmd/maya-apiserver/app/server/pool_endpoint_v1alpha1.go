/*
Copyright 2018 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pool "github.com/openebs/maya/pkg/storagepool"
)

type poolAPIOpsV1alpha1 struct {
	req  *http.Request
	resp http.ResponseWriter
}

// poolV1alpha1SpecificRequest is a http handler to handle HTTP
// requests to a OpenEBS pool.
func (s *HTTPServer) poolV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	glog.Infof("Cas template based pool request was received: method '%s'", req.Method)

	if req == nil {
		return nil, CodedError(400, "nil http request was received")
	}

	poolOp := &poolAPIOpsV1alpha1{
		req:  req,
		resp: resp,
	}

	switch req.Method {
	case "GET":
		return poolOp.httpGet()
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

func (p *poolAPIOpsV1alpha1) httpGet() (interface{}, error) {
	path := strings.TrimSpace(strings.TrimPrefix(p.req.URL.Path, "/latest/pools"))

	if path == "/" {
		return p.list()
	}
	return nil, nil
}

func (p *poolAPIOpsV1alpha1) list() (*v1alpha1.StoragePoolList, error) {
	glog.Infof("CAS template based storage pool list request was received")

	sOps, err := pool.NewStoragePoolOperation("")
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	pools, err := sOps.List()
	if err != nil {
		glog.Errorf("failed to list cas template based pools error: '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Cas template based pools were listed successfully")
	return pools, nil
}
