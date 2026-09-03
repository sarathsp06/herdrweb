<script lang="ts">
  import { goto } from '$app/navigation';
  let { path }: { path: string } = $props();
  const tabs = [
    { href: '/', label: 'Agents', glyph: '◉' },
    { href: '/spaces', label: 'Spaces', glyph: '⌗' },
    { href: '/settings', label: 'Settings', glyph: '⚙' }
  ];
  function active(href: string) {
    return href === '/' ? path === '/' : path.startsWith(href);
  }
</script>
<nav class="bar">
  {#each tabs as t}
    <button class="tab" class:active={active(t.href)} onclick={() => goto(t.href)}>
      <span class="g mono">{t.glyph}</span>
      <span class="l">{t.label}</span>
    </button>
  {/each}
</nav>
<style>
  .bar {
    position: fixed; left: 0; right: 0; bottom: 0; z-index: 40; display: flex;
    background: linear-gradient(180deg, transparent, var(--app-bg) 40%);
    padding: 6px 10px calc(6px + env(safe-area-inset-bottom));
  }
  .tab {
    flex: 1; min-height: 50px; background: none; border: none; border-radius: var(--r-btn);
    display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2px;
    color: var(--text-4); font-size: 10.5px;
  }
  .tab.active { background: var(--surface-tint-2); color: var(--text-1); }
  .g { font-size: 16px; }
</style>
