package command

import (
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openebs/maya/pkg/util"
	utiltesting "k8s.io/client-go/util/testing"
)

var (
	response = `{"metadata":{"annotations":{"vsm.openebs.io/targetportals":"10.98.65.136:3260","vsm.openebs.io/cluster-ips":"10.98.65.136","openebs.io/jiva-iqn":"iqn.2016-09.com.openebs.jiva:vol","deployment.kubernetes.io/revision":"1","openebs.io/storage-pool":"default","vsm.openebs.io/replica-count":"1","openebs.io/jiva-controller-status":"Running","openebs.io/volume-monitor":"false","openebs.io/replica-container-status":"Running","openebs.io/jiva-controller-cluster-ip":"10.98.65.136","openebs.io/jiva-replica-status":"Running","vsm.openebs.io/iqn":"iqn.2016-09.com.openebs.jiva:vol","openebs.io/capacity":"2G","openebs.io/jiva-controller-ips":"10.36.0.6","openebs.io/jiva-replica-ips":"10.36.0.7","vsm.openebs.io/replica-status":"Running","vsm.openebs.io/controller-status":"Running","openebs.io/controller-container-status":"Running","vsm.openebs.io/replica-ips":"10.36.0.7","openebs.io/jiva-target-portal":"10.98.65.136:3260","openebs.io/volume-type":"jiva","openebs.io/jiva-replica-count":"1","vsm.openebs.io/volume-size":"2G","vsm.openebs.io/controller-ips":"10.36.0.6"},"creationTimestamp":null,"labels":{},"name":"vol"},"status":{"Message":"","Phase":"Running","Reason":""}}`
)

func TestGetVolDetails(t *testing.T) {
	var (
		server     *httptest.Server
		annotation = Annotations{}
	)
	tests := map[string]struct {
		volumeName string

		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"500InternalServerError": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   500,
				ResponseBody: string(response),
				T:            t,
			},
			err:  util.InternalServerError,
			addr: "MAPI_ADDR",
		},
		"503ServerUnavailable": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   503,
				ResponseBody: string(response),
				T:            t,
			},
			err:  util.ServerUnavailable,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 400,
				T:          t,
			},
			err:  io.EOF,
			addr: "MAPI_ADDR",
		},
		"MAPIADDRSet": {
			volumeName: "vol",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(response),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"MAPIADDRNotSet": {
			volumeName: "vol",
			fakeHandler: utiltesting.FakeHandler{
				ResponseBody: string(response),
				StatusCode:   200,
				T:            t,
			},
			err:  util.MAPIADDRNotSet,
			addr: "",
		},
		"404NotFound": {
			volumeName: "vol",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 404,
				T:          t,
			},
			err:  util.PageNotFound,
			addr: "MAPI_ADDR",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server = httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			if got := annotation.GetVolAnnotations(tt.volumeName); got != tt.err {
				t.Fatalf("GetVolDetails(%v) => got %v, want %v ", tt.volumeName, got, tt.err)
			}
			defer os.Unsetenv(tt.addr)
			defer server.Close()

		})
	}
}
