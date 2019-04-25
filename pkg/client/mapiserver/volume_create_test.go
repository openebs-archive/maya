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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

func TestCreateVolume(t *testing.T) {
	tests := map[string]*struct {
		volumeName  string
		size        string
		namespace   string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			volumeName: "qwewretrytu",
			size:       "1G",
			namespace:  "appspace",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 200,
				//ResponseBody: "Volume 'qwewretrytu' deleted Successfully",
				T: t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			volumeName: "12324rty653423",
			size:       "1G",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 404,
				//ResponseBody: "Volume '12324rty653423' deleted Successfully",

				T: t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(404)),
			addr: "MAPI_ADDR",
		},
		"VolumeNameMissing": {
			volumeName: "",
			size:       "1G",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 400,
				//ResponseBody: "Volume name is missing",
				T: t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			got := CreateVolume(tt.volumeName, tt.size, tt.namespace)

			if !reflect.DeepEqual(got, tt.err) {
				t.Fatalf("CreateVolume(%v, %v) => got %v, want %v ", tt.volumeName, tt.size, got, tt.err)
			}
		})
	}
}
