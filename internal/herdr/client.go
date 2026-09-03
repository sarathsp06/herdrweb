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
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultSocketPath returns ~/.config/herdr/herdr.sock.
func DefaultSocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr", "herdr.sock")
}

// Client dials the Herdr Unix socket.
type Client struct {
	SocketPath  string
	DialTimeout time.Duration
	seq         int64

	// mu guards the persistent RPC connection, its pending waiters, and the
	// generation counter. Writes are serialized under mu; a single readLoop per
	// connection dispatches responses to waiters by id.
	mu      sync.Mutex
	conn    net.Conn
	pending map[string]chan call
	gen     uint64
}

// call carries a correlated reply (or a transport error) to a waiting Call.
type call struct {
	resp response
	err  error
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
	d := net.Dialer{Timeout: c.DialTimeout}
	return d.DialContext(ctx, "unix", c.SocketPath)
}

func (c *Client) nextID(prefix string) string {
	return fmt.Sprintf("%s:%d", prefix, atomic.AddInt64(&c.seq, 1))
}

// errClosed fails in-flight waiters when the persistent connection drops.
var errClosed = errors.New("herdr: connection closed")

// Call sends one request over the persistent multiplexed connection and returns
// its id-correlated result. Concurrent calls share one connection; a single
// reader dispatches replies by id. The connection is re-dialed lazily after a
// drop; in-flight calls at drop time fail with an error.
func (c *Client) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if params == nil {
		params = map[string]any{}
	}
	id := c.nextID("req")
	line, err := json.Marshal(request{ID: id, Method: method, Params: params})
	if err != nil {
		return nil, err
	}

	ch := make(chan call, 1)
	c.mu.Lock()
	if c.conn == nil {
		if err := c.connectLocked(ctx); err != nil {
			c.mu.Unlock()
			return nil, err
		}
	}
	conn, gen := c.conn, c.gen
	c.pending[id] = ch
	_, werr := conn.Write(append(line, '\n'))
	c.mu.Unlock()
	if werr != nil {
		c.dropConn(gen, werr)
		return nil, werr
	}

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case r := <-ch:
		if r.err != nil {
			return nil, r.err
		}
		if len(r.resp.Error) > 0 && string(r.resp.Error) != "null" {
			return nil, fmt.Errorf("herdr %s error: %s", method, string(r.resp.Error))
		}
		return r.resp.Result, nil
	}
}

// connectLocked dials a new RPC connection and starts its reader. Caller holds mu.
func (c *Client) connectLocked(ctx context.Context) error {
	conn, err := c.dial(ctx)
	if err != nil {
		return err
	}
	c.conn = conn
	c.pending = make(map[string]chan call)
	c.gen++
	go c.readLoop(conn, c.gen)
	return nil
}

// readLoop dispatches responses on conn to waiters by id until the connection
// errors, then drops it (failing every outstanding waiter).
func (c *Client) readLoop(conn net.Conn, gen uint64) {
	r := bufio.NewReaderSize(conn, 1<<20)
	for {
		raw, err := readLine(r)
		if err != nil {
			c.dropConn(gen, err)
			return
		}
		var resp response
		if err := json.Unmarshal(raw, &resp); err != nil {
			continue
		}
		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		if ok {
			delete(c.pending, resp.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- call{resp: resp} // buffered, never blocks
		}
	}
}

// dropConn tears down the given connection generation and fails its waiters.
// A no-op if the generation is already superseded (concurrent drop / reconnect).
func (c *Client) dropConn(gen uint64, cause error) {
	c.mu.Lock()
	if c.gen != gen || c.conn == nil {
		c.mu.Unlock()
		return
	}
	conn := c.conn
	pending := c.pending
	c.conn = nil
	c.pending = nil
	c.gen++ // invalidate this generation
	c.mu.Unlock()

	_ = conn.Close()
	if cause == nil {
		cause = errClosed
	}
	for _, ch := range pending {
		ch <- call{err: cause} // buffered, never blocks
	}
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
