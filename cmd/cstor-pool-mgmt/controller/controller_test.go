package controller

import (
	"testing"
	"time"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	//github.com/openebs/maya/vendor/k8s.io/client-go/kubernetes/fake
)

// TestCheckForCStorPoolCRD validates if CStorPool CRD operations
// can be done.
func TestCheckForCStorPoolCRD(t *testing.T) {
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	done := make(chan bool)

	go func(done chan bool) {
		checkForCStorPoolCRD(fakeOpenebsClient)
		done <- true
	}(done)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("Timeout - CStorPool is unknown")
	case <-done:
	}
}

// TestCheckForCStorVolumeReplicaCRD validates if CStorVolumeReplica CRD
// operations can be done.
func TestCheckForCStorVolumeReplicaCRD(t *testing.T) {
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	done := make(chan bool)

	go func(done chan bool) {
		checkForCStorVolumeReplicaCRD(fakeOpenebsClient)
		done <- true
	}(done)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("Timeout - CStorVolumeReplica is unknown")
	case <-done:
	}
}
