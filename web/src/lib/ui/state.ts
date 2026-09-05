import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { Config } from '$lib/protocol';

// View-local UI state (selection, draft, collapse, theme, toggles, toast, sheet).
export const draft = writable<string>('');
export const wrap = writable<boolean>(false);
export const toast = writable<string>('');

let toastTimer: ReturnType<typeof setTimeout> | null = null;
export function showToast(msg: string): void {
  toast.set(msg);
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toast.set(''), 2200);
}

const CONFIG_KEY = 'herdrweb.config';
const DEFAULT_CONFIG: Config = { theme: 'herdr-dark', notify: true, follow: true, ansi: true, devCaptions: false, fontScale: 1, navCorner: 'bottom-right' };

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

export interface Sheet {
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
}
export const sheet = writable<Sheet | null>(null);

export function openSheet(s: Sheet): void {
  sheet.set(s);
}
export function closeSheet(): void {
  sheet.set(null);
}

// Grouped/flat toggle for the agents section of the inbox sidebar.
export const agentsGrouped = writable<boolean>(true);

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

// Sidebar/nav visibility. Desktop: pushes content when open. Mobile: slide-in
// drawer overlay. Layout sets the initial value per breakpoint on mount.
export const navOpen = writable<boolean>(true);
