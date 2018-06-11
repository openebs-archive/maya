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
	snapshotResponse     = `{"actions":{},"id":"snapdemo1","links":{"self":"http://10.36.0.1:9501/v1/snapshotoutputs/snapdemo1"},"type":"snapshotOutput"}`
	volumeNameIsMissing  = errors.New("Volume name is missing")
	badReqErr            = errors.New(snapshotResponse)
	volNotFound          = errors.New("Volume not found")
	SnapshotListResponse = `{"volume-snap-snap1.img": {"name": "volume-snap-snap1.img", "parent": "", "children":[ "volume-snap-snap2.img", "volume-head-001.img"], "usercreated":true, "removed":false, "created": "2018-06-10T19:33:34Z", "size": "0"}, "volume-snap-snap2.img": {"name": "volume-snap-snap2.img", "parent": "volume-snap-snap1.img", "children":[ "volume-snap-snap3.img"], "created": "2018-06-10T19:33:34Z", "size": "0"}, "volume-snap-snap3.img": {"name": "volume-snap-snap3.img", "parent": "", "children":[], "usercreated":true, "removed":false, "created": "2018-06-10T19:33:34Z", "size": "0"}}`
	ZeroSnapshotResponse = `{"volume-head-001.img": {"name": "volume-head-001.img", "parent": "", "children":[], "usercreated":true, "removed":false, "created": "2018-06-10T19:33:34Z", "size": "0"}}`
	jsonError            = errors.New("unexpected end of JSON input")
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
				ResponseBody: string(SnapshotListResponse),
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
				ResponseBody: string(SnapshotListResponse),
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
				ResponseBody: string(SnapshotListResponse),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"NoResponse": {
			volumeName: "test",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 500,
				// ResponseBody: string(SnapshotListResponse),
				T: t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(500)),
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			volumeName: "12324rty653423",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(SnapshotListResponse),
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "234t5rgfgt-ht4",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(SnapshotListResponse),
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
				ResponseBody: string(SnapshotListResponse),
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
		"UnableToParseJSON": {
			volumeName: "foo",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(""),
				T:            t,
			},
			err:  fmt.Errorf("Failed to get the snapshot info, found error - %v", jsonError),
			addr: "MAPI_ADDR",
		},
		"ZeroSnapshots": {
			volumeName: "test",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(ZeroSnapshotResponse),
				T:            t,
			},
			err:  nil,
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
