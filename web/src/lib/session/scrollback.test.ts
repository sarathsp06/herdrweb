import { describe, it, expect, vi } from 'vitest';
import type { Call, CallResult } from '$lib/protocol';
import { sameLines, startScrollback } from './scrollback';

describe('sameLines', () => {
  it('detects equality and difference', () => {
    expect(sameLines(['a', 'b'], ['a', 'b'])).toBe(true);
    expect(sameLines(['a'], ['a', 'b'])).toBe(false);
    expect(sameLines(['a', 'b'], ['a', 'c'])).toBe(false);
  });
});

/** Capture the interval callback instead of running a real timer. */
function fakeTimer() {
  let cb: (() => void) | null = null;
  return {
    tick: async () => {
      await cb?.();
    },
    cleared: false as boolean,
    setInterval(fn: () => void): number {
      cb = fn;
      return 1;
    },
    clearInterval(this: { cleared: boolean }) {
      this.cleared = true;
    }
  };
}

describe('startScrollback', () => {
  it('reads once immediately and emits the split lines', async () => {
    const t = fakeTimer();
    const emitted: string[][] = [];
    const request = vi.fn(async (_c: Call): Promise<CallResult> => ({ read: { text: 'one\ntwo' } }));
    startScrollback('p1', {
      request,
      fallback: () => [],
      onLines: (l) => emitted.push(l),
      setInterval: t.setInterval,
      clearInterval: () => {}
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(request).toHaveBeenCalledWith({
      method: 'pane.read',
      params: { pane_id: 'p1', source: 'recent_unwrapped', lines: 200 }
    });
    expect(emitted).toEqual([['one', 'two']]);
  });

  it('suppresses emit when the text is unchanged', async () => {
    const t = fakeTimer();
    const emitted: string[][] = [];
    const request = vi.fn(async (): Promise<CallResult> => ({ read: { text: 'same' } }));
    startScrollback('p1', {
      request,
      fallback: () => [],
      onLines: (l) => emitted.push(l),
      setInterval: t.setInterval,
      clearInterval: () => {}
    });
    await Promise.resolve();
    await Promise.resolve();
    await t.tick(); // second read, identical text
    expect(emitted).toEqual([['same']]); // only one emit
  });

  it('falls back to the tail on empty text', async () => {
    const emitted: string[][] = [];
    startScrollback('p1', {
      request: async () => ({ read: { text: '' } }),
      fallback: () => ['tail-line'],
      onLines: (l) => emitted.push(l),
      setInterval: () => 1,
      clearInterval: () => {}
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(emitted).toEqual([['tail-line']]);
  });

  it('falls back to the tail when the request throws', async () => {
    const emitted: string[][] = [];
    startScrollback('p1', {
      request: async () => {
        throw new Error('socket down');
      },
      fallback: () => ['from-tail'],
      onLines: (l) => emitted.push(l),
      setInterval: () => 1,
      clearInterval: () => {}
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(emitted).toEqual([['from-tail']]);
  });

  it('stops emitting and clears the interval after teardown', async () => {
    let cleared = false;
    const emitted: string[][] = [];
    const ref: { cb: (() => void) | null } = { cb: null };
    const stop = startScrollback('p1', {
      request: async () => ({ read: { text: 'live' } }),
      fallback: () => [],
      onLines: (l) => emitted.push(l),
      setInterval: (fn) => {
        ref.cb = fn;
        return 7;
      },
      clearInterval: (h) => {
        cleared = h === 7;
      }
    });
    await Promise.resolve();
    await Promise.resolve();
    stop();
    expect(cleared).toBe(true);
    await ref.cb?.(); // a late tick must not emit after teardown
    await Promise.resolve();
    expect(emitted).toEqual([['live']]);
  });
});
