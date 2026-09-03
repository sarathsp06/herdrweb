import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { FIXTURE_SNAPSHOT } from '$lib/transport/fixture';

// The wire protocol is mirrored, not generated (AGENTS.md): the Go structs in
// internal/protocol are canonical and this TS twin + the fixture are hand-written.
// internal/protocol/testdata/snapshot_golden.json is the single captured contract.
// The Go test (internal/protocol/contract_test.go) pins the Go structs to it;
// this test pins the TS twin + fixture to the same file. FIXTURE_SNAPSHOT is
// typed `Snapshot`, so svelte-check already ties it to the TS types — asserting
// fixture ≡ golden therefore also ties the TS types to the golden. Drift on
// either side fails loud instead of shipping an unreadable snapshot.

// Walk decoded JSON collecting object field paths, array indices normalized away
// (e.g. "spaces.tabs.panes.status"), so key inventories compare structurally.
function collectPaths(v: unknown, prefix: string, set: Set<string>): void {
  if (Array.isArray(v)) {
    for (const e of v) collectPaths(e, prefix, set);
  } else if (v !== null && typeof v === 'object') {
    for (const [k, child] of Object.entries(v)) {
      const p = prefix ? `${prefix}.${k}` : k;
      set.add(p);
      collectPaths(child, p, set);
    }
  }
}

function pathsOf(v: unknown): Set<string> {
  const set = new Set<string>();
  collectPaths(v, '', set);
  return set;
}

describe('protocol wire contract', () => {
  const goldenPath = resolve(process.cwd(), '../internal/protocol/testdata/snapshot_golden.json');
  const golden = JSON.parse(readFileSync(goldenPath, 'utf8'));

  it('fixture shape matches the golden snapshot (minus the wire-only `type`)', () => {
    const goldenPaths = pathsOf(golden);
    goldenPaths.delete('type'); // TS `Snapshot` omits it; the SessionEvent union adds it on the wire
    const fixturePaths = pathsOf(FIXTURE_SNAPSHOT);
    expect([...fixturePaths].sort()).toEqual([...goldenPaths].sort());
  });

  it('golden carries the wire-only `type` discriminator', () => {
    expect(golden.type).toBe('snapshot');
  });
});
