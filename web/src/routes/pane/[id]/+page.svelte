<script lang="ts">
  import { tick } from 'svelte';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { findPaneIn } from '$lib/session/derive';
  import { lastPane, config } from '$lib/ui/state';
  import Composer from '$lib/chat/Composer.svelte';

  const s = session();
  const spaces = s.spaces;
  const paneId = $derived(decodeURIComponent($page.params.id ?? ''));
  const ref = $derived(findPaneIn($spaces, paneId));
  const blocked = $derived(ref?.pane.status === 'blocked');

  let scroller: HTMLElement | undefined = $state();
  let raw: string[] = $state([]);

  $effect(() => {
    lastPane.set(paneId);
  });

  // Raw scrollback via pane.read (Herdr returns {read:{text}}), refreshed as the
  // snapshot changes so the terminal tails live.
  $effect(() => {
    if (!ref) return;
    s.request({ method: 'pane.read', params: { pane_id: paneId, source: 'recent_unwrapped', lines: 200 } })
      .then((r) => {
        const text = r.read?.text ?? '';
        raw = text ? text.split('\n') : ref!.pane.tail;
      })
      .catch(() => (raw = ref!.pane.tail));
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
      <pre class="raw mono">{#each raw as line}<span class={rawClass(line)}>{line}
</span>{/each}</pre>
    </div>

    <Composer paneId={ref.pane.id} {blocked} />
  </section>
{:else}
  <div class="missing mono">pane {paneId} not found</div>
{/if}

<style>
  .chat { display: flex; flex-direction: column; height: 100%; }
  .scroll { flex: 1; overflow-y: auto; padding: 14px; }
  .cap { font-size: 10.5px; color: var(--text-4); margin-bottom: 8px; }
  .raw { margin: 0; font-size: 14px; line-height: 1.6; color: var(--text-1); white-space: pre; overflow-x: auto; }
  .raw .add { color: var(--done); } .raw .del { color: var(--blocked-badge-text); } .raw .ok { color: var(--working); }
  .missing { padding: 40px; color: var(--text-4); }
</style>
