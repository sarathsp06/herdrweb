<script lang="ts">
  import { tick } from 'svelte';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { findPaneIn } from '$lib/session/derive';
  import { mode, lastPane, config } from '$lib/ui/state';
  import { getTranscript } from '$lib/chat/transcripts';
  import Composer from '$lib/chat/Composer.svelte';
  import UserBubble from '$lib/chat/UserBubble.svelte';
  import Reasoning from '$lib/chat/Reasoning.svelte';
  import AgentText from '$lib/chat/AgentText.svelte';
  import ToolCall from '$lib/chat/ToolCall.svelte';
  import CodeBlock from '$lib/chat/CodeBlock.svelte';
  import DiffSummary from '$lib/chat/DiffSummary.svelte';
  import BlockedCard from '$lib/chat/BlockedCard.svelte';
  import Working from '$lib/chat/Working.svelte';

  const s = session();
  const spaces = s.spaces;
  const paneId = $derived(decodeURIComponent($page.params.id ?? ''));
  const ref = $derived(findPaneIn($spaces, paneId));
  const transcript = $derived(getTranscript(paneId));
  const blocked = $derived(ref?.pane.status === 'blocked');

  let scroller: HTMLElement | undefined = $state();
  let raw: string[] = $state([]);

  // Non-agent panes, and live agent panes with no transcript feed, open in raw
  // mode (raw scrollback is the only content until a live transcript exists).
  $effect(() => {
    lastPane.set(paneId);
    if (ref && (!ref.pane.agent || transcript.length === 0)) mode.set('raw');
  });

  // Raw scrollback via pane.read (Herdr returns {read:{text}}).
  $effect(() => {
    if ($mode !== 'raw' || !ref) return;
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
  // Reset to bottom-follow when switching pane or mode.
  $effect(() => {
    void paneId;
    void $mode;
    pinned = true;
  });
  // Auto-scroll after paint on content change, but only while pinned.
  $effect(() => {
    void transcript.length;
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
      {#if $mode === 'raw'}
        {#if $config.devCaptions}<div class="cap mono">pane.read · source=recent_unwrapped · lines=200</div>{/if}
        <pre class="raw mono">{#each raw as line}<span class={rawClass(line)}>{line}
</span>{/each}</pre>
      {:else}
        <div class="transcript">
          {#each transcript as m (m.id)}
            {#if m.kind === 'user'}<UserBubble text={m.text} caption={m.caption} />
            {:else if m.kind === 'reasoning'}<Reasoning summary={m.summary} body={m.body} />
            {:else if m.kind === 'agent'}<AgentText text={m.text} />
            {:else if m.kind === 'tool'}<ToolCall name={m.name} arg={m.arg} result={m.result} ok={m.ok} output={m.output} />
            {:else if m.kind === 'code'}<CodeBlock path={m.path} lang={m.lang} code={m.code} />
            {:else if m.kind === 'diff'}<DiffSummary paneId={ref.pane.id} files={m.files} add={m.add} del={m.del} preview={m.preview} />
            {:else if m.kind === 'blocked'}<BlockedCard paneId={ref.pane.id} question={m.question} terminal={m.terminal} age={m.age} />
            {:else if m.kind === 'working'}<Working label={m.label} />
            {/if}
          {/each}
        </div>
      {/if}
    </div>

    <Composer paneId={ref.pane.id} {blocked} />
  </section>
{:else}
  <div class="missing mono">pane {paneId} not found</div>
{/if}

<style>
  .chat { display: flex; flex-direction: column; height: 100%; }
  .scroll { flex: 1; overflow-y: auto; padding: 14px; }
  .transcript { display: flex; flex-direction: column; gap: 14px; max-width: 860px; margin: 0 auto; }
  .cap { font-size: 10.5px; color: var(--text-4); margin-bottom: 8px; }
  .raw { margin: 0; font-size: 14px; line-height: 1.6; color: var(--text-1); white-space: pre; overflow-x: auto; }
  .raw .add { color: var(--done); } .raw .del { color: var(--blocked-badge-text); } .raw .ok { color: var(--working); }
  .missing { padding: 40px; color: var(--text-4); }
</style>
