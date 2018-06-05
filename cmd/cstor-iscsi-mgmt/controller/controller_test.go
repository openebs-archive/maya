package controller

import (
	"testing"
	"time"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	//github.com/openebs/maya/vendor/k8s.io/client-go/kubernetes/fake
)

// TestCheckForCStorIscsiCRD validates if CStorIscsi CRD operations
// can be done.
func TestCheckForCStorIscsiCRD(t *testing.T) {
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	done := make(chan bool)

	go func(done chan bool) {
		checkForCStorIscsiCRD(fakeOpenebsClient)
		done <- true
	}(done)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("Timeout - CStorIscsi is unknown")
	case <-done:
	}
}

