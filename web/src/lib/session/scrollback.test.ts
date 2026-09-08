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

  it('does not fall back to the tail when the first read succeeds but is empty', async () => {
    const emitted: string[][] = [];
    const request = vi.fn(async (): Promise<CallResult> => ({ read: { text: '' } }));
    startScrollback('p1', {
      request,
      fallback: () => ['tail-line'],
      onLines: (l) => emitted.push(l),
      setInterval: () => 1,
      clearInterval: () => {}
    });
    await Promise.resolve();
    await Promise.resolve();
    // A successful empty reply is real content ("no output yet"), which
    // matches the poller's own [] starting state - nothing new to emit.
    expect(request).toHaveBeenCalledTimes(1);
    expect(emitted).toEqual([]);
  });

  it('emits the empty state when a successful reply reports the pane cleared', async () => {
    const t = fakeTimer();
    const emitted: string[][] = [];
    let text = 'one\ntwo\nthree';
    startScrollback('p1', {
      request: async () => ({ read: { text } }),
      fallback: () => ['tail-line'],
      onLines: (l) => emitted.push(l),
      setInterval: t.setInterval,
      clearInterval: () => {}
    });
    await Promise.resolve();
    await Promise.resolve();
    text = ''; // the backend affirmatively reports zero matching lines
    await t.tick();
    text = 'one\ntwo\nthree'; // and the pane resumes producing output
    await t.tick();
    // Both transitions are real content, never the fallback tail.
    expect(emitted).toEqual([['one', 'two', 'three'], [], ['one', 'two', 'three']]);
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

  it('keeps the last good read when the request throws after a real read landed', async () => {
    const t = fakeTimer();
    const emitted: string[][] = [];
    let fail = false;
    startScrollback('p1', {
      request: async () => {
        if (fail) throw new Error('socket closed');
        return { read: { text: 'live output' } };
      },
      fallback: () => ['tail-line'],
      onLines: (l) => emitted.push(l),
      setInterval: t.setInterval,
      clearInterval: () => {}
    });
    await Promise.resolve();
    await Promise.resolve();
    fail = true; // e.g. every pending call rejects on a WebSocket reconnect
    await t.tick();
    await t.tick();
    expect(emitted).toEqual([['live output']]);
  });

  it('skips a tick instead of overlapping when the previous request is still in flight', async () => {
    const t = fakeTimer();
    const emitted: string[][] = [];
    let calls = 0;
    const first = Promise.withResolvers<CallResult>();
    const request = vi.fn((): Promise<CallResult> => {
      calls++;
      if (calls === 1) return first.promise;
      return Promise.resolve({ read: { text: 'second' } });
    });
    startScrollback('p1', {
      request,
      fallback: () => [],
      onLines: (l) => emitted.push(l),
      setInterval: t.setInterval,
      clearInterval: () => {}
    });
    await Promise.resolve();
    // First read is still unresolved; a tick firing now must not start a
    // second overlapping request that could resolve out of order.
    await t.tick();
    expect(calls).toBe(1);
    first.resolve({ read: { text: 'first' } });
    await Promise.resolve();
    await Promise.resolve();
    expect(emitted).toEqual([['first']]);
    // Now the in-flight request has resolved; the next tick is free to fire.
    await t.tick();
    await Promise.resolve();
    await Promise.resolve();
    expect(calls).toBe(2);
    expect(emitted).toEqual([['first'], ['second']]);
  });
});
