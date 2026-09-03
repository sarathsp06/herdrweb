import { describe, it, expect } from 'vitest';
import { fitFontSize } from './fit';

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
