package herdr

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// muxServer accepts connections on a fresh unix socket, counting accepts, and
// hands each to handle. The accept counter proves the client reuses one RPC
// connection across many Calls.
func muxServer(t *testing.T, handle func(conn net.Conn, accepts *int64)) (sock string, accepts *int64) {
	t.Helper()
	sock = filepath.Join(t.TempDir(), "h.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	accepts = new(int64)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			atomic.AddInt64(accepts, 1)
			go handle(conn, accepts)
		}
	}()
	return sock, accepts
}

// echoConn serves many requests on one connection, replying {id,result:{echo:method}}.
// A per-request "delay" param lets a test force out-of-order replies; writes are
// serialized so concurrent replies don't interleave on the wire.
func echoConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	var wmu sync.Mutex
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return
		}
		var req struct {
			ID     string `json:"id"`
			Method string `json:"method"`
			Params struct {
				Delay int `json:"delay"`
			} `json:"params"`
		}
		if json.Unmarshal(line, &req) != nil {
			continue
		}
		go func(id, method string, delay int) {
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
			wmu.Lock()
			defer wmu.Unlock()
			_, _ = fmt.Fprintf(conn, "{\"id\":%q,\"result\":{\"echo\":%q}}\n", id, method)
		}(req.ID, req.Method, req.Params.Delay)
	}
}

func echoOf(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var out struct {
		Echo string `json:"echo"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("bad result %s: %v", raw, err)
	}
	return out.Echo
}

// Concurrent calls over one shared connection each get their id-correlated
// reply, and only a single connection is opened.
func TestMuxConcurrentCallsCorrelate(t *testing.T) {
	sock, accepts := muxServer(t, func(c net.Conn, _ *int64) { echoConn(c) })
	c := New(sock)
	ctx := context.Background()

	const n = 50
	var wg sync.WaitGroup
	got := make([]string, n)
	errs := make([]error, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			raw, err := c.Call(ctx, fmt.Sprintf("m%d", i), nil)
			if err != nil {
				errs[i] = err
				return
			}
			got[i] = echoOf(t, raw)
		}(i)
	}
	wg.Wait()

	for i := range n {
		if errs[i] != nil {
			t.Fatalf("call %d: %v", i, errs[i])
		}
		if want := fmt.Sprintf("m%d", i); got[i] != want {
			t.Fatalf("call %d correlated to %q, want %q", i, got[i], want)
		}
	}
	if a := atomic.LoadInt64(accepts); a != 1 {
		t.Fatalf("opened %d connections, want 1 (not multiplexed)", a)
	}
}

// Replies arriving in reverse order still correlate to the right caller.
func TestMuxOutOfOrderReplies(t *testing.T) {
	sock, _ := muxServer(t, func(c net.Conn, _ *int64) { echoConn(c) })
	c := New(sock)
	ctx := context.Background()

	// Larger delay on the earlier call => it replies last.
	delays := []int{60, 30, 0}
	var wg sync.WaitGroup
	got := make([]string, len(delays))
	for i, d := range delays {
		wg.Add(1)
		go func(i, d int) {
			defer wg.Done()
			raw, err := c.Call(ctx, fmt.Sprintf("m%d", i), map[string]any{"delay": d})
			if err != nil {
				t.Errorf("call %d: %v", i, err)
				return
			}
			got[i] = echoOf(t, raw)
		}(i, d)
	}
	wg.Wait()
	for i := range delays {
		if want := fmt.Sprintf("m%d", i); got[i] != want {
			t.Fatalf("call %d correlated to %q, want %q", i, got[i], want)
		}
	}
}

// A cancelled call returns promptly; its late reply is discarded and the shared
// reader keeps serving subsequent calls.
func TestMuxContextCancelDoesNotWedge(t *testing.T) {
	sock, accepts := muxServer(t, func(c net.Conn, _ *int64) { echoConn(c) })
	c := New(sock)

	ctx1, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if _, err := c.Call(ctx1, "slow", map[string]any{"delay": 400}); err == nil {
		t.Fatal("cancelled call returned nil error")
	}

	raw, err := c.Call(context.Background(), "after", nil)
	if err != nil {
		t.Fatalf("call after cancel failed (reader wedged?): %v", err)
	}
	if got := echoOf(t, raw); got != "after" {
		t.Fatalf("got %q, want %q", got, "after")
	}
	if a := atomic.LoadInt64(accepts); a != 1 {
		t.Fatalf("reconnected unexpectedly: %d accepts", a)
	}
}

// A dropped connection fails the in-flight call, and the next call transparently
// reconnects on a fresh connection.
func TestMuxReconnectAfterDrop(t *testing.T) {
	sock, accepts := muxServer(t, func(c net.Conn, acc *int64) {
		if atomic.LoadInt64(acc) == 1 { // first connection: read one line, then drop
			r := bufio.NewReader(c)
			_, _ = r.ReadBytes('\n')
			_ = c.Close()
			return
		}
		echoConn(c)
	})
	c := New(sock)
	ctx := context.Background()

	if _, err := c.Call(ctx, "willdrop", nil); err == nil {
		t.Fatal("in-flight call over a dropped connection returned nil error")
	}
	raw, err := c.Call(ctx, "reconnected", nil)
	if err != nil {
		t.Fatalf("call after drop failed to reconnect: %v", err)
	}
	if got := echoOf(t, raw); got != "reconnected" {
		t.Fatalf("got %q, want %q", got, "reconnected")
	}
	if a := atomic.LoadInt64(accepts); a != 2 {
		t.Fatalf("accepts = %d, want 2 (lazy reconnect)", a)
	}
}
