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

func TestCreateVolume(t *testing.T) {
	tests := map[string]struct {
		volumeName  string
		size        string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			volumeName: "qwewretrytu",
			size:       "1G",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: "Volume 'qwewretrytu' deleted Successfully",
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			volumeName: "12324rty653423",
			size:       "1G",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: "Volume '12324rty653423' deleted Successfully",

				T: t,
			},
			err:  fmt.Errorf("Status error: %v", http.StatusText(404)),
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "234t5rgfgt-ht4",
			size:       "1G",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: "Volume '12324rty653423' deleted Successfully",
			},
			err:  util.MAPIADDRNotSet,
			addr: "",
		},
		"VolumeNameMissing": {
			volumeName: "",
			size:       "1G",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume name is missing",
				T:            t,
			},
			err:  fmt.Errorf("Status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := CreateVolume(tt.volumeName, tt.size)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("CreateVolume(%v, %v) => got %v, want %v ", tt.volumeName, tt.size, got, tt.err)
			}
		})
	}
}
