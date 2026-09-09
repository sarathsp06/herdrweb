<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { agentsOf } from '$lib/session/derive';

  const s = session();
  const spaces = s.spaces;
  const blocked = $derived(agentsOf($spaces).filter((a) => a.pane.status === 'blocked').length);
  const path = $derived($page.url.pathname);

  const tabs: { href: string; label: string; glyph: string; match: (p: string) => boolean }[] = [
    { href: '/', label: 'Agents', glyph: '◉', match: (p) => p === '/' || p.startsWith('/pane/') },
    { href: '/spaces', label: 'Spaces', glyph: '⌗', match: (p) => p.startsWith('/spaces') },
    { href: '/settings', label: 'Settings', glyph: '⚙', match: (p) => p.startsWith('/settings') }
  ];
</script>

<nav class="tabbar" aria-label="primary">
  {#each tabs as t (t.href)}
    <button
      class="tab"
      class:active={t.match(path)}
      aria-current={t.match(path) ? 'page' : undefined}
      onclick={() => goto(t.href)}
    >
      <span class="glyph" aria-hidden="true">
        {t.glyph}
        {#if t.href === '/' && blocked > 0}<span class="badge mono">{blocked}</span>{/if}
      </span>
      <span class="label">{t.label}</span>
    </button>
  {/each}
</nav>

<style>
  .tabbar {
    flex: none; display: flex;
    background: var(--sidebar-bg); border-top: 1px solid var(--hairline);
    padding-bottom: env(safe-area-inset-bottom);
  }
  .tab {
    flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center;
    gap: 2px; min-height: 54px; padding: 6px 0 4px;
    background: none; border: none; color: var(--text-4);
  }
  .tab.active { color: var(--text-1); }
  .glyph { position: relative; font-size: 18px; line-height: 1; }
  .label { font-size: 10.5px; font-weight: 600; letter-spacing: 0.01em; }
  .badge {
    position: absolute; top: -5px; left: calc(100% - 2px);
    min-width: 15px; height: 15px; padding: 0 4px; border-radius: 999px;
    background: var(--blocked); color: var(--text-1);
    font-size: 9.5px; font-weight: 700; line-height: 15px; text-align: center;
  }
</style>
