<script lang="ts">
  import { goto } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { rollupOf, countsOf, monogram, primaryPaneOf } from '$lib/session/derive';
  import { openSheet, showToast } from '$lib/ui/state';
  import StatusDot from '$lib/ui/StatusDot.svelte';
  import Card from '$lib/ui/Card.svelte';
  import ButtonRow from '$lib/ui/ButtonRow.svelte';

  const s = session();
  const spaces = s.spaces;
  const desktop = $derived($width >= BREAKPOINT);

  $effect(() => {
    if (desktop && $spaces[0]) goto(`/spaces/${encodeURIComponent($spaces[0].id)}`, { replaceState: true });
  });

  function openChat(spaceId: string) {
    const paneId = primaryPaneOf($spaces, spaceId);
    if (paneId) goto(`/pane/${encodeURIComponent(paneId)}`);
    else goto(`/spaces/${encodeURIComponent(spaceId)}`);
  }

  function newSpace() {
    openSheet({
      kind: 'workspace.create', title: 'New space', body: 'Creates a workspace with one tab and a shell pane.',
      cta: 'Create', hasInput: true, hasCwd: true, inputLabel: 'Label', call: 'workspace.create { cwd, label, focus: false }',
      onConfirm: async (label, cwd) => { await s.request({ method: 'workspace.create', params: { cwd, label, focus: false } }).catch(() => {}); showToast('space created'); }
    });
  }
  function rename(id: string, label: string) {
    openSheet({ kind: 'workspace.rename', title: `Rename ${label}`, body: 'Renames this space everywhere it appears.', cta: 'Rename', hasInput: true, inputLabel: 'New label', call: 'workspace.rename { workspace_id, label }',
      onConfirm: async (l) => { await s.request({ method: 'workspace.rename', params: { workspace_id: id, label: l } }).catch(() => {}); showToast('space renamed'); } });
  }
  function close(id: string, label: string) {
    openSheet({ kind: 'workspace.close', title: `Close ${label}?`, body: 'Running processes are killed. A worktree checkout on disk is left alone.', cta: 'Close space', destructive: true, call: 'workspace.close { workspace_id }',
      onConfirm: async () => { await s.request({ method: 'workspace.close', params: { workspace_id: id } }).catch(() => {}); showToast('space closed'); } });
  }
  const rollupWord = (r: string) => (r === 'none' ? '—' : r);
</script>

<header class="hd">
  <h1 class="screen-title">Spaces</h1>
  <span class="sub mono">{$spaces.length} spaces · {$spaces.reduce((n, s2) => n + s2.tabs.length, 0)} tabs</span>
  <button class="new" onclick={newSpace}>＋ New</button>
</header>

<div class="list">
  {#each $spaces as sp (sp.id)}
    {@const c = countsOf($spaces, sp.id)}
    {@const r = rollupOf($spaces, sp.id)}
    <div class="wrap">
      <button class="body" onclick={() => openChat(sp.id)}>
        <span class="mono mono-square">{monogram(sp.label)}</span>
        <span class="col">
          <span class="mono label">{sp.label} {#if sp.worktree}<span class="wt mono">worktree</span>{/if}</span>
          <span class="mono cwd">{sp.cwd}</span>
        </span>
        <span class="right">
          <span class="rollup" data-status={r}><StatusDot status={r} /> {rollupWord(r)}</span>
          <span class="counts mono">{c.tabs} tabs · {c.panes} panes</span>
        </span>
      </button>
      <ButtonRow>
        <button onclick={() => goto(`/spaces/${encodeURIComponent(sp.id)}`)}>Tabs</button>
        <button onclick={() => rename(sp.id, sp.label)}>Rename</button>
        <button class="danger" onclick={() => close(sp.id, sp.label)}>Close</button>
      </ButtonRow>
    </div>
  {/each}
</div>

<style>
  .hd { display: flex; align-items: center; gap: 10px; padding: 16px 14px 8px; flex-wrap: wrap; }
  .sub { font-size: 11px; color: var(--text-4); }
  .new { margin-left: auto; background: var(--text-1); color: var(--text-on-light); border: none; border-radius: var(--r-btn); min-height: 44px; padding: 0 14px; font-weight: 600; }
  .list { padding: 6px 14px 96px; display: flex; flex-direction: column; gap: 12px; }
  .wrap { background: var(--card); border: 1px solid var(--hairline); border-radius: var(--r-card); overflow: hidden; }
  .body { width: 100%; display: flex; gap: 12px; align-items: center; padding: 12px; background: none; border: none; text-align: left; }
  .mono-square { width: 30px; height: 30px; border-radius: var(--r-chip); background: var(--surface-tint); display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 600; color: var(--text-2); flex: none; }
  .col { display: flex; flex-direction: column; min-width: 0; flex: 1; }
  .label { font-size: 13.5px; font-weight: 600; }
  .wt { font-size: 10px; color: var(--worktree-text); background: var(--worktree-bg); border-radius: var(--r-badge); padding: 1px 5px; margin-left: 4px; }
  .cwd { font-size: 11px; color: var(--text-4); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .right { display: flex; flex-direction: column; align-items: flex-end; gap: 2px; }
  .rollup { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-3); }
  .rollup[data-status='blocked'] { color: var(--blocked-badge-text); }
  .rollup[data-status='working'] { color: var(--working); }
  .counts { font-size: 10.5px; color: var(--text-4); }
</style>
