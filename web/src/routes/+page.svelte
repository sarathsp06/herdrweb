<script lang="ts">
  import { goto } from '$app/navigation';
  import { get } from 'svelte/store';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { lastPane, lastTabBySpace } from '$lib/ui/state';
  import { agentsOf, rollupOf, chatPaneForSpace, type AgentRef } from '$lib/session/derive';
  import type { ConnState } from '$lib/protocol';
  import AgentRow from '$lib/screens/AgentRow.svelte';
  import StatusGlyph from '$lib/ui/StatusGlyph.svelte';

  const s = session();
  const spaces = s.spaces;
  const connection = s.connection;
  const desktop = $derived($width >= BREAKPOINT);

  // Desktop: `/` resolves to the last-selected pane's chat (or the first agent).
  $effect(() => {
    if (!desktop) return;
    const target = $lastPane ?? agentsOf($spaces)[0]?.pane.id;
    if (target) goto(`/pane/${encodeURIComponent(target)}`, { replaceState: true });
  });

  const agents = $derived(agentsOf($spaces));
  // Urgency triage: blocked agents are the reason to open the app — they get
  // their own section on top; working next; idle/done last.
  const sections = $derived(
    (
      [
        { label: 'needs you', items: agents.filter((a) => a.pane.status === 'blocked') },
        { label: 'working', items: agents.filter((a) => a.pane.status === 'working') },
        { label: 'idle', items: agents.filter((a) => a.pane.status !== 'blocked' && a.pane.status !== 'working') }
      ] as { label: string; items: AgentRef[] }[]
    ).filter((sec) => sec.items.length > 0)
  );

  function openAgent(paneId: string) {
    lastPane.set(paneId);
    goto(`/pane/${encodeURIComponent(paneId)}`);
  }
  function openSpace(spaceId: string) {
    const paneId = chatPaneForSpace($spaces, spaceId, get(lastTabBySpace)[spaceId]);
    if (paneId) { openAgent(paneId); return; }
    goto(`/spaces/${encodeURIComponent(spaceId)}`);
  }
  const connColor: Record<ConnState, string> = {
    open: 'var(--done)', connecting: 'var(--working)', reconnecting: 'var(--working)', closed: 'var(--blocked)'
  };
</script>

{#if !desktop}
  <header class="flex items-center gap-2 px-3.5 pt-4 pb-2">
    <h1 class="screen-title">Agents</h1>
    <span
      class="h-2 w-2 rounded-full animate-[hpulse_2.6s_ease-in-out_infinite]"
      style="background: {connColor[$connection]}"
      title={$connection}
    ></span>
    <span class="mono text-[11px] text-muted-foreground">{$spaces.length} spaces · {$connection}</span>
  </header>

  <ul class="m-0 flex list-none gap-2 overflow-x-auto px-3.5 pt-1 pb-3" aria-label="spaces">
    {#each $spaces as sp (sp.id)}
      <li class="flex flex-none">
        <button
          class="flex min-h-11 flex-none items-center gap-[7px] rounded-(--r-chip) border border-border bg-card px-3.5 text-(--text-2) hover:bg-muted"
          onclick={() => openSpace(sp.id)}
        >
          <StatusGlyph status={rollupOf($spaces, sp.id)} />
          <span class="mono text-[12.5px] font-semibold">{sp.label}</span>
        </button>
      </li>
    {/each}
  </ul>

  <div class="list px-2 pb-6">
    {#each sections as sec (sec.label)}
      <div class="flex items-center gap-2 px-2 pt-3.5 pb-1.5">
        <span class="section-label">{sec.label}</span><span class="mono text-[10.5px] text-muted-foreground"
          >{sec.items.length}</span
        >
      </div>
      {#each sec.items as a (a.pane.id)}
        <AgentRow
          agent={a}
          sub={a.space.label + (a.space.branch ? ' · ' + a.space.branch : '')}
          onopen={() => openAgent(a.pane.id)}
        />
      {/each}
    {/each}
    {#if agents.length === 0}
      <div class="mono p-10 text-center text-muted-foreground">no agents running</div>
    {/if}
  </div>
{:else}
  <div class="mono p-10 text-center text-muted-foreground">select an agent</div>
{/if}
