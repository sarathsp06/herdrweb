<script lang="ts">
  import { get } from 'svelte/store';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { findPaneIn } from '$lib/session/derive';
  import { lastPane, config } from '$lib/ui/state';
  import { startScrollback } from '$lib/session/scrollback';
  import { fitToWidth } from '$lib/layout/fit';
  import { followScroll } from '$lib/layout/scrollFollow';
  import Composer from '$lib/chat/Composer.svelte';

  const s = session();
  const spaces = s.spaces;
  const paneId = $derived(decodeURIComponent($page.params.id ?? ''));
  const ref = $derived(findPaneIn($spaces, paneId));
  const blocked = $derived(ref?.pane.status === 'blocked');

  let raw: string[] = $state([]);

  $effect(() => {
    lastPane.set(paneId);
  });

  // Poll this pane's raw scrollback. The poller owns the interval, dedupe and
  // tail-fallback; keyed only on paneId so unrelated snapshot churn can't refetch.
  $effect(() =>
    startScrollback(paneId, {
      request: (call) => s.request(call),
      fallback: () => findPaneIn(get(spaces), paneId)?.pane.tail ?? [],
      onLines: (lines) => {
        raw = lines;
      }
    })
  );

  function rawClass(line: string): string {
    if (/^[+]/.test(line)) return 'add';
    if (/^[-?]/.test(line)) return 'del';
    if (/pass|ready|done|ok/i.test(line)) return 'ok';
    return '';
  }
</script>

{#if ref}
  <section class="chat">
    <div class="scroll" use:followScroll={{ deps: raw.length, key: paneId }}>
      {#if $config.devCaptions}<div class="cap mono">pane.read · source=recent_unwrapped · lines=200</div>{/if}
      <pre class="raw mono" use:fitToWidth={{ deps: raw }}>{#each raw as line}<span class={rawClass(line)}>{line}
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
