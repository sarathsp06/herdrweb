import { describe, it, expect } from 'vitest';
import { fitFontSize, widthAtBase } from './fit';

describe('fitFontSize', () => {
  it('returns null when content already fits', () => {
    expect(fitFontSize(400, 400)).toBeNull();
    expect(fitFontSize(400, 200)).toBeNull();
  });

  it('scales font-size down proportionally when content overflows', () => {
    // base 14 * (avail/content) = 14 * 300/420 = 10 (floored)
    expect(fitFontSize(300, 420)).toBe(10);
    // 14 * 362/402 = 12.6 -> 12
    expect(fitFontSize(362, 402)).toBe(12);
  });

  it('clamps to the readable floor for extreme overflow', () => {
    expect(fitFontSize(100, 1000)).toBe(10);
    expect(fitFontSize(100, 1000, 14, 8)).toBe(8);
  });

  it('honours a custom base size', () => {
    expect(fitFontSize(200, 400, 20)).toBe(10);
  });

  it('never scales for a zero-width container', () => {
    expect(fitFontSize(0, 500)).toBeNull();
  });
});

describe('widthAtBase', () => {
  it('is a no-op when currentPx already equals base', () => {
    expect(widthAtBase(420, 14, 14)).toBe(420);
  });

  it('rescales a measurement taken at a smaller applied size back up to base', () => {
    // measured 300px wide at 10px (already fitted); at 14px that content would be 420px
    expect(widthAtBase(300, 10, 14)).toBeCloseTo(420);
  });

  it('produces a stable fit target across repeated re-fits at an already-fitted size', () => {
    // Simulates the flicker bug: without a reset, each poll tick re-measures at
    // whatever size is currently applied. The normalized base-width — and thus
    // the computed target — must stay identical call over call when the
    // underlying content hasn't actually changed width.
    const avail = 300;
    let currentPx = 14;
    let scrollWidth = 420; // fixed underlying content width in real px
    for (let i = 0; i < 5; i++) {
      const content = widthAtBase(scrollWidth, currentPx, 14);
      const size = fitFontSize(avail, content, 14);
      expect(size).toBe(10);
      currentPx = size ?? 14;
      // Re-measuring the same content at the newly-applied size scales the
      // observed scrollWidth proportionally too (monospace advance is linear).
      scrollWidth = (420 * currentPx) / 14;
    }
  });
});
