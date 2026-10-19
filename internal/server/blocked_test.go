package server

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/sarathsp06/herdrweb/internal/herdr"
	"github.com/sarathsp06/herdrweb/internal/protocol"
)

// methodHerdr is a fake Herdr that answers each request by method name.
func methodHerdr(t *testing.T, answer func(method string, params map[string]any) any) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "hsock")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	sock := filepath.Join(dir, "h.sock")
	ln, err := herdr.Listen(sock)
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
			go func() {
				defer conn.Close()
				var mu sync.Mutex
				r := bufio.NewReader(conn)
				for {
					line, err := r.ReadBytes('\n')
					if err != nil {
						return
					}
					var req struct {
						ID     string         `json:"id"`
						Method string         `json:"method"`
						Params map[string]any `json:"params"`
					}
					_ = json.Unmarshal(line, &req)
					b, _ := json.Marshal(map[string]any{"id": req.ID, "result": answer(req.Method, req.Params)})
					mu.Lock()
					conn.Write(append(b, '\n'))
					mu.Unlock()
				}
			}()
		}
	}()
	return sock
}

func paneOf(t *testing.T, h *Hub, id string) protocol.Pane {
	t.Helper()
	var snap protocol.Snapshot
	if err := json.Unmarshal(h.snapshot, &snap); err != nil {
		t.Fatal(err)
	}
	for _, sp := range snap.Spaces {
		for _, tb := range sp.Tabs {
			for _, p := range tb.Panes {
				if p.ID == id {
					return p
				}
			}
		}
	}
	t.Fatalf("pane %s not in snapshot", id)
	return protocol.Pane{}
}

// TestRefreshAnnotatesBlockedPanes proves a blocked agent pane carries the
// question read from its detection screen and a stable blockedSince, and that
// both clear once the agent moves on.
func TestRefreshAnnotatesBlockedPanes(t *testing.T) {
	var mu sync.Mutex
	status := "blocked"
	var readSource any
	sock := methodHerdr(t, func(method string, params map[string]any) any {
		mu.Lock()
		defer mu.Unlock()
		switch method {
		case "session.snapshot":
			return map[string]any{"snapshot": map[string]any{
				"workspaces": []any{map[string]any{"workspace_id": "w1", "label": "api"}},
				"tabs":       []any{map[string]any{"tab_id": "w1:t1", "workspace_id": "w1", "label": "agents"}},
				"panes": []any{map[string]any{
					"pane_id": "w1:p1", "tab_id": "w1:t1", "workspace_id": "w1", "agent": "codex", "agent_status": status,
				}},
			}}
		case "pane.read":
			readSource = params["source"]
			return map[string]any{"read": map[string]any{"text": "? Apply patch to main.go? (y/n)\n"}}
		}
		return map[string]any{}
	})
	h := NewHub(herdr.New(sock), "", "test", nil)
	ctx := context.Background()

	if err := h.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	p := paneOf(t, h, "w1:p1")
	if p.Prompt == nil || p.Prompt.Question != "Apply patch to main.go?" || len(p.Prompt.Options) != 2 {
		t.Fatalf("prompt = %+v", p.Prompt)
	}
	if readSource != "detection" {
		t.Fatalf("pane.read source = %v, want detection", readSource)
	}
	since := p.BlockedSince
	if since == 0 {
		t.Fatal("blockedSince not set")
	}

	if err := h.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if got := paneOf(t, h, "w1:p1").BlockedSince; got != since {
		t.Fatalf("blockedSince moved from %d to %d while still blocked", since, got)
	}

	mu.Lock()
	status = "working"
	mu.Unlock()
	if err := h.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if p := paneOf(t, h, "w1:p1"); p.Prompt != nil || p.BlockedSince != 0 {
		t.Fatalf("working pane still annotated: %+v", p)
	}
	if len(h.blockedAt) != 0 {
		t.Fatalf("blockedAt not pruned: %v", h.blockedAt)
	}
}
