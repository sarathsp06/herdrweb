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
  <header class="hd">
    <h1 class="screen-title">Agents</h1>
    <span class="conn" style="--c: {connColor[$connection]}" title={$connection}></span>
    <span class="mono meta">{$spaces.length} spaces · {$connection}</span>
  </header>

  <ul class="chips" aria-label="spaces">
    {#each $spaces as sp (sp.id)}
      <li>
        <button class="chip" onclick={() => openSpace(sp.id)}>
          <StatusGlyph status={rollupOf($spaces, sp.id)} />
          <span class="mono clabel">{sp.label}</span>
        </button>
      </li>
    {/each}
  </ul>

  <div class="list">
    {#each sections as sec (sec.label)}
      <div class="ghead"><span class="section-label">{sec.label}</span><span class="count mono">{sec.items.length}</span></div>
      {#each sec.items as a (a.pane.id)}
        <AgentRow
          agent={a}
          sub={a.space.label + (a.space.branch ? ' · ' + a.space.branch : '')}
          onopen={() => openAgent(a.pane.id)}
        />
      {/each}
    {/each}
    {#if agents.length === 0}
      <div class="empty mono">no agents running</div>
    {/if}
  </div>
{:else}
  <div class="empty mono">select an agent</div>
{/if}

<style>
  .hd { display: flex; align-items: center; gap: 8px; padding: 16px 14px 8px; }
  .conn { width: 8px; height: 8px; border-radius: 50%; background: var(--c); animation: hpulse 2.6s ease-in-out infinite; }
  .meta { font-size: 11px; color: var(--text-4); }
  .chips { display: flex; gap: 8px; overflow-x: auto; padding: 4px 14px 12px; margin: 0; list-style: none; }
  .chips li { flex: none; display: flex; }
  .chip {
    flex: none; display: flex; align-items: center; gap: 7px; min-height: 44px;
    padding: 0 14px; border-radius: var(--r-chip); border: 1px solid var(--control);
    background: var(--card); color: var(--text-2);
  }
  .chip:hover { background: var(--surface-tint); }
  .clabel { font-size: 12.5px; font-weight: 600; }
  .list { padding: 0 8px 24px; }
  .ghead { display: flex; align-items: center; gap: 8px; padding: 14px 8px 6px; }
  .count { font-size: 10.5px; color: var(--text-4); }
  .empty { color: var(--text-4); padding: 40px; text-align: center; }
</style>
