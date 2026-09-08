import { describe, it, expect } from 'vitest';
import { resolveSwipe } from './swipe';

describe('resolveSwipe', () => {
  it('is null below the threshold (a tap, not a swipe)', () => {
    expect(resolveSwipe(10, -5, 32)).toBeNull();
  });

  it('counts exactly at the threshold as a swipe (reaching qualifies, not just exceeding)', () => {
    expect(resolveSwipe(32, 0, 32)).toBe('right');
  });

  it('reports right for a rightward-dominant swipe', () => {
    expect(resolveSwipe(50, 5, 32)).toBe('right');
  });

  it('reports left for a leftward-dominant swipe', () => {
    expect(resolveSwipe(-50, -5, 32)).toBe('left');
  });

  it('reports down for a downward-dominant swipe', () => {
    expect(resolveSwipe(5, 50, 32)).toBe('down');
  });

  it('reports up for an upward-dominant swipe', () => {
    expect(resolveSwipe(-5, -50, 32)).toBe('up');
  });

  it('breaks a horizontal/vertical tie toward the vertical axis', () => {
    // |dx| > |dy| is required to pick a horizontal direction; equal deltas
    // fall through to vertical, matching the dominant-axis "> not >=" check.
    expect(resolveSwipe(40, 40, 32)).toBe('down');
  });

  it('honours a custom threshold', () => {
    expect(resolveSwipe(10, 0, 5)).toBe('right');
    expect(resolveSwipe(10, 0, 20)).toBeNull();
  });
});
