package mapiserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/openebs/maya/pkg/util"
	utiltesting "k8s.io/client-go/util/testing"
)

func TestDeleteVolume(t *testing.T) {
	tests := map[string]struct {
		volumeName  string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			volumeName: "qwewretrytu",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: "Volume 'qwewretrytu' deleted Successfully",
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "234t5rgfgt-ht4",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: "Volume '12324rty653423' deleted Successfully",
			},
			err:  util.MAPIADDRNotSet,
			addr: "",
		},
		"VolumeNameMissing": {
			volumeName: "",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume name is missing",
				T:            t,
			},
			err:  fmt.Errorf("Status error: %v ", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"VolumeNotPresent": {
			volumeName: "volume",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: "Volume 'volume' not found",
				T:            t,
			},
			err:  fmt.Errorf("Status error: %v ", http.StatusText(404)),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := DeleteVolume(tt.volumeName)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("DeleteVolume(%v) => got %v, want %v ", tt.volumeName, got, tt.err)
			}
		})
	}
}
