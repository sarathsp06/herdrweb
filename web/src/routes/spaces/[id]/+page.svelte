<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { tabHasBlocked } from '$lib/session/derive';
  import { openSheet, showToast } from '$lib/ui/state';
  import StatusDot from '$lib/ui/StatusDot.svelte';
  import ButtonRow from '$lib/ui/ButtonRow.svelte';
  import { Button } from '$lib/components/ui/button';

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
  <header class="flex items-center gap-2.5 border-b border-(--hairline) px-3.5 py-3">
    <button class="h-8 w-8 text-[22px] text-(--text-2)" onclick={() => goto('/spaces')} aria-label="back">‹</button>
    <div class="flex flex-col">
      <span class="mono text-[15px] font-semibold">{space.label}</span><span
        class="mono text-[11px] text-muted-foreground">{space.cwd}</span
      >
    </div>
  </header>

  <div class="flex gap-1.5 overflow-x-auto border-b border-(--hairline) px-3.5 py-2.5">
    {#each space.tabs as t, i}
      <button
        class="mono flex min-h-10 flex-none items-center gap-1.5 rounded-(--r-chip) border px-3 text-xs {i === activeTab
          ? 'border-ring bg-accent text-foreground'
          : 'border-border bg-card text-(--text-3)'}"
        onclick={() => (activeTab = i)}
      >
        {t.label}
        {#if tabHasBlocked(t)}<span class="h-1.5 w-1.5 rounded-full bg-blocked"></span>{/if}
        <span class="text-muted-foreground">{t.panes.length}</span>
      </button>
    {/each}
    <button
      class="min-h-10 w-11 flex-none rounded-(--r-chip) border border-dashed border-border text-(--text-3)"
      onclick={addTab}>＋</button
    >
  </div>

  {#if tab}
    <div class="flex items-center px-3.5 pt-3 pb-1.5">
      <span class="mono text-[11px] tracking-[0.05em] text-(--text-3b)"
        >{tab.label.toUpperCase()} <span class="text-muted-foreground">{tab.id}</span></span
      >
      <div class="ml-auto flex gap-2">
        <Button variant="outline" size="sm" class="rounded-(--r-badge) text-[11.5px] text-(--text-2)" onclick={renameTab}
          >Rename</Button
        >
        <Button
          variant="outline"
          size="sm"
          class="rounded-(--r-badge) border-(--blocked-border) text-[11.5px] text-(--blocked-badge-text)"
          onclick={closeTab}>Close tab</Button
        >
      </div>
    </div>

    <div class="panes flex flex-col gap-3 px-3.5 pt-1.5 pb-24">
      {#each tab.panes as p (p.id)}
        <div class="overflow-hidden rounded-(--r-card) border border-(--hairline) bg-card">
          <div class="flex items-center gap-2.5 p-3">
            <StatusDot status={p.status} />
            <span class="flex flex-col">
              <span class="mono text-[13px] font-medium"
                >{p.label} <span class="text-[11px] text-muted-foreground">{p.id}</span></span
              >
              <span class="text-[11px] text-(--text-3b)">{p.sub}</span>
            </span>
          </div>
          <pre
            class="mono m-0 overflow-x-auto border-t border-(--hairline) bg-(--code-surface) px-3 py-2 text-[10.5px] leading-[1.6] whitespace-pre text-(--text-3)">{p.tail.join('\n')}</pre>
          <ButtonRow>
            <button onclick={() => goto(`/pane/${encodeURIComponent(p.id)}`)}>{p.agent ? 'Chat' : 'Read'}</button>
            <button onclick={() => renamePane(p.id, p.label)}>Rename</button>
            <button class="danger" onclick={() => closePane(p.id, p.label)}>Close</button>
          </ButtonRow>
        </div>
      {/each}
      <button
        class="min-h-[46px] rounded-(--r-card) border border-dashed border-border text-(--text-3)"
        onclick={split}>＋ Split a new pane</button
      >
    </div>
  {/if}
{:else}
  <div class="mono p-10 text-muted-foreground">space {spaceId} not found</div>
{/if}
