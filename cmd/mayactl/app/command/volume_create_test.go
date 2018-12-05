// Copyright © 2017 The OpenEBS Authors
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

package command

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

func TestIsVolumeExist(t *testing.T) {
	response := `{"items":[{"metadata":{"annotations":{"vsm.openebs.io/replica-status":"Pending,Running,Pending","openebs.io/jiva-replica-status":"Pending,Running,Pending","deployment.kubernetes.io/revision":"1","openebs.io/volume-monitor":"false","openebs.io/volume-type":"jiva","openebs.io/jiva-replica-count":"3","vsm.openebs.io/replica-ips":"nil,172.17.0.7,nil","openebs.io/jiva-replica-ips":"nil,172.17.0.7,nil","vsm.openebs.io/targetportals":"10.106.224.86:3260","vsm.openebs.io/cluster-ips":"10.106.224.86","openebs.io/storage-pool":"default","openebs.io/capacity":"5G","vsm.openebs.io/controller-ips":"172.17.0.3","openebs.io/jiva-controller-status":"Running","openebs.io/replica-container-status":"Running","openebs.io/jiva-iqn":"iqn.2016-09.com.openebs.jiva:test5","vsm.openebs.io/volume-size":"5G","openebs.io/controller-container-status":"Running","openebs.io/jiva-controller-cluster-ip":"10.106.224.86","vsm.openebs.io/iqn":"iqn.2016-09.com.openebs.jiva:test5","vsm.openebs.io/replica-count":"3","openebs.io/jiva-controller-ips":"172.17.0.3","vsm.openebs.io/controller-status":"Running","openebs.io/jiva-target-portal":"10.106.224.86:3260"},"creationTimestamp":null,"labels":{},"name":"test5"},"status":{"Message":"","Phase":"Running","Reason":""}}],"metadata":{}}`
	tests := map[string]*struct {
		volname        string
		fakeHandler    utiltesting.FakeHandler
		addr           string
		expectedOutput error
	}{
		"Creating new volume with volume name test1": {
			volname: "test1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(response),
				T:            t,
			},
			addr:           "MAPI_ADDR",
			expectedOutput: nil,
		},
		"Getting status error 400": {
			volname: "test2",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string("HTTP Error 400 : Bad Request"),
				T:            t,
			},
			addr:           "MAPI_ADDR",
			expectedOutput: fmt.Errorf("HTTP Error 400 : Bad Request"),
		},
		"Getting status error 404": {
			volname: "test3",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   404,
				ResponseBody: string("HTTP Error 404 : Not Found"),
				T:            t,
			},
			addr:           "MAPI_ADDR",
			expectedOutput: fmt.Errorf("HTTP Error 404 : Not Found"),
		},
		"Creating volume which already exist": {
			volname: "test5",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(response),
				T:            t,
			},
			addr:           "MAPI_ADDR",
			expectedOutput: fmt.Errorf("Volume creation failed : Volume %v already exist ", "test5"),
		},
	}

	for name, c := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&c.fakeHandler)
			defer os.Unsetenv(c.addr)
			defer server.Close()
			os.Setenv(c.addr, server.URL)
			err := IsVolumeExist(c.volname)
			if (err != nil && c.expectedOutput != nil) && string(err.Error()) != string(c.expectedOutput.Error()) {
				t.Errorf("\nExpected output was : %v \nbut got : %v", c.expectedOutput, err)
			} else if (err != nil && c.expectedOutput == nil) || (err == nil && c.expectedOutput != nil) {
				t.Errorf("\nExpected output was : %v \nbut got : %v", c.expectedOutput, err)
			}
		})
	}
}
