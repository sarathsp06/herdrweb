// Package attention detects the rising edge into an attention state across
// successive session snapshots: an agent pane entering Blocked or Done, or a
// working agent going Idle (finishing a turn). It is a pure in-memory leaf: it
// imports only protocol, does no I/O, and authors no user-facing copy — the
// caller decides what to do with a Hit (Web Push today, a terminal bell or
// metric tomorrow).
package attention

import (
	"sync"

	"github.com/sarathsp06/herdrweb/internal/protocol"
)

// Hit is one agent pane that just transitioned into an attention state.
type Hit struct {
	PaneID string
	Label  string
	Status protocol.Status // Blocked, Done, or Idle (a finished turn)
}

// Detector tracks each pane's last-observed status to find rising edges into
// Blocked or Done. The zero value is ready to use and is safe for concurrent
// use; a fresh Detector (e.g. after a process restart) never replays stale
// attention states, because its first Detect only seeds the baseline.
type Detector struct {
	mu   sync.Mutex
	prev map[string]protocol.Status // nil until the first Detect seeds it
}

// Detect walks snap deterministically (spaces -> tabs -> panes) and returns, in
// walk order, every agent pane whose status just crossed an attention edge since
// the previous call (see attentionEdge): entering Blocked or Done, or a working
// pane going Idle. Because session-detected agents (omp, pi) never reach Done and
// finish by going working->idle, that edge is what surfaces them. The first call
// on a Detector always returns nil: it captures the full status of every pane
// (agent or not) as the baseline without reporting anything. Detect mutates only
// the Detector's own baseline; it performs no I/O.
func (d *Detector) Detect(snap protocol.Snapshot) []Hit {
	d.mu.Lock()
	defer d.mu.Unlock()

	first := d.prev == nil
	cur := make(map[string]protocol.Status)
	var hits []Hit
	for _, sp := range snap.Spaces {
		for _, tb := range sp.Tabs {
			for _, p := range tb.Panes {
				cur[p.ID] = p.Status
				if first || !p.Agent {
					continue
				}
				prev := d.prev[p.ID]
				if prev != p.Status && attentionEdge(prev, p.Status) {
					hits = append(hits, Hit{PaneID: p.ID, Label: p.Label, Status: p.Status})
				}
			}
		}
	}
	d.prev = cur
	if first {
		return nil
	}
	return hits
}

// attentionEdge reports whether a status transition the operator should hear
// about just happened. The caller has already confirmed prev != cur. An agent
// entering Blocked or Done always qualifies; an agent going Idle qualifies only
// from Working — a finished turn — so idle panes reappearing (e.g. blocked->idle
// after the operator answers, or a freshly tracked idle pane) stay quiet.
func attentionEdge(prev, cur protocol.Status) bool {
	switch cur {
	case protocol.Blocked, protocol.Done:
		return true
	case protocol.Idle:
		return prev == protocol.Working
	default:
		return false
	}
}
