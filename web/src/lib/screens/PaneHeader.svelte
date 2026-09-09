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

<header class="flex flex-none items-center gap-2 border-b border-(--hairline) bg-(--sidebar-bg) px-2.5 py-1.5">
  {#if !desktop}
    <button
      class="-ml-1.5 h-11 w-11 flex-none rounded-(--r-chip) text-[26px] leading-none text-(--text-2) active:bg-muted"
      aria-label="back to agents"
      onclick={() => goto('/')}>‹</button
    >
  {/if}
  <button
    class="flex min-h-11 min-w-0 flex-1 flex-col justify-center gap-px rounded-(--r-chip) px-0.5 text-left {multiTab
      ? 'hover:bg-muted'
      : ''}"
    onclick={switchTab}
    aria-label={multiTab ? 'switch tab' : undefined}
  >
    <span class="mono truncate text-sm font-semibold text-foreground">{agentRef.pane.label}</span>
    <span class="truncate text-[11px] text-(--text-3b)"
      >{agentRef.space.label} · {agentRef.tab.label}{#if multiTab}<span class="text-muted-foreground" aria-hidden="true"> ▾</span>{/if}</span
    >
  </button>
  <StatusPill status={agentRef.pane.status} />
  <button
    class="-mr-1.5 h-11 w-11 flex-none rounded-(--r-chip) text-xl leading-none text-(--text-2) active:bg-muted"
    aria-label="pane actions"
    onclick={overflow}>⋯</button
  >
</header>
