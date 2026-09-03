<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { findPaneIn } from '$lib/session/derive';
  import { navOpen, mode } from '$lib/ui/state';
  import StatusPill from '$lib/ui/StatusPill.svelte';

  const s = session();
  const spaces = s.spaces;
  const path = $derived($page.url.pathname);

  interface Crumb { label: string; href?: string }

  const crumbs = $derived.by<Crumb[]>(() => {
    const out: Crumb[] = [{ label: '⌂', href: '/' }];
    if (path.startsWith('/pane/')) {
      const id = decodeURIComponent(path.slice('/pane/'.length).split('/')[0]);
      const ref = findPaneIn($spaces, id);
      if (ref) {
        out.push({ label: ref.space.label, href: `/spaces/${encodeURIComponent(ref.space.id)}` });
        out.push({ label: ref.pane.label });
      } else {
        out.push({ label: id });
      }
    } else if (path.startsWith('/spaces/')) {
      const id = decodeURIComponent(path.slice('/spaces/'.length).split('/')[0]);
      const sp = $spaces.find((x) => x.id === id);
      out.push({ label: 'spaces', href: '/spaces' });
      out.push({ label: sp?.label ?? id });
    } else if (path.startsWith('/spaces')) {
      out.push({ label: 'spaces' });
    } else if (path.startsWith('/settings')) {
      out.push({ label: 'settings' });
    }
    return out;
  });

  const onPane = $derived(path.startsWith('/pane/'));
  const paneRef = $derived(
    onPane ? findPaneIn($spaces, decodeURIComponent(path.slice('/pane/'.length).split('/')[0])) : undefined
  );
</script>

<header class="crumbbar">
  <button class="nav" aria-label="toggle navigation" title="navigation" onclick={() => navOpen.update((v) => !v)}>☰</button>
  <nav class="crumbs" aria-label="breadcrumb">
    {#each crumbs as c, i}
      {#if i > 0}<span class="sep" aria-hidden="true">/</span>{/if}
      {#if c.href && i < crumbs.length - 1}
        <button class="crumb link mono" onclick={() => goto(c.href!)}>{c.label}</button>
      {:else}
        <span class="crumb cur mono">{c.label}</span>
      {/if}
    {/each}
  </nav>
  <div class="right">
    {#if paneRef}<StatusPill status={paneRef.pane.status} />{/if}
    {#if onPane}
      <div class="switch">
        <button class:active={$mode === 'chat'} onclick={() => mode.set('chat')}>chat</button>
        <button class:active={$mode === 'raw'} onclick={() => mode.set('raw')}>raw</button>
      </div>
    {/if}
    <button class="gear" class:active={path.startsWith('/settings')} aria-label="settings" title="settings" onclick={() => goto('/settings')}>⚙</button>
  </div>
</header>

<style>
  .crumbbar {
    display: flex; align-items: center; gap: 8px; flex: none;
    height: 46px; padding: 0 10px;
    background: var(--sidebar-bg); border-bottom: 1px solid var(--hairline);
  }
  .nav { width: 34px; height: 34px; flex: none; border-radius: var(--r-chip); border: 1px solid var(--control); background: var(--card); color: var(--text-1); font-size: 16px; line-height: 1; }
  .nav:hover { background: var(--surface-tint); }
  .crumbs { display: flex; align-items: center; gap: 6px; min-width: 0; overflow-x: auto; flex: 1; }
  .sep { color: var(--text-4); font-size: 12px; }
  .crumb { flex: none; font-size: 13px; background: none; border: none; padding: 4px 4px; border-radius: var(--r-badge); white-space: nowrap; }
  .crumb.link { color: var(--text-3); }
  .crumb.link:hover { color: var(--text-1); background: var(--surface-tint); }
  .crumb.cur { color: var(--text-1); font-weight: 600; }
  .right { display: flex; align-items: center; gap: 8px; flex: none; }
  .switch { display: flex; background: var(--code-surface); border: 1px solid var(--control); border-radius: var(--r-chip); overflow: hidden; }
  .switch button { min-height: 32px; padding: 0 12px; background: none; border: none; color: var(--text-3); font-size: 12px; }
  .switch button.active { color: var(--app-bg); background: var(--text-1); }
  .gear { width: 34px; height: 34px; flex: none; border-radius: var(--r-chip); border: 1px solid var(--control); background: var(--card); color: var(--text-2); font-size: 15px; }
  .gear:hover, .gear.active { background: var(--surface-tint); color: var(--text-1); }
</style>
