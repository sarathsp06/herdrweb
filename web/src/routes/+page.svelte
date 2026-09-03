<script lang="ts">
  import { goto } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { lastPane } from '$lib/ui/state';
  import { agentsOf } from '$lib/session/derive';
  import Sidebar from '$lib/screens/Sidebar.svelte';

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
</script>

{#if !desktop}
  <header class="hd">
    <h1 class="screen-title">Agents</h1>
  </header>
  <Sidebar spaces={$spaces} connection={$connection} embedded />
{:else}
  <div class="empty mono">select an agent</div>
{/if}

<style>
  .hd { padding: 16px 14px 4px; }
  .empty { color: var(--text-4); padding: 40px; text-align: center; }
</style>
