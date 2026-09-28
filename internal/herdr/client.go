// Package herdr is a minimal client for the Herdr local socket API: newline
// delimited JSON, id-correlated request/response, plus a streaming event
// subscription. One connection per request (matching the CLI); a dedicated
// long-lived connection carries the event subscription.
package herdr

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"sync/atomic"
	"time"
)

// DefaultSocketPath returns the default session's socket in ConfigDir.
func DefaultSocketPath() string {
	return filepath.Join(ConfigDir(), "herdr.sock")
}

// SessionSocketPath returns the socket of a named Herdr session
// (`herdr --session <name>`); an empty name is the default session.
func SessionSocketPath(name string) string {
	if name == "" {
		return DefaultSocketPath()
	}
	return filepath.Join(ConfigDir(), "sessions", name, "herdr.sock")
}

// Client dials the Herdr socket (a named pipe on Windows).
type Client struct {
	SocketPath  string
	DialTimeout time.Duration
	seq         int64
}

// New returns a client for the given socket path (empty = default).
func New(socketPath string) *Client {
	if socketPath == "" {
		socketPath = DefaultSocketPath()
	}
	return &Client{SocketPath: socketPath, DialTimeout: 5 * time.Second}
}

type request struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

type response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
}

func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	return dialSocket(ctx, c.SocketPath, c.DialTimeout)
}

func (c *Client) nextID(prefix string) string {
	return fmt.Sprintf("%s:%d", prefix, atomic.AddInt64(&c.seq, 1))
}

// Call sends one request on its own connection and returns its result.
//
// Herdr answers exactly one request per connection and then closes it, as the
// CLI assumes, so connections cannot be shared. A second request written to a
// busy connection waits behind the first and then fails when Herdr closes it:
// behind an agent.prompt that waits for the agent, every pane.read and
// snapshot the bridge made stalled until the agent went idle.
func (c *Client) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if params == nil {
		params = map[string]any{}
	}
	line, err := json.Marshal(request{ID: c.nextID("req"), Method: method, Params: params})
	if err != nil {
		return nil, err
	}
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// Closing the connection is what unblocks the read when ctx ends.
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	_, err = conn.Write(append(line, '\n'))
	var raw []byte
	if err == nil {
		raw, err = readLine(bufio.NewReaderSize(conn, 1<<20))
	}
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	var resp response
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if len(resp.Error) > 0 && string(resp.Error) != "null" {
		return nil, fmt.Errorf("herdr %s error: %s", method, string(resp.Error))
	}
	return resp.Result, nil
}

// Snapshot fetches session.snapshot and unmarshals it (handling both the bare
// result and the {snapshot:{...}} wrapper the CLI uses).
func (c *Client) Snapshot(ctx context.Context, out any) error {
	res, err := c.Call(ctx, "session.snapshot", map[string]any{})
	if err != nil {
		return err
	}
	var wrap struct {
		Snapshot json.RawMessage `json:"snapshot"`
	}
	if json.Unmarshal(res, &wrap) == nil && len(wrap.Snapshot) > 0 {
		return json.Unmarshal(wrap.Snapshot, out)
	}
	return json.Unmarshal(res, out)
}

// Event is a subscription envelope: {event, data}.
type Event struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// AllEventTypes are the global (no pane_id) resource events the bridge
// subscribes to. Pane-scoped events (pane.agent_status_changed,
// pane.output_matched, pane.scroll_changed) require a pane_id and are not
// global subscriptions; the hub's periodic re-snapshot catches agent-status
// transitions instead.
var AllEventTypes = []string{
	"workspace.created", "workspace.updated", "workspace.renamed", "workspace.moved",
	"workspace.reordered", "workspace.closed", "workspace.focused",
	"worktree.created", "worktree.opened", "worktree.removed",
	"tab.created", "tab.closed", "tab.focused", "tab.renamed", "tab.moved",
	"pane.created", "pane.closed", "pane.updated", "pane.focused", "pane.moved",
	"pane.exited", "pane.agent_detected",
	"layout.updated",
}

// Subscribe opens a long-lived connection, subscribes to the given event types
// (empty = AllEventTypes) and calls onEvent for each envelope until ctx is done
// or the connection drops. It returns the terminating error (never nil unless
// ctx cancelled cleanly).
func (c *Client) Subscribe(ctx context.Context, types []string, onEvent func(Event)) error {
	if len(types) == 0 {
		types = AllEventTypes
	}
	conn, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	subs := make([]map[string]string, 0, len(types))
	for _, t := range types {
		subs = append(subs, map[string]string{"type": t})
	}
	req := request{ID: c.nextID("sub"), Method: "events.subscribe", Params: map[string]any{"subscriptions": subs}}
	line, _ := json.Marshal(req)
	if _, err := conn.Write(append(line, '\n')); err != nil {
		return err
	}

	r := bufio.NewReaderSize(conn, 1<<20)
	for {
		raw, err := readLine(r)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		var ev Event
		if err := json.Unmarshal(raw, &ev); err == nil && ev.Event != "" {
			onEvent(ev)
			continue
		}
		// The subscribe ack (a response with our id) is ignored.
	}
}

func readLine(r *bufio.Reader) ([]byte, error) {
	var buf []byte
	for {
		chunk, err := r.ReadBytes('\n')
		buf = append(buf, chunk...)
		if err == nil {
			return buf[:len(buf)-1], nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		return nil, err
	}
}
