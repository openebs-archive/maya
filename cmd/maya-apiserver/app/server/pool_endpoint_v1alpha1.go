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

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	pool "github.com/openebs/maya/pkg/storagepool"
	"k8s.io/klog/v2"
)

type poolAPIOpsV1alpha1 struct {
	req  *http.Request
	resp http.ResponseWriter
}

// poolV1alpha1SpecificRequest is a http handler
// to handle HTTP requests to a OpenEBS pool.
func (s *HTTPServer) poolV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req == nil {
		return nil, CodedError(400, "failed to handle storage pool request: nil http request received")
	}

	klog.Infof(" received storage pool request: method '%s'", req.Method)

	poolOp := &poolAPIOpsV1alpha1{
		req:  req,
		resp: resp,
	}

	switch req.Method {
	case "GET":
		return poolOp.httpGet()
	default:
		return nil, CodedError(405, http.StatusText(405))
	}
}

func (p *poolAPIOpsV1alpha1) httpGet() (interface{}, error) {
	path := strings.TrimSpace(strings.TrimPrefix(p.req.URL.Path, "/latest/pools"))

	if path == "/" {
		return p.list()
	}
	poolName := strings.TrimSpace(strings.TrimPrefix(path, "/"))
	return p.read(poolName)
}

func (p *poolAPIOpsV1alpha1) list() (*v1alpha1.CStorPoolList, error) {
	klog.Infof("received storage pool list request")
	sOps, err := pool.NewStoragePoolOperation("")
	if err != nil {
		return nil, CodedErrorWrap(400, err)
	}

	pools, err := sOps.List()
	if err != nil {
		klog.Errorf("failed to list storage pool: '%+v'", err)
		return nil, CodedErrorWrap(500, err)
	}

	klog.Infof("storage pools listed successfully")
	return pools, nil
}

func (p *poolAPIOpsV1alpha1) read(poolName string) (*v1alpha1.CStorPool, error) {
	klog.Infof("received storage pool read request: %s", poolName)
	sOps, err := pool.NewStoragePoolOperation(poolName)
	if err != nil {
		return nil, CodedErrorWrap(400, err)
	}

	pools, err := sOps.Read()
	if err != nil {
		klog.Errorf("failed to read storage pool '%s': %+v", poolName, err)
		if isNotFound(err) {
			return nil, CodedErrorWrapf(404, err, "pool '%s' not found", poolName)
		}
		return nil, CodedErrorWrapf(500, err, "failed to read storage pool '%s'", poolName)
	}

	klog.Infof("storage pool '%s' read successfully", poolName)
	return pools, nil
}
