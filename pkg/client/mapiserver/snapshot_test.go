package mapiserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/openebs/maya/pkg/util"
	utiltesting "k8s.io/client-go/util/testing"
)

var (
	snapshotResponse          = `{"actions":{},"id":"snapdemo1","links":{"self":"http://10.36.0.1:9501/v1/snapshotoutputs/snapdemo1"},"type":"snapshotOutput"}`
	volumeNameIsMissing       = errors.New("Volume name is missing")
	badReqErr                 = errors.New(snapshotResponse)
	volNotFound               = errors.New("Volume not found")
	NamespaceSnapshotResponse = `{volume-snap-testsnap.img":{"children":["volume-head-001.img"],"created":"2018-06-06T12:45:55Z","name":"volume-snap-testsnap.img","parent":"volume-snap-85437a6a-1945-427b-ba93-58011b34f634.img","removed":false,"size":"377753600","usercreated":true}`
)

func TestCreateSnapshot(t *testing.T) {
	tests := map[string]struct {
		volumeName  string
		snapName    string
		namespace   string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			volumeName: "qwewretrytu",
			snapName:   "fgdjhk",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			volumeName: "12324rty653423",
			snapName:   "134efvet454",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:  badReqErr,
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "234t5rgfgt-ht4",
			snapName:   "-09uhbvvbfghj",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
			},
			err:  util.MAPIADDRNotSet,
			addr: "",
		},
		"VolumeNameMissing": {
			volumeName: "",
			snapName:   "xfgcuio87654er",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume name is missing",
				T:            t,
			},
			err:  volumeNameIsMissing,
			addr: "MAPI_ADDR",
		},
		"VolumeNotFound": {
			volumeName: "fdghjk",
			snapName:   "xfgcuio87654er",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume not found",
				T:            t,
			},
			err:  volNotFound,
			addr: "MAPI_ADDR",
		},
		"AppNameSpaceVolume": {
			volumeName: "testvol",
			snapName:   "testsnap",
			namespace:  "app",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(NamespaceSnapshotResponse),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"WrongNameSpaceVolume": {
			volumeName: "testvol",
			snapName:   "testsnap",
			namespace:  "",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: string("Volume 'testvol' not found"),
				T:            t,
			},
			err:  fmt.Errorf("Volume 'testvol' not found"),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := CreateSnapshot(tt.volumeName, tt.snapName, tt.namespace)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("CreateSnapshot(%v, %v) => got %v, want %v ", tt.volumeName, tt.snapName, got, tt.err)
			}
		})
	}
}

func TestRevertSnapshot(t *testing.T) {
	tests := map[string]struct {
		volumeName  string
		snapName    string
		namespace   string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			volumeName: "qwewretrytu",
			snapName:   "fgdjhk",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			volumeName: "12324rty653423",
			snapName:   "134efvet454",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "234t5rgfgt-ht4",
			snapName:   "-09uhbvvbfghj",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
			},
			err:  util.MAPIADDRNotSet,
			addr: "",
		},
		"VolumeNameMissing": {
			volumeName: "",
			snapName:   "xfgcuio87654er",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume name is missing",
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"VolumeNotFound": {
			volumeName: "fdghjk",
			snapName:   "xfgcuio87654er",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume 'fdghjk' not found",
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"RevertAppNameSpaceVolume": {
			volumeName: "testvol",
			namespace:  "app",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(NamespaceSnapshotResponse),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"RevertWrongNameSpaceVolume": {
			volumeName: "testvol",
			namespace:  "",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: string("Volume 'testvol' not found"),
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(404)),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := RevertSnapshot(tt.volumeName, tt.snapName, tt.namespace)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("RevertSnapshot(%v, %v) => got %v, want %v ", tt.volumeName, tt.snapName, got, tt.err)
			}
		})
	}
}

func TestListSnapshot(t *testing.T) {
	tests := map[string]struct {
		volumeName  string
		namespace   string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			volumeName: "qwewretrytu",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			volumeName: "12324rty653423",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "234t5rgfgt-ht4",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
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
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"VolumeNotFound": {
			volumeName: "fdghjk",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume 'fdghjk' not found",
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"ListAppNameSpaceVolume": {
			volumeName: "testvol",
			namespace:  "app",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(NamespaceSnapshotResponse),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"ListWrongNameSpaceVolume": {
			volumeName: "testvol",
			namespace:  "",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: string("Volume 'testvol' not found"),
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(404)),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := ListSnapshot(tt.volumeName, tt.namespace)
			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("ListSnapshot(%v) => got %v, want %v ", tt.volumeName, got, tt.err)
			}
		})
	}
}
