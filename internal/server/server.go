// Package server wires the Herdr socket client to browser WebSocket clients and
// serves the embedded UI. The bridge owns exactly one Herdr connection, keeps a
// normalized snapshot cache, re-derives it on any Herdr event, and broadcasts
// the fresh snapshot to every connected browser. Browser calls are id-correlated
// and passed through to the Herdr socket.
package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sarathsp06/herdrweb/internal/attention"
	"github.com/sarathsp06/herdrweb/internal/config"
	"github.com/sarathsp06/herdrweb/internal/herdr"
	"github.com/sarathsp06/herdrweb/internal/protocol"
	"github.com/sarathsp06/herdrweb/internal/push"
	"github.com/sarathsp06/herdrweb/internal/webui"
)

// ConnState is the bridge<->Herdr connection lifecycle.
type ConnState string

const (
	Connecting   ConnState = "connecting"
	Open         ConnState = "open"
	Reconnecting ConnState = "reconnecting"
	Closed       ConnState = "closed"
)

// Hub coordinates the Herdr connection and browser clients.
type Hub struct {
	client  *herdr.Client
	cfgPath string
	version string
	push    *push.Manager

	mu       sync.RWMutex
	snapshot []byte // latest normalized snapshot JSON
	state    ConnState
	browsers map[*browser]struct{}

	attn attention.Detector // owns the rising-edge baseline + its own lock

	dirty chan struct{}
}

type browser struct {
	conn   *websocket.Conn
	send   chan []byte
	mu     sync.Mutex
	closed bool
}

// trySend delivers to the browser's write pump without blocking, and is safe
// against a concurrent close (the read loop may remove the browser while an
// in-flight handleCall goroutine is still replying).
func (b *browser) trySend(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	select {
	case b.send <- data:
	default: // slow client; drop
	}
}

// NewHub builds a hub. version is reported to the UI; pm may be nil to disable
// push notifications.
func NewHub(client *herdr.Client, cfgPath, version string, pm *push.Manager) *Hub {
	return &Hub{
		client:   client,
		cfgPath:  cfgPath,
		version:  version,
		push:     pm,
		state:    Connecting,
		browsers: map[*browser]struct{}{},
		dirty:    make(chan struct{}, 1),
	}
}

// Run maintains the Herdr connection until ctx is cancelled: bootstrap snapshot,
// subscribe, coalesce events into refreshes, reconnect with backoff.
func (h *Hub) Run(ctx context.Context) {
	go h.debouncer(ctx)
	go h.poller(ctx)
	backoff := 500 * time.Millisecond
	for ctx.Err() == nil {
		h.setState(Connecting)
		if err := h.refresh(ctx); err != nil {
			h.setState(Reconnecting)
			if !sleep(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}
		h.setState(Open)
		backoff = 500 * time.Millisecond
		// Subscribe blocks until the connection drops or ctx ends.
		err := h.client.Subscribe(ctx, nil, func(herdr.Event) { h.markDirty() })
		if ctx.Err() != nil {
			return
		}
		log.Printf("herdr subscription ended: %v; reconnecting", err)
		h.setState(Reconnecting)
		if !sleep(ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff)
	}
}

func (h *Hub) debouncer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-h.dirty:
			timer := time.NewTimer(120 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			_ = h.refresh(ctx)
		}
	}
}

// poller marks the snapshot dirty on a fixed cadence so pane-scoped changes that
// emit no global event (notably agent-status transitions) still surface live.
func (h *Hub) poller(ctx context.Context) {
	t := time.NewTicker(1500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			h.markDirty()
		}
	}
}

func (h *Hub) markDirty() {
	select {
	case h.dirty <- struct{}{}:
	default:
	}
}

func (h *Hub) refresh(ctx context.Context) error {
	rctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	var hs protocol.HerdrSnapshot
	if err := h.client.Snapshot(rctx, &hs); err != nil {
		return err
	}
	snap := protocol.Normalize(&hs)
	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	h.mu.Lock()
	h.snapshot = data
	h.mu.Unlock()
	h.broadcast(data)
	h.notifyAttention(ctx, snap)
	return nil
}

// notifyAttention pushes a Web Push notification for every agent pane that just
// transitioned into an attention state (blocked or done). The first snapshot
// only seeds the baseline so restarts never replay stale states.
func (h *Hub) notifyAttention(ctx context.Context, snap protocol.Snapshot) {
	if h.push == nil {
		return
	}
	hits := h.attn.Detect(snap)
	if len(hits) == 0 {
		return
	}
	if s, err := config.Load(h.cfgPath); err == nil && !s.Notify {
		return
	}
	for _, hit := range hits {
		label := hit.Label
		if label == "" {
			label = "An agent"
		}
		title, body := "Agent blocked", label+" needs you."
		if hit.Status == protocol.Done {
			title, body = "Agent done", label+" finished."
		}
		h.push.Notify(ctx, push.Notification{
			Title: title,
			Body:  body,
			URL:   "/pane/" + hit.PaneID,
		})
	}
}

func (h *Hub) setState(s ConnState) {
	h.mu.Lock()
	changed := h.state != s
	h.state = s
	h.mu.Unlock()
	if changed {
		log.Printf("herdr connection: %s", s)
	}
}

func (h *Hub) broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for b := range h.browsers {
		b.trySend(data)
	}
}

func (h *Hub) addBrowser(b *browser) {
	h.mu.Lock()
	h.browsers[b] = struct{}{}
	snap := h.snapshot
	h.mu.Unlock()
	if snap != nil {
		b.trySend(snap)
	}
}

func (h *Hub) removeBrowser(b *browser) {
	h.mu.Lock()
	delete(h.browsers, b)
	h.mu.Unlock()
	b.mu.Lock()
	if !b.closed {
		b.closed = true
		close(b.send)
	}
	b.mu.Unlock()
}

var upgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

type wsRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// maxInflightPerConn bounds concurrent in-flight browser RPC handlers on a
// single WebSocket connection; excess frames wait (per-connection backpressure)
// instead of spawning unbounded goroutines.
const maxInflightPerConn = 8

func (h *Hub) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	b := &browser{conn: conn, send: make(chan []byte, 16)}
	h.addBrowser(b)
	defer h.removeBrowser(b)

	go func() {
		for msg := range b.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	sem := make(chan struct{}, maxInflightPerConn)
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var req wsRequest
		if json.Unmarshal(data, &req) != nil || req.Method == "" {
			continue
		}
		h.dispatch(sem, func() { h.handleCall(r.Context(), b, req) })
	}
}

// dispatch runs fn in a goroutine, bounded by sem: it blocks the caller (the
// per-connection read loop) until a slot is free, then releases on completion.
func (h *Hub) dispatch(sem chan struct{}, fn func()) {
	sem <- struct{}{}
	go func() {
		defer func() { <-sem }()
		fn()
	}()
}

func (h *Hub) handleCall(ctx context.Context, b *browser, req wsRequest) {
	cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	var params any
	if len(req.Params) > 0 {
		_ = json.Unmarshal(req.Params, &params)
	}
	res, err := h.client.Call(cctx, req.Method, params)
	var out []byte
	if err != nil {
		out, _ = json.Marshal(map[string]any{"id": req.ID, "error": err.Error()})
	} else {
		out, _ = json.Marshal(map[string]any{"id": req.ID, "result": json.RawMessage(res)})
	}
	b.trySend(out)
}

func (h *Hub) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s, _ := config.Load(h.cfgPath)
		writeJSON(w, s)
	case http.MethodPut, http.MethodPost:
		var s config.Settings
		if json.NewDecoder(r.Body).Decode(&s) != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if err := config.Save(h.cfgPath, s); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Persisted -> ask Herdr to reload.
		_, _ = h.client.Call(r.Context(), "server.reload_config", map[string]any{})
		writeJSON(w, map[string]any{"ok": true})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Hub) handleHealth(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	state := h.state
	h.mu.RUnlock()
	writeJSON(w, map[string]any{"ok": true, "herdr": string(state), "version": h.version, "socket": h.client.SocketPath})
}

// handlePushKey returns the VAPID public key the browser subscribes with.
func (h *Hub) handlePushKey(w http.ResponseWriter, r *http.Request) {
	key := ""
	if h.push != nil {
		key = h.push.PublicKey()
	}
	writeJSON(w, map[string]any{"key": key})
}

// handlePushSubscribe records a browser push subscription posted by the UI.
func (h *Hub) handlePushSubscribe(w http.ResponseWriter, r *http.Request) {
	if h.push == nil {
		http.Error(w, "push disabled", http.StatusServiceUnavailable)
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var sub push.Subscription
	if json.NewDecoder(r.Body).Decode(&sub) != nil || sub.Endpoint == "" {
		http.Error(w, "bad subscription", http.StatusBadRequest)
		return
	}
	h.push.Add(sub)
	writeJSON(w, map[string]any{"ok": true})
}

// Handler returns the full HTTP handler (UI + /ws + /api).
func (h *Hub) Handler() (http.Handler, error) {
	ui, err := webui.Handler()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWS)
	mux.HandleFunc("/api/config", h.handleConfig)
	mux.HandleFunc("/api/health", h.handleHealth)
	mux.HandleFunc("/api/push/key", h.handlePushKey)
	mux.HandleFunc("/api/push/subscribe", h.handlePushSubscribe)
	mux.Handle("/", ui)
	return mux, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func nextBackoff(d time.Duration) time.Duration {
	d *= 2
	if d > 10*time.Second {
		return 10 * time.Second
	}
	return d
}
