<script lang="ts">
  import { goto } from '$app/navigation';
  import type { ConnState, Space } from '$lib/protocol';
  import { agentsOf, rollupOf, chatPaneForSpace } from '$lib/session/derive';
  import { get } from 'svelte/store';
  import { agentsGrouped, lastPane, lastTabBySpace } from '$lib/ui/state';
  import StatusGlyph from '$lib/ui/StatusGlyph.svelte';
  import AgentRow from '$lib/screens/AgentRow.svelte';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';

  let { spaces, connection }: { spaces: Space[]; connection: ConnState } = $props();

  const agents = $derived(agentsOf(spaces));

  function openAgent(paneId: string) {
    lastPane.set(paneId);
    goto(`/pane/${encodeURIComponent(paneId)}`);
  }
  function openSpace(spaceId: string) {
    const paneId = chatPaneForSpace(spaces, spaceId, get(lastTabBySpace)[spaceId]);
    if (paneId) { openAgent(paneId); return; }
    goto(`/spaces/${encodeURIComponent(spaceId)}`);
  }
  const connColor: Record<ConnState, string> = {
    open: 'var(--done)', connecting: 'var(--working)', reconnecting: 'var(--working)', closed: 'var(--blocked)'
  };
</script>

<aside
  class="sidebar h-[calc(100vh/var(--font-scale,1))] w-[328px] flex-none overflow-y-auto border-r border-(--hairline) bg-(--sidebar-bg) px-3 pt-4 pb-6"
>
  <header class="mb-1 flex items-center gap-2">
    <span class="mono text-[15px] font-bold tracking-[-0.02em]">herdr</span>
    <span
      class="h-2 w-2 rounded-full animate-[hpulse_2.6s_ease-in-out_infinite]"
      style="background: {connColor[connection]}"
      title={connection}
    ></span>
  </header>
  <div class="mono mb-4 text-[11px] text-muted-foreground">default session · {spaces.length} spaces · {connection}</div>

  <div class="mb-[18px]">
    <div class="flex items-center justify-between px-1.5 pb-2"><span class="section-label">spaces</span></div>
    {#each spaces as sp (sp.id)}
      <div class="flex items-center gap-1">
        <button
          class="flex min-h-11 min-w-0 flex-1 items-center gap-[9px] rounded-(--r-chip) px-1.5 py-1.5 text-left hover:bg-muted"
          onclick={() => openSpace(sp.id)}
        >
          <StatusGlyph status={rollupOf(spaces, sp.id)} />
          <span class="mono text-[13px] font-semibold">{sp.label}</span>
          {#if sp.branch}<Badge
              variant="outline"
              class="mono ml-auto h-auto rounded-(--r-badge) px-[5px] py-px text-[10.5px] font-normal text-muted-foreground"
              >{sp.branch}</Badge
            >{/if}
        </button>
        <Button
          variant="outline"
          class="mono min-h-9 flex-none rounded-(--r-chip) bg-card px-2.5 text-[11px] font-normal text-(--text-3) hover:bg-muted hover:text-(--text-2)"
          title="tabs & panes"
          aria-label="tabs in {sp.label}"
          onclick={() => goto(`/spaces/${encodeURIComponent(sp.id)}`)}>tabs</Button
        >
      </div>
    {/each}
  </div>

  <div class="mb-[18px]">
    <div class="flex items-center justify-between px-1.5 pb-2">
      <span class="section-label">agents</span>
      <button
        class="mono rounded-(--r-badge) border border-border px-1.5 py-px text-[10.5px] text-(--text-3b)"
        aria-label="toggle grouped"
        onclick={() => agentsGrouped.update((v) => !v)}>{$agentsGrouped ? 'grouped' : 'flat'}</button
      >
    </div>
    {#each agents as a (a.pane.id)}
      <AgentRow agent={a} sub={$agentsGrouped ? a.space.label : a.pane.sub} onopen={() => openAgent(a.pane.id)} />
    {/each}
  </div>

  <nav class="mt-1 flex gap-1.5 border-t border-(--hairline) pt-3">
    <button
      class="mono flex min-h-[42px] flex-1 items-center justify-center gap-1.5 rounded-lg border border-border bg-card text-xs text-(--text-2) hover:bg-muted"
      onclick={() => goto('/')}><span class="text-[13px]">◉</span> Agents</button
    >
    <button
      class="mono flex min-h-[42px] flex-1 items-center justify-center gap-1.5 rounded-lg border border-border bg-card text-xs text-(--text-2) hover:bg-muted"
      onclick={() => goto('/spaces')}><span class="text-[13px]">⌗</span> Spaces</button
    >
    <button
      class="mono flex min-h-[42px] flex-1 items-center justify-center gap-1.5 rounded-lg border border-border bg-card text-xs text-(--text-2) hover:bg-muted"
      onclick={() => goto('/settings')}><span class="text-[13px]">⚙</span> Settings</button
    >
  </nav>
</aside>
