package server

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestDispatchBoundsConcurrency proves the per-connection semaphore caps the
// number of concurrent in-flight handlers at maxInflightPerConn, applying
// backpressure to the firing loop instead of spawning unbounded goroutines.
func TestDispatchBoundsConcurrency(t *testing.T) {
	h := NewHub(nil, "", "test")
	const limit = maxInflightPerConn
	const total = limit * 3

	sem := make(chan struct{}, limit)
	var inflight, peak int64
	started := make(chan struct{}, total)
	release := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(total)

	run := func() {
		defer wg.Done()
		n := atomic.AddInt64(&inflight, 1)
		for {
			p := atomic.LoadInt64(&peak)
			if n <= p || atomic.CompareAndSwapInt64(&peak, p, n) {
				break
			}
		}
		started <- struct{}{}
		<-release
		atomic.AddInt64(&inflight, -1)
	}

	go func() {
		for range total {
			h.dispatch(sem, run)
		}
	}()

	// Exactly `limit` handlers should run and then block; the firing loop is
	// stalled acquiring the (limit+1)th token.
	for range limit {
		<-started
	}
	time.Sleep(20 * time.Millisecond)
	if got := atomic.LoadInt64(&inflight); got != limit {
		t.Fatalf("in-flight = %d, want %d (semaphore not bounding)", got, limit)
	}
	if got := atomic.LoadInt64(&peak); got != limit {
		t.Fatalf("peak = %d, want %d", got, limit)
	}

	close(release)
	wg.Wait()
	if got := atomic.LoadInt64(&peak); got != limit {
		t.Fatalf("peak after drain = %d, want %d (exceeded the cap)", got, limit)
	}
}
