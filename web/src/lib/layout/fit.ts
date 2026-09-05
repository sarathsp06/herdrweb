import type { Action } from 'svelte/action';
import { BREAKPOINT } from './responsive';

/**
 * The scaled font-size (px) that makes monospace `content` fit within `avail`,
 * or null when it already fits. `content` MUST be measured at `base` font-size —
 * monospace advance scales linearly with font-size, so measuring the widest line
 * at `base` gives an exact one-pass fit; clamp to a readable `floor` and let
 * anything still wider scroll horizontally.
 */
export function fitFontSize(avail: number, content: number, base = 14, floor = 10): number | null {
  if (avail > 0 && content > avail) {
    return Math.max(floor, Math.floor(base * (avail / content)));
  }
  return null;
}

/**
 * Rescale a width measured at `currentPx` to its equivalent at `base` font-size,
 * so `fitFontSize` can be fed a live measurement without first resetting the
 * node's font-size back to `base` (that reset-then-remeasure is what caused the
 * visible flicker: every reactive update forced the node back to full size,
 * read the layout, then reapplied the fit — a resize flash on every poll tick).
 */
export function widthAtBase(content: number, currentPx: number, base = 14): number {
  return currentPx > 0 ? (content * base) / currentPx : content;
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
  const base = 14;

  const apply = () => {
    if (window.innerWidth >= bp) {
      if (node.style.fontSize) node.style.fontSize = '';
      return;
    }
    // Measure at whatever size is already applied (or `base` on first run) and
    // normalize to base-equivalent width — never reset first, so an
    // already-fitted node doesn't flash back to full size every re-fit.
    const currentPx = parseFloat(node.style.fontSize) || base;
    const content = widthAtBase(node.scrollWidth, currentPx, base);
    const size = fitFontSize(node.clientWidth, content, base);
    const next = size !== null ? size + 'px' : '';
    if (node.style.fontSize !== next) node.style.fontSize = next;
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
