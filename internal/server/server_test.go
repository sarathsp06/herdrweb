package server

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sarathsp06/herdrweb/internal/herdr"
	"github.com/sarathsp06/herdrweb/internal/protocol"
)

// TestDispatchBoundsConcurrency proves the per-connection semaphore caps the
// number of concurrent in-flight handlers at maxInflightPerConn, applying
// backpressure to the firing loop instead of spawning unbounded goroutines.
func TestDispatchBoundsConcurrency(t *testing.T) {
	h := NewHub(nil, "", "test", nil)
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

// TestApplySnapshotSkipsBroadcastWhenUnchanged proves the poller's fixed
// 1.5s cadence does not force a broadcast (and the reactive churn it causes
// in every connected browser) when Herdr's state has not actually moved.
func TestApplySnapshotSkipsBroadcastWhenUnchanged(t *testing.T) {
	h := NewHub(nil, "", "test", nil)
	b := &browser{send: make(chan []byte, 4)}
	h.addBrowser(b)

	ctx := context.Background()
	snap := protocol.Snapshot{Spaces: []protocol.Space{{ID: "w1", Label: "one"}}}
	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}

	h.applySnapshot(ctx, snap, data)
	h.applySnapshot(ctx, snap, data)
	h.applySnapshot(ctx, snap, data)

	if got := len(b.send); got != 1 {
		t.Fatalf("broadcasts received = %d, want 1 (repeat identical snapshots must not re-broadcast)", got)
	}

	changed := protocol.Snapshot{Spaces: []protocol.Space{{ID: "w1", Label: "renamed"}}}
	changedData, err := json.Marshal(changed)
	if err != nil {
		t.Fatal(err)
	}
	h.applySnapshot(ctx, changed, changedData)

	if got := len(b.send); got != 2 {
		t.Fatalf("broadcasts received = %d, want 2 after a real change", got)
	}
}

// fakeHerdrServer answers session.snapshot requests on one persistent
// connection (matching herdr.Client's multiplexed connection), each with a
// per-arrival-order controllable delay and payload - lets a test force an
// earlier call to resolve after a later one. respond is invoked synchronously
// right after a request line is read and numbered (before its delayed
// response is even scheduled), so a test can use it to signal "the wire has
// this request" instead of guessing at a sleep duration.
func fakeHerdrServer(t *testing.T, respond func(callNum int) (delay time.Duration, workspaceLabel string)) string {
	t.Helper()
	// A short, test-name-independent dir: t.TempDir() embeds the (long) test
	// name and can overflow the unix sockaddr_un path limit on macOS.
	dir, err := os.MkdirTemp("", "hsock")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	sock := filepath.Join(dir, "h.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var writeMu sync.Mutex
		r := bufio.NewReader(conn)
		callNum := 0
		for {
			line, err := r.ReadBytes('\n')
			if err != nil {
				return
			}
			var req struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal(line, &req)
			callNum++
			n := callNum
			delay, label := respond(n)
			go func() {
				time.Sleep(delay)
				b, _ := json.Marshal(map[string]any{"id": req.ID, "result": map[string]any{
					"snapshot": map[string]any{"workspaces": []any{map[string]any{"workspace_id": "w1", "label": label}}},
				}})
				writeMu.Lock()
				conn.Write(append(b, '\n'))
				writeMu.Unlock()
			}()
		}
	}()
	return sock
}

// TestRefreshSerializesAgainstOutOfOrderCompletion proves refresh() cannot
// regress the cached snapshot to stale data when an earlier-started Herdr
// Snapshot RPC is slower than a later one. Without serializing, both RPCs
// fire concurrently, the fast/later one applies first, and the slow/earlier
// one then overwrites it with older data on completion - a momentary
// backward flash of the UI. Serialized, the second refresh's RPC cannot even
// start until the first's completes, so completion order matches start
// order and the cache always ends on the freshest data.
func TestRefreshSerializesAgainstOutOfOrderCompletion(t *testing.T) {
	firstArrived := make(chan struct{})
	sock := fakeHerdrServer(t, func(callNum int) (time.Duration, string) {
		if callNum == 1 {
			close(firstArrived)
			return 150 * time.Millisecond, "first"
		}
		return 10 * time.Millisecond, "second"
	})
	h := NewHub(herdr.New(sock), "", "test", nil)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := h.refresh(context.Background()); err != nil {
			t.Errorf("first refresh: %v", err)
		}
	}()
	select {
	case <-firstArrived: // the first refresh's Snapshot request has reached the wire
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the first refresh's request to reach the fake server")
	}
	go func() {
		defer wg.Done()
		if err := h.refresh(context.Background()); err != nil {
			t.Errorf("second refresh: %v", err)
		}
	}()
	wg.Wait()

	var got protocol.Snapshot
	if err := json.Unmarshal(h.snapshot, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Spaces) != 1 || got.Spaces[0].Label != "second" {
		t.Fatalf("cached snapshot label = %+v, want the later refresh's data (\"second\")", got.Spaces)
	}
}
