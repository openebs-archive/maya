package common

import (
	"testing"
	"time"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
)

// TestCheckForCStorVolumeCRD validates if CStorVolume CRD operations
// can be done.
func TestCheckForCStorVolumeCRD(t *testing.T) {
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	done := make(chan bool)
	defer close(done)
	go func(done chan bool) {
		CheckForCStorVolumeCRD(fakeOpenebsClient)
		done <- true
	}(done)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("Timeout - CStorVolume is unknown")
	case <-done:
	}
}
