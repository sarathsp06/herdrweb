import type { Pane, Rollup, Space, Status, Tab } from '$lib/protocol';
import { STATUS_ORDER } from '$lib/protocol';

export interface AgentRef { pane: Pane; space: Space; tab: Tab }

export function agentsOf(spaces: Space[]): AgentRef[] {
  const out: AgentRef[] = [];
  for (const space of spaces)
    for (const tab of space.tabs)
      for (const pane of tab.panes) if (pane.agent) out.push({ pane, space, tab });
  return out.sort((a, b) => STATUS_ORDER[a.pane.status] - STATUS_ORDER[b.pane.status]);
}

export function findPaneIn(spaces: Space[], paneId: string): AgentRef | undefined {
  for (const space of spaces)
    for (const tab of space.tabs) {
      const pane = tab.panes.find((p) => p.id === paneId);
      if (pane) return { pane, space, tab };
    }
  return undefined;
}

/** The pane to open when a whole space is selected: the most attention-worthy
 *  agent pane (blocked > working > done > idle), else the space's first pane. */
export function primaryPaneOf(spaces: Space[], spaceId: string): string | undefined {
  const space = spaces.find((s) => s.id === spaceId);
  if (!space) return undefined;
  const panes = space.tabs.flatMap((t) => t.panes);
  const agents = panes
    .filter((p) => p.agent)
    .sort((a, b) => STATUS_ORDER[a.status] - STATUS_ORDER[b.status]);
  return (agents[0] ?? panes[0])?.id;
}

/** Primary pane of a single tab: most attention-worthy agent pane, else first pane. */
export function primaryPaneOfTab(tab: Tab): string | undefined {
  const agents = tab.panes
    .filter((p) => p.agent)
    .sort((a, b) => STATUS_ORDER[a.status] - STATUS_ORDER[b.status]);
  return (agents[0] ?? tab.panes[0])?.id;
}

/** Pane to open when a space is selected: the remembered tab's primary pane when
 *  that tab still exists, else the space-wide primary pane. */
export function chatPaneForSpace(
  spaces: Space[],
  spaceId: string,
  lastTabId?: string | null
): string | undefined {
  if (lastTabId) {
    const tab = spaces.find((s) => s.id === spaceId)?.tabs.find((t) => t.id === lastTabId);
    const paneId = tab && primaryPaneOfTab(tab);
    if (paneId) return paneId;
  }
  return primaryPaneOf(spaces, spaceId);
}

export function rollupOf(spaces: Space[], spaceId: string): Rollup {
  const space = spaces.find((s) => s.id === spaceId);
  if (!space) return 'none';
  const agentPanes = space.tabs.flatMap((t) => t.panes).filter((p) => p.agent);
  if (agentPanes.length === 0) return 'none';
  if (agentPanes.some((p) => p.status === 'blocked')) return 'blocked';
  if (agentPanes.some((p) => p.status === 'working')) return 'working';
  return 'idle';
}

export function countsOf(spaces: Space[], spaceId: string): { tabs: number; panes: number } {
  const space = spaces.find((s) => s.id === spaceId);
  if (!space) return { tabs: 0, panes: 0 };
  return { tabs: space.tabs.length, panes: space.tabs.reduce((n, t) => n + t.panes.length, 0) };
}

export function tabHasBlocked(tab: Tab): boolean {
  return tab.panes.some((p) => p.agent && p.status === 'blocked');
}

export function monogram(label: string): string {
  const cleaned = label.replace(/[^a-zA-Z0-9]/g, '');
  return (cleaned.slice(0, 2) || label.slice(0, 2)).toUpperCase();
}

/** Status glyph mirroring the Herdr TUI sidebar (shape carries status, not colour alone). */
export function glyph(status: Status | Rollup): string {
  switch (status) {
    case 'blocked': return '◉'; // ◉ filled target — demands attention
    case 'working': return '◐'; // ◐ half — in progress
    case 'done': return '●';    // ● solid — finished
    case 'idle': return '○';    // ○ hollow — ready
    case 'none': return '—';    // — no agents
    default: return '·';        // · unknown
  }
}
