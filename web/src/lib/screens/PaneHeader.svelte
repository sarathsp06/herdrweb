<script lang="ts">
  import { goto } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { openActions, openSheet, showToast } from '$lib/ui/state';
  import { primaryPaneOfTab, tabHasBlocked, type AgentRef } from '$lib/session/derive';
  import StatusPill from '$lib/ui/StatusPill.svelte';

  let { agentRef }: { agentRef: AgentRef } = $props();
  const s = session();
  const desktop = $derived($width >= BREAKPOINT);
  const paneId = $derived(agentRef.pane.id);
  const multiTab = $derived(agentRef.space.tabs.length > 1);

  // Tab switching left the screen as a chip strip; the header title is the
  // switcher — tapping it lists the space's tabs in a bottom sheet.
  function switchTab() {
    if (!multiTab) return;
    openActions(
      'tab.switch',
      agentRef.space.label,
      agentRef.space.tabs.map((t) => ({
        label: t.label,
        hint: `${t.panes.length} ${t.panes.length === 1 ? 'pane' : 'panes'}`,
        glyph: tabHasBlocked(t) ? '◉' : '·',
        active: t.id === agentRef.tab.id,
        onSelect: () => {
          const target = primaryPaneOfTab(t);
          if (target && target !== paneId) goto(`/pane/${encodeURIComponent(target)}`);
        }
      }))
    );
  }

  function overflow() {
    openActions('pane.actions', agentRef.pane.label, [
      { label: 'View diff', glyph: '±', onSelect: () => goto(`/pane/${encodeURIComponent(paneId)}/diff`) },
      {
        label: 'Rename pane',
        glyph: '✎',
        onSelect: () =>
          openSheet({
            kind: 'pane.rename', title: `Rename ${agentRef.pane.label}`, body: 'Renames this pane.', cta: 'Rename',
            hasInput: true, inputLabel: 'New label', call: 'pane.rename { pane_id, label }',
            onConfirm: async (l) => { await s.request({ method: 'pane.rename', params: { pane_id: paneId, label: l } }).catch(() => {}); showToast('pane renamed'); }
          })
      },
      {
        label: 'Close pane',
        glyph: '✕',
        destructive: true,
        onSelect: () =>
          openSheet({
            kind: 'pane.close', title: `Close ${agentRef.pane.label}?`, body: 'The pane and its process are terminated.', cta: 'Close pane',
            destructive: true, call: 'pane.close { pane_id }',
            onConfirm: async () => { await s.request({ method: 'pane.close', params: { pane_id: paneId } }).catch(() => {}); showToast('pane closed'); goto('/'); }
          })
      }
    ]);
  }
</script>

<header class="phd">
  {#if !desktop}
    <button class="back" aria-label="back to agents" onclick={() => goto('/')}>‹</button>
  {/if}
  <button class="title" class:tappable={multiTab} onclick={switchTab} aria-label={multiTab ? 'switch tab' : undefined}>
    <span class="mono name">{agentRef.pane.label}</span>
    <span class="sub">{agentRef.space.label} · {agentRef.tab.label}{#if multiTab}<span class="chev" aria-hidden="true"> ▾</span>{/if}</span>
  </button>
  <StatusPill status={agentRef.pane.status} />
  <button class="more" aria-label="pane actions" onclick={overflow}>⋯</button>
</header>

<style>
  .phd {
    flex: none; display: flex; align-items: center; gap: 8px;
    padding: 6px 10px; border-bottom: 1px solid var(--hairline); background: var(--sidebar-bg);
  }
  .back {
    flex: none; width: 44px; height: 44px; margin-left: -6px;
    border: none; background: none; color: var(--text-2); font-size: 26px; line-height: 1;
    border-radius: var(--r-chip);
  }
  .back:active, .more:active { background: var(--surface-tint); }
  .title {
    flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px;
    min-height: 44px; justify-content: center; text-align: left;
    background: none; border: none; padding: 0 2px; border-radius: var(--r-chip);
  }
  .title.tappable:hover { background: var(--surface-tint); }
  .name { font-size: 14px; font-weight: 600; color: var(--text-1); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .sub { font-size: 11px; color: var(--text-3b); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .chev { color: var(--text-4); }
  .more {
    flex: none; width: 44px; height: 44px; margin-right: -6px;
    border: none; background: none; color: var(--text-2); font-size: 20px; line-height: 1;
    border-radius: var(--r-chip);
  }
</style>
