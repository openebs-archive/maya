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
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

func TestInitialize(t *testing.T) {
	tests := map[string]struct {
		key   string
		value string
	}{
		"MAPI_ADDRSet":    {"MAPI_ADDR", "127.0.0.1"},
		"MAPI_ADDRNotSet": {"MAPI_ADDR", ""},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv(tt.key, tt.value)
			defer os.Unsetenv(tt.key)
			Initialize()
		})
	}
}

func TestGetURL(t *testing.T) {
	cases := map[string]*struct {
		addr, port     string
		envaddr        string
		expectedoutput string
	}{
		"Environment variable set": {
			envaddr:        "192.168.0.2",
			expectedoutput: "192.168.0.2",
		},
		"Environment variable not set": {
			addr:           "192.168.0.1",
			port:           "5656",
			expectedoutput: "http://192.168.0.1:5656",
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			if len(tt.envaddr) > 0 {
				os.Setenv("MAPI_ADDR", tt.envaddr)

			} else {
				MAPIAddr = tt.addr
				MAPIAddrPort = tt.port
			}

			got := GetURL()
			os.Unsetenv("MAPI_ADDR")
			MAPIAddr = ""
			if !reflect.DeepEqual(got, tt.expectedoutput) {
				t.Fatalf("GetURL => got %v, want %v ", got, tt.expectedoutput)
			}

		})
	}
}

func TestGetRequest(t *testing.T) {
	testcases := map[string]*struct {
		fakeHandler utiltesting.FakeHandler
		namespace   string
		err         error
		chkbody     bool
		usefakeurl  bool
		invalidurl  string
	}{
		"Invalid Url with no namespace": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(""),
				T:            t,
			},
			namespace:  "",
			err:        errors.New("Invalid URL"),
			chkbody:    false,
			usefakeurl: false,
			invalidurl: "",
		},
		"No response with no namespace": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "",
			err:        errors.New("Server status error: Bad Request"),
			chkbody:    false,
			usefakeurl: true,
		},
		"Url not reachable": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "",
			err:        errors.New(`Get "invalid": unsupported protocol scheme ""`),
			chkbody:    false,
			usefakeurl: false,
			invalidurl: "invalid",
		},
		"not parsable url": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "",
			err:        errors.New(`parse "*:": first path segment in URL cannot contain colon`),
			chkbody:    false,
			usefakeurl: false,
			invalidurl: "*:",
		},
		"chk body is enabled": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "",
			err:        errors.New(response),
			chkbody:    true,
			usefakeurl: true,
		},
		"chk body is enabled with namespace": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "default",
			err:        errors.New(response),
			chkbody:    true,
			usefakeurl: true,
		},
	}

	for name, tt := range testcases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			url := ""
			if tt.usefakeurl {
				url = server.URL
			} else {
				url = tt.invalidurl
			}
			defer server.Close()
			_, got := getRequest(url, tt.namespace, tt.chkbody)
			if got.Error() != tt.err.Error() {
				t.Fatalf("\nTest Name :%v \n  got => %v \n  want => %v \n", name, got, tt.err)
			}
		})
	}
}

func TestPostRequest(t *testing.T) {
	testcases := map[string]*struct {
		fakeHandler utiltesting.FakeHandler
		namespace   string
		err         error
		chkbody     bool
		usefakeurl  bool
		invalidurl  string
	}{
		"Invalid Url with no namespace": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(""),
				T:            t,
			},
			namespace:  "",
			err:        errors.New("Invalid URL"),
			chkbody:    false,
			usefakeurl: false,
			invalidurl: "",
		},
		"No response with no namespace": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "",
			err:        errors.New("Server status error: Bad Request"),
			chkbody:    false,
			usefakeurl: true,
		},
		"Url not reachable": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "",
			err:        errors.New(`Post "invalid": unsupported protocol scheme ""`),
			chkbody:    false,
			usefakeurl: false,
			invalidurl: "invalid",
		},
		"not parsable url": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "",
			err:        errors.New(`parse "*:": first path segment in URL cannot contain colon`),
			chkbody:    false,
			usefakeurl: false,
			invalidurl: "*:",
		},
		"chk body is enabled": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "",
			err:        errors.New(response),
			chkbody:    true,
			usefakeurl: true,
		},
		"chk body is enabled with namespace": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(response),
				T:            t,
			},
			namespace:  "default",
			err:        errors.New(response),
			chkbody:    true,
			usefakeurl: true,
		},
	}

	for name, tt := range testcases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			url := ""
			if tt.usefakeurl {
				url = server.URL
			} else {
				url = tt.invalidurl
			}
			defer server.Close()
			r := []byte{23, 4, 23}
			_, got := postRequest(url, r, tt.namespace, tt.chkbody)
			if got.Error() != tt.err.Error() {
				t.Fatalf("\nTest Name :%v \n  got => %v \n  want => %v \n", name, got, tt.err)
			}
		})
	}
}
