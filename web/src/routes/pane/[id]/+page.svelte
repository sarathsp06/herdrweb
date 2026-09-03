<script lang="ts">
  import { tick } from 'svelte';
  import { get } from 'svelte/store';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { findPaneIn } from '$lib/session/derive';
  import { lastPane, config } from '$lib/ui/state';
  import { BREAKPOINT } from '$lib/layout/responsive';
  import Composer from '$lib/chat/Composer.svelte';

  const s = session();
  const spaces = s.spaces;
  const paneId = $derived(decodeURIComponent($page.params.id ?? ''));
  const ref = $derived(findPaneIn($spaces, paneId));
  const blocked = $derived(ref?.pane.status === 'blocked');

  let scroller: HTMLElement | undefined = $state();
  let pre: HTMLElement | undefined = $state();
  let raw: string[] = $state([]);

  $effect(() => {
    lastPane.set(paneId);
  });

  // Fit ASCII/box output to the viewport on phones so a diagram keeps its
  // columns on every screen size instead of soft-wrapping into garbage.
  // Monospace advance scales linearly with font-size, so measuring the widest
  // line at base size gives an exact fit in one pass; clamp to a readable floor
  // and let anything still wider scroll horizontally. Desktop keeps full columns.
  function fitRaw(): void {
    const el = pre;
    if (!el) return;
    el.style.fontSize = '';
    if (window.innerWidth >= BREAKPOINT) return;
    const avail = el.clientWidth;
    const content = el.scrollWidth; // reading layout flushes the reset above
    if (avail > 0 && content > avail) {
      el.style.fontSize = Math.max(10, Math.floor(14 * (avail / content))) + 'px';
    }
  }
  // Re-fit on content change (after paint) and on container resize (rotation,
  // split, a differently-sized phone). Observe the scroller, not the <pre> —
  // the pre's height tracks font-size and would feed back into a loop.
  $effect(() => {
    void raw;
    void paneId;
    void tick().then(fitRaw);
  });
  $effect(() => {
    const box = scroller;
    if (!box) return;
    const ro = new ResizeObserver(() => fitRaw());
    ro.observe(box);
    return () => ro.disconnect();
  });

  // Raw scrollback via pane.read (Herdr returns {read:{text}}). There is no push
  // event for plain terminal output, so we poll on a fixed cadence keyed ONLY on
  // paneId — never on the session store — so unrelated snapshot churn (other
  // panes, status ticks) cannot re-fetch or re-render this view. `raw` is only
  // reassigned when the text actually changed, so the terminal never refreshes
  // unconditionally.
  function sameLines(a: string[], b: string[]): boolean {
    if (a.length !== b.length) return false;
    for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
    return true;
  }
  $effect(() => {
    const id = paneId;
    let alive = true;
    const readOnce = async () => {
      let next: string[];
      try {
        const r = await s.request({ method: 'pane.read', params: { pane_id: id, source: 'recent_unwrapped', lines: 200 } });
        const text = r.read?.text ?? '';
        next = text ? text.split('\n') : (findPaneIn(get(spaces), id)?.pane.tail ?? []);
      } catch {
        next = findPaneIn(get(spaces), id)?.pane.tail ?? [];
      }
      if (alive && !sameLines(next, raw)) raw = next;
    };
    void readOnce();
    const iv = setInterval(readOnce, 1000);
    return () => { alive = false; clearInterval(iv); };
  });

  // Follow new output only when the user is already at the bottom, so scrolling
  // up to read history is not yanked back down by live refreshes.
  let pinned = $state(true);
  function onScroll() {
    const el = scroller;
    if (!el) return;
    pinned = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
  }
  // Reset to bottom-follow when switching pane.
  $effect(() => {
    void paneId;
    pinned = true;
  });
  // Auto-scroll after paint on content change, but only while pinned.
  $effect(() => {
    void raw.length;
    if (!pinned) return;
    tick().then(() => { if (scroller) scroller.scrollTop = scroller.scrollHeight; });
  });

  function rawClass(line: string): string {
    if (/^[+]/.test(line)) return 'add';
    if (/^[-?]/.test(line)) return 'del';
    if (/pass|ready|done|ok/i.test(line)) return 'ok';
    return '';
  }
</script>

{#if ref}
  <section class="chat">
    <div class="scroll" bind:this={scroller} onscroll={onScroll}>
      {#if $config.devCaptions}<div class="cap mono">pane.read · source=recent_unwrapped · lines=200</div>{/if}
      <pre class="raw mono" bind:this={pre}>{#each raw as line}<span class={rawClass(line)}>{line}
</span>{/each}</pre>
    </div>

    <Composer paneId={ref.pane.id} {blocked} agent={ref.pane.agent} />
  </section>
{:else}
  <div class="missing mono">pane {paneId} not found</div>
{/if}

<style>
  .chat { display: flex; flex-direction: column; height: 100%; }
  .scroll { flex: 1; overflow-y: auto; padding: 14px; }
  .cap { font-size: 10.5px; color: var(--text-4); margin-bottom: 8px; }
  /* Terminal scrollback must render exact code points — kill Fira Code's
     contextual ligatures (calt) so ASCII art (-> == != |=> box rules) stays literal. */
  .raw { margin: 0; font-size: 14px; line-height: 1.6; color: var(--text-1); white-space: pre; overflow-x: auto; font-variant-ligatures: none; font-feature-settings: 'liga' 0, 'calt' 0, 'tnum' 1; }
  .raw .add { color: var(--done); } .raw .del { color: var(--blocked-badge-text); } .raw .ok { color: var(--working); }
  .missing { padding: 40px; color: var(--text-4); }
</style>
