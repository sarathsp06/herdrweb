import { readable } from 'svelte/store';
import { browser } from '$app/environment';

export const BREAKPOINT = 880;

export const width = readable(browser ? window.innerWidth : 1200, (set) => {
  if (!browser) return;
  let last = window.innerWidth;
  set(last);
  const onResize = () => {
    if (Math.abs(window.innerWidth - last) < 8) return;
    last = window.innerWidth;
    set(last);
  };
  window.addEventListener('resize', onResize);
  return () => window.removeEventListener('resize', onResize);
});
