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

// volumeV1alpha1SpecificRequest is a http handler to handle HTTP
// requests to a OpenEBS volume.
func (s *HTTPServer) poolV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	glog.Infof("cas template based pool request was received: method '%s'", req.Method)

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

	poolName := strings.TrimPrefix(path, "/")
	glog.Infof(poolName)
	return p.read(poolName)
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

	glog.Infof("cas template based pools were listed successfully")
	return pools, nil
}

func (p *poolAPIOpsV1alpha1) read(poolName string) (*v1alpha1.StoragePool, error) {
	glog.Infof("CAS template based storage pool read request was received")

	sOps, err := pool.NewStoragePoolOperation(poolName)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	pool, err := sOps.Read()
	if err != nil {
		glog.Errorf("failed to read cas template based pools error: '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("cas template based pools were readed successfully")
	return pool, nil
}
