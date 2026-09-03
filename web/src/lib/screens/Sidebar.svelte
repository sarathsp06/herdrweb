<script lang="ts">
  import { goto } from '$app/navigation';
  import type { ConnState, Space } from '$lib/protocol';
  import { agentsOf, rollupOf, monogram, primaryPaneOf } from '$lib/session/derive';
  import { agentsGrouped, lastPane } from '$lib/ui/state';
  import StatusGlyph from '$lib/ui/StatusGlyph.svelte';
  import StatusDot from '$lib/ui/StatusDot.svelte';

  let { spaces, connection, embedded = false, onselect, onclose }:
    { spaces: Space[]; connection: ConnState; embedded?: boolean; onselect?: () => void; onclose?: () => void } = $props();

  const agents = $derived(agentsOf(spaces));

  function openAgent(paneId: string) {
    lastPane.set(paneId);
    goto(`/pane/${encodeURIComponent(paneId)}`);
    onselect?.();
  }
  function openSpace(spaceId: string) {
    const paneId = primaryPaneOf(spaces, spaceId);
    if (paneId) { openAgent(paneId); return; }
    goto(`/spaces/${encodeURIComponent(spaceId)}`);
    onselect?.();
  }
  function openSpaceDetail(spaceId: string) {
    goto(`/spaces/${encodeURIComponent(spaceId)}`);
    onselect?.();
  }
  function go(path: string) {
    goto(path);
    onselect?.();
  }
  const connColor: Record<ConnState, string> = {
    open: 'var(--done)', connecting: 'var(--working)', reconnecting: 'var(--working)', closed: 'var(--blocked)'
  };
</script>

<aside class="sidebar" class:embedded>
  <header class="brand">
    <span class="mark mono">herdr</span>
    <span class="conn" style="--c: {connColor[connection]}" title={connection}></span>
    {#if onclose}
      <button class="navclose" aria-label="hide navigation" title="hide navigation" onclick={() => onclose()}>⟨</button>
    {/if}
  </header>
  <div class="session mono">default session · {spaces.length} spaces · {connection}</div>

  <div class="group">
    <div class="ghead"><span class="section-label">spaces</span></div>
    {#each spaces as sp (sp.id)}
      <div class="srow">
        <button class="srow-main" onclick={() => openSpace(sp.id)}>
          <StatusGlyph status={rollupOf(spaces, sp.id)} />
          <span class="mono label">{sp.label}</span>
          {#if sp.branch}<span class="mono branch">{sp.branch}</span>{/if}
        </button>
        <button class="tabsbtn mono" title="tabs & panes" aria-label="tabs in {sp.label}" onclick={() => openSpaceDetail(sp.id)}>tabs</button>
      </div>
    {/each}
  </div>

  <div class="group">
    <div class="ghead">
      <span class="section-label">agents</span>
      <button class="toggle mono" onclick={() => agentsGrouped.update((v) => !v)}>{$agentsGrouped ? 'grouped' : 'flat'}</button>
    </div>
    {#each agents as a (a.pane.id)}
      <button class="arow" onclick={() => openAgent(a.pane.id)}>
        <StatusGlyph status={a.pane.status} />
        <span class="col">
          <span class="mono name">{a.pane.label} <span class="pid">{a.pane.id}</span></span>
          <span class="sub">{$agentsGrouped ? a.space.label : a.pane.sub}</span>
        </span>
        <span class="word" data-status={a.pane.status}>{a.pane.status}</span>
      </button>
    {/each}
  </div>

  <nav class="foot">
    <button class="frow mono" onclick={() => go('/')}><span class="fg">◉</span> Agents</button>
    <button class="frow mono" onclick={() => go('/spaces')}><span class="fg">⌗</span> Spaces</button>
    <button class="frow mono" onclick={() => go('/settings')}><span class="fg">⚙</span> Settings</button>
  </nav>
</aside>

<style>
  .sidebar { width: 328px; flex: none; background: var(--sidebar-bg); border-right: 1px solid var(--hairline); height: 100vh; overflow-y: auto; padding: 16px 12px 24px; }
  .sidebar.embedded { width: 100%; height: auto; border-right: none; padding: 14px 14px 96px; background: transparent; }
  .brand { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
  .mark { font-size: 15px; font-weight: 700; letter-spacing: -0.02em; }
  .conn { width: 8px; height: 8px; border-radius: 50%; background: var(--c); animation: hpulse 2.6s ease-in-out infinite; }
  .navclose { margin-left: auto; width: 30px; height: 30px; border-radius: var(--r-chip); border: 1px solid var(--control); background: var(--card); color: var(--text-2); font-size: 15px; line-height: 1; }
  .navclose:hover { background: var(--surface-tint); }
  .session { font-size: 11px; color: var(--text-4); margin-bottom: 16px; }
  .group { margin-bottom: 18px; }
  .ghead { display: flex; align-items: center; justify-content: space-between; padding: 0 6px 8px; }
  .toggle { font-size: 10.5px; color: var(--text-3b); background: none; border: 1px solid var(--control); border-radius: var(--r-badge); padding: 1px 6px; }
  .srow { display: flex; align-items: center; gap: 4px; }
  .srow-main, .arow { flex: 1; min-width: 0; width: 100%; display: flex; align-items: center; gap: 9px; min-height: 44px; background: none; border: none; border-radius: var(--r-chip); padding: 6px 6px; text-align: left; }
  .srow-main:hover, .arow:hover { background: var(--surface-tint); }
  .tabsbtn { flex: none; min-height: 30px; padding: 0 10px; border-radius: var(--r-chip); border: 1px solid var(--control); background: var(--card); color: var(--text-3); font-size: 11px; }
  .tabsbtn:hover { background: var(--surface-tint); color: var(--text-2); }
  .foot { display: flex; gap: 6px; margin-top: 4px; border-top: 1px solid var(--hairline); padding-top: 12px; }
  .frow { flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px; min-height: 42px; border-radius: var(--r-btn); border: 1px solid var(--control); background: var(--card); color: var(--text-2); font-size: 12px; }
  .frow:hover { background: var(--surface-tint); }
  .fg { font-size: 13px; }
  .label { font-size: 13px; font-weight: 600; }
  .branch { font-size: 10.5px; color: var(--text-4); border: 1px solid var(--control); border-radius: var(--r-badge); padding: 1px 5px; margin-left: auto; }
  .col { display: flex; flex-direction: column; min-width: 0; flex: 1; }
  .name { font-size: 13px; font-weight: 500; }
  .pid { color: var(--text-4); font-size: 11px; }
  .sub { font-size: 11px; color: var(--text-3b); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .word { font-size: 11px; font-family: var(--font-mono); color: var(--text-3); }
  .word[data-status='blocked'] { color: var(--blocked-badge-text); }
  .word[data-status='working'] { color: var(--working); }
  .word[data-status='done'] { color: var(--done); }
</style>
