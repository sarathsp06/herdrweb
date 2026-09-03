package server

import (
	"sync"
	"testing"
)

// TestBrowserSendCloseRace reproduces the "send on closed channel" panic: a
// browser is removed (its send channel closed) while in-flight handleCall
// goroutines still reply. trySend must be safe against the concurrent close.
func TestBrowserSendCloseRace(t *testing.T) {
	closeGuarded := func(b *browser) {
		b.mu.Lock()
		if !b.closed {
			b.closed = true
			close(b.send)
		}
		b.mu.Unlock()
	}

	for range 50 {
		b := &browser{send: make(chan []byte, 8)}

		drained := make(chan struct{})
		go func() {
			for range b.send {
			}
			close(drained)
		}()

		var wg sync.WaitGroup
		for range 20 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 100 {
					b.trySend([]byte("x"))
				}
			}()
		}

		// Close concurrently with the senders, via the same guarded path
		// removeBrowser uses.
		go closeGuarded(b)

		wg.Wait()
		closeGuarded(b) // ensure closed so the drainer exits
		<-drained
	}
}
