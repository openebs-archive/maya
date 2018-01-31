package mapiserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	utiltesting "k8s.io/client-go/util/testing"
)

var (
	response = `{"items":[{"metadata":{"annotations":{"openebs.io/jiva-controller-status":"Running","vsm.openebs.io/replica-ips":"10.42.0.2,10.44.0.2","vsm.openebs.io/replica-status":"Running,Running","openebs.io/jiva-target-portal":"10.109.180.113:3260","openebs.io/jiva-controller-cluster-ip":"10.109.180.113","vsm.openebs.io/controller-ips":"10.42.0.1","openebs.io/volume-type":"jiva","vsm.openebs.io/replica-count":"2","openebs.io/jiva-replica-count":"2","vsm.openebs.io/volume-size":"1G","openebs.io/jiva-controller-ips":"10.42.0.1","openebs.io/jiva-replica-status":"Running,Running","openebs.io/replica-container-status":"Running,Running","deployment.kubernetes.io/revision":"1","vsm.openebs.io/iqn":"iqn.2016-09.com.openebs.jiva:vol2","vsm.openebs.io/cluster-ips":"10.109.180.113","openebs.io/capacity":"1G","openebs.io/controller-container-status":"Running","openebs.io/jiva-iqn":"iqn.2016-09.com.openebs.jiva:vol2","openebs.io/storage-pool":"default","vsm.openebs.io/controller-status":"Running","openebs.io/jiva-replica-ips":"10.42.0.2,10.44.0.2","vsm.openebs.io/targetportals":"10.109.180.113:3260","openebs.io/volume-monitor":"false"},"creationTimestamp":null,"labels":{},"name":"vol2"},"status":{"Message":"","Phase":"Running","Reason":""}}],"metadata":{}}`
)

func TestListVolumes(t *testing.T) {
	var (
		vsm v1.VolumeList
	)
	tests := map[string]struct {
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(response),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: string(response),
				T:            t,
			},
			err:  fmt.Errorf("Status Error: %v", http.StatusText(404)),
			addr: "MAPI_ADDR",
		},
		"EmptyResponse": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 204,
				T:          t,
			},
			err:  fmt.Errorf("Status Error: %v", http.StatusText(204)),
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(response),
				T:            t,
			},
			err:  util.MAPIADDRNotSet,
			addr: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := ListVolumes(&vsm)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("ListVolumes(%v) => got %v, want %v ", vsm, got, tt.err)
			}
			defer os.Unsetenv(tt.addr)
			defer server.Close()
		})
	}
}
