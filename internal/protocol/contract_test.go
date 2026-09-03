package protocol

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
)

// The wire protocol is mirrored, not generated (see AGENTS.md): protocol.go is
// canonical and web/src/lib/protocol/index.ts + fixture.ts are hand-written
// twins. testdata/snapshot_golden.json is the single captured contract both
// languages verify against. This test pins the Go structs to the golden; the
// TS test (web/src/lib/protocol/contract.test.ts) pins the TS twin + fixture to
// the same file. Rename/remove/add a field on one side and one of the two fails
// loud in CI instead of silently shipping a snapshot the other side can't read.

// collectPaths walks decoded JSON and returns the set of object field paths,
// with array indices normalized away (e.g. "spaces.tabs.panes.status"), so the
// key inventory can be compared structurally regardless of element count.
func collectPaths(v any, prefix string, set map[string]bool) {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			set[p] = true
			collectPaths(child, p, set)
		}
	case []any:
		for _, e := range t {
			collectPaths(e, prefix, set)
		}
	}
}

func pathsOf(t *testing.T, raw []byte) map[string]bool {
	t.Helper()
	var m any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	set := map[string]bool{}
	collectPaths(m, "", set)
	return set
}

func TestSnapshotContract(t *testing.T) {
	golden, err := os.ReadFile("testdata/snapshot_golden.json")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	// 1. The golden must decode into Snapshot with no unknown keys — catches a
	// field the golden (contract) carries that the Go structs no longer know.
	dec := json.NewDecoder(bytes.NewReader(golden))
	dec.DisallowUnknownFields()
	var snap Snapshot
	if err := dec.Decode(&snap); err != nil {
		t.Fatalf("golden has a field the Go structs reject: %v", err)
	}

	// 2. Re-marshalling the decoded value must reproduce exactly the golden's
	// key inventory — catches a field the Go structs emit that the golden lacks.
	out, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	goSet := pathsOf(t, out)
	goldenSet := pathsOf(t, golden)
	if !reflect.DeepEqual(goSet, goldenSet) {
		t.Fatalf("Go snapshot shape drifted from golden\n  go:     %v\n  golden: %v", sortedKeys(goSet), sortedKeys(goldenSet))
	}
}

func sortedKeys(m map[string]bool) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
