import type { Action } from 'svelte/action';
import { BREAKPOINT } from './responsive';

/**
 * The scaled font-size (px) that makes monospace `content` fit within `avail`,
 * or null when it already fits. Monospace advance scales linearly with
 * font-size, so measuring the widest line at `base` gives an exact one-pass fit;
 * clamp to a readable `floor` and let anything still wider scroll horizontally.
 */
export function fitFontSize(avail: number, content: number, base = 14, floor = 10): number | null {
  if (avail > 0 && content > avail) {
    return Math.max(floor, Math.floor(base * (avail / content)));
  }
  return null;
}

export interface FitParams {
  /** Container whose width bounds the fit; defaults to the node's parent. */
  observe?: HTMLElement;
  /** Below this viewport width fitting applies; at/above it, full columns. */
  breakpoint?: number;
  /** Reference the reactive content so the action re-fits when lines change. */
  deps?: unknown;
}

/**
 * Svelte action: fit ASCII/box output to the viewport on phones so a diagram
 * keeps its columns on every screen size instead of soft-wrapping into garbage.
 * Observes the container (font-independent width) — never the node itself, whose
 * height tracks font-size and would feed back into a resize loop. Desktop
 * (viewport >= breakpoint) keeps full-size columns.
 */
export const fitToWidth: Action<HTMLElement, FitParams | undefined> = (node, params) => {
  let bp = params?.breakpoint ?? BREAKPOINT;
  let observed: HTMLElement = params?.observe ?? node.parentElement ?? node;

  const apply = () => {
    node.style.fontSize = '';
    if (window.innerWidth >= bp) return;
    const size = fitFontSize(node.clientWidth, node.scrollWidth); // reading flushes the reset
    if (size !== null) node.style.fontSize = size + 'px';
  };
  const refit = () => requestAnimationFrame(apply);

  let ro = new ResizeObserver(refit);
  ro.observe(observed);
  refit();

  return {
    update(next?: FitParams) {
      bp = next?.breakpoint ?? BREAKPOINT;
      const target = next?.observe ?? node.parentElement ?? node;
      if (target !== observed) {
        ro.disconnect();
        observed = target;
        ro = new ResizeObserver(refit);
        ro.observe(observed);
      }
      refit();
    },
    destroy() {
      ro.disconnect();
    }
  };
};
