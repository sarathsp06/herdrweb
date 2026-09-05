<script lang="ts">
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { findPaneIn, primaryPaneOfTab, tabHasBlocked } from '$lib/session/derive';
  import { lastPane, rememberTab, config } from '$lib/ui/state';
  import { startScrollback } from '$lib/session/scrollback';
  import { fitToWidth } from '$lib/layout/fit';
  import { followScroll } from '$lib/layout/scrollFollow';
  import Composer from '$lib/chat/Composer.svelte';
  import { parseAnsiLines, segStyle } from '$lib/term/ansi';

  const s = session();
  const spaces = s.spaces;
  const paneId = $derived(decodeURIComponent($page.params.id ?? ''));
  const ref = $derived(findPaneIn($spaces, paneId));
  const blocked = $derived(ref?.pane.status === 'blocked');

  let raw: string[] = $state([]);

  $effect(() => {
    lastPane.set(paneId);
    if (ref) rememberTab(ref.space.id, ref.tab.id);
  });

  // Switch tabs from the chat header: jump to the target tab's primary pane.
  function openTab(tabId: string) {
    const ref2 = findPaneIn(get(spaces), paneId);
    const tab = ref2?.space.tabs.find((t) => t.id === tabId);
    const target = tab && primaryPaneOfTab(tab);
    if (target && target !== paneId) goto(`/pane/${encodeURIComponent(target)}`);
  }

  // Poll this pane's raw scrollback. The poller owns the interval, dedupe and
  // tail-fallback; keyed only on paneId so unrelated snapshot churn can't refetch.
  $effect(() =>
    startScrollback(paneId, {
      request: (call) => s.request(call),
      fallback: () => findPaneIn(get(spaces), paneId)?.pane.tail ?? [],
      ansi: $config.ansi,
      onLines: (lines) => {
        raw = lines;
      }
    })
  );

  // When the ANSI toggle is on we requested `format: 'ansi'`, so parse SGR into
  // styled segments (state carries across lines). Off = plain text, no colour.
  const rows = $derived($config.ansi ? parseAnsiLines(raw) : null);
</script>

{#if ref}
  <section class="chat">
    <div class="tabs">
      {#each ref.space.tabs as t (t.id)}
        <button class="tab mono" class:active={t.id === ref.tab.id} onclick={() => openTab(t.id)}>
          {t.label}
          {#if tabHasBlocked(t)}<span class="bdot"></span>{/if}
          <span class="pc">{t.panes.length}</span>
        </button>
      {/each}
    </div>

    <div class="scroll" use:followScroll={{ deps: raw.length, key: paneId }}>
      {#if $config.devCaptions}<div class="cap mono">pane.read · source=recent_unwrapped · lines=200{$config.ansi ? ' · format=ansi' : ''}</div>{/if}
      <pre class="raw mono" use:fitToWidth={{ deps: raw }}>{#if rows}{#each rows as segs}<span class="ln">{#each segs as seg}<span style={segStyle(seg.sgr)}>{seg.text}</span>{/each}
</span>{/each}{:else}{#each raw as line}<span class="ln">{line}
</span>{/each}{/if}</pre>
    </div>

    <Composer paneId={ref.pane.id} {blocked} agent={ref.pane.agent} />
  </section>
{:else}
  <div class="missing mono">pane {paneId} not found</div>
{/if}

<style>
  .tabs { flex: none; display: flex; gap: 6px; overflow-x: auto; padding: 10px 14px; border-bottom: 1px solid var(--hairline); }
  .tab { flex: none; display: flex; align-items: center; gap: 6px; min-height: 40px; padding: 0 12px; border-radius: var(--r-chip); border: 1px solid var(--control); background: var(--card); color: var(--text-3); font-size: 12px; }
  .tab.active { border-color: var(--control-selected-2); background: var(--surface-tint-2); color: var(--text-1); }
  .bdot { width: 6px; height: 6px; border-radius: 50%; background: var(--blocked); }
  .pc { color: var(--text-4); }
  .chat { display: flex; flex-direction: column; height: 100%; }
  .scroll { flex: 1; overflow-y: auto; padding: 14px; }
  .cap { font-size: 10.5px; color: var(--text-4); margin-bottom: 8px; }
  /* Terminal scrollback must render exact code points — kill Fira Code's
     contextual ligatures (calt) so ASCII art (-> == != |=> box rules) stays literal. */
  .raw { margin: 0; font-size: 14px; line-height: 1.6; color: var(--text-1); white-space: pre; overflow-x: auto; font-variant-ligatures: none; font-feature-settings: 'liga' 0, 'calt' 0, 'tnum' 1; }
  .raw .ln { display: inline; }
  .missing { padding: 40px; color: var(--text-4); }
</style>
