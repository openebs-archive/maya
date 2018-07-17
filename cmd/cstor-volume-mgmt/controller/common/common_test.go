package common

import (
	"testing"
	"time"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
)

// TestCheckForCStorVolumeCR validates if CStorVolume CR operations
// can be done.
func TestCheckForCStorVolumeCR(t *testing.T) {
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	done := make(chan bool)
	defer close(done)
	go func(done chan bool) {
		//CheckForCStorVolumeCR tries to find the volume CR and if is is not found
		// it will wait for 10 seconds and continue trying in the loop.
		// as we are already passing the fake CR, it has to find it immediately
		// if not, it means the code is not working properly
		CheckForCStorVolumeCR(fakeOpenebsClient)
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
