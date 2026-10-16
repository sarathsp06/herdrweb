import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { Config } from '$lib/protocol';

// View-local UI state (selection, draft, collapse, theme, toggles, toast, sheet).
export const wrap = writable<boolean>(false);
export const toast = writable<string>('');

let toastTimer: ReturnType<typeof setTimeout> | null = null;
export function showToast(msg: string): void {
  toast.set(msg);
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toast.set(''), 2200);
}

// Composer drafts, keyed by pane id, so text typed in one pane can never be
// sent to another after switching. Persisted: mobile browsers drop background
// tabs, and a half-written prompt shouldn't die with the reload.
const DRAFTS_KEY = 'herdrweb.drafts';
function loadDrafts(): Record<string, string> {
  if (!browser) return {};
  try {
    const raw = localStorage.getItem(DRAFTS_KEY);
    if (raw) return JSON.parse(raw) as Record<string, string>;
  } catch {
    // ignore corrupt storage
  }
  return {};
}
export const drafts = writable<Record<string, string>>(loadDrafts());
if (browser) {
  drafts.subscribe((m) => {
    try {
      localStorage.setItem(DRAFTS_KEY, JSON.stringify(m));
    } catch {
      // storage unavailable; in-memory only
    }
  });
}
/** Set one pane's draft; an empty draft drops the key so the map stays small. */
export function setDraft(paneId: string, text: string): void {
  drafts.update((m) => {
    if ((m[paneId] ?? '') === text) return m;
    const { [paneId]: _, ...rest } = m;
    return text ? { ...rest, [paneId]: text } : rest;
  });
}

// Sent-prompt history per pane, newest last, so the operator can recall and
// resend/edit instead of retyping on a phone keyboard. Persisted like drafts.
const HISTORY_KEY = 'herdrweb.history';
export const HISTORY_MAX = 50;
function loadHistory(): Record<string, string[]> {
  if (!browser) return {};
  try {
    const raw = localStorage.getItem(HISTORY_KEY);
    if (raw) return JSON.parse(raw) as Record<string, string[]>;
  } catch {
    // ignore corrupt storage
  }
  return {};
}
export const history = writable<Record<string, string[]>>(loadHistory());
if (browser) {
  history.subscribe((m) => {
    try {
      localStorage.setItem(HISTORY_KEY, JSON.stringify(m));
    } catch {
      // storage unavailable; in-memory only
    }
  });
}
/** Record a sent prompt; an immediate repeat is not stored twice. */
export function pushHistory(paneId: string, text: string): void {
  history.update((m) => {
    const list = m[paneId] ?? [];
    if (list[list.length - 1] === text) return m;
    return { ...m, [paneId]: [...list, text].slice(-HISTORY_MAX) };
  });
}

const CONFIG_KEY = 'herdrweb.config';
const DEFAULT_CONFIG: Config = { theme: 'herdr-dark', notify: true, follow: true, ansi: true, devCaptions: false, fontScale: 1 };

function loadConfig(): Config {
  if (!browser) return { ...DEFAULT_CONFIG };
  try {
    const raw = localStorage.getItem(CONFIG_KEY);
    if (raw) return { ...DEFAULT_CONFIG, ...JSON.parse(raw) };
  } catch {
    // ignore corrupt storage
  }
  return { ...DEFAULT_CONFIG };
}

export const config = writable<Config>(loadConfig());
if (browser) {
  config.subscribe((c) => {
    try {
      localStorage.setItem(CONFIG_KEY, JSON.stringify(c));
    } catch {
      // storage unavailable; in-memory only
    }
  });
}

export interface SheetAction {
  label: string;
  hint?: string;
  glyph?: string;
  destructive?: boolean;
  active?: boolean;
  onSelect: () => void;
}

/** Confirmation sheet: every mutation routes through one before firing. */
export interface ConfirmSheet {
  kind: string;
  title: string;
  body: string;
  cta: string;
  destructive?: boolean;
  hasInput?: boolean;
  hasCwd?: boolean;
  inputLabel?: string;
  call: string;
  onConfirm: (label: string, cwd: string) => void;
  actions?: undefined;
}

/** Action sheet: a tap-to-pick list (tab switcher, per-card overflow menus). */
export interface ActionSheet {
  kind: string;
  title: string;
  actions: SheetAction[];
}

export type Sheet = ConfirmSheet | ActionSheet;
export const sheet = writable<Sheet | null>(null);

export function openSheet(s: ConfirmSheet): void {
  sheet.set(s);
}
export function openActions(kind: string, title: string, actions: SheetAction[]): void {
  sheet.set({ kind, title, actions });
}
export function closeSheet(): void {
  sheet.set(null);
}

// Grouped/flat toggle for the agents section of the inbox sidebar.
export const agentsGrouped = writable<boolean>(true);

// Desktop terminal focus: hide the docked sidebar on `/pane/*` so the terminal
// gets the full width. Persisted; other routes always keep the sidebar docked.
const SIDEBAR_KEY = 'herdrweb.sidebarCollapsed';
function loadSidebarCollapsed(): boolean {
  if (!browser) return false;
  try {
    return localStorage.getItem(SIDEBAR_KEY) === '1';
  } catch {
    return false; // storage unavailable
  }
}
export const sidebarCollapsed = writable<boolean>(loadSidebarCollapsed());
if (browser) {
  sidebarCollapsed.subscribe((v) => {
    try {
      localStorage.setItem(SIDEBAR_KEY, v ? '1' : '0');
    } catch {
      // storage unavailable; in-memory only
    }
  });
}
// While collapsed, the sidebar can be shown over the terminal (it does not
// resize it); any navigation closes it again.
export const sidebarPeek = writable<boolean>(false);

// Last-selected pane, so desktop `/` resolves to its chat.
export const lastPane = writable<string | null>(null);

// Last-selected tab per space, so re-opening a space returns to that tab's chat.
const LAST_TAB_KEY = 'herdrweb.lastTab';
function loadLastTab(): Record<string, string> {
  if (!browser) return {};
  try {
    const raw = localStorage.getItem(LAST_TAB_KEY);
    if (raw) return JSON.parse(raw) as Record<string, string>;
  } catch {
    // ignore corrupt storage
  }
  return {};
}
export const lastTabBySpace = writable<Record<string, string>>(loadLastTab());
if (browser) {
  lastTabBySpace.subscribe((m) => {
    try {
      localStorage.setItem(LAST_TAB_KEY, JSON.stringify(m));
    } catch {
      // storage unavailable; in-memory only
    }
  });
}
export function rememberTab(spaceId: string, tabId: string): void {
  lastTabBySpace.update((m) => (m[spaceId] === tabId ? m : { ...m, [spaceId]: tabId }));
}
