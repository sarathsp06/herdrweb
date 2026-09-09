<script lang="ts">
  import { get } from 'svelte/store';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { findPaneIn } from '$lib/session/derive';
  import { lastPane, rememberTab, config, showToast } from '$lib/ui/state';
  import { startScrollback } from '$lib/session/scrollback';
  import { fitToWidth } from '$lib/layout/fit';
  import { followScroll } from '$lib/layout/scrollFollow';
  import { swipe, type SwipeDirection } from '$lib/layout/swipe';
  import Composer from '$lib/chat/Composer.svelte';
  import PaneHeader from '$lib/screens/PaneHeader.svelte';
  import { parseAnsiLines, segStyle } from '$lib/term/ansi';
  import type { Call } from '$lib/protocol';

  const s = session();
  const spaces = s.spaces;
  const paneId = $derived(decodeURIComponent($page.params.id ?? ''));
  const ref = $derived(findPaneIn($spaces, paneId));
  const blocked = $derived(ref?.pane.status === 'blocked');

  let raw: string[] = $state([]);
  // Direct control: while on, swiping the transcript sends arrow keys to the
  // pane instead of scrolling - one swipe per keypress, no interpretation of
  // what's on screen. Off by default, and reset on every pane switch so a
  // gesture never accidentally reaches a different pane.
  let controlMode = $state(false);
  // Recomputed on every ref change (unlike controlMode itself): if the pane
  // stops being an agent mid-session (process exits, replaced) while control
  // mode is still on, this drops immediately so the transcript reverts to
  // native scroll instead of silently eating touches with nowhere to send them.
  const swipeEnabled = $derived(controlMode && !!ref?.pane.agent);

  // Reset only when the pane identity actually changes - not on every
  // snapshot/ref update (agent status ticks, tab renames, ...), which would
  // otherwise silently flip direct control back off mid-use.
  $effect(() => {
    paneId;
    controlMode = false;
  });

  $effect(() => {
    lastPane.set(paneId);
    if (ref) rememberTab(ref.space.id, ref.tab.id);
  });

  // Direct control is agent-only: it exists to speed up navigating an
  // agent's interactive TUI (e.g. `/model`), so it only ever talks to
  // `agent.send_keys`, never a plain terminal pane.
  function sendSwipeKey(dir: SwipeDirection) {
    if (!ref?.pane.agent) return;
    const call: Call = { method: 'agent.send_keys', params: { target: paneId, keys: [dir] } };
    void s.request(call).catch(() => {});
    showToast(`swipe → ${dir}`);
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
  // Raw terminal output assumes a dark terminal; on a light theme, segStyle
  // neutralises the near-black backgrounds it paints for its own UI chrome.
  const isLight = $derived($config.theme === 'solarized-light' || $config.theme === 'paper');
</script>

{#if ref}
  <section class="flex h-full flex-col">
    <PaneHeader agentRef={ref} />

    <div
      class="scroll flex-1 overflow-y-auto p-3.5 {swipeEnabled ? 'touch-none outline-2 outline-ring -outline-offset-2' : ''}"
      use:followScroll={{ deps: raw.length, key: paneId }}
      use:swipe={{ enabled: swipeEnabled, onSwipe: sendSwipeKey }}
    >
      {#if $config.devCaptions}<div class="mono mb-2 text-[10.5px] text-muted-foreground">pane.read · source=recent_unwrapped · lines=200{$config.ansi ? ' · format=ansi' : ''}</div>{/if}
      <pre class="raw mono" use:fitToWidth={{ deps: raw }}>{#if rows}{#each rows as segs}<span class="ln">{#each segs as seg}<span style={segStyle(seg.sgr, isLight)}>{seg.text}</span>{/each}
</span>{/each}{:else}{#each raw as line}<span class="ln">{line}
</span>{/each}{/if}</pre>
    </div>

    <Composer
      paneId={ref.pane.id}
      {blocked}
      agent={ref.pane.agent}
      {controlMode}
      ontogglecontrol={() => (controlMode = !controlMode)}
    />
  </section>
{:else}
  <div class="mono p-10 text-muted-foreground">pane {paneId} not found</div>
{/if}

<style>
  /* Terminal scrollback must render exact code points — kill Fira Code's
     contextual ligatures (calt) so ASCII art (-> == != |=> box rules) stays literal. */
  .raw { margin: 0; font-size: 14px; line-height: 1.6; color: var(--text-1); white-space: pre; overflow-x: auto; font-variant-ligatures: none; font-feature-settings: 'liga' 0, 'calt' 0, 'tnum' 1; }
  .raw .ln { display: inline; }
</style>
