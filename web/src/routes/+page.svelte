<script lang="ts">
  import { goto } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { lastPane } from '$lib/ui/state';
  import { agentsOf, inboxSections, plural } from '$lib/session/derive';
  import { connLabel } from '$lib/chat/answer';
  import type { ConnState } from '$lib/protocol';
  import AgentRow from '$lib/screens/AgentRow.svelte';
  import BlockedCard from '$lib/screens/BlockedCard.svelte';

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
  // their own section on top, each showing what it asks; working next; idle
  // and done last. The desktop sidebar renders the same sections.
  const sections = $derived(inboxSections(agents));

  function openAgent(paneId: string) {
    lastPane.set(paneId);
    goto(`/pane/${encodeURIComponent(paneId)}`);
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
      title={connLabel($connection)}
    ></span>
    <span class="mono text-[11px] text-muted-foreground"
      >{connLabel($connection)} · {plural($spaces.length, 'space')} · {plural(agents.length, 'agent')}</span
    >
  </header>

  <div class="list px-2 pb-6">
    {#each sections as sec (sec.label)}
      <div class="flex items-center gap-2 px-2 pt-3.5 pb-1.5">
        <span class="section-label">{sec.label}</span><span class="mono text-[10.5px] text-muted-foreground"
          >{sec.items.length}</span
        >
      </div>
      {#if sec.label === 'needs you'}
        <div class="flex flex-col gap-2 px-1.5">
          {#each sec.items as a (a.pane.id)}
            <BlockedCard agent={a} onopen={() => openAgent(a.pane.id)} />
          {/each}
        </div>
      {:else}
        {#each sec.items as a (a.pane.id)}
          <AgentRow
            agent={a}
            sub={a.space.label + (a.space.branch ? ' · ' + a.space.branch : '')}
            onopen={() => openAgent(a.pane.id)}
          />
        {/each}
      {/if}
    {/each}
    {#if agents.length === 0}
      <div class="mono p-10 text-center text-muted-foreground">no agents running</div>
    {/if}
  </div>
{:else}
  <div class="mono p-10 text-center text-muted-foreground">select an agent</div>
{/if}
