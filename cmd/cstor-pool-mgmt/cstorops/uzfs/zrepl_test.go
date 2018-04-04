package uzfs

import (
	"testing"
	"time"
)

// TestCheckForZrepl tests if zrepl is running or not with timeout
func TestCheckForZrepl(t *testing.T) {
	done := make(chan bool)
	go func() {
		CheckForZrepl()
		done <- true
	}()
	select {
	case <-time.After(12 * time.Second):
		t.Fatalf("Timeout error")
	case <-done:
	}
}
