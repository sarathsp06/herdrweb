import { readable, writable } from 'svelte/store';
import type { Call, CallResult, ConnState, Prompt } from '$lib/protocol';

/** Wall clock for "waiting 4m" labels, ticking twice a minute. */
export const now = readable(Date.now(), (set) => {
  const t = setInterval(() => set(Date.now()), 30_000);
  return () => clearInterval(t);
});

/** "4m", "1h 5m", "just now" — how long an agent has been waiting. */
export function waitedFor(blockedSince: number | undefined, now: number): string {
  if (!blockedSince) return '';
  const min = Math.floor(Math.max(0, now - blockedSince) / 60_000);
  if (min < 1) return 'just now';
  if (min < 60) return `${min}m`;
  return `${Math.floor(min / 60)}h ${min % 60}m`;
}

/** Connection state in the words a person reads. */
export function connLabel(c: ConnState): string {
  return { open: 'live', connecting: 'connecting…', reconnecting: 'reconnecting…', closed: 'offline' }[c];
}

/** Answers short enough to offer straight from the inbox: one question, at
 *  most three choices, none of them a write-in. Anything else opens the pane. */
export function inlineAnswers(p: Prompt | undefined) {
  if (!p || p.tabs?.length || !p.nav || p.options.length > 3) return null;
  if (p.options.some((o) => o.needsText || o.keys.length === 0)) return null;
  return p.options;
}

/** A short key hint for an option: its digit/letter, or nothing for arrow picks. */
export function keyHint(p: Prompt, keys: string[]): string {
  return p.nav === 'digits' || p.nav === 'keys' ? keys.join(' ') : '';
}

/** Send an answer's keys as one ordered agent.send_keys call. */
export function sendAnswer(request: (c: Call) => Promise<CallResult>, paneId: string, keys: string[]) {
  return request({ method: 'agent.send_keys', params: { target: paneId, keys } });
}

// ---- Multi-question forms --------------------------------------------------
// The screen only shows the current question, so the answers chosen so far are
// remembered here (per pane, this tab only) to drive the progress strip and the
// review before Submit.
export interface Answered {
  question: string;
  answer: string;
}
export const answered = writable<Record<string, Answered[]>>({});

export function recordAnswer(paneId: string, question: string, answer: string) {
  answered.update((all) => {
    const list = (all[paneId] ?? []).filter((a) => a.question !== question);
    return { ...all, [paneId]: [...list, { question, answer }] };
  });
}

export function clearAnswers(paneId: string) {
  answered.update((all) => {
    if (!all[paneId]) return all;
    const next = { ...all };
    delete next[paneId];
    return next;
  });
}
