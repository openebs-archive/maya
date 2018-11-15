package mapiserver

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

// returns true when both errors are true or else returns false
func checkErr(err1, err2 error) bool {
	if (err1 != nil && err2 == nil) || (err1 == nil && err2 != nil) || (err1 != nil && err2 != nil && err1.Error() != err2.Error()) {
		return false
	}
	return true
}

func TestListPool(t *testing.T) {
	tests := map[string]*struct {
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: `{"items":[{"apiVersion":"openebs.io/v1alpha1","kind":"StoragePool","metadata":{"creationTimestamp":"2018-11-15T07:53:56Z","generation":1,"labels":{"openebs.io/cas-type":"cstor","openebs.io/cstor-pool":"cstor-sparse-pool-g5pi","openebs.io/storage-pool-claim":"cstor-sparse-pool","openebs.io/version":"0.7.0","kubernetes.io/hostname":"127.0.0.1","openebs.io/cas-template-name":"cstor-pool-create-default-0.7.0"},"name":"cstor-sparse-pool-g5pi","resourceVersion":"580","selfLink":"/apis/openebs.io/v1alpha1/storagepools/cstor-sparse-pool-g5pi","uid":"9a6a4b68-e8ab-11e8-b96a-b4b686bd0cff"},"spec":{"disks":{"diskList":["sparse-5a92ced3e2ee21eac7b930f670b5eab5"]},"format":"","message":"","mountpoint":"","name":"","nodename":"","path":"","poolSpec":{"cacheFile":"/tmp/cstor-sparse-pool.cache","overProvisioning":false,"poolType":"striped"}}}],"metadata":{"resourceVersion":"658","selfLink":"/apis/openebs.io/v1alpha1/storagepools"}}`,
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: "HTTP Error 404 : Not Found",
				T:            t,
			},
			err:  fmt.Errorf("Server status error: Not Found"),
			addr: "MAPI_ADDR",
		},
		"EmptyResponse": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 200,
				T:          t,
			},
			err:  fmt.Errorf("unexpected end of JSON input"),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			_, got := ListPools()

			if !checkErr(got, tt.err) {
				t.Fatalf("TestName: %v | ListVolumes() => Got: %v | Want: %v ", name, got, tt.err)
			}
		})
	}
}
