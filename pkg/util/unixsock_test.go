/*
Copyright 2018 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"testing"
)

func TestIsResponseEOD(t *testing.T) {
	tests := map[string]struct {
		resp           []string
		cmd            string
		expectedresult bool
	}{
		"empty response": {
			resp:           []string{},
			cmd:            "STATUS",
			expectedresult: false,
		},
		"only header line in response": {
			resp:           []string{"iSCSI Target Controller version istgt:0.5.20121028:08:47:01:Aug 28 2018 on  from\r\n"},
			cmd:            "STATUS",
			expectedresult: false,
		},
		"header line along with incomplete output in response": {
			resp:           []string{"iSCSI Target Controller version istgt:0.5.20121028:08:47:01:Aug 28 2018 on  from\r\n", "STATUS iqn FAKE\r\n"},
			cmd:            "STATUS",
			expectedresult: false,
		},
		"header line with complete output in response": {
			resp:           []string{"iSCSI Target Controller version istgt:0.5.20121028:08:47:01:Aug 28 2018 on  from\r\n", "STATUS iqn FAKE\r\n", "OK STATUS\r\n"},
			cmd:            "STATUS",
			expectedresult: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsResponseEOD(mock.resp, mock.cmd)
			if mock.expectedresult != result {
				t.Fatalf("failed test '%s' - expected result '%v': actual result '%v'", name, mock.expectedresult, result)
			}
		})
	}
}
