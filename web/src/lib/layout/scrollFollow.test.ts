import { describe, it, expect } from 'vitest';
import { isPinned } from './scrollFollow';

describe('isPinned', () => {
  it('is pinned at the exact bottom', () => {
    expect(isPinned(1000, 800, 200)).toBe(true);
  });

  it('is pinned within the default threshold of the bottom', () => {
    expect(isPinned(1000, 770, 200)).toBe(true); // 30px from bottom < 40
  });

  it('is not pinned when scrolled up beyond the threshold', () => {
    expect(isPinned(1000, 700, 200)).toBe(false); // 100px from bottom
  });

  it('honours a custom threshold', () => {
    expect(isPinned(1000, 700, 200, 150)).toBe(true);
  });
});
