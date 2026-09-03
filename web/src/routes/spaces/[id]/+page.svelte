<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { tabHasBlocked } from '$lib/session/derive';
  import { openSheet, showToast } from '$lib/ui/state';
  import StatusDot from '$lib/ui/StatusDot.svelte';
  import ButtonRow from '$lib/ui/ButtonRow.svelte';

  const s = session();
  const spaces = s.spaces;
  const spaceId = $derived(decodeURIComponent($page.params.id ?? ''));
  const space = $derived($spaces.find((sp) => sp.id === spaceId));
  let activeTab = $state(0);
  const tab = $derived(space?.tabs[Math.min(activeTab, (space?.tabs.length ?? 1) - 1)]);

  function addTab() {
    openSheet({ kind: 'tab.create', title: 'New tab', body: 'Adds a tab with one shell pane to this space.', cta: 'Create tab', hasInput: true, inputLabel: 'Label', call: 'tab.create { workspace_id, label }',
      onConfirm: async (l) => { await s.request({ method: 'tab.create', params: { workspace_id: spaceId, label: l } }).catch(() => {}); showToast('tab created'); } });
  }
  function renameTab() {
    if (!tab) return;
    openSheet({ kind: 'tab.rename', title: `Rename ${tab.label}`, body: 'Renames this tab.', cta: 'Rename', hasInput: true, inputLabel: 'New label', call: 'tab.rename { tab_id, label }',
      onConfirm: async (l) => { await s.request({ method: 'tab.rename', params: { tab_id: tab.id, label: l } }).catch(() => {}); showToast('tab renamed'); } });
  }
  function closeTab() {
    if (!tab) return;
    openSheet({ kind: 'tab.close', title: `Close ${tab.label}?`, body: 'Panes in this tab are closed and their processes killed.', cta: 'Close tab', destructive: true, call: 'tab.close { tab_id }',
      onConfirm: async () => { await s.request({ method: 'tab.close', params: { tab_id: tab.id } }).catch(() => {}); showToast('tab closed'); } });
  }
  function split() {
    openSheet({ kind: 'pane.split', title: 'Split a new pane', body: 'Opens a new shell pane to the right at 50%.', cta: 'Split', call: 'pane.split { direction: "right", ratio: 0.5 }',
      onConfirm: async () => { const p = tab?.panes[0]; if (p) await s.request({ method: 'pane.split', params: { pane_id: p.id, direction: 'right', ratio: 0.5 } }).catch(() => {}); showToast('pane split'); } });
  }
  function renamePane(id: string, label: string) {
    openSheet({ kind: 'pane.rename', title: `Rename ${label}`, body: 'Renames this pane.', cta: 'Rename', hasInput: true, inputLabel: 'New label', call: 'pane.rename { pane_id, label }',
      onConfirm: async (l) => { await s.request({ method: 'pane.rename', params: { pane_id: id, label: l } }).catch(() => {}); showToast('pane renamed'); } });
  }
  function closePane(id: string, label: string) {
    openSheet({ kind: 'pane.close', title: `Close ${label}?`, body: 'The pane and its process are terminated.', cta: 'Close pane', destructive: true, call: 'pane.close { pane_id }',
      onConfirm: async () => { await s.request({ method: 'pane.close', params: { pane_id: id } }).catch(() => {}); showToast('pane closed'); } });
  }
</script>

{#if space}
  <header class="bar">
    <button class="back" onclick={() => goto('/spaces')} aria-label="back">‹</button>
    <div class="title"><span class="mono label">{space.label}</span><span class="mono cwd">{space.cwd}</span></div>
  </header>

  <div class="tabs">
    {#each space.tabs as t, i}
      <button class="tab mono" class:active={i === activeTab} onclick={() => (activeTab = i)}>
        {t.label}
        {#if tabHasBlocked(t)}<span class="bdot"></span>{/if}
        <span class="pc">{t.panes.length}</span>
      </button>
    {/each}
    <button class="addtab" onclick={addTab}>＋</button>
  </div>

  {#if tab}
    <div class="tabhead">
      <span class="mono up">{tab.label.toUpperCase()} <span class="tid">{tab.id}</span></span>
      <div class="acts">
        <button class="mini" onclick={renameTab}>Rename</button>
        <button class="mini danger" onclick={closeTab}>Close tab</button>
      </div>
    </div>

    <div class="panes">
      {#each tab.panes as p (p.id)}
        <div class="pane">
          <div class="ph">
            <StatusDot status={p.status} />
            <span class="col">
              <span class="mono pl">{p.label} <span class="pid">{p.id}</span></span>
              <span class="sub">{p.sub}</span>
            </span>
          </div>
          <pre class="tail mono">{p.tail.join('\n')}</pre>
          <ButtonRow>
            <button onclick={() => goto(`/pane/${encodeURIComponent(p.id)}`)}>{p.agent ? 'Chat' : 'Read'}</button>
            <button onclick={() => renamePane(p.id, p.label)}>Rename</button>
            <button class="danger" onclick={() => closePane(p.id, p.label)}>Close</button>
          </ButtonRow>
        </div>
      {/each}
      <button class="split" onclick={split}>＋ Split a new pane</button>
    </div>
  {/if}
{:else}
  <div class="missing mono">space {spaceId} not found</div>
{/if}

<style>
  .bar { display: flex; align-items: center; gap: 10px; padding: 12px 14px; border-bottom: 1px solid var(--hairline); }
  .back { width: 32px; height: 32px; border: none; background: none; color: var(--text-2); font-size: 22px; }
  .title { display: flex; flex-direction: column; }
  .title .label { font-size: 15px; font-weight: 600; }
  .title .cwd { font-size: 11px; color: var(--text-4); }
  .tabs { display: flex; gap: 6px; overflow-x: auto; padding: 10px 14px; border-bottom: 1px solid var(--hairline); }
  .tab { flex: none; display: flex; align-items: center; gap: 6px; min-height: 40px; padding: 0 12px; border-radius: var(--r-chip); border: 1px solid var(--control); background: var(--card); color: var(--text-3); font-size: 12px; }
  .tab.active { border-color: var(--control-selected-2); background: var(--surface-tint-2); color: var(--text-1); }
  .bdot { width: 6px; height: 6px; border-radius: 50%; background: var(--blocked); }
  .pc { color: var(--text-4); }
  .addtab { flex: none; min-height: 40px; width: 44px; border: 1px dashed var(--control); border-radius: var(--r-chip); background: none; color: var(--text-3); }
  .tabhead { display: flex; align-items: center; padding: 12px 14px 6px; }
  .up { font-size: 11px; letter-spacing: 0.05em; color: var(--text-3b); }
  .tid { color: var(--text-4); }
  .acts { margin-left: auto; display: flex; gap: 8px; }
  .mini { font-size: 11.5px; padding: 5px 10px; border-radius: var(--r-badge); border: 1px solid var(--control); background: none; color: var(--text-2); }
  .mini.danger { color: var(--blocked-badge-text); border-color: var(--blocked-border); }
  .panes { padding: 6px 14px 96px; display: flex; flex-direction: column; gap: 12px; }
  .pane { background: var(--card); border: 1px solid var(--hairline); border-radius: var(--r-card); overflow: hidden; }
  .ph { display: flex; align-items: center; gap: 10px; padding: 12px; }
  .col { display: flex; flex-direction: column; }
  .pl { font-size: 13px; font-weight: 500; }
  .pid { color: var(--text-4); font-size: 11px; }
  .sub { font-size: 11px; color: var(--text-3b); }
  .tail { margin: 0; padding: 8px 12px; background: var(--code-surface); border-top: 1px solid var(--hairline); font-size: 10.5px; line-height: 1.6; color: var(--text-3); white-space: pre; overflow-x: auto; }
  .split { min-height: 46px; border: 1px dashed var(--control); border-radius: var(--r-card); background: none; color: var(--text-3); }
  .missing { padding: 40px; color: var(--text-4); }
</style>
