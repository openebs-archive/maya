package command

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
)

var (
	response = `{"metadata":{"annotations":{"vsm.openebs.io/targetportals":"10.98.65.136:3260","vsm.openebs.io/cluster-ips":"10.98.65.136","openebs.io/jiva-iqn":"iqn.2016-09.com.openebs.jiva:vol","deployment.kubernetes.io/revision":"1","openebs.io/storage-pool":"default","vsm.openebs.io/replica-count":"1","openebs.io/jiva-controller-status":"Running","openebs.io/volume-monitor":"false","openebs.io/replica-container-status":"Running","openebs.io/jiva-controller-cluster-ip":"10.98.65.136","openebs.io/jiva-replica-status":"Running","vsm.openebs.io/iqn":"iqn.2016-09.com.openebs.jiva:vol","openebs.io/capacity":"2G","openebs.io/jiva-controller-ips":"10.36.0.6","openebs.io/jiva-replica-ips":"10.36.0.7","vsm.openebs.io/replica-status":"Running","vsm.openebs.io/controller-status":"Running","openebs.io/controller-container-status":"Running","vsm.openebs.io/replica-ips":"10.36.0.7","openebs.io/jiva-target-portal":"10.98.65.136:3260","openebs.io/volume-type":"jiva","openebs.io/jiva-replica-count":"1","vsm.openebs.io/volume-size":"2G","vsm.openebs.io/controller-ips":"10.36.0.6"},"creationTimestamp":null,"labels":{},"name":"vol"},"status":{"Message":"","Phase":"Running","Reason":""}}`
)

func TestGetVolDetails(t *testing.T) {
	var (
		volume v1.Volume
		server *httptest.Server
	)
	tests := map[string]struct {
		volumeName string
		resp       interface{}
		err        error
		addr       string
	}{
		"MAPIADDRSet":    {"vol", response, nil, "MAPI_ADDR"},
		"MAPIADDRNotSet": {"vol", response, util.MAPIADDRNotSet, ""},
		"EmptyResponse":  {"vol", "", io.EOF, "MAPI_ADDR"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, tt.resp)
			}))
			os.Setenv(tt.addr, server.URL)
			if got := GetVolDetails(tt.volumeName, &volume); got != tt.err {
				t.Fatalf("GetVolDetails(%v) => got %v, want %v ", tt.volumeName, got, tt.err)
			}
			defer os.Unsetenv(tt.addr)
			defer server.Close()

		})
	}
}
