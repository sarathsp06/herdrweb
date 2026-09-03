// Package attention detects the rising edge into an attention state (an agent
// pane entering Blocked or Done) across successive session snapshots. It is a
// pure in-memory leaf: it imports only protocol, does no I/O, and authors no
// user-facing copy — the caller decides what to do with a Hit (Web Push today,
// a terminal bell or metric tomorrow).
package attention

import (
	"sync"

	"github.com/sarathsp06/herdrweb/internal/protocol"
)

// Hit is one agent pane that just transitioned into an attention state.
type Hit struct {
	PaneID string
	Label  string
	Status protocol.Status // always protocol.Blocked or protocol.Done
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
// walk order, every agent pane whose status just rose into Blocked or Done
// since the previous call. A transition fires when the pane's prior status
// differs from its current attention status, so Blocked->Done (and Done->Blocked)
// fire again. The first call on a Detector always returns nil: it captures the
// full status of every pane (agent or not) as the baseline without reporting
// anything. Detect mutates only the Detector's own baseline; it performs no I/O.
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
				if !p.Agent || (p.Status != protocol.Blocked && p.Status != protocol.Done) {
					continue
				}
				if !first && d.prev[p.ID] != p.Status {
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
