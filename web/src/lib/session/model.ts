import type { Pane, Snapshot, SessionEvent, Space, Tab } from '$lib/protocol';

export interface AgentRef {
  pane: Pane;
  space: Space;
  tab: Tab;
}

/** Pure, framework-free session state. Bootstraps from a snapshot and
 *  live-patches from events. Unit-tested directly. */
export class SessionModel {
  spaces: Space[] = [];
  focus: { spaceId?: string; tabId?: string; paneId?: string } = {};

  // Re-snapshot fires on every poll tick (~1.5s) regardless of whether
  // anything actually changed, so this merges bottom-up against the previous
  // tree: a space/tab/pane whose own fields are unchanged keeps its prior
  // object reference. Views bound to an unaffected space/tab/pane (e.g. the
  // pane currently open) then see no prop change and skip re-rendering —
  // only the subtree that actually changed gets new identity.
  applySnapshot(snap: Snapshot): void {
    const prev = this.spaces;
    this.spaces = snap.spaces.map((s) => mergeSpace(prev.find((p) => p.id === s.id), s));
    this.focus = { ...snap.focus };
  }

  apply(ev: SessionEvent): void {
    switch (ev.type) {
      case 'snapshot':
        this.applySnapshot(ev);
        break;
      case 'workspace.updated':
        this.upsertSpace(cloneSpace(ev.space));
        break;
      case 'workspace.closed':
        this.spaces = this.spaces.filter((s) => s.id !== ev.spaceId);
        break;
      case 'tab.updated':
        this.upsertTab(ev.spaceId, ev.tab);
        break;
      case 'pane.updated':
        this.upsertPane(ev.spaceId, ev.tabId, ev.pane);
        break;
      case 'pane.agent_status_changed': {
        const found = this.findPane(ev.paneId);
        if (found) found.pane.status = ev.status;
        break;
      }
      case 'pane.output_matched': {
        const found = this.findPane(ev.paneId);
        if (found) found.pane.tail = [...found.pane.tail, ev.line].slice(-200);
        break;
      }
    }
  }

  private upsertSpace(space: Space): void {
    const i = this.spaces.findIndex((s) => s.id === space.id);
    if (i === -1) this.spaces.push(space);
    else this.spaces[i] = space;
  }

  private upsertTab(spaceId: string, tab: Tab): void {
    const space = this.space(spaceId);
    if (!space) return;
    const i = space.tabs.findIndex((t) => t.id === tab.id);
    if (i === -1) space.tabs.push(tab);
    else space.tabs[i] = tab;
  }

  private upsertPane(spaceId: string, tabId: string, pane: Pane): void {
    const space = this.space(spaceId);
    const tab = space?.tabs.find((t) => t.id === tabId);
    if (!tab) return;
    const i = tab.panes.findIndex((p) => p.id === pane.id);
    if (i === -1) tab.panes.push(pane);
    else tab.panes[i] = pane;
  }

  space(id: string): Space | undefined {
    return this.spaces.find((s) => s.id === id);
  }

  findPane(paneId: string): AgentRef | undefined {
    for (const space of this.spaces) {
      for (const tab of space.tabs) {
        const pane = tab.panes.find((p) => p.id === paneId);
        if (pane) return { pane, space, tab };
      }
    }
    return undefined;
  }

}

function cloneSpace(s: Space): Space {
  return {
    ...s,
    tabs: s.tabs.map((t) => ({ ...t, panes: t.panes.map((p) => ({ ...p, tail: [...p.tail] })) }))
  };
}

function sameTail(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}

function paneFieldsEqual(a: Pane, b: Pane): boolean {
  return (
    a.label === b.label &&
    a.sub === b.sub &&
    a.status === b.status &&
    a.agent === b.agent &&
    sameTail(a.tail, b.tail)
  );
}

/** Reuse `old` when `next`'s content is identical; else clone `next` fresh
 *  (never returning a reference into `next` itself — callers must not alias
 *  the incoming snapshot). */
function mergePane(old: Pane | undefined, next: Pane): Pane {
  if (old && paneFieldsEqual(old, next)) return old;
  return { ...next, tail: [...next.tail] };
}

function mergeTab(old: Tab | undefined, next: Tab): Tab {
  const panes = next.panes.map((p) => mergePane(old?.panes.find((op) => op.id === p.id), p));
  if (
    old &&
    old.label === next.label &&
    old.panes.length === panes.length &&
    old.panes.every((p, i) => p === panes[i])
  ) {
    return old;
  }
  return { ...next, panes };
}

function mergeSpace(old: Space | undefined, next: Space): Space {
  const tabs = next.tabs.map((t) => mergeTab(old?.tabs.find((ot) => ot.id === t.id), t));
  if (
    old &&
    old.label === next.label &&
    old.cwd === next.cwd &&
    old.branch === next.branch &&
    old.worktree === next.worktree &&
    old.tabs.length === tabs.length &&
    old.tabs.every((t, i) => t === tabs[i])
  ) {
    return old;
  }
  return { ...next, tabs };
}
