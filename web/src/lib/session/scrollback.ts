import type { Call, CallResult } from '$lib/protocol';

/** True when two line sets are element-for-element identical. */
export function sameLines(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}

export interface ScrollbackDeps {
  /** Fire a socket call and await its result. */
  request(call: Call): Promise<CallResult>;
  /** Lines to show before any read has ever succeeded, if the first read errors (e.g. the snapshot tail). */
  fallback(): string[];
  /** Invoked only when the deduped line set actually changes. */
  onLines(lines: string[]): void;
  /** Request ANSI-formatted output (SGR escapes preserved) instead of plain text. */
  ansi?: boolean;
  /** Injectable timers (tests); default to the globals. */
  setInterval?(fn: () => void, ms: number): number;
  clearInterval?(handle: number): void;
}

/**
 * Poll a pane's raw scrollback via `pane.read`. There is no push event for plain
 * terminal output, so this owns the interval, the dedupe (only emits on change),
 * and the fallback to the snapshot tail. Returns a stop function; the poller
 * keeps its own last-emitted set, so a fresh call = a fresh baseline.
 *
 * Two hazards this guards against (both produced a visible content flash —
 * a hundred-plus line transcript collapsing to the short snapshot tail and
 * back a tick later):
 *  - Ticks are not awaited against each other. If a round trip ever outlives
 *    `intervalMs` (slow request, WebSocket reconnect backlog), a later tick
 *    can start before an earlier one resolves; with no ordering guarantee on
 *    the wire, the earlier tick's stale reply can land last and overwrite
 *    fresher content. Fixed by skipping a tick outright while one is in flight
 *    rather than letting them race.
 *  - `pane.read` *erroring* is normal during a reconnect (every in-flight call
 *    rejects when the socket drops) or a transient backend hiccup - it says
 *    nothing about the pane's actual content, so once a real read has landed
 *    an error is "no update this tick" and the last good content stays; the
 *    snapshot-tail fallback only ever seeds the very first read, before
 *    anything real has been shown. A *successful* reply with empty text is
 *    different: it is the backend affirmatively saying the pane has no
 *    matching output right now (freshly cleared, freshly spawned), so it is
 *    always authoritative and always replaces whatever was shown before.
 */
export function startScrollback(paneId: string, deps: ScrollbackDeps, intervalMs = 1000): () => void {
  const si = deps.setInterval ?? ((fn, ms) => setInterval(fn, ms) as unknown as number);
  const ci = deps.clearInterval ?? ((h) => clearInterval(h));
  let alive = true;
  let current: string[] = [];
  let haveResult = false; // true once a genuine (non-fallback) read has landed
  let inFlight = false; // serializes ticks so overlapping requests can't race

  const readOnce = async () => {
    if (inFlight) return; // previous request still outstanding; skip this tick
    inFlight = true;
    try {
      let next: string[];
      try {
        const r = await deps.request({
          method: 'pane.read',
          params: {
            pane_id: paneId,
            source: 'recent_unwrapped',
            lines: 200,
            ...(deps.ansi ? { format: 'ansi' as const } : {})
          }
        });
        const text = r.read?.text ?? '';
        next = text ? text.split('\n') : []; // successful reply is authoritative, even when empty
        haveResult = true;
      } catch {
        if (haveResult) return; // transient error; keep showing the last good read
        next = deps.fallback(); // no live read has ever succeeded; show the best known state
      }
      if (alive && !sameLines(next, current)) {
        current = next;
        deps.onLines(next);
      }
    } finally {
      inFlight = false;
    }
  };

  void readOnce();
  const handle = si(readOnce, intervalMs);
  return () => {
    alive = false;
    ci(handle);
  };
}
