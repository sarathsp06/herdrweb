<script lang="ts">
  import { goto } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { rollupOf, countsOf, monogram, chatPaneForSpace } from '$lib/session/derive';
  import { get } from 'svelte/store';
  import { openSheet, openActions, showToast, lastTabBySpace } from '$lib/ui/state';
  import StatusDot from '$lib/ui/StatusDot.svelte';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';

  const s = session();
  const spaces = s.spaces;
  const desktop = $derived($width >= BREAKPOINT);

  $effect(() => {
    if (desktop && $spaces[0]) goto(`/spaces/${encodeURIComponent($spaces[0].id)}`, { replaceState: true });
  });

  function openChat(spaceId: string) {
    const paneId = chatPaneForSpace($spaces, spaceId, get(lastTabBySpace)[spaceId]);
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
  function overflow(id: string, label: string) {
    openActions('workspace.actions', label, [
      { label: 'Tabs & panes', glyph: '⌗', onSelect: () => goto(`/spaces/${encodeURIComponent(id)}`) },
      { label: 'Rename space', glyph: '✎', onSelect: () => rename(id, label) },
      { label: 'Close space', glyph: '✕', destructive: true, onSelect: () => close(id, label) }
    ]);
  }
  const rollupWord = (r: string) => (r === 'none' ? '—' : r);
  const rollupColor: Record<string, string> = {
    blocked: 'text-(--blocked-badge-text)',
    working: 'text-working'
  };
</script>

<header class="flex flex-wrap items-center gap-2.5 px-3.5 pt-4 pb-2">
  <h1 class="screen-title">Spaces</h1>
  <span class="mono text-[11px] text-muted-foreground"
    >{$spaces.length} spaces · {$spaces.reduce((n, s2) => n + s2.tabs.length, 0)} tabs</span
  >
  <Button class="ml-auto min-h-11 rounded-lg px-3.5 font-semibold" onclick={newSpace}>＋ New</Button>
</header>

<div class="list flex flex-col gap-3 px-3.5 pt-1.5 pb-6">
  {#each $spaces as sp (sp.id)}
    {@const c = countsOf($spaces, sp.id)}
    {@const r = rollupOf($spaces, sp.id)}
    <div class="flex items-stretch overflow-hidden rounded-(--r-card) border border-(--hairline) bg-card">
      <button class="flex min-w-0 flex-1 items-center gap-3 p-3 text-left" onclick={() => openChat(sp.id)}>
        <span
          class="mono flex h-[30px] w-[30px] flex-none items-center justify-center rounded-(--r-chip) bg-muted text-xs font-semibold text-(--text-2)"
          >{monogram(sp.label)}</span
        >
        <span class="flex min-w-0 flex-1 flex-col">
          <span class="mono text-[13.5px] font-semibold"
            >{sp.label}
            {#if sp.worktree}<Badge
                variant="outline"
                class="mono ml-1 h-auto rounded-(--r-badge) border-transparent bg-(--worktree-bg) px-[5px] py-px text-[10px] font-normal text-(--worktree-text)"
                >worktree</Badge
              >{/if}</span
          >
          <span class="mono truncate text-[11px] text-muted-foreground">{sp.cwd}</span>
        </span>
        <span class="flex flex-col items-end gap-0.5">
          <span class="flex items-center gap-1.5 text-xs {rollupColor[r] ?? 'text-(--text-3)'}" data-status={r}
            ><StatusDot status={r} /> {rollupWord(r)}</span
          >
          <span class="mono text-[10.5px] text-muted-foreground">{c.tabs} tabs · {c.panes} panes</span>
        </span>
      </button>
      <button
        class="w-11 flex-none border-l border-(--hairline) text-xl text-(--text-3) hover:bg-muted hover:text-foreground"
        aria-label="actions for {sp.label}"
        onclick={() => overflow(sp.id, sp.label)}>⋯</button
      >
    </div>
  {/each}
</div>
