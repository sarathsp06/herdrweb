import type { Action } from 'svelte/action';

/**
 * True when the viewport is within `threshold` px of the bottom — i.e. the user
 * is following live output rather than reading history further up.
 */
export function isPinned(scrollHeight: number, scrollTop: number, clientHeight: number, threshold = 40): boolean {
  return scrollHeight - scrollTop - clientHeight < threshold;
}

export interface FollowParams {
  /** Reference the reactive content so the action scrolls after new output. */
  deps?: unknown;
  /** When this changes (e.g. pane id), re-pin to the bottom. */
  key?: unknown;
  /** Distance from the bottom still counted as "following". */
  threshold?: number;
}

/**
 * Svelte action: follow new output only while the user is already at the bottom,
 * so scrolling up to read history is not yanked back down by live refreshes.
 * Changing `key` (pane switch) re-pins to the bottom.
 */
export const followScroll: Action<HTMLElement, FollowParams | undefined> = (node, params) => {
  let threshold = params?.threshold ?? 40;
  let key = params?.key;
  let pinned = true;

  const onScroll = () => {
    pinned = isPinned(node.scrollHeight, node.scrollTop, node.clientHeight, threshold);
  };
  const toBottom = () =>
    requestAnimationFrame(() => {
      if (pinned) node.scrollTop = node.scrollHeight;
    });

  node.addEventListener('scroll', onScroll, { passive: true });
  toBottom();

  return {
    update(next?: FollowParams) {
      threshold = next?.threshold ?? 40;
      if (next?.key !== key) {
        key = next?.key;
        pinned = true;
      }
      toBottom();
    },
    destroy() {
      node.removeEventListener('scroll', onScroll);
    }
  };
};
