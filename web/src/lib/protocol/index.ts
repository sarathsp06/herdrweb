// Canonical UI protocol types. Go structs in bridge/internal/protocol are the
// source of truth; these are the hand-mirrored TypeScript twin (design Q1).

export type Status = 'working' | 'blocked' | 'done' | 'idle' | 'unknown';

export const STATUS_ORDER: Record<Status, number> = {
  blocked: 0,
  working: 1,
  done: 2,
  idle: 3,
  unknown: 4
};

/** Rollup adds 'none' for spaces with no agent panes (rendered as an em dash). */
export type Rollup = Status | 'none';

export interface Pane {
  id: string; // e.g. "w1:p2"
  label: string;
  sub: string; // subtitle, e.g. "claude · editing PaneCell.svelte"
  status: Status;
  agent: boolean; // true when Herdr recognizes a coding agent in the pane
  tail: string[]; // recent output lines (may be empty until read)
}

export interface Tab {
  id: string; // e.g. "w1:t1"
  label: string;
  panes: Pane[];
}

export interface Space {
  id: string; // e.g. "w1"
  label: string;
  cwd: string;
  branch: string;
  worktree: string | null; // worktree checkout path when this space is a worktree
  tabs: Tab[];
}

export interface Snapshot {
  spaces: Space[];
  focus: { spaceId?: string; tabId?: string; paneId?: string };
}

// ---- Live event union the session store patches from ----
export type SessionEvent =
  | ({ type: 'snapshot' } & Snapshot)
  | { type: 'workspace.updated'; space: Space }
  | { type: 'workspace.closed'; spaceId: string }
  | { type: 'tab.updated'; spaceId: string; tab: Tab }
  | { type: 'pane.updated'; spaceId: string; tabId: string; pane: Pane }
  | { type: 'pane.agent_status_changed'; paneId: string; status: Status }
  | { type: 'pane.output_matched'; paneId: string; line: string };

// ---- Connection lifecycle ----
export type ConnState = 'connecting' | 'open' | 'reconnecting' | 'closed';

// ---- Outbound socket calls (mirrors Herdr socket API surface used by the UI) ----
export type SendKey = 'y' | 'a' | 'n' | 'esc';

export interface PromptWait {
  until: Status[];
  timeout_ms: number;
}

export type Call =
  | { method: 'pane.read'; params: { pane_id: string; source: 'recent_unwrapped'; lines: number } }
  | { method: 'agent.prompt'; params: { target: string; text: string; wait: PromptWait } }
  | { method: 'agent.send_keys'; params: { target: string; keys: string[] } }
  | { method: 'workspace.create'; params: { cwd: string; label: string; focus: boolean } }
  | { method: 'workspace.rename'; params: { workspace_id: string; label: string } }
  | { method: 'workspace.close'; params: { workspace_id: string } }
  | { method: 'tab.create'; params: { workspace_id: string; label: string } }
  | { method: 'tab.rename'; params: { tab_id: string; label: string } }
  | { method: 'tab.close'; params: { tab_id: string } }
  | { method: 'pane.split'; params: { pane_id: string; direction: 'right' | 'down'; ratio: number } }
  | { method: 'pane.rename'; params: { pane_id: string; label: string } }
  | { method: 'pane.close'; params: { pane_id: string } }
  | { method: 'pane.send_text'; params: { pane_id: string; text: string } }
  | { method: 'pane.send_keys'; params: { pane_id: string; keys: string[] } }
  | { method: 'server.reload_config'; params: Record<string, never> };

export type CallResult = { ok?: boolean; lines?: string[]; read?: { text?: string }; [k: string]: unknown };

export interface Config {
  theme: 'herdr-dark' | 'ash' | 'gruvbox' | 'solarized-light';
  notify: boolean; // push when blocked
  follow: boolean; // follow focused pane
  ansi: boolean; // keep ANSI colors in raw
  devCaptions: boolean; // show socket-call captions (developer setting, off by default)
  fontScale: number; // webapp text-size multiplier (applied as document zoom)
  navCorner: 'top' | 'bottom-right' | 'bottom-left'; // phone nav-toggle placement
}

export interface ServerInfo {
  bridge: string;
  socket: string;
  version: string;
  protocol: string;
  uptime: string;
}
