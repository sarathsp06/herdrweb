package attention

import (
	"reflect"
	"testing"

	"github.com/sarathsp06/herdrweb/internal/protocol"
)

// snap builds a one-space one-tab snapshot from the given panes.
func snap(panes ...protocol.Pane) protocol.Snapshot {
	return protocol.Snapshot{
		Type:   "snapshot",
		Spaces: []protocol.Space{{ID: "w1", Tabs: []protocol.Tab{{ID: "w1:t1", Panes: panes}}}},
	}
}

func agentPane(id, label string, status protocol.Status) protocol.Pane {
	return protocol.Pane{ID: id, Label: label, Status: status, Agent: true}
}

func TestFirstCallSeedsBaselineOnly(t *testing.T) {
	var d Detector
	// A pane already Blocked on the very first snapshot must NOT fire — a fresh
	// process never replays a stale attention state.
	if hits := d.Detect(snap(agentPane("w1:p1", "claude", protocol.Blocked))); hits != nil {
		t.Fatalf("first call should seed baseline only, got %v", hits)
	}
}

func TestRisingEdgeIntoBlockedAndDone(t *testing.T) {
	var d Detector
	d.Detect(snap(agentPane("w1:p1", "claude", protocol.Working))) // baseline

	hits := d.Detect(snap(agentPane("w1:p1", "claude", protocol.Blocked)))
	if want := []Hit{{PaneID: "w1:p1", Label: "claude", Status: protocol.Blocked}}; !reflect.DeepEqual(hits, want) {
		t.Fatalf("working->blocked: got %v want %v", hits, want)
	}

	// blocked -> done fires again.
	hits = d.Detect(snap(agentPane("w1:p1", "claude", protocol.Done)))
	if want := []Hit{{PaneID: "w1:p1", Label: "claude", Status: protocol.Done}}; !reflect.DeepEqual(hits, want) {
		t.Fatalf("blocked->done: got %v want %v", hits, want)
	}
}

func TestNoFireWhenStatusUnchanged(t *testing.T) {
	var d Detector
	d.Detect(snap(agentPane("w1:p1", "claude", protocol.Blocked))) // baseline (no fire)
	if hits := d.Detect(snap(agentPane("w1:p1", "claude", protocol.Blocked))); hits != nil {
		t.Fatalf("unchanged blocked should not re-fire, got %v", hits)
	}
}

func TestNonAgentAndNonAttentionExcluded(t *testing.T) {
	var d Detector
	d.Detect(snap(
		protocol.Pane{ID: "term", Status: protocol.Working, Agent: false},
		agentPane("w1:p1", "claude", protocol.Working),
	))
	// A non-agent pane entering "blocked" and an agent pane entering idle/working
	// must not fire; only agent panes entering Blocked/Done do.
	hits := d.Detect(snap(
		protocol.Pane{ID: "term", Status: protocol.Blocked, Agent: false},
		agentPane("w1:p1", "claude", protocol.Idle),
	))
	if hits != nil {
		t.Fatalf("non-agent/non-attention transitions must not fire, got %v", hits)
	}
}

func TestBaselineTracksAllPanesNotJustAttention(t *testing.T) {
	var d Detector
	// Baseline records the agent pane as Working even though it is not an
	// attention state; the later rising edge must diff against that stored value.
	d.Detect(snap(agentPane("w1:p1", "claude", protocol.Working)))
	d.Detect(snap(agentPane("w1:p1", "claude", protocol.Idle))) // still tracked, no fire
	hits := d.Detect(snap(agentPane("w1:p1", "claude", protocol.Blocked)))
	if len(hits) != 1 || hits[0].Status != protocol.Blocked {
		t.Fatalf("idle->blocked should fire once, got %v", hits)
	}
}

func TestResultOrderFollowsWalk(t *testing.T) {
	var d Detector
	d.Detect(snap(
		agentPane("w1:p1", "a", protocol.Working),
		agentPane("w1:p2", "b", protocol.Working),
		agentPane("w1:p3", "c", protocol.Working),
	))
	hits := d.Detect(snap(
		agentPane("w1:p1", "a", protocol.Blocked),
		agentPane("w1:p2", "b", protocol.Done),
		agentPane("w1:p3", "c", protocol.Blocked),
	))
	got := []string{}
	for _, h := range hits {
		got = append(got, h.PaneID)
	}
	if want := []string{"w1:p1", "w1:p2", "w1:p3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("order: got %v want %v", got, want)
	}
}
