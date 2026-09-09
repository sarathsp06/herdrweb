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

<nav
  class="tabbar flex flex-none border-t border-(--hairline) bg-(--sidebar-bg) pb-[env(safe-area-inset-bottom)]"
  aria-label="primary"
>
  {#each tabs as t (t.href)}
    <button
      class="flex min-h-[54px] flex-1 flex-col items-center justify-center gap-0.5 pt-1.5 pb-1 {t.match(path)
        ? 'text-foreground'
        : 'text-muted-foreground'}"
      aria-current={t.match(path) ? 'page' : undefined}
      onclick={() => goto(t.href)}
    >
      <span class="relative text-lg leading-none" aria-hidden="true">
        {t.glyph}
        {#if t.href === '/' && blocked > 0}<span
            class="mono absolute -top-[5px] left-[calc(100%-2px)] h-[15px] min-w-[15px] rounded-full bg-blocked px-1 text-center text-[9.5px] leading-[15px] font-bold text-foreground"
            >{blocked}</span
          >{/if}
      </span>
      <span class="text-[10.5px] font-semibold tracking-[0.01em]">{t.label}</span>
    </button>
  {/each}
</nav>
