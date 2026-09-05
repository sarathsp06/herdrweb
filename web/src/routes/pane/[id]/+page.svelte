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
  import { parseAnsiLines, segStyle } from '$lib/term/ansi';

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
  .chat { display: flex; flex-direction: column; height: 100%; }
  .scroll { flex: 1; overflow-y: auto; padding: 14px; }
  .cap { font-size: 10.5px; color: var(--text-4); margin-bottom: 8px; }
  /* Terminal scrollback must render exact code points — kill Fira Code's
     contextual ligatures (calt) so ASCII art (-> == != |=> box rules) stays literal. */
  .raw { margin: 0; font-size: 14px; line-height: 1.6; color: var(--text-1); white-space: pre; overflow-x: auto; font-variant-ligatures: none; font-feature-settings: 'liga' 0, 'calt' 0, 'tnum' 1; }
  .raw .ln { display: inline; }
  .missing { padding: 40px; color: var(--text-4); }
</style>
