package herdr

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// fakeServer answers one request per connection, mimicking Herdr's socket.
func fakeServer(t *testing.T) string {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "h.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleConn(conn)
		}
	}()
	return sock
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	line, err := r.ReadBytes('\n')
	if err != nil {
		return
	}
	var req struct {
		ID     string          `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	_ = json.Unmarshal(line, &req)
	switch req.Method {
	case "ping":
		writeLine(conn, map[string]any{"id": req.ID, "result": map[string]any{"type": "pong"}})
	case "session.snapshot":
		writeLine(conn, map[string]any{"id": req.ID, "result": map[string]any{
			"snapshot": map[string]any{"workspaces": []any{map[string]any{"workspace_id": "w1", "label": "demo"}}},
		}})
	case "events.subscribe":
		writeLine(conn, map[string]any{"id": req.ID, "result": map[string]any{"ok": true}})
		writeLine(conn, map[string]any{"event": "pane.updated", "data": map[string]any{"pane_id": "w1:p1"}})
		time.Sleep(50 * time.Millisecond)
	default:
		writeLine(conn, map[string]any{"id": req.ID, "error": "unknown"})
	}
}

func writeLine(conn net.Conn, v any) {
	b, _ := json.Marshal(v)
	conn.Write(append(b, '\n'))
}

func TestCallCorrelatesResponse(t *testing.T) {
	c := New(fakeServer(t))
	res, err := c.Call(context.Background(), "ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(res, &got) != nil || got.Type != "pong" {
		t.Fatalf("bad ping result: %s", res)
	}
}

func TestSnapshotUnwraps(t *testing.T) {
	c := New(fakeServer(t))
	var out struct {
		Workspaces []struct {
			WorkspaceID string `json:"workspace_id"`
		} `json:"workspaces"`
	}
	if err := c.Snapshot(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Workspaces) != 1 || out.Workspaces[0].WorkspaceID != "w1" {
		t.Fatalf("snapshot not unwrapped: %+v", out)
	}
}

func TestSubscribeStreamsEvents(t *testing.T) {
	c := New(fakeServer(t))
	ctx, cancel := context.WithCancel(context.Background())
	var mu sync.Mutex
	var events []string
	done := make(chan struct{})
	go func() {
		_ = c.Subscribe(ctx, []string{"pane.updated"}, func(e Event) {
			mu.Lock()
			events = append(events, e.Event)
			mu.Unlock()
			cancel()
		})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("subscribe did not return")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(events) == 0 || events[0] != "pane.updated" {
		t.Fatalf("no events streamed: %+v", events)
	}
}
