// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mapiserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

var (
	snapshotResponse        = `{"actions":{},"id":"snapdemo1","links":{"self":"http://10.36.0.1:9501/v1/snapshotoutputs/snapdemo1"},"type":"snapshotOutput"}`
	errVolumeNameIsMissing  = errors.New("Volume name is missing")
	errBadReqErr            = errors.New(snapshotResponse)
	errVolNotFound          = errors.New("Volume not found")
	SnapshotListResponse    = `{"volume-snap-snap1.img": {"name": "volume-snap-snap1.img", "parent": "", "children":[ "volume-snap-snap2.img", "volume-head-001.img"], "usercreated":true, "removed":false, "created": "2018-06-12T19:33:34Z", "size": "0"}, "volume-snap-snap2.img": {"name": "volume-snap-snap2.img", "parent": "volume-snap-snap1.img", "children":[ "volume-snap-snap3.img"], "created": "2018-06-10T19:33:34Z", "size": "0"}, "volume-snap-snap3.img": {"name": "volume-snap-snap3.img", "parent": "", "children":[], "usercreated":true, "removed":false, "created": "2018-06-10T19:33:34Z", "size": "0"}, "volume-head-01.img": {"name": "volume-head-01.img", "parent": "", "children":[ ], "usercreated":true, "removed":false, "created": "2018-06-12T19:33:34Z", "size": "0"}}`
	ZeroSnapshotResponse    = `{"volume-head-001.img": {"name": "volume-head-001.img", "parent": "", "children":[], "usercreated":true, "removed":false, "created": "2018-06-10T19:33:34Z", "size": "0"}}`
	WrongDateFormatResponse = `{"volume-snap-snap1.img": {"name": "volume-snap-snap1.img", "parent": "", "children":[ "volume-snap-snap2.img", "volume-head-001.img"], "usercreated":true, "removed":false, "created": "2018-06-10T19:33:", "size": "0"}, "volume-snap-snap2.img": {"name": "volume-snap-snap2.img", "parent": "volume-snap-snap1.img", "children":[ "volume-snap-snap3.img"], "created": "2018-06-10T19:33:34Z", "size": "0"}, "volume-head-01.img": {"name": "volume-head-01.img", "parent": "", "children":[ ], "usercreated":true, "removed":false, "created": "2018-06-12T19:33:34Z", "size": "0"}}`
	errJSONError            = errors.New("unexpected end of JSON input")
	errdateParse            = errors.New("parsing time \"2018-06-10T19:33:\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"05\"")
)

func TestCreateSnapshot(t *testing.T) {
	tests := map[string]*struct {
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
			err:  errBadReqErr,
			addr: "MAPI_ADDR",
		},
		"VolumeNameMissing": {
			volumeName: "",
			snapName:   "xfgcuio87654er",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume name is missing",
				T:            t,
			},
			err:  errVolumeNameIsMissing,
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
			err:  errVolNotFound,
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
	tests := map[string]*struct {
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
	tests := map[string]*struct {
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
			err:  fmt.Errorf("Failed to get the snapshot info, found error - %v", errJSONError),
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
		"ParsingErrorWrongDateFormat": {
			volumeName: "dateFormat",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(WrongDateFormatResponse),
				T:            t,
			},
			err:  fmt.Errorf("Error changing date format to UnixDate, found error - %v", errdateParse),
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
