<script lang="ts">
  import { tick } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { findPaneIn } from '$lib/session/derive';
  import { mode, lastPane, config } from '$lib/ui/state';
  import { getTranscript } from '$lib/chat/transcripts';
  import StatusPill from '$lib/ui/StatusPill.svelte';
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

  // Non-agent panes open directly in raw mode.
  $effect(() => {
    lastPane.set(paneId);
    if (ref && !ref.pane.agent) mode.set('raw');
  });

  // Raw scrollback via pane.read.
  $effect(() => {
    if ($mode !== 'raw' || !ref) return;
    s.request({ method: 'pane.read', params: { target: paneId, source: 'recent-unwrapped', lines: 80 } })
      .then((r) => (raw = (r.lines as string[]) ?? ref!.pane.tail))
      .catch(() => (raw = ref!.pane.tail));
  });

  // Auto-scroll to bottom after paint on pane / mode / transcript change.
  $effect(() => {
    void paneId; void $mode; void transcript.length; void raw.length;
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
    <header class="bar">
      <button class="back" onclick={() => goto('/')} aria-label="back">‹</button>
      <div class="title">
        <div class="line1"><span class="name">{ref.pane.label}</span> <StatusPill status={ref.pane.status} /></div>
        <div class="line2 mono">{ref.space.label} · {ref.pane.id}</div>
      </div>
      <div class="switch">
        <button class:active={$mode === 'chat'} onclick={() => mode.set('chat')} disabled={!ref.pane.agent}>chat</button>
        <button class:active={$mode === 'raw'} onclick={() => mode.set('raw')}>raw</button>
      </div>
    </header>

    <div class="scroll" bind:this={scroller}>
      {#if $mode === 'raw'}
        {#if $config.devCaptions}<div class="cap mono">pane.read · source=recent-unwrapped · lines=80</div>{/if}
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
  .chat { display: flex; flex-direction: column; height: 100vh; }
  .bar { display: flex; align-items: center; gap: 10px; padding: 10px 14px; border-bottom: 1px solid var(--hairline); }
  .back { width: 32px; height: 32px; border-radius: var(--r-chip); border: none; background: none; color: var(--text-2); font-size: 22px; }
  .title { flex: 1; min-width: 0; }
  .line1 { display: flex; align-items: center; gap: 8px; font-size: 14px; font-weight: 600; }
  .line2 { font-size: 11px; color: var(--text-4); }
  .switch { display: flex; background: var(--code-surface); border: 1px solid var(--control); border-radius: var(--r-chip); overflow: hidden; }
  .switch button { min-height: 32px; padding: 0 12px; background: none; border: none; color: var(--text-3); font-size: 12px; }
  .switch button.active { background: var(--text-on-light); color: var(--text-1); background: #18181b; }
  .scroll { flex: 1; overflow-y: auto; padding: 14px; }
  .transcript { display: flex; flex-direction: column; gap: 14px; max-width: 860px; margin: 0 auto; }
  .cap { font-size: 10.5px; color: var(--text-4); margin-bottom: 8px; }
  .raw { margin: 0; font-size: 10.5px; line-height: 1.6; color: var(--text-3); white-space: pre; overflow-x: auto; }
  .raw .add { color: var(--done); } .raw .del { color: var(--blocked-badge-text); } .raw .ok { color: var(--working); }
  .missing { padding: 40px; color: var(--text-4); }
  @media (min-width: 880px) { .chat { height: 100vh; } }
</style>
