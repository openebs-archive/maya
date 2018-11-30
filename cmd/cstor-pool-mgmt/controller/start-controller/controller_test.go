/*
Copyright 2018 The OpenEBS Authors.

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

package startcontroller

import (
	"os"
	"testing"
	"time"
)

func TestGetSyncInterval(t *testing.T) {
	tests := map[string]struct {
		resyncInterval string
		expectedResult time.Duration
	}{
		"resync environment variable is missing": {
			resyncInterval: "",
			expectedResult: 30 * time.Second,
		},
		"resync environment variable is non numeric": {
			resyncInterval: "sfdgg",
			expectedResult: 30 * time.Second,
		},
		"resync interval is set to zero(0)": {
			resyncInterval: "0",
			expectedResult: 30 * time.Second,
		},
		"resync interval is correct": {
			resyncInterval: "13",
			expectedResult: 13 * time.Second,
		},
	}

	for name, mock := range tests {
		os.Setenv("RESYNC_INTERVAL", mock.resyncInterval)
		defer os.Unsetenv("RESYNC_INTERVAL")
		t.Run(name, func(t *testing.T) {
			interval := getSyncInterval()
			if interval != mock.expectedResult {
				t.Errorf("unable to get correct resync interval, expected: %v got %v", mock.expectedResult, interval)
			}
		})
	}
}
