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

  applySnapshot(snap: Snapshot): void {
    this.spaces = snap.spaces.map(cloneSpace);
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
