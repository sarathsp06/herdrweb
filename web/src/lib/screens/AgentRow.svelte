<script lang="ts">
  import type { AgentRef } from '$lib/session/derive';
  import StatusGlyph from '$lib/ui/StatusGlyph.svelte';

  let { agent, sub, onopen }: { agent: AgentRef; sub: string; onopen: () => void } = $props();
</script>

<button class="arow" onclick={onopen}>
  <StatusGlyph status={agent.pane.status} />
  <span class="col">
    <span class="mono name">{agent.pane.label} <span class="pid">{agent.pane.id}</span></span>
    <span class="sub">{sub}</span>
  </span>
  <span class="word" data-status={agent.pane.status}>{agent.pane.status}</span>
</button>

<style>
  .arow {
    width: 100%; display: flex; align-items: center; gap: 10px;
    min-height: 52px; padding: 6px 8px;
    background: none; border: none; border-radius: var(--r-chip); text-align: left;
  }
  .arow:hover { background: var(--surface-tint); }
  .col { display: flex; flex-direction: column; min-width: 0; flex: 1; gap: 1px; }
  .name { font-size: 13.5px; font-weight: 500; }
  .pid { color: var(--text-4); font-size: 11px; }
  .sub { font-size: 11.5px; color: var(--text-3b); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .word { font-size: 11px; font-family: var(--font-mono); color: var(--text-3); flex: none; }
  .word[data-status='blocked'] { color: var(--blocked-badge-text); }
  .word[data-status='working'] { color: var(--working); }
  .word[data-status='done'] { color: var(--done); }
</style>
