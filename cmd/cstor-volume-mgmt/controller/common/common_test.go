// Copyright © 2017-2019 The OpenEBS Authors
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

package common

import (
	"testing"
	"time"

	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
)

// TestCheckForCStorVolumeCRD validates if CStorVolume CRD operations
// can be done.
func TestCheckForCStorVolumeCRD(t *testing.T) {
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	done := make(chan bool)
	defer close(done)
	go func(done chan bool) {
		//CheckForCStorVolumeCR tries to find the volume CR and if is is not found
		// it will wait for 10 seconds and continue trying in the loop.
		// as we are already passing the fake CR, it has to find it immediately
		// if not, it means the code is not working properly
		CheckForCStorVolumeCRD(fakeOpenebsClient)
		//this below line will get executed only when CheckForCStorVolumeCR has
		//found the CR. Otherwise, the function will not return and we timeout
		// in the below select block and fail the testcase.
		done <- true
	}(done)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("Timeout - CStorVolume is unknown")
	case <-done:
	}
}
