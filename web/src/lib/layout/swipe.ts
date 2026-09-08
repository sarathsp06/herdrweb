import type { Action } from 'svelte/action';

export type SwipeDirection = 'up' | 'down' | 'left' | 'right';

export interface SwipeParams {
  /** Reported once per completed gesture, dominant axis wins. */
  onSwipe: (dir: SwipeDirection) => void;
  /** While false, native touch behaviour (scroll/zoom/tap) passes through untouched. */
  enabled: boolean;
  /** Minimum travel (px) before a touch counts as a swipe rather than a tap. */
  threshold?: number;
}

/**
 * Pure gesture classification: given the total travel of one touch and a
 * minimum-distance threshold, returns the dominant-axis direction, or null
 * when the travel was too small to count as a swipe (a tap, or scroll noise).
 */
export function resolveSwipe(dx: number, dy: number, threshold: number): SwipeDirection | null {
  if (Math.max(Math.abs(dx), Math.abs(dy)) < threshold) return null;
  if (Math.abs(dx) > Math.abs(dy)) return dx > 0 ? 'right' : 'left';
  return dy > 0 ? 'down' : 'up';
}

/**
 * Svelte action: while `enabled`, single-finger touches on `node` are
 * captured as directional swipes instead of scrolling the page, and reported
 * via `onSwipe`. This never inspects `node`'s content - it is pure
 * gesture-to-direction translation, meant to gate a temporary "direct
 * control" mode so raw terminal navigation (an interactive picker, `fzf`,
 * `vim`, ...) takes one swipe instead of one tap per arrow key. Disabled by
 * default so ordinary scrolling and tapping are untouched until the caller
 * opts in.
 */
export const swipe: Action<HTMLElement, SwipeParams> = (node, params) => {
  let p = params;
  let startX = 0;
  let startY = 0;
  let tracking = false;

  const onTouchStart = (e: TouchEvent) => {
    if (!p.enabled || e.touches.length !== 1) {
      tracking = false;
      return;
    }
    tracking = true;
    startX = e.touches[0].clientX;
    startY = e.touches[0].clientY;
  };
  const onTouchMove = (e: TouchEvent) => {
    if (!p.enabled || !tracking) return;
    e.preventDefault(); // suppress native scroll while a gesture is live
  };
  const onTouchEnd = (e: TouchEvent) => {
    if (!p.enabled || !tracking) return;
    tracking = false;
    const t = e.changedTouches[0];
    if (!t) return;
    const dir = resolveSwipe(t.clientX - startX, t.clientY - startY, p.threshold ?? 32);
    if (dir) p.onSwipe(dir);
  };
  const onTouchCancel = () => {
    tracking = false;
  };

  node.addEventListener('touchstart', onTouchStart, { passive: true });
  node.addEventListener('touchmove', onTouchMove, { passive: false });
  node.addEventListener('touchend', onTouchEnd, { passive: true });
  node.addEventListener('touchcancel', onTouchCancel, { passive: true });

  return {
    update(next: SwipeParams) {
      p = next;
    },
    destroy() {
      node.removeEventListener('touchstart', onTouchStart);
      node.removeEventListener('touchmove', onTouchMove);
      node.removeEventListener('touchend', onTouchEnd);
      node.removeEventListener('touchcancel', onTouchCancel);
    }
  };
};
