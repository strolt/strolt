package api

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOperationTrackerWaitsAndStopsAdmission(t *testing.T) {
	tracker := &operationTracker{mu: sync.Mutex{}, wg: sync.WaitGroup{}, stopping: false}
	release, ok := tracker.begin()
	require.True(t, ok)

	stopped := make(chan struct{})
	go func() {
		tracker.stopAndWait()
		close(stopped)
	}()

	require.Eventually(t, func() bool {
		tracker.mu.Lock()
		defer tracker.mu.Unlock()
		return tracker.stopping
	}, time.Second, time.Millisecond)
	_, ok = tracker.begin()
	require.False(t, ok)

	release()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not wait for the active operation")
	}
}
