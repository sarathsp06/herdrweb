import { describe, it, expect } from 'vitest';
import { SessionModel } from './model';
import { agentsOf, rollupOf, findPaneIn, tabHasBlocked } from './derive';
import { FIXTURE_SNAPSHOT, FixtureTransport } from '$lib/transport/fixture';
import type { Snapshot } from '$lib/protocol';

function fresh(): SessionModel {
  const m = new SessionModel();
  m.applySnapshot(FIXTURE_SNAPSHOT);
  return m;
}

describe('SessionModel bootstrap + rollup', () => {
  it('bootstraps every space/tab/pane from a snapshot', () => {
    const m = fresh();
    expect(m.spaces.map((s) => s.id)).toEqual(['w1', 'w2', 'w3', 'w4']);
    expect(m.findPane('w1:p2')?.pane.status).toBe('blocked');
  });

  it('derives rollup: blocked dominates working', () => {
    const m = fresh();
    expect(rollupOf(m.spaces, 'w1')).toBe('blocked'); // codex blocked beats claude working
    expect(rollupOf(m.spaces, 'w2')).toBe('idle'); // codex done, no working/blocked -> idle
    expect(rollupOf(m.spaces, 'w3')).toBe('blocked');
    expect(rollupOf(m.spaces, 'w4')).toBe('none'); // no agent panes
  });

  it('sorts agents blocked-first', () => {
    const m = fresh();
    const order = agentsOf(m.spaces).map((a) => a.pane.status);
    expect(order[0]).toBe('blocked');
    // never a non-blocked before a blocked
    const firstNonBlocked = order.findIndex((s) => s !== 'blocked');
    expect(order.slice(firstNonBlocked).includes('blocked')).toBe(false);
  });

  it('only agent panes appear in the agents list', () => {
    const m = fresh();
    const ids = agentsOf(m.spaces).map((a) => a.pane.id);
    expect(ids).not.toContain('w1:p3'); // dev server pane
    expect(ids).not.toContain('w4:p1'); // zsh
    expect(ids).toContain('w1:p2');
  });
});

describe('SessionModel live patching', () => {
  it('pane.agent_status_changed updates status in place', () => {
    const m = fresh();
    m.apply({ type: 'pane.agent_status_changed', paneId: 'w1:p2', status: 'idle' });
    expect(m.findPane('w1:p2')?.pane.status).toBe('idle');
    expect(rollupOf(m.spaces, 'w1')).toBe('working'); // now only claude working
  });

  it('pane.updated upserts a pane', () => {
    const m = fresh();
    m.apply({
      type: 'pane.updated',
      spaceId: 'w1',
      tabId: 'w1:t1',
      pane: { id: 'w1:p2', label: 'codex', sub: 'codex · idle', status: 'idle', agent: true, tail: [] }
    });
    expect(m.findPane('w1:p2')?.pane.sub).toBe('codex · idle');
  });

  it('pane.output_matched appends to the tail', () => {
    const m = fresh();
    m.apply({ type: 'pane.output_matched', paneId: 'w1:p1', line: 'new log line' });
    expect(m.findPane('w1:p1')?.pane.tail.at(-1)).toBe('new log line');
  });

  it('workspace.closed removes the space', () => {
    const m = fresh();
    m.apply({ type: 'workspace.closed', spaceId: 'w4' });
    expect(m.spaces.map((s) => s.id)).not.toContain('w4');
  });

  it('re-snapshot replaces state (reconnect semantics)', () => {
    const m = fresh();
    m.apply({ type: 'workspace.closed', spaceId: 'w1' });
    const snap: Snapshot = FIXTURE_SNAPSHOT;
    m.apply({ type: 'snapshot', ...snap });
    expect(m.spaces.map((s) => s.id)).toEqual(['w1', 'w2', 'w3', 'w4']);
  });

  it('mutating a re-snapshotted model does not mutate the source fixture', () => {
    const m = fresh();
    m.findPane('w1:p2')!.pane.status = 'done';
    expect(FIXTURE_SNAPSHOT.spaces[0].tabs[0].panes[1].status).toBe('blocked');
  });
});

describe('helpers', () => {
  it('tabHasBlocked flags a tab with a blocked agent pane', () => {
    const m = fresh();
    const t = m.space('w1')!.tabs[0];
    expect(tabHasBlocked(t)).toBe(true);
    expect(tabHasBlocked(m.space('w2')!.tabs[0])).toBe(false);
  });

  it('FixtureTransport emits a snapshot on subscribe and answers pane.read', async () => {
    const t = new FixtureTransport();
    let got: string[] = [];
    t.subscribe((ev) => {
      if (ev.type === 'snapshot') got = ev.spaces.map((s) => s.id);
    });
    expect(got).toEqual(['w1', 'w2', 'w3', 'w4']);
    const r = await t.request({ method: 'pane.read', params: { pane_id: 'w1:p1', source: 'recent_unwrapped', lines: 80 } });
    expect(r.ok).toBe(true);
    expect(typeof r.read?.text).toBe('string');
  });
});
