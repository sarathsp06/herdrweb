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
  /** Lines to show when the read fails or returns empty (e.g. the snapshot tail). */
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
 * and the fallback to the snapshot tail on empty/error. Returns a stop function;
 * the poller keeps its own last-emitted set, so a fresh call = a fresh baseline.
 */
export function startScrollback(paneId: string, deps: ScrollbackDeps, intervalMs = 1000): () => void {
  const si = deps.setInterval ?? ((fn, ms) => setInterval(fn, ms) as unknown as number);
  const ci = deps.clearInterval ?? ((h) => clearInterval(h));
  let alive = true;
  let current: string[] = [];

  const readOnce = async () => {
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
      next = text ? text.split('\n') : deps.fallback();
    } catch {
      next = deps.fallback();
    }
    if (alive && !sameLines(next, current)) {
      current = next;
      deps.onLines(next);
    }
  };

  void readOnce();
  const handle = si(readOnce, intervalMs);
  return () => {
    alive = false;
    ci(handle);
  };
}
